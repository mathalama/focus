package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultBackendURL = "http://localhost:8080"
	defaultAppURL     = "http://localhost:5173"
	defaultPollWait   = 30 * time.Second
	defaultAPIAddr    = ":8091"
	botAuthHeaderName = "X-Telegram-Bot-Auth"

	callbackNotifyOn  = "notify_on"
	callbackNotifyOff = "notify_off"
	callbackStatus    = "show_status"
	callbackHelp      = "show_help"
)

type config struct {
	token           string
	backendURL      string
	appURL          string
	botAuth         string
	pollTimeout     time.Duration
	internalAPIAddr string
}

type app struct {
	cfg    config
	client *http.Client
	apiURL string
}

type telegramAPIResponse[T any] struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      T      `json:"result"`
}

type telegramUpdate struct {
	UpdateID      int                    `json:"update_id"`
	Message       *telegramMessage       `json:"message"`
	CallbackQuery *telegramCallbackQuery `json:"callback_query"`
}

type telegramMessage struct {
	Chat telegramChat `json:"chat"`
	From telegramUser `json:"from"`
	Text string       `json:"text"`
}

type telegramChat struct {
	ID int64 `json:"id"`
}

type telegramCallbackQuery struct {
	ID      string           `json:"id"`
	From    telegramUser     `json:"from"`
	Message *telegramMessage `json:"message"`
	Data    string           `json:"data"`
}

type telegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type sendMessageRequest struct {
	ChatID      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *inlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type inlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

type inlineKeyboardMarkup struct {
	InlineKeyboard [][]inlineKeyboardButton `json:"inline_keyboard"`
}

type answerCallbackQueryRequest struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
}

type backendErrorResponse struct {
	Error string `json:"error"`
}

type backendTelegramLinkRequest struct {
	Code             string `json:"code"`
	TelegramUserID   int64  `json:"telegram_user_id"`
	TelegramUsername string `json:"telegram_username"`
	TelegramFirst    string `json:"telegram_first_name"`
	TelegramLast     string `json:"telegram_last_name"`
}

type backendLinkedUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type backendTelegramStatusResponse struct {
	Linked               bool               `json:"linked"`
	NotificationsEnabled bool               `json:"notifications_enabled"`
	User                 *backendLinkedUser `json:"user"`
}

type backendTelegramNotificationsRequest struct {
	TelegramUserID int64 `json:"telegram_user_id"`
	Enabled        bool  `json:"enabled"`
}

type internalNotifyRequest struct {
	TelegramUserID int64  `json:"telegram_user_id"`
	Message        string `json:"message"`
	DisableButtons bool   `json:"disable_buttons"`
	Force          bool   `json:"force"`
}

func main() {
	if err := loadDotEnvIfPresent(".env"); err != nil {
		log.Fatalf("config error: failed to load .env: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := &app{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.pollTimeout + 15*time.Second,
		},
		apiURL: fmt.Sprintf("https://api.telegram.org/bot%s", cfg.token),
	}

	log.Printf("telegram bot started (backend=%s)", cfg.backendURL)
	if err := app.run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped with error: %v", err)
	}
}

