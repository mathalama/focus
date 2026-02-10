package httpapi

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"mathalama-focus/backend/internal/domain"
	"mathalama-focus/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

const telegramBotAuthHeader = "X-Telegram-Bot-Auth"

func (h *Handler) validateBotAuth(c *gin.Context) bool {
	if h.telegramBotAuthToken == "" {
		respondError(c, http.StatusServiceUnavailable, "telegram integration is not configured")
		return false
	}
	botAuth := strings.TrimSpace(c.GetHeader(telegramBotAuthHeader))
	if subtle.ConstantTimeCompare([]byte(botAuth), []byte(h.telegramBotAuthToken)) != 1 {
		respondError(c, http.StatusUnauthorized, "invalid bot auth token")
		return false
	}
	return true
}

func (h *Handler) CreateTelegramLinkCode(c *gin.Context) {
	userID := getUserID(c)

	code, expiresAt, err := h.telegram.CreateLinkCode(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create telegram link code")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": code, "expires_at": expiresAt.UTC()})
	h.trackEvent(c.Request.Context(), userID, "telegram.link_code.created", nil)
}

func (h *Handler) GetTelegramIdentity(c *gin.Context) {
	userID := getUserID(c)

	identity, err := h.telegram.GetIdentity(c.Request.Context(), userID)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusOK, gin.H{"identity": nil})
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get telegram identity")
		return
	}

	c.JSON(http.StatusOK, gin.H{"identity": identity})
}

func (h *Handler) UnlinkTelegram(c *gin.Context) {
	userID := getUserID(c)
	if err := h.telegram.Unlink(c.Request.Context(), userID); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to unlink telegram account")
		return
	}
	c.JSON(http.StatusOK, gin.H{"unlinked": true})
	h.trackEvent(c.Request.Context(), userID, "telegram.unlinked", nil)
}

func (h *Handler) TelegramStatusByUserID(c *gin.Context) {
	if !h.validateBotAuth(c) {
		return
	}

	rawID := strings.TrimSpace(c.Query("telegram_user_id"))
	telegramUserID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || telegramUserID <= 0 {
		respondError(c, http.StatusBadRequest, "telegram_user_id must be a positive integer")
		return
	}

	user, notif, err := h.telegram.GetStatus(c.Request.Context(), telegramUserID)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusOK, gin.H{"linked": false, "notifications_enabled": false})
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to resolve telegram link status")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"linked":                true,
		"notifications_enabled": notif,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

type telegramNotificationsRequest struct {
	TelegramUserID int64 `json:"telegram_user_id"`
	Enabled        bool  `json:"enabled"`
}

func (h *Handler) TelegramSetNotificationsByUserID(c *gin.Context) {
	if !h.validateBotAuth(c) {
		return
	}

	var req telegramNotificationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TelegramUserID <= 0 {
		respondError(c, http.StatusBadRequest, "telegram_user_id must be a positive integer")
		return
	}

	err := h.telegram.SetNotifications(c.Request.Context(), req.TelegramUserID, req.Enabled)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "telegram account is not linked")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update telegram notifications")
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true, "notifications_enabled": req.Enabled})
	h.trackEvent(c.Request.Context(), "", "telegram.notifications.updated", map[string]any{
		"telegram_user_id":      req.TelegramUserID,
		"notifications_enabled": req.Enabled,
	})
}

type telegramLinkByCodeRequest struct {
	Code             string `json:"code"`
	TelegramUserID   int64  `json:"telegram_user_id"`
	TelegramUsername string `json:"telegram_username"`
	TelegramFirst    string `json:"telegram_first_name"`
	TelegramLast     string `json:"telegram_last_name"`
}

func (h *Handler) TelegramLinkByCode(c *gin.Context) {
	if !h.validateBotAuth(c) {
		return
	}

	var req telegramLinkByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	req.TelegramUsername = strings.TrimSpace(req.TelegramUsername)
	req.TelegramFirst = strings.TrimSpace(req.TelegramFirst)
	req.TelegramLast = strings.TrimSpace(req.TelegramLast)

	if req.Code == "" || req.TelegramUserID <= 0 {
		respondError(c, http.StatusBadRequest, "code and telegram_user_id are required")
		return
	}

	user, err := h.telegram.LinkByCode(c.Request.Context(), usecase.TelegramLinkInput{
		Code:             req.Code,
		TelegramUserID:   req.TelegramUserID,
		TelegramUsername: req.TelegramUsername,
		TelegramFirst:    req.TelegramFirst,
		TelegramLast:     req.TelegramLast,
	})
	if errors.Is(err, domain.ErrTelegramCodeInvalid) {
		respondError(c, http.StatusBadRequest, "invalid or expired code")
		return
	}
	if errors.Is(err, domain.ErrTelegramAlreadyLinked) {
		respondError(c, http.StatusConflict, "telegram account is already linked to another user")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to link telegram account")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"linked_user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
	h.trackEvent(c.Request.Context(), user.ID, "telegram.linked", map[string]any{
		"telegram_user_id": req.TelegramUserID,
	})
}
