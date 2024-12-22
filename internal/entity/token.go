package entity

import "time"

type RefreshToken struct {
	ID           string
	UserID       string
	RefreshHash  string
	AccessExpiry time.Time
	IssuedAt     time.Time
	ExpiresAt    time.Time
	ClientIP     string
	Used         bool
}

type Tokens struct {
	AccessToken  string `json:"access_token" validate:"required,jwt"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}
