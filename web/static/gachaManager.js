// in /static/gachaManager.js
import { GachaAnimator } from './gachaAnimator.js';
export class GachaManager {
    constructor(socket) {
        if (!socket) {
            throw new Error("GachaManager requires a WebSocket instance.");
        }
        this.socket = socket;
        this.firstSelectedSlotId = null;
        this.animator = new GachaAnimator(); // Create an instance of the animator


        this.gachaContainer = document.getElementById('gacha-game-container');
        this.rollButtonVisual = document.getElementById('gacha-roll-button-visual');
        this.displayCircleContainer = document.getElementById('item-display-circle-container');

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

       this.animator.animateRollButtonPress(this.rollButtonVisual, () => {
            // When the button is down, shake the container
            this.animator.animateGachaShake(this.gachaContainer);
            
            // Then, send the request to the server
            const message = {
                type: "start_roll",
                payload: { severity: this.currentSeverity }
            };
            this.socket.send(JSON.stringify(message));
        });
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
                  this.gachaSlotData.item = prizeItem;
                  this.animator.animateRollButtonReturn(this.rollButtonVisual);
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
    this.gachaSlotData = payload.gachaSlot;
    this.inventorySlotsData = payload.inventorySlots;
    let totalLuck = 0;

    // Update Gacha Slot
    const gachaSlotElement = this.slotElements['gacha_slot'];
    // 7. Use empty string instead of '-'
    gachaSlotElement.textContent = this.gachaSlotData.item ? this.gachaSlotData.item.emoji : '';
    this._updateLuckDisplay(gachaSlotElement, this.gachaSlotData.item);

    // Update Inventory Slots
    this.inventorySlotsData.forEach((slotData) => {
        const slotElement = this.slotElements[slotData.id];
        if (slotElement) {
            slotElement.textContent = slotData.item ? slotData.item.emoji : '';
            this._updateLuckDisplay(slotElement, slotData.item);
            if (slotData.item) {
                totalLuck += slotData.item.luckValue; // Add to total luck
            }
        }
    });

    // 5. Update Total Luck Display
    document.getElementById('total-luck-display').textContent = `Total Luck: ${totalLuck}`;

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
            this.animator.animateDisplayCircle(this.displayCircleContainer, false); // Animate out
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
            this.animator.animateDisplayCircle(this.displayCircleContainer, true); // Animate in
            this.itemDisplayEmoji.textContent = selectedItem.emoji;
            let infoText = `Name: ${selectedItem.name}\n`;
            infoText += `Luck: ${selectedItem.luckValue}\n`;
            infoText += `Rarity: ${selectedItem.rarity.name}`;
            document.getElementById('item-display-info').textContent = infoText;
        } else {
            this.animator.animateDisplayCircle(this.displayCircleContainer, false); // Animate out
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

    const currentSection = this.severitySections[sectionIndex];
    const color = window.getComputedStyle(currentSection).backgroundColor;

    // 1. Update the handle's color
    this.severityHandle.style.setProperty('--handle-color', color);

    // 2. --- CORRECTED FILL LOGIC ---
    // The fill should start at the top (top: 0) and have a height
    // that reaches down to the current section.
    const fillHeightPercentage = (sectionIndex + 1) * 20; // 20% per section from the top
    this.gachaBarFill.style.height = `${fillHeightPercentage}%`;
    this.gachaBarFill.style.backgroundColor = color;
}
    
}
