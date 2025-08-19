// in /static/gachaManager.js

export class GachaManager {
    constructor(socket) {
        if (!socket) {
            throw new Error("GachaManager requires a WebSocket instance.");
        }
        this.socket = socket;
        this.firstSelectedSlotId = null;

        this.rollButton = document.getElementById('gacha-roll-button');
        this.viewportElement = document.getElementById('gacha-viewport');
        this.reelElement = document.getElementById('gacha-reel');
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
                // Start the animation with the data from the server
                this._runRollAnimation(data.payload.reel, data.payload.winnerIndex);
                break;
            case "player_state_update":
                this.updateAllSlots(data.payload);
                break;
            case "error":
                console.error("Gacha Error:", data.payload.message);
                alert(`Gacha Error: ${data.payload.message}`);
                this.isRolling = false; // Re-enable button on error
                this.rollButton.disabled = false;
                break;
        }
    }
    sendRollRequest() {
        if (this.isRolling) return; // Don't allow rolling while animating
        this.isRolling = true;
        this.rollButton.disabled = true; // Disable button during roll

        const message = {
            type: "start_roll",
            payload: {
                severity: 1
            }
        };
        this.socket.send(JSON.stringify(message));
    }


    _runRollAnimation(reelItems, winnerIndex) {
        // 1. Setup the reel
        this.reelElement.innerHTML = ''; // Clear previous items
        this.reelElement.style.transform = 'translateX(0px)'; // Reset position

        // Populate the reel with new emoji items
        reelItems.forEach(itemData => {
            const itemDiv = document.createElement('div');
            itemDiv.className = 'reel-item';
            itemDiv.textContent = itemData.emoji;
            this.reelElement.appendChild(itemDiv);
        });

        const winnerElement = this.reelElement.children[winnerIndex];
        if (!winnerElement) {
            console.error("Winner element not found!");
            this.isRolling = false;
            this.rollButton.disabled = false;
            return;
        }
        
        // 2. Calculate the target position
        const viewportWidth = this.viewportElement.offsetWidth;
        // We want to center the winner element inside the viewport
        const targetX = (viewportWidth / 2) - (winnerElement.offsetLeft + winnerElement.offsetWidth / 2);

        // 3. Animate
        const duration = 4000; // 4 seconds
        const startTime = performance.now();

        const animate = (currentTime) => {
            const elapsedTime = currentTime - startTime;
            if (elapsedTime >= duration) {
                // Animation finished
                this.reelElement.style.transform = `translateX(${targetX}px)`;
                this.isRolling = false;
                this.rollButton.disabled = false;
                return;
            }

            // Easing function (easeOutCubic) - starts fast, ends slow
            const progress = 1 - Math.pow(1 - (elapsedTime / duration), 3);
            const currentX = progress * targetX;
            
            this.reelElement.style.transform = `translateX(${currentX}px)`;

            requestAnimationFrame(animate); // Request the next frame
        };

        requestAnimationFrame(animate); // Start the animation loop
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
        const reelText = payload.reel.map(itemInstance => itemInstance.emoji).join('');
        this.reelDisplay.textContent = reelText;
    }

    updateAllSlots(payload) {
        // Update the Gacha Slot based on the full player state
        const gachaSlotElement = this.slotElements['gacha_slot'];
        gachaSlotElement.textContent = payload.gachaSlot.item ? payload.gachaSlot.item.emoji : '-';

        // Update Inventory Slots
        payload.inventorySlots.forEach((slotData) => {
            const slotElement = this.slotElements[slotData.id];
            if (slotElement) {
                slotElement.textContent = slotData.item ? slotData.item.emoji : '-';
            }
        });
    }
}