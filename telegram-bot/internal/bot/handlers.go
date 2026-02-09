package bot

import (
	"context"
	"log"
	"net/http"
	"strings"

	"mathalama-focus/telegram-bot/internal/backend"
	"mathalama-focus/telegram-bot/internal/telegram"
)

func (b *Bot) handleMessage(ctx context.Context, msg telegram.Message) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	command, arg := parseCommand(text)
	switch command {
	case "/start":
		code := parseStartCode(arg)
		if code == "" {
			b.sendLinkStatus(ctx, msg)
			return
		}
		b.tryLink(ctx, msg, code)
	case "/status":
		b.sendLinkStatus(ctx, msg)
	case "/help":
		b.sendMessage(ctx, msg.Chat.ID, helpText())
	case "/notify", "/notifications":
		b.handleNotifyCommand(ctx, msg, arg)
	case "/subscribe":
		b.setNotifications(ctx, msg, true)
	case "/unsubscribe":
		b.setNotifications(ctx, msg, false)
	case "/link":
		if arg == "" {
			b.sendMessage(ctx, msg.Chat.ID, "Использование: /link <CODE>")
			return
		}
		b.tryLink(ctx, msg, arg)
	default:
		b.sendMessage(ctx, msg.Chat.ID, helpText())
	}
}

func (b *Bot) handleCallbackQuery(ctx context.Context, query telegram.CallbackQuery) {
	if query.ID == "" {
		return
	}

	chatID := int64(0)
	if query.Message != nil {
		chatID = query.Message.Chat.ID
	}
	if chatID <= 0 {
		b.tg.AnswerCallbackQuery(ctx, query.ID, "Чат недоступен")
		return
	}

	action := strings.TrimSpace(query.Data)
	switch action {
	case telegram.CallbackNotifyOn:
		b.setNotificationsByUser(ctx, query.From, chatID, true)
		b.tg.AnswerCallbackQuery(ctx, query.ID, "Уведомления: ON")
	case telegram.CallbackNotifyOff:
		b.setNotificationsByUser(ctx, query.From, chatID, false)
		b.tg.AnswerCallbackQuery(ctx, query.ID, "Уведомления: OFF")
	case telegram.CallbackStatus:
		b.sendLinkStatusByUser(ctx, query.From, chatID)
		b.tg.AnswerCallbackQuery(ctx, query.ID, "Статус обновлен")
	case telegram.CallbackHelp:
		b.sendMessage(ctx, chatID, helpText())
		b.tg.AnswerCallbackQuery(ctx, query.ID, "Открываю помощь")
	default:
		b.tg.AnswerCallbackQuery(ctx, query.ID, "Неизвестная кнопка")
	}
}

func (b *Bot) handleNotifyCommand(ctx context.Context, msg telegram.Message, arg string) {
	enabled, ok := parseNotifyArg(arg)
	if !ok {
		b.sendMessage(ctx, msg.Chat.ID, "Использование: /notify on или /notify off")
		return
	}
	b.setNotifications(ctx, msg, enabled)
}

func (b *Bot) sendLinkStatus(ctx context.Context, msg telegram.Message) {
	b.sendLinkStatusByUser(ctx, msg.From, msg.Chat.ID)
}

func (b *Bot) sendLinkStatusByUser(ctx context.Context, from telegram.User, chatID int64) {
	status, statusCode, backendErr := b.backend.GetTelegramStatus(ctx, from.ID)
	if statusCode == http.StatusOK {
		if status.Linked {
			b.sendMessage(ctx, chatID, alreadyLinkedText(status.User, status.NotificationsEnabled))
			return
		}
		b.sendMessage(ctx, chatID, notLinkedText(b.cfg.AppURL))
		return
	}

	if statusCode == http.StatusUnauthorized || statusCode == http.StatusServiceUnavailable {
		log.Printf("telegram status check rejected (status=%d): %s", statusCode, backendErr)
		b.sendMessage(ctx, chatID, "Интеграция Telegram временно недоступна. Попробуй позже.")
		return
	}

	log.Printf("telegram status check failed (status=%d): %s", statusCode, backendErr)
	b.sendMessage(ctx, chatID, "Не удалось проверить статус привязки. Попробуй позже.")
}

func (b *Bot) setNotifications(ctx context.Context, msg telegram.Message, enabled bool) {
	b.setNotificationsByUser(ctx, msg.From, msg.Chat.ID, enabled)
}

