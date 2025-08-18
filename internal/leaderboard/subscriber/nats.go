package subscriber

import (
	"context"
	"encoding/json"
	"go-confess-sins-api/pkg/models"
	"log/slog"

	"github.com/nats-io/nats.go"
)

// Store is an interface that our subscriber depends on.
// This decouples it from the concrete store implementation.
type Store interface {
	UpdateSinFromEvent(ctx context.Context, sin models.Sin) error
}

// NatsSubscriber listens for messages and updates the store.
type NatsSubscriber struct {
	nc    *nats.Conn
	store Store
}

func NewNatsSubscriber(nc *nats.Conn, store Store) *NatsSubscriber {
	return &NatsSubscriber{nc: nc, store: store}
}

// Start begins listening for "sins.updated" messages.
func (s *NatsSubscriber) Start() error {
	_, err := s.nc.Subscribe("sins.updated", func(msg *nats.Msg) {
		slog.Info("Received an update from NATS")
		var sin models.Sin
		if err := json.Unmarshal(msg.Data, &sin); err != nil {
			slog.Error("Error decoding NATS message", "error", err)
			return
		}

		// Update the leaderboard with the new data
		if err := s.store.UpdateSinFromEvent(context.Background(), sin); err != nil {
			slog.Error("Error updating leaderboard from event", "error", err)
		}
	})
	if err != nil {
		return err
	}
	slog.Info("Listening for sin updates from NATS...")
	return nil
}