func loadConfig() (config, error) {
	cfg := config{
		token:           strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		backendURL:      strings.TrimRight(strings.TrimSpace(envOrDefault("BACKEND_URL", defaultBackendURL)), "/"),
		appURL:          strings.TrimRight(strings.TrimSpace(envOrDefault("APP_URL", defaultAppURL)), "/"),
		botAuth:         strings.TrimSpace(os.Getenv("TELEGRAM_BOT_AUTH_TOKEN")),
		pollTimeout:     parseDurationSeconds(os.Getenv("TELEGRAM_POLL_TIMEOUT_SECONDS"), defaultPollWait),
		internalAPIAddr: strings.TrimSpace(envOrDefault("BOT_INTERNAL_API_ADDR", defaultAPIAddr)),
	}

	if cfg.token == "" {
		return config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.botAuth == "" {
		return config{}, errors.New("TELEGRAM_BOT_AUTH_TOKEN is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func loadDotEnvIfPresent(path string) error {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, rawValue, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value := normalizeEnvValue(rawValue)
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func normalizeEnvValue(raw string) string {
	value := strings.TrimSpace(raw)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func parseDurationSeconds(raw string, fallback time.Duration) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func (a *app) run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- a.runPolling(ctx)
	}()

	go func() {
		errCh <- a.runInternalAPIServer(ctx)
	}()

	var firstErr error
	for i := 0; i < cap(errCh); i++ {
		err := <-errCh
		if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed) {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (a *app) runPolling(ctx context.Context) error {
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := a.getUpdates(ctx, offset)
		if err != nil {
			log.Printf("getUpdates error: %v", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			if update.CallbackQuery != nil {
				a.handleCallbackQuery(ctx, *update.CallbackQuery)
				continue
			}
			if update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
				continue
			}

			a.handleMessage(ctx, *update.Message)
		}
	}
}

func (a *app) runInternalAPIServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.handleInternalHealth)
	mux.HandleFunc("/internal/notify", a.handleInternalNotify)

	srv := &http.Server{
		Addr:              a.cfg.internalAPIAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("internal api shutdown error: %v", err)
		}
	}()

	log.Printf("telegram internal api listening on %s", a.cfg.internalAPIAddr)
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (a *app) handleInternalHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (a *app) handleInternalNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(r.Header.Get(botAuthHeaderName))), []byte(a.cfg.botAuth)) != 1 {
		writeJSONError(w, http.StatusUnauthorized, "invalid bot auth token")
		return
	}

	var req internalNotifyRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Message = strings.TrimSpace(req.Message)
	if req.TelegramUserID <= 0 || req.Message == "" {
		writeJSONError(w, http.StatusBadRequest, "telegram_user_id and message are required")
		return
	}

	if !req.Force {
		status, statusCode, statusErr := a.getTelegramStatus(r.Context(), req.TelegramUserID)
		if statusCode != http.StatusOK {
			log.Printf("internal notify status check failed (status=%d): %s", statusCode, statusErr)
			writeJSONError(w, http.StatusBadGateway, "failed to resolve telegram link status")
			return
		}
		if !status.Linked {
			writeJSONError(w, http.StatusNotFound, "telegram account is not linked")
			return
		}
		if !status.NotificationsEnabled {
			writeJSONError(w, http.StatusConflict, "telegram notifications are disabled")
			return
		}
	}

	replyMarkup := a.actionKeyboard()
	if req.DisableButtons {
		replyMarkup = nil
	}

	if err := a.sendMessageWithMarkupResult(r.Context(), req.TelegramUserID, req.Message, replyMarkup); err != nil {
		log.Printf("internal notify send failed: %v", err)
		writeJSONError(w, http.StatusBadGateway, "failed to send telegram message")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"sent":                 true,
		"telegram_user_id":     req.TelegramUserID,
		"notifications_forced": req.Force,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response failed: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]any{"error": strings.TrimSpace(message)})
}

func (a *app) getUpdates(ctx context.Context, offset int) ([]telegramUpdate, error) {
	updatesURL, err := url.Parse(a.apiURL + "/getUpdates")
	if err != nil {
		return nil, err
	}

	query := updatesURL.Query()
	query.Set("offset", strconv.Itoa(offset))
	query.Set("timeout", strconv.Itoa(int(a.cfg.pollTimeout.Seconds())))
	updatesURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updatesURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload telegramAPIResponse[[]telegramUpdate]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.OK {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", payload.Description)
	}

	return payload.Result, nil
}

