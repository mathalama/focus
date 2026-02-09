package backend

type ErrorResponse struct {
	Error string `json:"error"`
}

type TelegramLinkRequest struct {
	Code             string `json:"code"`
	TelegramUserID   int64  `json:"telegram_user_id"`
	TelegramUsername string `json:"telegram_username"`
	TelegramFirst    string `json:"telegram_first_name"`
	TelegramLast     string `json:"telegram_last_name"`
}

type LinkedUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TelegramStatusResponse struct {
	Linked               bool        `json:"linked"`
	NotificationsEnabled bool        `json:"notifications_enabled"`
	User                 *LinkedUser `json:"user"`
}

type TelegramNotificationsRequest struct {
	TelegramUserID int64 `json:"telegram_user_id"`
	Enabled        bool  `json:"enabled"`
}
