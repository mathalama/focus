package bot

import (
	"context"
	"log"
	"strings"
	"time"
)

func (b *Bot) runPolling(ctx context.Context) error {
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := b.tg.GetUpdates(ctx, offset, b.cfg.PollTimeout)
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
				b.handleCallbackQuery(ctx, *update.CallbackQuery)
				continue
			}
			if update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
				continue
			}

			b.handleMessage(ctx, *update.Message)
		}
	}
}