func (a *app) handleMessage(ctx context.Context, msg telegramMessage) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	command, arg := parseCommand(text)
	switch command {
	case "/start":
		code := parseStartCode(arg)
		if code == "" {
			a.sendLinkStatus(ctx, msg)
			return
		}
		a.tryLink(ctx, msg, code)
	case "/status":
		a.sendLinkStatus(ctx, msg)
	case "/help":
		a.sendMessage(ctx, msg.Chat.ID, a.helpText())
	case "/notify", "/notifications":
		a.handleNotifyCommand(ctx, msg, arg)
	case "/subscribe":
		a.setNotifications(ctx, msg, true)
	case "/unsubscribe":
		a.setNotifications(ctx, msg, false)
	case "/link":
		if arg == "" {
			a.sendMessage(ctx, msg.Chat.ID, "Использование: /link <CODE>")
			return
		}
		a.tryLink(ctx, msg, arg)
	default:
		a.sendMessage(ctx, msg.Chat.ID, a.helpText())
	}
}

func (a *app) actionKeyboard() *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{
		InlineKeyboard: [][]inlineKeyboardButton{
			{
				{Text: "Включить", CallbackData: callbackNotifyOn},
				{Text: "Выключить", CallbackData: callbackNotifyOff},
			},
			{
				{Text: "Статус", CallbackData: callbackStatus},
				{Text: "Помощь", CallbackData: callbackHelp},
			},
		},
	}
}

func (a *app) handleCallbackQuery(ctx context.Context, query telegramCallbackQuery) {
	if query.ID == "" {
		return
	}

	chatID := int64(0)
	if query.Message != nil {
		chatID = query.Message.Chat.ID
	}
	if chatID <= 0 {
		a.answerCallbackQuery(ctx, query.ID, "Чат недоступен")
		return
	}

	action := strings.TrimSpace(query.Data)
	switch action {
	case callbackNotifyOn:
		a.setNotificationsByUser(ctx, query.From, chatID, true)
		a.answerCallbackQuery(ctx, query.ID, "Уведомления: ON")
	case callbackNotifyOff:
		a.setNotificationsByUser(ctx, query.From, chatID, false)
		a.answerCallbackQuery(ctx, query.ID, "Уведомления: OFF")
	case callbackStatus:
		a.sendLinkStatusByUser(ctx, query.From, chatID)
		a.answerCallbackQuery(ctx, query.ID, "Статус обновлен")
	case callbackHelp:
		a.sendMessage(ctx, chatID, a.helpText())
		a.answerCallbackQuery(ctx, query.ID, "Открываю помощь")
	default:
		a.answerCallbackQuery(ctx, query.ID, "Неизвестная кнопка")
	}
}

func parseCommand(text string) (string, string) {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", ""
	}

	command := strings.ToLower(parts[0])
	if at := strings.Index(command, "@"); at > 0 {
		command = command[:at]
	}
	if len(parts) == 1 {
		return command, ""
	}
	return command, strings.TrimSpace(strings.Join(parts[1:], " "))
}

func parseStartCode(arg string) string {
	value := strings.TrimSpace(arg)
	if value == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(value), "link_") {
		value = value[5:]
	}
	return strings.TrimSpace(value)
}

func (a *app) helpText() string {
	return strings.Join([]string{
		"Команды бота:",
		"/start или /status — показать статус привязки",
		"/link <CODE> — привязать аккаунт по коду из приложения",
		"/notify on — включить уведомления",
		"/notify off — выключить уведомления",
		"/subscribe — включить уведомления",
		"/unsubscribe — выключить уведомления",
		"/help — показать команды",
	}, "\n")
}

func parseNotifyArg(arg string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(arg)) {
	case "on", "true", "1", "enable", "enabled":
		return true, true
	case "off", "false", "0", "disable", "disabled":
		return false, true
	default:
		return false, false
	}
}

func (a *app) sendLinkStatus(ctx context.Context, msg telegramMessage) {
	a.sendLinkStatusByUser(ctx, msg.From, msg.Chat.ID)
}

