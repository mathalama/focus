package postgres

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"strings"

	"mathalama-focus/backend/internal/domain"
)

func normalizeSessionStatus(session *domain.FocusSession) {
	if session == nil {
		return
	}
	if session.Status == "cancelled" {
		session.Status = "abandoned"
	}
}

func clampScore(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return roundToTenth(value)
}

func roundToTenth(value float64) float64 {
	return math.Round(value*10) / 10
}


func generateRandomToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", errors.New("invalid token length")
	}

	raw, err := randomBytes(byteLength)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func randomBytes(length int) ([]byte, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}
