package bot

import (
	"fmt"
	"strings"

	"mathalama-focus/telegram-bot/internal/backend"
	"mathalama-focus/telegram-bot/internal/telegram"
)

func helpText() string {
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

func notLinkedText(appURL string) string {
	lines := []string{
		"Telegram пока не привязан к аккаунту.",
		"",
		"Что сделать:",
	}

	if appURL != "" {
		lines = append(lines, "1) Открой приложение: "+appURL)
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

func alreadyLinkedText(user *backend.LinkedUser, notificationsEnabled bool) string {
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

// ActionKeyboard returns the default inline keyboard shown with every message.
func ActionKeyboard() *telegram.InlineKeyboardMarkup {
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "Включить", CallbackData: telegram.CallbackNotifyOn},
				{Text: "Выключить", CallbackData: telegram.CallbackNotifyOff},
			},
			{
				{Text: "Статус", CallbackData: telegram.CallbackStatus},
				{Text: "Помощь", CallbackData: telegram.CallbackHelp},
			},
		},
	}
}
