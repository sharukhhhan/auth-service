package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"medods-tz/internal/entity"
	"medods-tz/internal/repository"
	"medods-tz/internal/repository/repoerrors"
	"medods-tz/internal/sender"
	"time"
)

type TokenClaims struct {
	jwt.StandardClaims
	ClientIP string
	UserID   string
}

type Auth struct {
	userRepo        repository.UserRepository
	tokenRepo       repository.TokenRepository
	signKey         string
	tokenTTL        time.Duration
	refreshTokenTTL time.Duration
	securityLog     *logrus.Logger
	emailSender     sender.Email
}

func NewAuth(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	tokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	signKey string,
	securityLog *logrus.Logger,
	emailSender sender.Email) *Auth {
	return &Auth{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		signKey:         signKey,
		tokenTTL:        tokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		securityLog:     securityLog,
		emailSender:     emailSender,
	}
}

func (s *Auth) CreateTokens(ctx context.Context, userID, clientIP string) (*entity.Tokens, error) {

	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("error while trying to find user: %w", err)
	}

	accessToken, err := s.generateAccessToken(clientIP, userID)
	if err != nil {
		return nil, fmt.Errorf("error while generating access token: %w", err)
	}

	refreshToken, refreshTokenHash, err := s.generateRefreshToken()

	refreshTokenEntiry := entity.RefreshToken{
		UserID:      userID,
		RefreshHash: refreshTokenHash,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(s.refreshTokenTTL),
		ClientIP:    clientIP,
		Used:        false,
	}

	err = s.tokenRepo.CreateRefreshToken(ctx, refreshTokenEntiry)
	if err != nil {
		if errors.Is(err, repoerrors.ErrAlreadyExists) {
			return nil, ErrSessionAlreadyExists
		}

		return nil, fmt.Errorf("error while creating refresh token: %w", err)
	}

	tokens := entity.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return &tokens, nil
}

func (s *Auth) RefreshTokens(ctx context.Context, refreshToken, accessToken string) (*entity.Tokens, error) {
	claims, err := s.parseAccessToken(accessToken)
	if err != nil && !errors.Is(err, ErrAccessTokenExpired) {
		return nil, fmt.Errorf("%w: %w", ErrParsingAccessToken, err)
	}

	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("error while trying to find user: %w", err)
	}

	refreshTokenEntities, err := s.tokenRepo.GetRefreshTokenEntitiesByUserID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("error while getting refresh token by userID: %w", err)
	}
	if len(refreshTokenEntities) < 1 {
		return nil, ErrNoSessionsFoundWithThisUserID
	}

	token, err := s.findMatchingRefreshTokens(refreshToken, refreshTokenEntities)
	if err != nil {
		if !errors.Is(err, ErrRefreshTokenNotFound) {
			return nil, fmt.Errorf("error while comaring token_hash and input_token: %w", err)
		}

		return nil, err
	}

	if token.Used {
		return nil, ErrRefreshTokenAlreadyUsed
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, ErrRefreshTokenExpired
	}

	if claims.ClientIP != token.ClientIP {
		err = s.emailSender.SendWarningEmail(user.Email, "Suspicious login", fmt.Sprintf("Warning! Someone logged in from this IP: %s", claims.ClientIP))
		if err != nil {
			s.securityLog.Errorf("error while sending warning emailSender to user_id=%s", user.ID)
		}
	}

	err = s.tokenRepo.MarkRefreshTokenUsed(ctx, token.ID)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return nil, ErrRefreshTokenNotFound
		}

		return nil, fmt.Errorf("error while marking refresh token as used: %w", err)
	}

	accessToken, err = s.generateAccessToken(claims.ClientIP, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("error while generating access token: %w", err)
	}

	refreshToken, refreshTokenHash, err := s.generateRefreshToken()

	refreshTokenEntiry := entity.RefreshToken{
		UserID:      claims.UserID,
		RefreshHash: refreshTokenHash,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(s.refreshTokenTTL),
		ClientIP:    claims.ClientIP,
		Used:        false,
	}

	err = s.tokenRepo.CreateRefreshToken(ctx, refreshTokenEntiry)
	if err != nil {
		if errors.Is(err, repoerrors.ErrAlreadyExists) {
			return nil, ErrSessionAlreadyExists
		}

		return nil, fmt.Errorf("error while creating refresh token: %w", err)
	}

	tokens := entity.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return &tokens, nil
}

func (s *Auth) generateAccessToken(clientIP string, userID string) (string, error) {
	claims := TokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		clientIP,
		userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessToken, err := token.SignedString([]byte(s.signKey))
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (s *Auth) generateRefreshToken() (string, string, error) {
	// generating string for refresh token
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	refreshToken := base64.RawURLEncoding.EncodeToString(bytes)[:32]

	refreshTokenHash, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return refreshToken, string(refreshTokenHash), nil
}

func (s *Auth) findMatchingRefreshTokens(inputToken string, refreshTokenEntities []entity.RefreshToken) (*entity.RefreshToken, error) {
	for _, token := range refreshTokenEntities {
		err := bcrypt.CompareHashAndPassword([]byte(token.RefreshHash), []byte(inputToken))
		if err != nil {
			if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return nil, err
			}

			continue
		}

		return &token, nil
	}

	return nil, ErrRefreshTokenNotFound
}

func (s *Auth) parseAccessToken(accessToken string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	_, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.signKey), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			// check for token expiration
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return claims, ErrAccessTokenExpired
			}
			return nil, fmt.Errorf("token validation error: %w", err)
		}
		return nil, fmt.Errorf("unexpected token parsing error: %w", err)
	}

	return claims, nil
}
