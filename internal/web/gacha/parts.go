package gacha

import (
	"sync"
	"time"
)

// GetDropTablePayload is the payload for a "get_droptable_info" command.
type GetDropTablePayload struct {
	Severity int `json:"severity"`
}

// DropTableInfoItem represents a single item in the droptable view.
type DropTableInfoItem struct {
	Item          Item    `json:"Item"`
	DropChance    float64 `json:"DropChance"`
	TimesObtained int     `json:"TimesObtained"`
}

// Rarity and Item structs remain the same as they define static game data.
type Rarity struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Color            string  `json:"color"`
	WeightMultiplier float64 `json:"weightMultiplier"`
}

type Item struct {
	ID            string  `json:"id"`
	Emoji         string  `json:"emoji"`
	Name          string  `json:"name"`
	RarityID      string  `json:"rarityId"`
	BaseWeight    float64 `json:"baseWeight"`
	BaseLuckValue int     `json:"baseLuckValue"`
}

// ItemInstance is a unique instance of an item that a player owns.
type ItemInstance struct {
	InstanceID string `json:"instanceId"`
	*Item
	Rarity    *Rarity `json:"rarity"`
	LuckValue int     `json:"luckValue"`
}

// --- NEW AND REFACTORED STRUCTS ---

// UnlockCheck is a function type for our custom locking logic.
// It returns 'true' if the slot should be considered unlocked.
type UnlockCheck func() bool

// Slot represents a single inventory or machine slot that can hold an item.
type Slot struct {
	ID           string        `json:"id"`
	Item         *ItemInstance `json:"item"`
	isUnlockable UnlockCheck   // This is our flexible, unexported lock function.
}

// IsLocked is the public method to check if a slot is accessible.
// It executes the custom unlock logic if it exists.
func (s *Slot) IsLocked() bool {
	if s.isUnlockable != nil {
		// If the custom check says it's NOT unlockable, then it's locked.
		return !s.isUnlockable()
	}
	// If no custom check exists, the slot is never locked.
	return false
}

// PlayerStatus defines the player's current state (e.g., rolling, idle).
type PlayerStatus string

// Player represents the complete state of a single user in the game.
type Player struct {
	mu              sync.Mutex
	ID              string         `json:"id"`
	NextRollTime    time.Time      `json:"-"` // Time after which the player can roll again.
	InventorySlots  [3]*Slot       `json:"inventorySlots"`
	GachaSlot       *Slot          `json:"gachaSlot"`
	Luck            float64        `json:"luck"`
	DiscoveredItems map[string]int `json:"-"`
}
