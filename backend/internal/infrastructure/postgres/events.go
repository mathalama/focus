package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mathalama-focus/backend/internal/domain"
)

func (r *Repository) TrackEvent(
	ctx context.Context,
	userID, eventName, source string,
	properties map[string]any,
) error {
	eventName = strings.TrimSpace(eventName)
	if eventName == "" {
		return nil
	}
	if strings.TrimSpace(source) == "" {
		source = "api"
	}

	var propertiesRaw []byte
	var err error
	if len(properties) == 0 {
		propertiesRaw = []byte(`{}`)
	} else {
		propertiesRaw, err = json.Marshal(properties)
		if err != nil {
			return fmt.Errorf("marshal product event properties: %w", err)
		}
	}

	const query = `
		INSERT INTO product_events (user_id, event_name, source, properties)
		VALUES (NULLIF($1, '')::uuid, $2, $3, $4::jsonb)
	`
	if _, err := r.pool.Exec(ctx, query, strings.TrimSpace(userID), eventName, source, string(propertiesRaw)); err != nil {
		return fmt.Errorf("insert product event: %w", err)
	}
	return nil
}

func (r *Repository) ListRecentEvents(ctx context.Context, limit int, eventName string) ([]domain.ProductEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	eventName = strings.TrimSpace(eventName)
	query := `
		SELECT id, user_id::text, event_name, source, properties, created_at
		FROM product_events
	`
	args := make([]any, 0, 2)
	if eventName != "" {
		query += ` WHERE event_name = $1`
		args = append(args, eventName)
	}

	if len(args) == 0 {
		query += ` ORDER BY created_at DESC LIMIT $1`
		args = append(args, limit)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2`
		args = append(args, limit)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list recent product events: %w", err)
	}
	defer rows.Close()

	events := make([]domain.ProductEvent, 0, limit)
	for rows.Next() {
		var item domain.ProductEvent
		var userID string
		var rawProperties []byte
		if err := rows.Scan(&item.ID, &userID, &item.EventName, &item.Source, &rawProperties, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan product event: %w", err)
		}

		if strings.TrimSpace(userID) != "" {
			item.UserID = &userID
		}

		item.Properties = make(map[string]any)
		if len(rawProperties) > 0 {
			if err := json.Unmarshal(rawProperties, &item.Properties); err != nil {
				item.Properties = map[string]any{"raw": string(rawProperties)}
			}
		}
		events = append(events, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product events: %w", err)
	}

	return events, nil
}
