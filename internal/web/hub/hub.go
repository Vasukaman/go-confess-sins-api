package hub

import (
	"github.com/olahol/melody"
)

// Hub now only manages WebSocket connections.
type Hub struct {
	Melody *melody.Melody
}

func New() *Hub {
	return &Hub{
		Melody: melody.New(),
	}
}

// Broadcast is a simple method that can be called from anywhere.
func (h *Hub) Broadcast(msg []byte) {
	h.Melody.Broadcast(msg)
}
