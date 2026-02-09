package postgres

import (
	"context"
	"errors"
	"fmt"

	"mathalama-focus/backend/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) ListItems(ctx context.Context) ([]domain.Item, error) {
	const query = `SELECT id, name, description, cost, type FROM items ORDER BY cost ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Item, 0)
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Cost, &item.Type); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) BuyItem(ctx context.Context, userID, itemID string) (domain.UserItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.UserItem{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var cost int
	if err := tx.QueryRow(ctx, `SELECT cost FROM items WHERE id = $1`, itemID).Scan(&cost); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserItem{}, domain.ErrNotFound
		}
		return domain.UserItem{}, fmt.Errorf("get item cost: %w", err)
	}

	var balance int
	if err := tx.QueryRow(ctx, `SELECT nectar_balance FROM users WHERE id = $1`, userID).Scan(&balance); err != nil {
		return domain.UserItem{}, fmt.Errorf("get user balance: %w", err)
	}

	if balance < cost {
		return domain.UserItem{}, domain.ErrInsufficientBalance
	}

	if _, err := tx.Exec(ctx, `UPDATE users SET nectar_balance = nectar_balance - $1 WHERE id = $2`, cost, userID); err != nil {
		return domain.UserItem{}, fmt.Errorf("deduct balance: %w", err)
	}

	var userItem domain.UserItem
	if err := tx.QueryRow(ctx,
		`INSERT INTO user_items (user_id, item_id) VALUES ($1, $2) RETURNING id, user_id, item_id, purchased_at`,
		userID, itemID,
	).Scan(&userItem.ID, &userItem.UserID, &userItem.ItemID, &userItem.PurchasedAt); err != nil {
		return domain.UserItem{}, fmt.Errorf("add item to user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.UserItem{}, fmt.Errorf("commit transaction: %w", err)
	}

	return userItem, nil
}

func (r *Repository) GetLeaderboard(ctx context.Context) ([]domain.LeaderboardEntry, error) {
	const query = `
		SELECT id, name, total_nectar_earned,
			RANK() OVER (ORDER BY total_nectar_earned DESC)::INT as rank
		FROM users
		ORDER BY total_nectar_earned DESC
		LIMIT 10
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]domain.LeaderboardEntry, 0)
	for rows.Next() {
		var entry domain.LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Name, &entry.TotalNectarEarned, &entry.Rank); err != nil {
			return nil, fmt.Errorf("scan leaderboard entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
