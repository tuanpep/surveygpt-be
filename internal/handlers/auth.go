package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/pkg/httputil"
	"github.com/surveyflow/be/internal/pkg/validator"
	"github.com/surveyflow/be/internal/services"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authSvc *services.AuthService
	validator *validator.Validator
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authSvc:   authSvc,
		validator: validator.New(),
	}
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c echo.Context) error {
	var input services.SignUpInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	if errs := h.validator.Validate(input); errs != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{
			"code":    "VALIDATION_FAILED",
			"message": "validation failed",
			"errors":  errs,
		})
	}

	output, err := h.authSvc.SignUp(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return httputil.Created(c, output)
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c echo.Context) error {
	var input services.SignInInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	if errs := h.validator.Validate(input); errs != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{
			"code":    "VALIDATION_FAILED",
			"message": "validation failed",
			"errors":  errs,
		})
	}

	output, err := h.authSvc.SignIn(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return httputil.OK(c, output)
}

// Login2FA handles POST /auth/login/2fa.
func (h *AuthHandler) Login2FA(c echo.Context) error {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	if body.Code == "" {
		return apperr.BadRequest("2FA code is required")
	}

	output, err := h.authSvc.SignInWith2FA(c.Request().Context(), body.Email, body.Code)
	if err != nil {
		return err
	}

	return httputil.OK(c, output)
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(c echo.Context) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	if body.RefreshToken == "" {
		return apperr.BadRequest("refresh_token is required")
	}

	tokens, err := h.authSvc.RefreshToken(c.Request().Context(), body.RefreshToken)
	if err != nil {
		return err
	}

	return httputil.OK(c, tokens)
}

// ForgotPassword handles POST /auth/forgot-password.
func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	if body.Email == "" {
		return apperr.BadRequest("email is required")
	}

	if err := h.authSvc.ForgotPassword(c.Request().Context(), body.Email); err != nil {
		return err
	}

	return httputil.Message(c, "if an account with that email exists, a password reset link has been sent")
}

// ResetPassword handles POST /auth/reset-password.
func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	if body.Token == "" || body.NewPassword == "" {
		return apperr.BadRequest("token and new_password are required")
	}

	if err := h.authSvc.ResetPassword(c.Request().Context(), body.Token, body.NewPassword); err != nil {
		return err
	}

	return httputil.Message(c, "password has been reset successfully")
}

// GetMe handles GET /me.
func (h *AuthHandler) GetMe(c echo.Context) error {
	userID := middleware.GetUserIDFromContext(c)
	orgID := middleware.GetOrgIDFromContext(c)

	user, err := h.authSvc.GetCurrentUser(c.Request().Context(), userID, orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, user)
}

// UpdateMe handles PUT /me.
func (h *AuthHandler) UpdateMe(c echo.Context) error {
	var body struct {
		Name      string  `json:"name"`
		AvatarURL *string `json:"avatar_url"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	userID := middleware.GetUserIDFromContext(c)
	user, err := h.authSvc.UpdateProfile(c.Request().Context(), userID, body.Name, body.AvatarURL)
	if err != nil {
		return err
	}

	return httputil.OK(c, user)
}

// ChangePassword handles PUT /me/password.
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	var input services.ChangePasswordInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	if errs := h.validator.Validate(input); errs != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{
			"code":    "VALIDATION_FAILED",
			"message": "validation failed",
			"errors":  errs,
		})
	}

	userID := middleware.GetUserIDFromContext(c)
	if err := h.authSvc.ChangePassword(c.Request().Context(), userID, input.CurrentPassword, input.NewPassword); err != nil {
		return err
	}

	return httputil.Message(c, "password changed successfully")
}
