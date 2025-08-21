package gacha

import (
	"fmt"
	"log"
	"math"
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
	//Temp code, later use json
	sev1Table := []Item{
		// Common (Nature/Outdoors) - BaseLuckValue: 1-5
		{ID: "leaf", Emoji: "🍁", Name: "Fallen Leaf", RarityID: "common", BaseWeight: 100, BaseLuckValue: 1},
		{ID: "seedling", Emoji: "🌱", Name: "Tiny Seedling", RarityID: "common", BaseWeight: 95, BaseLuckValue: 2},
		{ID: "cloud", Emoji: "☁️", Name: "Fluffy Cloud", RarityID: "common", BaseWeight: 90, BaseLuckValue: 3},
		{ID: "rock", Emoji: "🪨", Name: "Smooth Rock", RarityID: "common", BaseWeight: 85, BaseLuckValue: 4},
		{ID: "rain", Emoji: "🌧️", Name: "Gentle Rain", RarityID: "common", BaseWeight: 80, BaseLuckValue: 5},
		{ID: "mushroom", Emoji: "🍄", Name: "Forest Mushroom", RarityID: "common", BaseWeight: 75, BaseLuckValue: 5},
		{ID: "snail", Emoji: "🐌", Name: "Slow Snail", RarityID: "common", BaseWeight: 70, BaseLuckValue: 4},
		{ID: "worm", Emoji: "🪱", Name: "Muddy Worm", RarityID: "common", BaseWeight: 65, BaseLuckValue: 3},

		// Uncommon (Wildlife) - BaseLuckValue: 6-15
		{ID: "butterfly", Emoji: "🦋", Name: "Monarch Butterfly", RarityID: "uncommon", BaseWeight: 50, BaseLuckValue: 6},
		{ID: "bee", Emoji: "🐝", Name: "Fuzzy Bee", RarityID: "uncommon", BaseWeight: 45, BaseLuckValue: 7},
		{ID: "owl", Emoji: "🦉", Name: "Wise Owl", RarityID: "uncommon", BaseWeight: 40, BaseLuckValue: 9},
		{ID: "deer", Emoji: "🦌", Name: "Majestic Deer", RarityID: "uncommon", BaseWeight: 35, BaseLuckValue: 12},
		{ID: "fox", Emoji: "🦊", Name: "Clever Fox", RarityID: "uncommon", BaseWeight: 30, BaseLuckValue: 15},

		// Rare (Mythical Creatures) - BaseLuckValue: 16-30
		{ID: "unicorn", Emoji: "🦄", Name: "Shining Unicorn", RarityID: "rare", BaseWeight: 20, BaseLuckValue: 18},
		{ID: "dragon", Emoji: "🐉", Name: "Green Dragon", RarityID: "rare", BaseWeight: 15, BaseLuckValue: 22},
		{ID: "phoenix", Emoji: "🔥", Name: "Reborn Phoenix", RarityID: "rare", BaseWeight: 10, BaseLuckValue: 28},

		// Mythical (Cosmic) - BaseLuckValue: 31-50
		{ID: "star", Emoji: "⭐", Name: "Falling Star", RarityID: "mythical", BaseWeight: 5, BaseLuckValue: 35},
		{ID: "planet", Emoji: "🪐", Name: "Distant Planet", RarityID: "mythical", BaseWeight: 3, BaseLuckValue: 45},

		// Legendary (Galactic Phenomena) - BaseLuckValue: 51-100
		{ID: "supernova", Emoji: "💥", Name: "Supernova", RarityID: "legendary", BaseWeight: 1, BaseLuckValue: 75},
		{ID: "blackhole", Emoji: "🕳️", Name: "Singular Black Hole", RarityID: "legendary", BaseWeight: 0.5, BaseLuckValue: 100},
	}

	sev2Table := []Item{
		// Common (Daily Life) - BaseLuckValue: 5-10
		{ID: "coffee", Emoji: "☕", Name: "Morning Coffee", RarityID: "common", BaseWeight: 100, BaseLuckValue: 5},
		{ID: "book", Emoji: "📖", Name: "Open Book", RarityID: "common", BaseWeight: 95, BaseLuckValue: 6},
		{ID: "key", Emoji: "🔑", Name: "House Key", RarityID: "common", BaseWeight: 90, BaseLuckValue: 7},
		{ID: "phone", Emoji: "📱", Name: "Smartphone", RarityID: "common", BaseWeight: 85, BaseLuckValue: 8},
		{ID: "headphones", Emoji: "🎧", Name: "Headphones", RarityID: "common", BaseWeight: 80, BaseLuckValue: 9},
		{ID: "soda", Emoji: "🥤", Name: "Ice Cold Soda", RarityID: "common", BaseWeight: 75, BaseLuckValue: 10},
		{ID: "pen", Emoji: "🖊️", Name: "Ink Pen", RarityID: "common", BaseWeight: 70, BaseLuckValue: 9},
		{ID: "camera", Emoji: "📸", Name: "Old Camera", RarityID: "common", BaseWeight: 65, BaseLuckValue: 8},
		{ID: "wallet", Emoji: "👛", Name: "Leather Wallet", RarityID: "common", BaseWeight: 60, BaseLuckValue: 7},
		{ID: "watch", Emoji: "⌚", Name: "Wrist Watch", RarityID: "common", BaseWeight: 55, BaseLuckValue: 6},

		// Uncommon (Urban Exploration) - BaseLuckValue: 11-20
		{ID: "subway", Emoji: "🚇", Name: "Subway Train", RarityID: "uncommon", BaseWeight: 40, BaseLuckValue: 12},
		{ID: "skyscraper", Emoji: "🏙️", Name: "Skyscraper", RarityID: "uncommon", BaseWeight: 35, BaseLuckValue: 14},
		{ID: "streetart", Emoji: "🎨", Name: "Vibrant Street Art", RarityID: "uncommon", BaseWeight: 30, BaseLuckValue: 16},
		{ID: "taxi", Emoji: "🚕", Name: "Yellow Taxi", RarityID: "uncommon", BaseWeight: 25, BaseLuckValue: 18},
		{ID: "bridge", Emoji: "🌉", Name: "City Bridge", RarityID: "uncommon", BaseWeight: 20, BaseLuckValue: 20},

		// Rare (Future Tech) - BaseLuckValue: 21-40
		{ID: "robot", Emoji: "🤖", Name: "Service Robot", RarityID: "rare", BaseWeight: 15, BaseLuckValue: 25},
		{ID: "drone", Emoji: "🚁", Name: "Delivery Drone", RarityID: "rare", BaseWeight: 10, BaseLuckValue: 30},
		{ID: "cyborg", Emoji: "🦾", Name: "Cybernetic Arm", RarityID: "rare", BaseWeight: 8, BaseLuckValue: 35},

		// Mythical (Interdimensional) - BaseLuckValue: 41-70
		{ID: "portal", Emoji: "🌀", Name: "Interdimensional Portal", RarityID: "mythical", BaseWeight: 4, BaseLuckValue: 50},
		{ID: "alien", Emoji: "👽", Name: "Friendly Alien", RarityID: "mythical", BaseWeight: 2, BaseLuckValue: 60},

		// Legendary (Reality-Bending) - BaseLuckValue: 71-150
		{ID: "glitch", Emoji: "👾", Name: "Reality Glitch", RarityID: "legendary", BaseWeight: 1, BaseLuckValue: 120},
	}

	sev3Table := []Item{
		// Common (Food & Drink) - BaseLuckValue: 10-20
		{ID: "pizza", Emoji: "🍕", Name: "Slice of Pizza", RarityID: "common", BaseWeight: 100, BaseLuckValue: 10},
		{ID: "sushi", Emoji: "🍣", Name: "Sushi Roll", RarityID: "common", BaseWeight: 95, BaseLuckValue: 12},
		{ID: "taco", Emoji: "🌮", Name: "Crunchy Taco", RarityID: "common", BaseWeight: 90, BaseLuckValue: 14},
		{ID: "burger", Emoji: "🍔", Name: "Cheeseburger", RarityID: "common", BaseWeight: 85, BaseLuckValue: 16},
		{ID: "ramen", Emoji: "🍜", Name: "Bowl of Ramen", RarityID: "common", BaseWeight: 80, BaseLuckValue: 18},
		{ID: "donut", Emoji: "🍩", Name: "Glazed Donut", RarityID: "common", BaseWeight: 75, BaseLuckValue: 20},
		{ID: "icecream", Emoji: "🍦", Name: "Ice Cream Cone", RarityID: "common", BaseWeight: 70, BaseLuckValue: 19},
		{ID: "milk", Emoji: "🥛", Name: "Glass of Milk", RarityID: "common", BaseWeight: 65, BaseLuckValue: 18},
		{ID: "popcorn", Emoji: "🍿", Name: "Popcorn Bucket", RarityID: "common", BaseWeight: 60, BaseLuckValue: 17},
		{ID: "cake", Emoji: "🎂", Name: "Birthday Cake", RarityID: "common", BaseWeight: 55, BaseLuckValue: 16},

		// Uncommon (Tools & Crafts) - BaseLuckValue: 21-35
		{ID: "hammer", Emoji: "🔨", Name: "Steel Hammer", RarityID: "uncommon", BaseWeight: 40, BaseLuckValue: 25},
		{ID: "paintbrush", Emoji: "🖌️", Name: "Artist's Paintbrush", RarityID: "uncommon", BaseWeight: 35, BaseLuckValue: 28},
		{ID: "scissors", Emoji: "✂️", Name: "Sharp Scissors", RarityID: "uncommon", BaseWeight: 30, BaseLuckValue: 30},
		{ID: "thread", Emoji: "🧵", Name: "Needle and Thread", RarityID: "uncommon", BaseWeight: 25, BaseLuckValue: 32},
		{ID: "pottery", Emoji: "🏺", Name: "Ancient Pottery", RarityID: "uncommon", BaseWeight: 20, BaseLuckValue: 35},

		// Rare (Treasures) - BaseLuckValue: 36-60
		{ID: "gem", Emoji: "💎", Name: "Sparkling Gem", RarityID: "rare", BaseWeight: 15, BaseLuckValue: 45},
		{ID: "crown", Emoji: "👑", Name: "Royal Crown", RarityID: "rare", BaseWeight: 10, BaseLuckValue: 50},
		{ID: "treasurechest", Emoji: " chests", Name: "Treasure Chest", RarityID: "rare", BaseWeight: 8, BaseLuckValue: 55},

		// Mythical (Artifacts) - BaseLuckValue: 61-100
		{ID: "lamp", Emoji: "🧞", Name: "Genie's Lamp", RarityID: "mythical", BaseWeight: 4, BaseLuckValue: 80},
		{ID: "sword", Emoji: "⚔️", Name: "Excalibur", RarityID: "mythical", BaseWeight: 2, BaseLuckValue: 90},

		// Legendary (Cosmic Entities) - BaseLuckValue: 101-200
		{ID: "god", Emoji: "✨", Name: "Celestial Being", RarityID: "legendary", BaseWeight: 1, BaseLuckValue: 175},
	}

	sev4Table := []Item{
		// Common (Sports) - BaseLuckValue: 15-30
		{ID: "soccerball", Emoji: "⚽", Name: "Soccer Ball", RarityID: "common", BaseWeight: 100, BaseLuckValue: 15},
		{ID: "basketball", Emoji: "🏀", Name: "Basketball", RarityID: "common", BaseWeight: 95, BaseLuckValue: 17},
		{ID: "baseball", Emoji: "⚾", Name: "Baseball", RarityID: "common", BaseWeight: 90, BaseLuckValue: 19},
		{ID: "football", Emoji: "🏈", Name: "American Football", RarityID: "common", BaseWeight: 85, BaseLuckValue: 21},
		{ID: "helmet", Emoji: " helmets", Name: "Protective Helmet", RarityID: "common", BaseWeight: 80, BaseLuckValue: 23},
		{ID: "gloves", Emoji: " gloves", Name: "Boxing Gloves", RarityID: "common", BaseWeight: 75, BaseLuckValue: 25},
		{ID: "medal", Emoji: "🥇", Name: "Gold Medal", RarityID: "common", BaseWeight: 70, BaseLuckValue: 27},
		{ID: "trophy", Emoji: "🏆", Name: "Small Trophy", RarityID: "common", BaseWeight: 65, BaseLuckValue: 29},
		{ID: "whistle", Emoji: " whistle", Name: "Coach's Whistle", RarityID: "common", BaseWeight: 60, BaseLuckValue: 30},
		{ID: "swim", Emoji: "🏊", Name: "Swimmer", RarityID: "common", BaseWeight: 55, BaseLuckValue: 28},

		// Uncommon (Exploration & Adventure) - BaseLuckValue: 31-50
		{ID: "compass", Emoji: "🧭", Name: "Old Compass", RarityID: "uncommon", BaseWeight: 40, BaseLuckValue: 35},
		{ID: "map", Emoji: "🗺️", Name: "Ancient Map", RarityID: "uncommon", BaseWeight: 35, BaseLuckValue: 38},
		{ID: "tent", Emoji: "⛺", Name: "Camping Tent", RarityID: "uncommon", BaseWeight: 30, BaseLuckValue: 41},
		{ID: "hiking", Emoji: "🥾", Name: "Hiking Boots", RarityID: "uncommon", BaseWeight: 25, BaseLuckValue: 45},
		{ID: "parachute", Emoji: "🪂", Name: "Parachute", RarityID: "uncommon", BaseWeight: 20, BaseLuckValue: 50},

		// Rare (Vehicles) - BaseLuckValue: 51-80
		{ID: "car", Emoji: "🚗", Name: "Sports Car", RarityID: "rare", BaseWeight: 15, BaseLuckValue: 60},
		{ID: "spaceship", Emoji: "🚀", Name: "Rocket Ship", RarityID: "rare", BaseWeight: 10, BaseLuckValue: 65},
		{ID: "submarine", Emoji: " submarine", Name: "Submarine", RarityID: "rare", BaseWeight: 8, BaseLuckValue: 70},

		// Mythical (Fantasy Travel) - BaseLuckValue: 81-120
		{ID: "flyingcarpet", Emoji: "🧞‍♂️", Name: "Flying Carpet", RarityID: "mythical", BaseWeight: 4, BaseLuckValue: 100},
		{ID: "teleporter", Emoji: "⚛️", Name: "Quantum Teleporter", RarityID: "mythical", BaseWeight: 2, BaseLuckValue: 110},

		// Legendary (Travel to other dimensions) - BaseLuckValue: 121-250
		{ID: "wormhole", Emoji: "🌀", Name: "Stable Wormhole", RarityID: "legendary", BaseWeight: 1, BaseLuckValue: 220},
	}

	sev5Table := []Item{
		// Common (Music) - BaseLuckValue: 20-40
		{ID: "notes", Emoji: "🎶", Name: "Musical Notes", RarityID: "common", BaseWeight: 100, BaseLuckValue: 20},
		{ID: "microphone", Emoji: "🎤", Name: "Microphone", RarityID: "common", BaseWeight: 95, BaseLuckValue: 22},
		{ID: "guitar", Emoji: "🎸", Name: "Electric Guitar", RarityID: "common", BaseWeight: 90, BaseLuckValue: 24},
		{ID: "drums", Emoji: "🥁", Name: "Drum Set", RarityID: "common", BaseWeight: 85, BaseLuckValue: 26},
		{ID: "speaker", Emoji: "🔊", Name: "Loud Speaker", RarityID: "common", BaseWeight: 80, BaseLuckValue: 28},
		{ID: "piano", Emoji: "🎹", Name: "Grand Piano", RarityID: "common", BaseWeight: 75, BaseLuckValue: 30},
		{ID: "violin", Emoji: "🎻", Name: "Wooden Violin", RarityID: "common", BaseWeight: 70, BaseLuckValue: 32},
		{ID: "trumpet", Emoji: "🎺", Name: "Shiny Trumpet", RarityID: "common", BaseWeight: 65, BaseLuckValue: 34},
		{ID: "saxophone", Emoji: "🎷", Name: "Cool Saxophone", RarityID: "common", BaseWeight: 60, BaseLuckValue: 36},
		{ID: "headset", Emoji: "🎧", Name: "DJ Headset", RarityID: "common", BaseWeight: 55, BaseLuckValue: 38},

		// Uncommon (Performance) - BaseLuckValue: 41-60
		{ID: "clapper", Emoji: "🎬", Name: "Film Clapper", RarityID: "uncommon", BaseWeight: 40, BaseLuckValue: 45},
		{ID: "mask", Emoji: "🎭", Name: "Theater Mask", RarityID: "uncommon", BaseWeight: 35, BaseLuckValue: 48},
		{ID: "spotlight", Emoji: "💡", Name: "Spotlight", RarityID: "uncommon", BaseWeight: 30, BaseLuckValue: 52},
		{ID: "ballet", Emoji: "🩰", Name: "Ballet Shoes", RarityID: "uncommon", BaseWeight: 25, BaseLuckValue: 55},
		{ID: "circus", Emoji: "🎪", Name: "Circus Tent", RarityID: "uncommon", BaseWeight: 20, BaseLuckValue: 60},

		// Rare (Fine Arts) - BaseLuckValue: 61-100
		{ID: "sculpture", Emoji: "🗿", Name: "Abstract Sculpture", RarityID: "rare", BaseWeight: 15, BaseLuckValue: 75},
		{ID: "statue", Emoji: "🗽", Name: "Statue of Liberty", RarityID: "rare", BaseWeight: 10, BaseLuckValue: 85},
		{ID: "painting", Emoji: "🖼️", Name: "Masterpiece Painting", RarityID: "rare", BaseWeight: 8, BaseLuckValue: 95},

		// Mythical (Creative Concepts) - BaseLuckValue: 101-150
		{ID: "idea", Emoji: "💡", Name: "Brilliant Idea", RarityID: "mythical", BaseWeight: 4, BaseLuckValue: 120},
		{ID: "muse", Emoji: "🎨", Name: "Creative Muse", RarityID: "mythical", BaseWeight: 2, BaseLuckValue: 140},

		// Legendary (Inspiration) - BaseLuckValue: 151-300
		{ID: "spark", Emoji: "✨", Name: "Divine Spark of Creation", RarityID: "legendary", BaseWeight: 1, BaseLuckValue: 270},
	}

	source := rand.NewSource(time.Now().UnixNano())

	return &GachaMachine{
		sender:   sender,
		players:  make(map[string]*Player),
		rarities: raritiesMap,
		dropTables: map[int][]Item{
			1: sev1Table,
			2: sev2Table,
			3: sev3Table,
			4: sev4Table,
			5: sev5Table},
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
	player.DiscoveredItems[visualRoll[winnerIndex].ID]++
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

	gm.recalculatePlayerLuck(player)
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
		ID:              userID,
		Luck:            0,
		DiscoveredItems: make(map[string]int),
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
	finalWeight := item.BaseWeight * rarity.WeightMultiplier // 0.5*0.1=0.05

	if player.Luck > 0 {
		if rarity.ID == "rare" || rarity.ID == "legendary" || rarity.ID == "mythical" {
			// Increase chance for higher tier items
			finalWeight *= (1 + player.Luck/100.0) // Using 100 for a more balanced percentage
		} else {
			// Decrease chance for lower tier items, but never let it go below zero
			multiplier := math.Max(0, 1-player.Luck/100.0)
			finalWeight *= multiplier
		}
	}

	return finalWeight
}

func (gm *GachaMachine) HandleGetDropTableInfo(userID string, severity int) error {
	// ADD THIS LOG
	log.Printf("[DEBUG] User %s: Entering HandleGetDropTableInfo for severity %d", userID, severity)

	player, ok := gm.players[userID]
	if !ok {
		return fmt.Errorf("player %s not found", userID)
	}

	table, ok := gm.dropTables[severity]
	if !ok {
		return fmt.Errorf("no drop table for severity %d", severity)
	}

	player.mu.Lock()
	defer player.mu.Unlock()

	totalWeight := 0.0
	for _, item := range table {
		totalWeight += gm.calculateFinalWeight(item, player)
	}

	// ADD THIS LOG
	log.Printf("[DEBUG] User %s: Calculated totalWeight: %f", userID, totalWeight)

	if totalWeight == 0 {
		return fmt.Errorf("drop table for severity %d has zero total weight", severity)
	}

	payload := make([]DropTableInfoItem, len(table))
	for i, item := range table {
		finalWeight := gm.calculateFinalWeight(item, player)
		chance := finalWeight / totalWeight

		payload[i] = DropTableInfoItem{
			Item:          item,
			DropChance:    chance,
			TimesObtained: player.DiscoveredItems[item.ID],
		}
	}

	// ADD THIS LOG
	log.Printf("[DEBUG] User %s: Successfully built payload, about to send.", userID)

	gm.sendDropTableInfo(userID, payload)
	return nil
}

func (gm *GachaMachine) recalculatePlayerLuck(player *Player) {
	var totalLuck float64 = 0
	for _, slot := range player.InventorySlots {
		if slot.Item != nil {
			totalLuck += float64(slot.Item.LuckValue)
		}
	}
	player.Luck = totalLuck
}
