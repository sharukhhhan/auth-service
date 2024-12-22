package service

import "errors"

var (
	ErrUserNotFound                  = errors.New("user not found")
	ErrSessionAlreadyExists          = errors.New("session with this refresh_token and user_id already exists")
	ErrRefreshTokenNotFound          = errors.New("refresh token not found")
	ErrRefreshTokenAlreadyUsed       = errors.New("refresh token already used")
	ErrRefreshTokenExpired           = errors.New("refresh token expired")
	ErrAccessTokenExpired            = errors.New("token is expired")
	ErrParsingAccessToken            = errors.New("error parsing access token")
	ErrNoSessionsFoundWithThisUserID = errors.New("no sessions found with this user_id")
)
