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
        this.itemDisplayEmoji = document.getElementById('item-display-emoji');
        this.itemDisplayLuck = document.getElementById('item-display-luck');
        
          this.severityBar = document.getElementById('severity-selector-bar');
        this.severityHandle = document.getElementById('severity-handle-wrapper');

        this.severityBar = document.getElementById('severity-selector-bar');
        this.severityHandle = document.getElementById('severity-handle-wrapper');
        this.severityHandleOutline = this.severityHandle.querySelector('::before'); // For color
        this.severityHandleGrip = document.getElementById('severity-handle-grip');
        this.gachaBarFill = document.getElementById('gacha-bar-fill');
        this.severitySections = this.severityBar.querySelectorAll('.severity-section');
        this.isDraggingSeverity = false;


    }

    // Call this to set up all event listeners
    init() {
        this.rollButton.addEventListener('click', () => this.sendRollRequest());
         this._setupSeveritySelector();     
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
              // in handleServerMessage
                this._runRollAnimation(data.payload.reel, data.payload.winnerIndex, data.payload.prize);
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
                severity: this.currentSeverity
            }
        };
        this.socket.send(JSON.stringify(message));
    }


    _runRollAnimation(reelItems, winnerIndex, prizeItem) {
        // 1. Setup the reel
        this.reelElement.innerHTML = '';
        this.reelElement.style.transform = 'translateX(0px)';
        this.slotElements['gacha_slot'].textContent = '-'; // Clear gacha slot

        reelItems.forEach(itemData => {
            const itemDiv = document.createElement('div');
            itemDiv.className = 'reel-item';
            itemDiv.textContent = itemData.emoji;
            this.reelElement.appendChild(itemDiv);
        });

        const winnerElement = this.reelElement.children[winnerIndex];
        if (!winnerElement) { /* ... error handling ... */ return; }
        
        // 2. Calculate target position
        const viewportWidth = this.viewportElement.offsetWidth;
        const targetX = (viewportWidth / 2) - (winnerElement.offsetLeft + winnerElement.offsetWidth / 2);

        // 3. Animate
        const duration = 4000;
        const startTime = performance.now();

        const animate = (currentTime) => {
            const elapsedTime = currentTime - startTime;
            if (elapsedTime >= duration) {
                // Animation finished
                this.reelElement.style.transform = `translateX(${targetX}px)`;
                this.isRolling = false;
                this.rollButton.disabled = false;
                
                // --- NEW LOGIC ON COMPLETION ---
                // 1. Hide the prize item in the reel
                winnerElement.style.opacity = '0';
                // 2. Place the prize emoji in the gacha slot button
                this.slotElements['gacha_slot'].textContent = prizeItem.emoji;
                // --- END NEW LOGIC ---
                return;
            }

            const progress = 1 - Math.pow(1 - (elapsedTime / duration), 3);
            const currentX = progress * targetX;
            this.reelElement.style.transform = `translateX(${currentX}px)`;
            requestAnimationFrame(animate);
        };
        requestAnimationFrame(animate);
    }


    handleSlotClick(event) {
        const clickedSlotId = event.currentTarget.id;

        if (!this.firstSelectedSlotId) {
            // This is the first click
            this.firstSelectedSlotId = clickedSlotId;
            event.currentTarget.classList.add('selected');
              this._updateDisplayCircle();
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
            this._updateDisplayCircle();
        }
    }

    resetSelection() {
        if (this.firstSelectedSlotId) {
            this.slotElements[this.firstSelectedSlotId]?.classList.remove('selected');
        }
        this.firstSelectedSlotId = null;
          this._updateDisplayCircle(); 
    }


    updateReel(payload) {
        const reelText = payload.reel.map(itemInstance => itemInstance.emoji).join('');
        this.reelDisplay.textContent = reelText;

          this._updateDisplayCircle();
    }

     updateAllSlots(payload) {
        // We need to store the full slot data to access luck values later
        this.gachaSlotData = payload.gachaSlot;
        this.inventorySlotsData = payload.inventorySlots;

        // Update Gacha Slot
        const gachaSlotElement = this.slotElements['gacha_slot'];
        gachaSlotElement.textContent = this.gachaSlotData.item ? this.gachaSlotData.item.emoji : '-';
        this._updateLuckDisplay(gachaSlotElement, this.gachaSlotData.item);

        // Update Inventory Slots
        this.inventorySlotsData.forEach((slotData) => {
            const slotElement = this.slotElements[slotData.id];
            if (slotElement) {
                slotElement.textContent = slotData.item ? slotData.item.emoji : '-';
                this._updateLuckDisplay(slotElement, slotData.item);
            }
        });
        this._updateDisplayCircle(); 
    }

     _updateLuckDisplay(slotElement, item) {
        const wrapper = slotElement.parentElement;
        const luckDisplay = wrapper.querySelector('.luck-display');

        if (!luckDisplay) {
            return;
    }
        if (item) {
            luckDisplay.textContent = item.luckValue; 
            luckDisplay.style.display = 'flex';
        } else {
            luckDisplay.style.display = 'none';
        }
    }

    _updateDisplayCircle() {
        if (!this.firstSelectedSlotId) {
            this.itemDisplayEmoji.textContent = '-';
            this.itemDisplayLuck.textContent = 'Luck: ?';
            return;
        }

        let selectedItem = null;
        if (this.firstSelectedSlotId === 'gacha_slot') {
            selectedItem = this.gachaSlotData?.item;
        } else {
            const slotData = this.inventorySlotsData.find(s => s.id === this.firstSelectedSlotId);
            selectedItem = slotData?.item;
        }

        if (selectedItem) {
            this.itemDisplayEmoji.textContent = selectedItem.emoji;
            this.itemDisplayLuck.textContent = `Luck: ${selectedItem.luckValue}`;
        } else {
            this.itemDisplayEmoji.textContent = '-';
            this.itemDisplayLuck.textContent = 'Luck: ?';
        }
    }

      _setupSeveritySelector() {
        const handleDragStart = (e) => {
            this.isDraggingSeverity = true;
            e.preventDefault();
        };

        const handleDragMove = (e) => {
            if (!this.isDraggingSeverity) return;

            const barRect = this.severityBar.getBoundingClientRect();
            const clientY = e.touches ? e.touches[0].clientY : e.clientY;
            const relativeY = clientY - barRect.top;

            // Move the handle visually without transition
            this.severityHandle.style.transition = 'none';
            const handleTop = relativeY - (this.severityHandle.offsetHeight / 2);
            this.severityHandle.style.top = `${handleTop}px`;
            
            // Update visuals in real-time while dragging
            this._updateSeverityVisuals(handleTop);
        };

        const handleDragEnd = () => {
            if (!this.isDraggingSeverity) return;
            this.isDraggingSeverity = false;

            const barRect = this.severityBar.getBoundingClientRect();
            const sectionHeight = barRect.height / 5;
            const handleTop = this.severityHandle.offsetTop;
            const middleOfHandle = handleTop + (this.severityHandle.offsetHeight / 2);
            const sectionIndex = Math.floor(middleOfHandle / sectionHeight);
            const clampedIndex = Math.max(0, Math.min(4, sectionIndex));
            
            this.currentSeverity = 5 - clampedIndex;
            console.log(`Severity set to: ${this.currentSeverity}`);

            // Snap the handle and visuals to the final position
            const finalTop = clampedIndex * sectionHeight;
            this.severityHandle.style.transition = 'top 0.1s ease-out';
            this.severityHandle.style.top = `${finalTop}px`;
            this._updateSeverityVisuals(finalTop);
        };

        // Mouse and Touch events
        this.severityHandle.addEventListener('mousedown', handleDragStart);
        window.addEventListener('mousemove', handleDragMove);
        window.addEventListener('mouseup', handleDragEnd);
        this.severityHandle.addEventListener('touchstart', handleDragStart);
        window.addEventListener('touchmove', handleDragMove);
        window.addEventListener('touchend', handleDragEnd);

        // Set initial position and visuals
        const initialTop = 4 * (this.severityBar.offsetHeight / 5);
        this.severityHandle.style.top = `${initialTop}px`;
        this._updateSeverityVisuals(initialTop);
    }


      _updateSeverityVisuals(handleTop) {
        const barRect = this.severityBar.getBoundingClientRect();
        const sectionHeight = barRect.height / 5;
        const middleOfHandle = handleTop + (this.severityHandle.offsetHeight / 2);
        let sectionIndex = Math.floor(middleOfHandle / sectionHeight);
        sectionIndex = Math.max(0, Math.min(4, sectionIndex));

        // Get the color from the current severity section
        const currentSection = this.severitySections[sectionIndex];
        const color = window.getComputedStyle(currentSection).backgroundColor;

        // 1. Update the handle's color
        // Note: We can't directly style ::before, so we use a CSS variable
        this.severityHandle.style.setProperty('--handle-color', color);
        this.severityHandleGrip.style.backgroundColor = color;

        // 2. Update the fill element's color and position
        this.gachaBarFill.style.backgroundColor = color;
        this.gachaBarFill.style.top = `${sectionIndex * sectionHeight}px`;

        const fillHeightPercentage = (5 - sectionIndex) * 20; // 20% height per section
        this.gachaBarFill.style.height = `${fillHeightPercentage}%`;
    }
}