func (a *app) sendLinkStatusByUser(ctx context.Context, from telegramUser, chatID int64) {
	status, statusCode, backendErr := a.getTelegramStatus(ctx, from.ID)
	if statusCode == http.StatusOK {
		if status.Linked {
			a.sendMessage(ctx, chatID, a.alreadyLinkedText(status.User, status.NotificationsEnabled))
			return
		}
		a.sendMessage(ctx, chatID, a.notLinkedText())
		return
	}

	if statusCode == http.StatusUnauthorized || statusCode == http.StatusServiceUnavailable {
		log.Printf("telegram status check rejected (status=%d): %s", statusCode, backendErr)
		a.sendMessage(ctx, chatID, "Интеграция Telegram временно недоступна. Попробуй позже.")
		return
	}

	log.Printf("telegram status check failed (status=%d): %s", statusCode, backendErr)
	a.sendMessage(ctx, chatID, "Не удалось проверить статус привязки. Попробуй позже.")
}

func (a *app) handleNotifyCommand(ctx context.Context, msg telegramMessage, arg string) {
	enabled, ok := parseNotifyArg(arg)
	if !ok {
		a.sendMessage(ctx, msg.Chat.ID, "Использование: /notify on или /notify off")
		return
	}
	a.setNotifications(ctx, msg, enabled)
}

func (a *app) setNotifications(ctx context.Context, msg telegramMessage, enabled bool) {
	a.setNotificationsByUser(ctx, msg.From, msg.Chat.ID, enabled)
}

func (a *app) setNotificationsByUser(ctx context.Context, from telegramUser, chatID int64, enabled bool) {
	status, statusCode, backendErr := a.getTelegramStatus(ctx, from.ID)
	if statusCode == http.StatusOK && !status.Linked {
		a.sendMessage(ctx, chatID, a.notLinkedText())
		return
	}
	if statusCode == http.StatusOK && status.Linked && status.NotificationsEnabled == enabled {
		if enabled {
			a.sendMessage(ctx, chatID, "Уведомления уже включены.\nЧтобы выключить: /notify off")
		} else {
			a.sendMessage(ctx, chatID, "Уведомления уже выключены.\nЧтобы включить: /notify on")
		}
		return
	}
	if statusCode != http.StatusOK {
		log.Printf("telegram status check before notify update failed (status=%d): %s", statusCode, backendErr)
		a.sendMessage(ctx, chatID, "Не удалось проверить статус привязки. Попробуй позже.")
		return
	}

	updateStatusCode, updateErr := a.setTelegramNotifications(ctx, from.ID, enabled)
	switch updateStatusCode {
	case http.StatusOK:
		updatedStatus, updatedStatusCode, updatedStatusErr := a.getTelegramStatus(ctx, from.ID)
		if updatedStatusCode == http.StatusOK && updatedStatus.Linked {
			a.sendMessage(ctx, chatID, a.alreadyLinkedText(updatedStatus.User, updatedStatus.NotificationsEnabled))
			return
		}
		if updatedStatusCode != http.StatusOK {
			log.Printf("telegram status check after notify update failed (status=%d): %s", updatedStatusCode, updatedStatusErr)
		}
		if enabled {
			a.sendMessage(ctx, chatID, "Уведомления включены.")
		} else {
			a.sendMessage(ctx, chatID, "Уведомления выключены.")
		}
	case http.StatusNotFound:
		a.sendMessage(ctx, chatID, a.notLinkedText())
	case http.StatusUnauthorized, http.StatusServiceUnavailable:
		log.Printf("telegram notify update rejected (status=%d): %s", updateStatusCode, updateErr)
		a.sendMessage(ctx, chatID, "Интеграция Telegram временно недоступна. Попробуй позже.")
	default:
		log.Printf("telegram notify update failed (status=%d): %s", updateStatusCode, updateErr)
		a.sendMessage(ctx, chatID, "Не удалось обновить настройки уведомлений. Попробуй позже.")
	}
}

func (a *app) notLinkedText() string {
	lines := []string{
		"Telegram пока не привязан к аккаунту.",
		"",
		"Что сделать:",
	}

	if a.cfg.appURL != "" {
		lines = append(lines, "1) Открой приложение: "+a.cfg.appURL)
		lines = append(lines, "2) Войди в аккаунт")
		lines = append(lines, "3) В профиле открой блок Telegram и сгенерируй код")
		lines = append(lines, "4) Отправь сюда: /link <CODE>")
	} else {
		lines = append(lines, "1) Войди в приложение")
		lines = append(lines, "2) В профиле открой блок Telegram и сгенерируй код")
		lines = append(lines, "3) Отправь сюда: /link <CODE>")
	}

	lines = append(lines, "", "Пример: /link ABCD2345", "После привязки уведомления по умолчанию выключены.")
	return strings.Join(lines, "\n")
}

