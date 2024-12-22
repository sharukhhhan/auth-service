package v1

import (
	"errors"
	"github.com/labstack/echo/v4"
	"medods-tz/internal/entity"
	"medods-tz/internal/service"
	"net/http"
)

type authRoutes struct {
	authService service.AuthService
}

func newAuthRoutes(g *echo.Group, authService service.AuthService) {
	r := &authRoutes{
		authService: authService,
	}

	g.POST("/token", r.createTokens)
	g.POST("/refresh", r.refreshTokens)
}

type createTokensInput struct {
	UserId   string `json:"user_id" validate:"required,uuid"`
	ClientIP string `json:"client_ip" validate:"required,ip"`
}

func (r *authRoutes) createTokens(c echo.Context) error {
	var input createTokensInput

	if err := c.Bind(&input); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := c.Validate(&input); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, err)
	}

	tokens, err := r.authService.CreateTokens(c.Request().Context(), input.UserId, input.ClientIP)
	if err != nil {
		if errors.Is(err, service.ErrSessionAlreadyExists) || errors.Is(err, service.ErrUserNotFound) {
			return newErrorResponse(c, http.StatusBadRequest, err)
		}

		return newErrorResponse(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, tokens)
}

func (r *authRoutes) refreshTokens(c echo.Context) error {
	var input entity.Tokens

	if err := c.Bind(&input); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := c.Validate(&input); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, err)
	}

	tokens, err := r.authService.RefreshTokens(c.Request().Context(), input.RefreshToken, input.AccessToken)
	if err != nil {
		if errors.Is(err, service.ErrParsingAccessToken) ||
			errors.Is(err, service.ErrRefreshTokenNotFound) ||
			errors.Is(err, service.ErrRefreshTokenAlreadyUsed) ||
			errors.Is(err, service.ErrRefreshTokenExpired) ||
			errors.Is(err, service.ErrUserNotFound) ||
			errors.Is(err, service.ErrNoSessionsFoundWithThisUserID) {

			return newErrorResponse(c, http.StatusBadRequest, err)
		}

		return newErrorResponse(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, tokens)
}
