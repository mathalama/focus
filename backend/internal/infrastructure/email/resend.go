package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ResendService struct {
	apiKey    string
	fromEmail string
	client    *http.Client
	ttlMin    int
}

func NewResendService(apiKey, fromEmail string, ttlMin int) *ResendService {
	return &ResendService{
		apiKey:    strings.TrimSpace(apiKey),
		fromEmail: strings.TrimSpace(fromEmail),
		client:    &http.Client{Timeout: 30 * time.Second},
		ttlMin:    ttlMin,
	}
}

func (s *ResendService) SendVerificationEmail(ctx context.Context, toEmail, toName, verifyLink, idempotencyKey string) error {
	htmlBody := s.buildEmailHTML(
		fmt.Sprintf("Hello %s,", htmlEscape(strings.TrimSpace(toName))),
		"Welcome to Mathalama Focus! Please confirm your email address to activate your account and start your focus journey.",
		verifyLink,
		"Verify Email Address",
		fmt.Sprintf("This link will expire in %d minutes.", s.ttlMin),
	)

	payload := map[string]any{
		"from":    s.fromEmail,
		"to":      []string{toEmail},
		"subject": "Verify your Mathalama Focus account",
		"html":    htmlBody,
	}

	return s.send(ctx, payload, idempotencyKey)
}

func (s *ResendService) SendPasswordResetEmail(ctx context.Context, toEmail, toName, resetLink, idempotencyKey string) error {
	htmlBody := s.buildEmailHTML(
		fmt.Sprintf("Hello %s,", htmlEscape(strings.TrimSpace(toName))),
		"We received a request to reset your password. Click the button below to choose a new one.",
		resetLink,
		"Reset Password",
		"This link will expire in 30 minutes. If you didn't request this, you can safely ignore this email.",
	)

	payload := map[string]any{
		"from":    s.fromEmail,
		"to":      []string{toEmail},
		"subject": "Reset your Mathalama Focus password",
		"html":    htmlBody,
	}

	return s.send(ctx, payload, idempotencyKey)
}

func (s *ResendService) send(ctx context.Context, payload map[string]any, idempotencyKey string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("resend api error (status=%d): %s", resp.StatusCode, strings.TrimSpace(string(raw)))
}

func (s *ResendService) buildEmailHTML(greeting, message, actionURL, actionText, footerInfo string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mathalama Focus</title>
    <style>
        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: #050505;
            color: #e5e5e5;
            margin: 0;
            padding: 0;
            line-height: 1.6;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            padding: 20px;
        }
        .card {
            background-color: #121212;
            border: 1px solid #262626;
            border-radius: 12px;
            padding: 40px;
            text-align: center;
        }
        .logo {
            font-family: 'JetBrains Mono', monospace;
            font-weight: bold;
            font-size: 24px;
            letter-spacing: -1px;
            color: #ffffff;
            margin-bottom: 30px;
            text-transform: uppercase;
        }
        .greeting {
            font-size: 18px;
            font-weight: 600;
            color: #ffffff;
            margin-bottom: 16px;
        }
        .message {
            font-size: 16px;
            color: #a3a3a3;
            margin-bottom: 32px;
        }
        .button {
            display: inline-block;
            background-color: #ffffff;
            color: #000000 !important;
            text-decoration: none;
            padding: 14px 32px;
            border-radius: 8px;
            font-weight: 700;
            font-size: 16px;
            transition: opacity 0.2s;
        }
        .footer {
            margin-top: 32px;
            font-size: 13px;
            color: #525252;
            text-align: center;
        }
        .divider {
            height: 1px;
            background-color: #262626;
            margin: 32px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">MATHALAMA FOCUS</div>
            <div class="greeting">%s</div>
            <div class="message">%s</div>
            <a href="%s" class="button">%s</a>
            <div class="divider"></div>
            <div class="footer">
                %s<br>
                &copy; 2026 Mathalama. All rights reserved.
            </div>
        </div>
    </div>
</body>
</html>
`, greeting, message, actionURL, actionText, footerInfo)
}

func htmlEscape(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(value)
}
