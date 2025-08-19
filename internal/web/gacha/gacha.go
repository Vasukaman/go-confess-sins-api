package gacha

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MessageSender is the interface our GachaMachine uses to communicate
// with the outside world (e.g., a WebSocket server).
type MessageSender interface {
	SendToUser(userID string, payload []byte) error
}

// GachaMachine handles the game's core logic for all players.
type GachaMachine struct {
	mu           sync.Mutex
	sender       MessageSender
	players      map[string]*Player // Map of userID -> Player state
	rarities     map[string]*Rarity
	dropTables   map[int][]Item
	rng          *rand.Rand
	rollDuration time.Duration // The time cost of a roll.
}

// NewGachaMachine creates and initializes the machine.
// It now requires a MessageSender for communication.
func NewGachaMachine(sender MessageSender) *GachaMachine {

	// --- DEFINE RARITIES ---
	common := &Rarity{ID: "common", Name: "Common", Color: "#9e9e9e", WeightMultiplier: 10.0}
	uncommon := &Rarity{ID: "uncommon", Name: "Uncommon", Color: "#4caf50", WeightMultiplier: 5.0}
	rare := &Rarity{ID: "rare", Name: "Rare", Color: "#2196f3", WeightMultiplier: 2.0}
	mythical := &Rarity{ID: "mythical", Name: "Mythical", Color: "#9c27b0", WeightMultiplier: 0.5}
	legendary := &Rarity{ID: "legendary", Name: "Legendary", Color: "#ff9800", WeightMultiplier: 0.1}

	raritiesMap := map[string]*Rarity{
		"common":    common,
		"uncommon":  uncommon,
		"rare":      rare,
		"mythical":  mythical,
		"legendary": legendary,
	}

	// --- DEFINE ITEMS ---
	sev1Table := []Item{
		{ID: "bug", Emoji: "🐛", Name: "Simple Bug", RarityID: "common", BaseWeight: 100, BaseLuckValue: 1},
		{ID: "typo", Emoji: "🤦", Name: "Facepalm Typo", RarityID: "common", BaseWeight: 100, BaseLuckValue: 1},
		{ID: "performance", Emoji: "🔥", Name: "Performance Issue", RarityID: "uncommon", BaseWeight: 50, BaseLuckValue: 5},
		{ID: "refactor", Emoji: "🤔", Name: "Needless Refactor", RarityID: "rare", BaseWeight: 20, BaseLuckValue: 10},
		{ID: "deadsoul", Emoji: "💀", Name: "Dead Soul", RarityID: "legendary", BaseWeight: 1, BaseLuckValue: 100},
	}

	source := rand.NewSource(time.Now().UnixNano())

	return &GachaMachine{
		sender:       sender,
		players:      make(map[string]*Player),
		rarities:     raritiesMap,
		dropTables:   map[int][]Item{1: sev1Table},
		rng:          rand.New(source),
		rollDuration: 3 * time.Second, // The base duration of a roll.
	}
}

// Roll starts the gacha roll process for a specific user.
func (gm *GachaMachine) Roll(userID string, severity int) error {
	player, ok := gm.players[userID]
	if !ok {
		return fmt.Errorf("player %s not found", userID)
	}

	player.mu.Lock()
	defer player.mu.Unlock()

	if time.Now().Before(player.NextRollTime) {
		return fmt.Errorf("roll is on cooldown")
	}

	table, ok := gm.dropTables[severity]
	if !ok {
		return fmt.Errorf("no drop table for severity %d", severity)
	}

	unlockTime := time.Now().Add(gm.rollDuration)
	player.NextRollTime = unlockTime

	prizeBlueprint, _ := gm.rollOne(table, player)
	prizeInstance := gm.createItemInstance(prizeBlueprint)
	player.GachaSlot.Item = prizeInstance

	player.GachaSlot.isUnlockable = func() bool {
		return time.Now().After(unlockTime)
	}

	visualRoll := gm.createVisualRoll(table, player)
	winnerIndex := 24 // The "winning" item is the 25th in the list
	visualRoll[winnerIndex] = *prizeInstance

	// Use the communication helper to send the result
	gm.sendRollResult(userID, visualRoll, winnerIndex, prizeInstance)
	return nil
}

