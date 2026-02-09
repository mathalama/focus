package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"mathalama-focus/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type devLoginRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

func (h *Handler) DevLogin(c *gin.Context) {
	var req devLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, token, err := h.auth.DevLogin(c.Request.Context(), req.Email, req.Name)
	if errors.Is(err, domain.ErrInvalidEmail) {
		respondError(c, http.StatusBadRequest, "email is required")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to login")
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	out, err := h.auth.Register(c.Request.Context(), req.Email, req.Name, req.Password)
	if errors.Is(err, domain.ErrEmailNotConfigured) {
		respondError(c, http.StatusServiceUnavailable, "email verification is not configured")
		return
	}
	if errors.Is(err, domain.ErrInvalidEmail) {
		respondError(c, http.StatusBadRequest, "valid email is required")
		return
	}
	if errors.Is(err, domain.ErrWeakPassword) {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, domain.ErrAlreadyExists) {
		respondError(c, http.StatusConflict, "email is already registered")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to register user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"registered":                  true,
		"user":                        out.User,
		"requires_email_verification": true,
		"verification_email_sent":     out.VerificationEmailSent,
		"verification_expires_at":     out.VerificationExpiresAt,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, token, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, domain.ErrInvalidEmail) {
		respondError(c, http.StatusBadRequest, "valid email is required")
		return
	}
	if errors.Is(err, domain.ErrInvalidCredentials) {
		respondError(c, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if errors.Is(err, domain.ErrEmailNotVerified) {
		respondError(c, http.StatusForbidden, "email is not verified")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to authenticate")
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *Handler) ResendVerificationEmail(c *gin.Context) {
	var req resendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	out, err := h.auth.ResendVerification(c.Request.Context(), req.Email)
	if errors.Is(err, domain.ErrEmailNotConfigured) {
		respondError(c, http.StatusServiceUnavailable, "email verification is not configured")
		return
	}
	if errors.Is(err, domain.ErrInvalidEmail) {
		respondError(c, http.StatusBadRequest, "valid email is required")
		return
	}
	if err != nil {
		respondError(c, http.StatusBadGateway, "failed to send verification email")
		return
	}

	if out.AlreadyVerified {
		c.JSON(http.StatusOK, gin.H{"resent": false, "already_verified": true})
		return
	}

	resp := gin.H{"resent": out.Resent}
	if !out.ExpiresAt.IsZero() {
		resp["verification_expires_at"] = out.ExpiresAt
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		h.renderVerifyResult(c, false, "verification token is required")
		return
	}

	_, err := h.auth.VerifyEmail(c.Request.Context(), token)
	if errors.Is(err, domain.ErrEmailTokenInvalid) {
		h.renderVerifyResult(c, false, "verification token is invalid or expired")
		return
	}
	if err != nil {
		h.renderVerifyResult(c, false, "failed to verify email")
		return
	}

	successURL := h.auth.SuccessRedirectURL()
	if successURL != "" {
		c.Redirect(http.StatusFound, successURL)
		return
	}
	c.JSON(http.StatusOK, gin.H{"verified": true})
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := getUserID(c)

	user, err := h.auth.GetUser(c.Request.Context(), userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get user")
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) renderVerifyResult(c *gin.Context, success bool, errorMessage string) {
	if success {
		successURL := h.auth.SuccessRedirectURL()
		if successURL != "" {
			c.Redirect(http.StatusFound, successURL)
			return
		}
		c.JSON(http.StatusOK, gin.H{"verified": true})
		return
	}

	failureURL := h.auth.FailureRedirectURL()
	if failureURL != "" {
		parsed, err := url.Parse(failureURL)
		if err == nil {
			query := parsed.Query()
			if strings.TrimSpace(errorMessage) != "" {
				query.Set("reason", errorMessage)
			}
			parsed.RawQuery = query.Encode()
			failureURL = parsed.String()
		}
		c.Redirect(http.StatusFound, failureURL)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"verified": false, "error": errorMessage})
}
