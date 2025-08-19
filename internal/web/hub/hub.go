package hub

import (
	"go-confess-sins-api/internal/web/gacha"
	"log"

	"github.com/olahol/melody"
)

type MelodySender struct {
	m *melody.Melody
}

// SendToUser finds the correct user's session and sends the message.
func (s *MelodySender) SendToUser(userID string, payload []byte) error {
	// Find all sessions that match the userID and send the message.
	// In practice, there should only be one.
	return s.m.BroadcastFilter(payload, func(q *melody.Session) bool {
		val, exists := q.Get("userID")
		if !exists {
			return false
		}
		return val.(string) == userID
	})
}

// Hub now only manages WebSocket connections.
type Hub struct {
	Melody       *melody.Melody
	GachaMachine *gacha.GachaMachine
}

func New() *Hub {
	m := melody.New()

	// 1. Create the sender bridge.
	sender := &MelodySender{m: m}

	// 2. Create the GachaMachine, injecting the sender.
	gachaMachine := gacha.NewGachaMachine(sender)

	return &Hub{
		Melody:       m,
		GachaMachine: gachaMachine,
	}
}

// Broadcast is a simple method that can be called from anywhere.
func (h *Hub) Broadcast(msg []byte) {
	h.Melody.Broadcast(msg)
}

func (h *Hub) SetupHandlers() {
	// Runs when a new user connects.
	h.Melody.HandleConnect(func(s *melody.Session) {
		userID, exists := s.Get("userID")
		if !exists {
			log.Println("User connected without a userID, disconnecting.")
			s.Close()
			return
		}
		log.Printf("User %s connected.", userID.(string))
		h.GachaMachine.AddPlayer(userID.(string))
	})

	// Runs when a user disconnects.
	h.Melody.HandleDisconnect(func(s *melody.Session) {
		userID, exists := s.Get("userID")
		if !exists {
			return
		}
		log.Printf("User %s disconnected.", userID.(string))
		h.GachaMachine.RemovePlayer(userID.(string))
	})

	// Runs for every message received.
	h.Melody.HandleMessage(func(s *melody.Session, msg []byte) {
		userID, exists := s.Get("userID")
		if !exists {
			return // Cannot process message without a user.
		}
		// Route the raw message directly to the gacha machine.
		h.GachaMachine.HandleMessage(userID.(string), msg)
	})
}