// SwapItems handles a player's request to swap items between two slots.
// SwapItems handles a player's request to swap items between two slots.
func (gm *GachaMachine) SwapItems(userID, sourceSlotID, destSlotID string) error {
	player, ok := gm.players[userID]
	if !ok {
		return fmt.Errorf("player %s not found", userID)
	}

	player.mu.Lock()
	defer player.mu.Unlock()

	sourceSlot := gm.findSlot(player, sourceSlotID)
	destSlot := gm.findSlot(player, destSlotID)

	if sourceSlot == nil || destSlot == nil {
		return fmt.Errorf("invalid slot ID")
	}

	if sourceSlot.IsLocked() || destSlot.IsLocked() {
		return fmt.Errorf("one or more slots are locked")
	}

	sourceSlot.Item, destSlot.Item = destSlot.Item, sourceSlot.Item

	// If we moved an item OUT of the gacha slot, it should no longer be time-locked.
	// This makes the slot instantly usable for the next roll.
	if sourceSlot.ID == "gacha_slot" {
		sourceSlot.isUnlockable = nil
	}
	if destSlot.ID == "gacha_slot" {
		destSlot.isUnlockable = nil
	}

	// Use the communication helper to send the updated state
	gm.sendPlayerStateUpdate(userID)
	return nil
}

func (gm *GachaMachine) AddPlayer(userID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Don't overwrite an existing player
	if _, exists := gm.players[userID]; exists {
		return
	}

	player := &Player{
		ID:   userID,
		Luck: 0,
		InventorySlots: [3]*Slot{
			{ID: "inventory_slot_0"},
			{ID: "inventory_slot_1"},
			{ID: "inventory_slot_2"},
		},
		GachaSlot: &Slot{ID: "gacha_slot"},
	}
	gm.players[userID] = player

	// Send the initial state to the new player
	gm.sendPlayerStateUpdate(userID)
}

// RemovePlayer cleans up a player's state when they disconnect.
func (gm *GachaMachine) RemovePlayer(userID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	delete(gm.players, userID)
}

// (Private helper functions like rollOne, createItemInstance, findSlot etc. would go here)
func (gm *GachaMachine) rollOne(table []Item, player *Player) (Item, bool) {
	if len(table) == 0 {
		return Item{}, false
	}

	// Step 1: Calculate the total weight of all items in the table.
	totalWeight := 0.0
	for _, item := range table {
		// We'll use our new flexible weight calculation method here.
		totalWeight += gm.calculateFinalWeight(item, player)
	}

	if totalWeight == 0 {
		// Avoid division by zero if all weights are somehow 0.
		return table[0], true
	}

	// Step 2: Pick a random number between 0 and the total weight.
	roll := gm.rng.Float64() * totalWeight

	// Step 3: Iterate through the items, subtracting their weight from the roll
	// until the roll is less than or equal to 0. That's our winner.
	for _, item := range table {
		roll -= gm.calculateFinalWeight(item, player)
		if roll <= 0 {
			return item, true
		}
	}

	// Fallback in case of floating point inaccuracies. Should rarely, if ever, be hit.
	return table[len(table)-1], true
}

func (gm *GachaMachine) createItemInstance(blueprint Item) *ItemInstance {
	return &ItemInstance{
		InstanceID: uuid.NewString(),
		Item:       &blueprint,
		Rarity:     gm.rarities[blueprint.RarityID],
		LuckValue:  blueprint.BaseLuckValue,
	}
}
func (gm *GachaMachine) createVisualRoll(table []Item, player *Player) []ItemInstance {
	var visualRoll []ItemInstance
	for i := 0; i < 30; i++ {
		itemBlueprint, _ := gm.rollOne(table, player)
		instance := gm.createItemInstance(itemBlueprint)
		visualRoll = append(visualRoll, *instance)
	}
	return visualRoll
}

func (gm *GachaMachine) findSlot(player *Player, slotID string) *Slot {
	if slotID == "gacha_slot" {
		return player.GachaSlot
	}
	for _, slot := range player.InventorySlots {
		if slot.ID == slotID {
			return slot
		}
	}
	return nil
}

func (gm *GachaMachine) calculateFinalWeight(item Item, player *Player) float64 {
	rarity, ok := gm.rarities[item.RarityID]
	if !ok {
		return 0 // Item has an unknown rarity, so it can't be rolled.
	}

	// Base calculation: Item's own weight multiplied by the rarity's modifier.
	finalWeight := item.BaseWeight * rarity.WeightMultiplier

	// --- FUTURE LOGIC GOES HERE ---
	// This is where you would add logic based on player.Luck.
	// For example:
	// if player.Luck > 50 {
	//     if rarity.ID == "rare" || rarity.ID == "legendary" {
	//         finalWeight *= 1.5 // 50% boost to high-tier items for lucky players
	//     }
	// }

	return finalWeight
}