func (b *Bot) setNotificationsByUser(ctx context.Context, from telegram.User, chatID int64, enabled bool) {
	status, statusCode, backendErr := b.backend.GetTelegramStatus(ctx, from.ID)
	if statusCode == http.StatusOK && !status.Linked {
		b.sendMessage(ctx, chatID, notLinkedText(b.cfg.AppURL))
		return
	}
	if statusCode == http.StatusOK && status.Linked && status.NotificationsEnabled == enabled {
		if enabled {
			b.sendMessage(ctx, chatID, "Уведомления уже включены.\nЧтобы выключить: /notify off")
		} else {
			b.sendMessage(ctx, chatID, "Уведомления уже выключены.\nЧтобы включить: /notify on")
		}
		return
	}
	if statusCode != http.StatusOK {
		log.Printf("telegram status check before notify update failed (status=%d): %s", statusCode, backendErr)
		b.sendMessage(ctx, chatID, "Не удалось проверить статус привязки. Попробуй позже.")
		return
	}

	updateStatusCode, updateErr := b.backend.SetTelegramNotifications(ctx, from.ID, enabled)
	switch updateStatusCode {
	case http.StatusOK:
		updatedStatus, updatedStatusCode, updatedStatusErr := b.backend.GetTelegramStatus(ctx, from.ID)
		if updatedStatusCode == http.StatusOK && updatedStatus.Linked {
			b.sendMessage(ctx, chatID, alreadyLinkedText(updatedStatus.User, updatedStatus.NotificationsEnabled))
			return
		}
		if updatedStatusCode != http.StatusOK {
			log.Printf("telegram status check after notify update failed (status=%d): %s", updatedStatusCode, updatedStatusErr)
		}
		if enabled {
			b.sendMessage(ctx, chatID, "Уведомления включены.")
		} else {
			b.sendMessage(ctx, chatID, "Уведомления выключены.")
		}
	case http.StatusNotFound:
		b.sendMessage(ctx, chatID, notLinkedText(b.cfg.AppURL))
	case http.StatusUnauthorized, http.StatusServiceUnavailable:
		log.Printf("telegram notify update rejected (status=%d): %s", updateStatusCode, updateErr)
		b.sendMessage(ctx, chatID, "Интеграция Telegram временно недоступна. Попробуй позже.")
	default:
		log.Printf("telegram notify update failed (status=%d): %s", updateStatusCode, updateErr)
		b.sendMessage(ctx, chatID, "Не удалось обновить настройки уведомлений. Попробуй позже.")
	}
}

func (b *Bot) tryLink(ctx context.Context, msg telegram.Message, rawCode string) {
	code := strings.ToUpper(strings.TrimSpace(rawCode))
	if code == "" {
		b.sendMessage(ctx, msg.Chat.ID, "Код не найден. Использование: /link <CODE>")
		return
	}

	status, statusCode, backendErr := b.backend.GetTelegramStatus(ctx, msg.From.ID)
	if statusCode == http.StatusOK && status.Linked {
		b.sendMessage(ctx, msg.Chat.ID, alreadyLinkedText(status.User, status.NotificationsEnabled))
		return
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNotFound && statusCode != http.StatusBadRequest {
		log.Printf("telegram status check before link failed (status=%d): %s", statusCode, backendErr)
		b.sendMessage(ctx, msg.Chat.ID, "Не удалось проверить текущую привязку. Попробуй позже.")
		return
	}

	statusCode, backendErr = b.backend.LinkTelegram(ctx, backend.TelegramLinkRequest{
		Code:             code,
		TelegramUserID:   msg.From.ID,
		TelegramUsername: msg.From.Username,
		TelegramFirst:    msg.From.FirstName,
		TelegramLast:     msg.From.LastName,
	})
	switch statusCode {
	case http.StatusOK:
		linkedStatus, linkedStatusCode, linkedErr := b.backend.GetTelegramStatus(ctx, msg.From.ID)
		if linkedStatusCode == http.StatusOK && linkedStatus.Linked {
			b.sendMessage(ctx, msg.Chat.ID, "Готово.\n"+alreadyLinkedText(linkedStatus.User, linkedStatus.NotificationsEnabled))
			return
		}
		if linkedStatusCode != http.StatusOK {
			log.Printf("telegram status check after link failed (status=%d): %s", linkedStatusCode, linkedErr)
		}
		b.sendMessage(ctx, msg.Chat.ID, "Готово. Telegram аккаунт успешно привязан.")
	case http.StatusBadRequest:
		b.sendMessage(ctx, msg.Chat.ID, "Код недействителен или истек. Запроси новый код в приложении.")
	case http.StatusConflict:
		b.sendMessage(ctx, msg.Chat.ID, "Этот Telegram уже привязан к другому пользователю. Сначала отвяжи его в профиле того аккаунта.")
	default:
		log.Printf("telegram link failed (status=%d): %s", statusCode, backendErr)
		b.sendMessage(ctx, msg.Chat.ID, "Не удалось привязать аккаунт. Попробуй позже.")
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
