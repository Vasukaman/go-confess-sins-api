// in /static/gachaManager.js

export class GachaManager {
    constructor(socket) {
        if (!socket) {
            throw new Error("GachaManager requires a WebSocket instance.");
        }
        this.socket = socket;
        this.firstSelectedSlotId = null;

        // Get references to all the UI elements
        this.rollButton = document.getElementById('gacha-roll-button');
        this.reelDisplay = document.getElementById('gacha-reel');
        this.slotElements = {
            'gacha_slot': document.getElementById('gacha_slot'),
            'inventory_slot_0': document.getElementById('inventory_slot_0'),
            'inventory_slot_1': document.getElementById('inventory_slot_1'),
            'inventory_slot_2': document.getElementById('inventory_slot_2'),
        };
    }

    // Call this to set up all event listeners
    init() {
        this.rollButton.addEventListener('click', () => this.sendRollRequest());

        // Add a single click listener for all slots
        for (const slotId in this.slotElements) {
            this.slotElements[slotId].addEventListener('click', (event) => this.handleSlotClick(event));
        }
    }

    // Call this from your main WebSocket's onmessage handler to route messages
    handleServerMessage(data) {
        console.log("GachaManager received:", data);
        switch (data.type) {
            case "roll_result":
                this.updateReel(data.payload);
                break;
            case "player_state_update":
                this.updateAllSlots(data.payload);
                break;
            case "error":
                console.error("Gacha Error:", data.payload.message);
                alert(`Gacha Error: ${data.payload.message}`);
                break;
        }
    }

    sendRollRequest() {
        const message = {
            type: "start_roll",
            payload: {
                severity: 1 // Hardcoded for testing
            }
        };
        this.socket.send(JSON.stringify(message));
    }

    handleSlotClick(event) {
        const clickedSlotId = event.currentTarget.id;

        if (!this.firstSelectedSlotId) {
            // This is the first click
            this.firstSelectedSlotId = clickedSlotId;
            event.currentTarget.classList.add('selected');
        } else {
            // This is the second click
            if (this.firstSelectedSlotId === clickedSlotId) {
                // User clicked the same slot twice, so deselect it
                this.resetSelection();
                return;
            }

            // Send the swap request to the server
            const message = {
                type: "swap_items",
                payload: {
                    sourceSlotId: this.firstSelectedSlotId,
                    destSlotId: clickedSlotId,
                }
            };
            this.socket.send(JSON.stringify(message));
            this.resetSelection();
        }
    }

    resetSelection() {
        if (this.firstSelectedSlotId) {
            this.slotElements[this.firstSelectedSlotId]?.classList.remove('selected');
        }
        this.firstSelectedSlotId = null;
    }

    updateReel(payload) {
        const reelText = payload.reel.map(itemInstance => itemInstance.Item.emoji).join('');
        this.reelDisplay.textContent = reelText;
    }

    updateAllSlots(payload) {
        // Update the Gacha Slot based on the full player state
        const gachaSlotElement = this.slotElements['gacha_slot'];
        gachaSlotElement.textContent = payload.gachaSlot.item ? payload.gachaSlot.item.Item.emoji : '-';

        // Update Inventory Slots
        payload.inventorySlots.forEach((slotData) => {
            const slotElement = this.slotElements[slotData.id];
            if (slotElement) {
                slotElement.textContent = slotData.item ? slotData.item.Item.emoji : '-';
            }
        });
    }
}