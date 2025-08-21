// in /static/gachaManager.js
import { GachaAnimator } from './gachaAnimator.js';
import { elements } from './domElementsGacha.js';
import { UIUpdater } from './uiUpdaterGacha.js';

export class GachaManager {
    constructor(socket) {
        this.socket = socket;
        this.animator = new GachaAnimator();
        this.ui = new UIUpdater(elements); // Give the UI updater the elements

        // Game State
        this.isRolling = false;
        this.firstSelectedSlotId = null;
        this.currentSeverity = 1;
        this.playerState = { gachaSlot: {}, inventorySlots: [] };
    }

    init() {
        this._setupInputHandlers();
        this.ui.updateSeverityVisuals(elements.severity.handle.offsetTop); // Set initial visuals
       
    }

    handleServerMessage(data) {
        
        console.log("GachaManager received:", data);
        switch (data.type) {
            case "roll_result":
                this._handleRollResult(data.payload);
                break;
            case "player_state_update":
                this.playerState = data.payload;
                this.ui.updateAllSlots(this.playerState);
                this._updateDisplayCircle();
                this.requestCollectionData(); 
                break;

            case "droptable_info_update":
            this.renderCollection(data.payload);
            break;
            case "error":
                this._handleError(data.payload.message);
                break;
        }
    }

    // --- Private Logic Methods ---

    _setupInputHandlers() {
        elements.rollButton.addEventListener('click', () => this._requestRoll());
        
        for (const slotId in elements.slots) {
            elements.slots[slotId].addEventListener('click', (event) => this._handleSlotClick(event));
        }
        
        // Severity Selector Drag Logic
     let isDragging = false;
        const handleDragStart = (e) => { isDragging = true; e.preventDefault(); };
        
        const handleDragMove = (e) => {
            if (!isDragging) return;
            const barRect = elements.severity.bar.getBoundingClientRect();
            const clientY = e.touches ? e.touches[0].clientY : e.clientY;
            const relativeY = clientY - barRect.top;
            const handleTop = relativeY - (elements.severity.handle.offsetHeight / 2);
            elements.severity.handle.style.transition = 'none';
            elements.severity.handle.style.top = `${handleTop}px`;
            this.ui.updateSeverityVisuals(handleTop);

            const sectionHeight = barRect.height / 5;
            const sectionIndex = Math.max(0, Math.min(4, Math.floor((handleTop + elements.severity.handle.offsetHeight / 2) / sectionHeight)));
            const hoverSeverity = 5 - sectionIndex;
            
            if (hoverSeverity !== this.lastRequestedSeverity) {
                this.lastRequestedSeverity = hoverSeverity;
                this.requestCollectionData(); // Request new data as we drag
            }
        };
        const handleDragEnd = () => {
            if (!isDragging) return;
            isDragging = false;
            const barRect = elements.severity.bar.getBoundingClientRect();
            const sectionHeight = barRect.height / 5;
            const handleTop = elements.severity.handle.offsetTop;
            const middleOfHandle = handleTop + (elements.severity.handle.offsetHeight / 2);
            const sectionIndex = Math.max(0, Math.min(4, Math.floor(middleOfHandle / sectionHeight)));
            
            this.currentSeverity = 5 - sectionIndex;
            const finalTop = sectionIndex * sectionHeight;
            elements.severity.handle.style.transition = 'top 0.1s ease-out';
            elements.severity.handle.style.top = `${finalTop}px`;
            this.ui.updateSeverityVisuals(finalTop);
                this.requestCollectionData(); 
        };
        
        elements.severity.handle.addEventListener('mousedown', handleDragStart);
        window.addEventListener('mousemove', handleDragMove);
        window.addEventListener('mouseup', handleDragEnd);
        // Add touch events for mobile
    }

    _requestRoll() {
        if (this.isRolling) return;
        this.isRolling = true;
        elements.rollButton.disabled = true;
         this._resetSelection();

        this.ui.updateSlot(elements.slots['gacha_slot'], null);
         elements.prizeInfo.prizeNameDisplay.classList.remove('visible');
        elements.prizeInfo.prizeChanceDisplay.classList.remove('visible');
        elements.prizeInfo.prizeRarityDisplay.classList.remove('visible');
        this.animator.animateRollButtonPress(elements.rollButtonVisual, () => {
            this.animator.animateGachaShake(elements.gachaContainer);
            this.socket.send(JSON.stringify({
                type: "start_roll",
                payload: { severity: this.currentSeverity }
            }));
        });
    }

