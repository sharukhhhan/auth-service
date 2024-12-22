package service

import (
	"context"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"medods-tz/internal/entity"
	"testing"
	"time"
)

func TestAuth_CreateTokens(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(mockUserRepo)
	mockTokenRepo := new(mockTokenRepo)
	mockSender := new(mockEmail)

	log := logrus.New()
	auth := NewAuth(
		mockUserRepo,
		mockTokenRepo,
		time.Minute*15,
		time.Hour*24,
		"test-sign-key",
		log,
		mockSender,
	)

	mockUserRepo.On("GetUserByID", ctx, "user-id").Return(&entity.User{ID: "user-id", Email: "test@example.com"}, nil)
	mockTokenRepo.On("CreateRefreshToken", ctx, mock.Anything).Return(nil)

	tokens, err := auth.CreateTokens(ctx, "user-id", "127.0.0.1")

	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	mockUserRepo.AssertCalled(t, "GetUserByID", ctx, "user-id")
	mockTokenRepo.AssertCalled(t, "CreateRefreshToken", ctx, mock.Anything)
}

func TestAuth_RefreshTokens(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(mockUserRepo)
	mockTokenRepo := new(mockTokenRepo)
	mockEmail := new(mockEmail)

	log := logrus.New()
	auth := NewAuth(
		mockUserRepo,
		mockTokenRepo,
		time.Minute*15,
		time.Hour*24,
		"test-sign-key",
		log,
		mockEmail,
	)

	claims := TokenClaims{
		ClientIP: "127.0.0.1",
		UserID:   "user-id",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-time.Minute).Unix(),
			IssuedAt:  time.Now().Add(-time.Hour).Unix(),
		},
	}

	accessToken, _ := auth.generateAccessToken(claims.ClientIP, claims.UserID)

	mockUserRepo.On("GetUserByID", ctx, "user-id").Return(&entity.User{ID: "user-id", Email: "test@example.com"}, nil)
	mockTokenRepo.On("GetRefreshTokenEntitiesByUserID", ctx, "user-id").Return([]entity.RefreshToken{
		{
			ID:          "token-id",
			UserID:      "user-id",
			RefreshHash: string(hashRefreshToken("valid-refresh-token")),
			ClientIP:    "127.0.0.1",
			ExpiresAt:   time.Now().Add(time.Hour),
			Used:        false,
		},
	}, nil)
	mockTokenRepo.On("MarkRefreshTokenUsed", ctx, "token-id").Return(nil)
	mockTokenRepo.On("CreateRefreshToken", ctx, mock.Anything).Return(nil)

	tokens, err := auth.RefreshTokens(ctx, "valid-refresh-token", accessToken)

	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	mockUserRepo.AssertCalled(t, "GetUserByID", ctx, "user-id")
	mockTokenRepo.AssertCalled(t, "GetRefreshTokenEntitiesByUserID", ctx, "user-id")
	mockTokenRepo.AssertCalled(t, "MarkRefreshTokenUsed", ctx, "token-id")
	mockTokenRepo.AssertCalled(t, "CreateRefreshToken", ctx, mock.Anything)
}

func hashRefreshToken(token string) []byte {
	hash, _ := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	return hash
}