func (a *app) alreadyLinkedText(user *backendLinkedUser, notificationsEnabled bool) string {
	notificationState := "выключены"
	nextNotificationAction := "Чтобы включить: /notify on"
	if notificationsEnabled {
		notificationState = "включены"
		nextNotificationAction = "Чтобы выключить: /notify off"
	}

	if user == nil {
		return strings.Join([]string{
			"Telegram уже привязан к аккаунту.",
			fmt.Sprintf("Уведомления: %s.", notificationState),
			nextNotificationAction,
			"Если нужно перепривязать — сначала отвяжи Telegram в профиле приложения.",
		}, "\n")
	}

	userLabel := strings.TrimSpace(user.Name)
	if userLabel == "" {
		userLabel = strings.TrimSpace(user.Email)
	}
	if userLabel == "" {
		userLabel = "аккаунт"
	}

	return strings.Join([]string{
		"Telegram уже привязан.",
		fmt.Sprintf("Аккаунт: %s", userLabel),
		fmt.Sprintf("Уведомления: %s.", notificationState),
		nextNotificationAction,
		"Если нужно перепривязать — сначала отвяжи Telegram в профиле приложения, потом отправь новый /link код.",
	}, "\n")
}

func (a *app) tryLink(ctx context.Context, msg telegramMessage, rawCode string) {
	code := strings.ToUpper(strings.TrimSpace(rawCode))
	if code == "" {
		a.sendMessage(ctx, msg.Chat.ID, "Код не найден. Использование: /link <CODE>")
		return
	}

	status, statusCode, backendErr := a.getTelegramStatus(ctx, msg.From.ID)
	if statusCode == http.StatusOK && status.Linked {
		a.sendMessage(ctx, msg.Chat.ID, a.alreadyLinkedText(status.User, status.NotificationsEnabled))
		return
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNotFound && statusCode != http.StatusBadRequest {
		log.Printf("telegram status check before link failed (status=%d): %s", statusCode, backendErr)
		a.sendMessage(ctx, msg.Chat.ID, "Не удалось проверить текущую привязку. Попробуй позже.")
		return
	}

	statusCode, backendErr = a.linkTelegram(ctx, backendTelegramLinkRequest{
		Code:             code,
		TelegramUserID:   msg.From.ID,
		TelegramUsername: msg.From.Username,
		TelegramFirst:    msg.From.FirstName,
		TelegramLast:     msg.From.LastName,
	})
	switch statusCode {
	case http.StatusOK:
		linkedStatus, linkedStatusCode, linkedErr := a.getTelegramStatus(ctx, msg.From.ID)
		if linkedStatusCode == http.StatusOK && linkedStatus.Linked {
			a.sendMessage(ctx, msg.Chat.ID, "Готово.\n"+a.alreadyLinkedText(linkedStatus.User, linkedStatus.NotificationsEnabled))
			return
		}
		if linkedStatusCode != http.StatusOK {
			log.Printf("telegram status check after link failed (status=%d): %s", linkedStatusCode, linkedErr)
		}
		a.sendMessage(ctx, msg.Chat.ID, "Готово. Telegram аккаунт успешно привязан.")
	case http.StatusBadRequest:
		a.sendMessage(ctx, msg.Chat.ID, "Код недействителен или истек. Запроси новый код в приложении.")
	case http.StatusConflict:
		a.sendMessage(ctx, msg.Chat.ID, "Этот Telegram уже привязан к другому пользователю. Сначала отвяжи его в профиле того аккаунта.")
	default:
		log.Printf("telegram link failed (status=%d): %s", statusCode, backendErr)
		a.sendMessage(ctx, msg.Chat.ID, "Не удалось привязать аккаунт. Попробуй позже.")
	}
}

func (a *app) getTelegramStatus(ctx context.Context, telegramUserID int64) (backendTelegramStatusResponse, int, string) {
	if telegramUserID <= 0 {
		return backendTelegramStatusResponse{}, http.StatusBadRequest, "telegram_user_id is invalid"
	}

	statusURL := fmt.Sprintf("%s/api/v1/integrations/telegram/status?telegram_user_id=%d", a.cfg.backendURL, telegramUserID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return backendTelegramStatusResponse{}, 0, err.Error()
	}
	req.Header.Set("X-Telegram-Bot-Auth", a.cfg.botAuth)

	resp, err := a.client.Do(req)
	if err != nil {
		return backendTelegramStatusResponse{}, 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var payload backendTelegramStatusResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return backendTelegramStatusResponse{}, resp.StatusCode, err.Error()
		}
		return payload, resp.StatusCode, ""
	}

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed backendErrorResponse
	if err := json.Unmarshal(raw, &parsed); err == nil && parsed.Error != "" {
		return backendTelegramStatusResponse{}, resp.StatusCode, parsed.Error
	}

	return backendTelegramStatusResponse{}, resp.StatusCode, strings.TrimSpace(string(raw))
}