    _handleRollResult(payload) {
        const { reel, winnerIndex, prize, prizeDropChance} = payload;
        
        // Setup reel for animation
        elements.reel.innerHTML = '';
        reel.forEach(itemData => {
            const itemDiv = document.createElement('div');
            itemDiv.className = 'reel-item';
            itemDiv.textContent = itemData.emoji;
            elements.reel.appendChild(itemDiv);
        });
        
        const winnerElement = elements.reel.children[winnerIndex];
        const targetX = (elements.viewport.offsetWidth / 2) - (winnerElement.offsetLeft + winnerElement.offsetWidth / 2);

        // Animate using requestAnimationFrame (or Anime.js)
        const duration = 4000;
        const startTime = performance.now();
        const animate = (currentTime) => {
            const elapsedTime = currentTime - startTime;
            if (elapsedTime >= duration) {
                elements.reel.style.transform = `translateX(${targetX}px)`;
                this.isRolling = false;
                elements.rollButton.disabled = false;
                winnerElement.style.opacity = '0';
                
                // Update state and UI
                this.playerState.gachaSlot.item = prize;
                this.ui.updateSlot(elements.slots['gacha_slot'], prize);
                elements.prizeInfo.prizeNameDisplay.textContent = prize.name;
                elements.prizeInfo.prizeChanceDisplay.textContent = `(${(prizeDropChance * 100).toFixed(2)}%)`;
                
                elements.prizeInfo.prizeRarityDisplay.textContent = prize.rarity.name;
                elements.prizeInfo.prizeRarityDisplay.style.color = prize.rarity.color;
                
                elements.prizeInfo.prizeRarityDisplay.classList.add('visible');
                elements.prizeInfo.prizeNameDisplay.classList.add('visible');
                elements.prizeInfo.prizeChanceDisplay.classList.add('visible');

                this.requestCollectionData(); 
                
                //Deselect everything and select new item
                 this._resetSelection();
                elements.slots['gacha_slot'].click(); 

                this.animator.animateRollButtonReturn(elements.rollButtonVisual);
                return;
            }
            const progress = 1 - Math.pow(1 - (elapsedTime / duration), 3);
            elements.reel.style.transform = `translateX(${progress * targetX}px)`;
            requestAnimationFrame(animate);
        };
        requestAnimationFrame(animate);
    }

    _handleSlotClick(event) {
        if (this.isRolling) return;
        const clickedSlot = event.currentTarget;
        const clickedSlotId = clickedSlot.id;

        if (!this.firstSelectedSlotId) {
            this.firstSelectedSlotId = clickedSlotId;
            clickedSlot.classList.add('selected');
            this.animator.animateSlotSelect(clickedSlot);
        } else {
            const sourceSlotId = this.firstSelectedSlotId;
            this._resetSelection();

            if (sourceSlotId !== clickedSlotId) {

                   const sourceElement = elements.slots[sourceSlotId];
                   const destElement = elements.slots[clickedSlotId];

      
                   this.animator.animateSwap(sourceElement, destElement, () => {
                    this.socket.send(JSON.stringify({
                        type: "swap_items",
                        payload: { sourceSlotId, destSlotId: clickedSlotId }
                    }));
                });
            }
        }
        this._updateDisplayCircle();
    }

    _resetSelection() {
        if (this.firstSelectedSlotId) {
            const selectedSlot = elements.slots[this.firstSelectedSlotId];
            if (selectedSlot) {
                selectedSlot.classList.remove('selected');
                this.animator.animateSlotDeselect(selectedSlot);
            }
        }
        this.firstSelectedSlotId = null;
        this._updateDisplayCircle()
    }

    _updateDisplayCircle() {
        let selectedItem = null;
        if (this.firstSelectedSlotId) {
            const slotData = this.firstSelectedSlotId === 'gacha_slot'
                ? this.playerState.gachaSlot
                : this.playerState.inventorySlots.find(s => s.id === this.firstSelectedSlotId);
            selectedItem = slotData?.item;
        }
        
        this.ui.updateDisplayCircle(selectedItem);
        this.animator.animateDisplayCircle(elements.displayCircleContainer, !!selectedItem);
    }
    
    _handleError(message) {
        console.error("Gacha Error:", message);
        alert(`Gacha Error: ${message}`);
        this.isRolling = false;
        elements.rollButton.disabled = false;
    }

    requestCollectionData() {
    if (this.isRolling) return; // Don't request while a roll is happening
    
    elements.codex.collectionSeveritySpan.textContent = this.currentSeverity;
    const message = {
        type: "get_droptable_info",
        payload: { severity: this.currentSeverity }
    };
    console.log("Sending to server:", JSON.stringify(message)); 
    this.socket.send(JSON.stringify(message));
}

// 2. This method renders the data received from the server
renderCollection(items) {
    elements.codex.collectionGrid.innerHTML = ''; // Clear out old items

    items.forEach(itemData => {
        const itemDiv = document.createElement('div');
        itemDiv.className = 'collection-item';
        if (itemData.TimesObtained > 0) {
            itemDiv.classList.add('discovered');
        }

        const emojiDiv = document.createElement('div');
        emojiDiv.className = 'collection-item-emoji';
        emojiDiv.textContent = itemData.Item.emoji;

        const chanceSpan = document.createElement('span');
        chanceSpan.className = 'collection-item-chance';
        
        // --- NEW PROBABILITY LOGIC ---
        const chance = itemData.DropChance * 100;
        if (chance > 0 && chance < 0.01) {
            // If the chance is tiny but not zero, show a minimum value
            chanceSpan.textContent = `<0.01%`;
        } else {
            chanceSpan.textContent = `${chance.toFixed(2)}%`;
        }
        
        // --- NEW COUNTER LOGIC ---
        if (itemData.TimesObtained > 0) {
            const counterDiv = document.createElement('div');
            counterDiv.className = 'collection-item-counter';
            counterDiv.textContent = itemData.TimesObtained;
            itemDiv.appendChild(counterDiv); // Add the counter to the item
        }

        itemDiv.appendChild(emojiDiv);
        itemDiv.appendChild(chanceSpan);
        elements.codex.collectionGrid.appendChild(itemDiv);
    });
}
}