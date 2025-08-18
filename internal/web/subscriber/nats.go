package subscriber

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

// Hub is an interface that our listener depends on.
type Hub interface {
	Broadcast(msg []byte)
}

// NatsListener listens for messages and tells the hub to broadcast.
type NatsListener struct {
	nc  *nats.Conn
	hub Hub
}

func New(natsURL string, hub Hub) (*NatsListener, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	return &NatsListener{nc: nc, hub: hub}, nil
}

// Start begins listening for "sins.updated" messages.
func (s *NatsListener) Start() {
	_, err := s.nc.Subscribe("sins.updated", func(msg *nats.Msg) {
		slog.Info("NATS update received, telling hub to broadcast.")
		s.hub.Broadcast([]byte(`{"type": "update"}`))
	})
	if err != nil {
		slog.Error("Failed to subscribe to NATS", "error", err)
	}
	slog.Info("NATS listener started.")
}

func (s *NatsListener) Close() {
	s.nc.Close()
}
