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
		client:    &http.Client{Timeout: 10 * time.Second},
		ttlMin:    ttlMin,
	}
}

func (s *ResendService) SendVerificationEmail(ctx context.Context, toEmail, toName, verifyLink string) error {
	payload := map[string]any{
		"from":    s.fromEmail,
		"to":      []string{toEmail},
		"subject": "Verify your Mathalama Focus account",
		"html": fmt.Sprintf(
			`<p>Hello %s,</p><p>Confirm your email to activate your account:</p><p><a href="%s">Verify email</a></p><p>This link expires in %d minutes.</p>`,
			htmlEscape(strings.TrimSpace(toName)),
			htmlEscape(strings.TrimSpace(verifyLink)),
			s.ttlMin,
		),
	}

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

func (s *ResendService) SendPasswordResetEmail(ctx context.Context, toEmail, toName, resetLink string) error {
	payload := map[string]any{
		"from":    s.fromEmail,
		"to":      []string{toEmail},
		"subject": "Reset your Mathalama Focus password",
		"html": fmt.Sprintf(
			`<p>Hello %s,</p><p>Click the link below to reset your password:</p><p><a href="%s">Reset password</a></p><p>This link expires in 30 minutes. If you didn't request this, you can safely ignore this email.</p>`,
			htmlEscape(strings.TrimSpace(toName)),
			htmlEscape(strings.TrimSpace(resetLink)),
		),
	}

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