func (a *app) setTelegramNotifications(ctx context.Context, telegramUserID int64, enabled bool) (int, string) {
	body, err := json.Marshal(backendTelegramNotificationsRequest{
		TelegramUserID: telegramUserID,
		Enabled:        enabled,
	})
	if err != nil {
		return 0, err.Error()
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		a.cfg.backendURL+"/api/v1/integrations/telegram/notifications",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-Bot-Auth", a.cfg.botAuth)

	resp, err := a.client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, ""
	}

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed backendErrorResponse
	if err := json.Unmarshal(raw, &parsed); err == nil && parsed.Error != "" {
		return resp.StatusCode, parsed.Error
	}
	return resp.StatusCode, strings.TrimSpace(string(raw))
}

func (a *app) linkTelegram(ctx context.Context, payload backendTelegramLinkRequest) (int, string) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err.Error()
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.cfg.backendURL+"/api/v1/integrations/telegram/link",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-Bot-Auth", a.cfg.botAuth)

	resp, err := a.client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, ""
	}

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed backendErrorResponse
	if err := json.Unmarshal(raw, &parsed); err == nil && parsed.Error != "" {
		return resp.StatusCode, parsed.Error
	}
	return resp.StatusCode, strings.TrimSpace(string(raw))
}

func (a *app) sendMessage(ctx context.Context, chatID int64, text string) {
	a.sendMessageWithMarkup(ctx, chatID, text, a.actionKeyboard())
}

func (a *app) sendMessageWithMarkup(ctx context.Context, chatID int64, text string, markup *inlineKeyboardMarkup) {
	if err := a.sendMessageWithMarkupResult(ctx, chatID, text, markup); err != nil {
		log.Printf("sendMessage failed: %v", err)
	}
}

func (a *app) sendMessageWithMarkupResult(ctx context.Context, chatID int64, text string, markup *inlineKeyboardMarkup) error {
	payload, err := json.Marshal(sendMessageRequest{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		return fmt.Errorf("marshal sendMessage payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.apiURL+"/sendMessage",
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("build sendMessage request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("sendMessage request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("telegram sendMessage failed (status=%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func (a *app) answerCallbackQuery(ctx context.Context, queryID, text string) {
	if strings.TrimSpace(queryID) == "" {
		return
	}

	payload, err := json.Marshal(answerCallbackQueryRequest{
		CallbackQueryID: queryID,
		Text:            strings.TrimSpace(text),
	})
	if err != nil {
		log.Printf("marshal answerCallbackQuery payload: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.apiURL+"/answerCallbackQuery",
		bytes.NewReader(payload),
	)
	if err != nil {
		log.Printf("build answerCallbackQuery request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		log.Printf("answerCallbackQuery request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		log.Printf("answerCallbackQuery failed (status=%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}
