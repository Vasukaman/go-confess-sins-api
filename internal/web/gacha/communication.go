// In gacha/communication.go

package gacha

import (
	"encoding/json"
	"fmt"
	"log"
)

// --- 1. Define Message Structures (The API Contract) ---

// OutgoingMessage is a generic wrapper for all messages sent from the server.
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// PlayerStateUpdatePayload is the payload for updating the client's view.
type PlayerStateUpdatePayload struct {
	InventorySlots [3]*Slot `json:"inventorySlots"`
	GachaSlot      *Slot    `json:"gachaSlot"`
	Luck           float64  `json:"luck"`
}

// RollResultPayload is the payload sent when a roll is initiated.
type RollResultPayload struct {
	Reel        []ItemInstance `json:"reel"`
	WinnerIndex int            `json:"winnerIndex"`
	Prize       *ItemInstance  `json:"prize"`
	RollTime    float64        `json:"rollTime"` // Send duration in seconds
}

// ErrorPayload is used to send errors back to the client.
type ErrorPayload struct {
	Message string `json:"message"`
}

// IncomingCommand is a generic wrapper for messages from the client.
type IncomingCommand struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SwapPayload is the expected payload for a "swap_items" command.
type SwapPayload struct {
	SourceSlotID string `json:"sourceSlotId"`
	DestSlotID   string `json:"destSlotId"`
}

// RollPayload is the expected payload for a "start_roll" command.
type RollPayload struct {
	Severity int `json:"severity"`
}

// --- 2. Create the Public Message Handler ---

// HandleMessage is the single entry point for all incoming WebSocket messages.
// It parses the command and routes it to the correct internal function.
func (gm *GachaMachine) HandleMessage(userID string, rawMessage []byte) {
	var cmd IncomingCommand
	if err := json.Unmarshal(rawMessage, &cmd); err != nil {
		gm.sendError(userID, "Invalid command format")
		return
	}

	switch cmd.Type {
	case "start_roll":
		var p RollPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			gm.sendError(userID, "Invalid payload for start_roll")
			return
		}
		// The Roll function will call the appropriate send methods on success
		if err := gm.Roll(userID, p.Severity); err != nil {
			gm.sendError(userID, err.Error())
		}

	case "swap_items":
		var p SwapPayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			gm.sendError(userID, "Invalid payload for swap_items")
			return
		}
		// The SwapItems function will call the appropriate send methods on success
		if err := gm.SwapItems(userID, p.SourceSlotID, p.DestSlotID); err != nil {
			gm.sendError(userID, err.Error())
		}

	case "get_droptable_info":
		// ADD THIS LOG to see if we're entering the correct case
		log.Printf("[DEBUG] User %s: Handling 'get_droptable_info'", userID)

		var p GetDropTablePayload
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			log.Printf("[ERROR] User %s: Failed to parse get_droptable_info payload: %v", userID, err)
			gm.sendError(userID, "Invalid payload for get_droptable_info")
			return
		}
		if err := gm.HandleGetDropTableInfo(userID, p.Severity); err != nil {
			// This will log any error returned from the main logic function
			log.Printf("[ERROR] User %s: Error in HandleGetDropTableInfo: %v", userID, err)
			gm.sendError(userID, err.Error())
		}

	default:
		gm.sendError(userID, fmt.Sprintf("Unknown command type: %s", cmd.Type))
	}
}

func (gm *GachaMachine) sendDropTableInfo(userID string, payload []DropTableInfoItem) {
	gm.send(userID, "droptable_info_update", payload)
}

// send is a private generic helper to marshal and send any message.
func (gm *GachaMachine) send(userID, msgType string, payload interface{}) {
	msg := OutgoingMessage{
		Type:    msgType,
		Payload: payload,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message for user %s: %v", userID, err)
		return
	}

	if err := gm.sender.SendToUser(userID, bytes); err != nil {
		log.Printf("Error sending message to user %s: %v", userID, err)
	}
}

// sendPlayerStateUpdate sends the complete, current state of the player's slots.
func (gm *GachaMachine) sendPlayerStateUpdate(userID string) {
	player, ok := gm.players[userID]
	if !ok {
		return // Can't send update for a non-existent player
	}

	payload := PlayerStateUpdatePayload{
		InventorySlots: player.InventorySlots,
		GachaSlot:      player.GachaSlot,
		Luck:           player.Luck,
	}
	gm.send(userID, "player_state_update", payload)
}

// sendRollResult sends the outcome of a gacha roll.
func (gm *GachaMachine) sendRollResult(userID string, reel []ItemInstance, winnerIndex int, prize *ItemInstance) {
	payload := RollResultPayload{
		Reel:        reel,
		WinnerIndex: winnerIndex,
		Prize:       prize,
		RollTime:    gm.rollDuration.Seconds(),
	}
	gm.send(userID, "roll_result", payload)
}

// sendError sends a structured error message to the client.
func (gm *GachaMachine) sendError(userID string, message string) {
	payload := ErrorPayload{
		Message: message,
	}
	gm.send(userID, "error", payload)
}
