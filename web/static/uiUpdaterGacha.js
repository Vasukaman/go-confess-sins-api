// in /static/uiUpdater.js

export class UIUpdater {
    constructor(elements) {
        this.elements = elements;
    }

    updateAllSlots(playerState) {
        let totalLuck = 0;
        const { gachaSlot, inventorySlots } = playerState;

        // Update Gacha Slot
        this.updateSlot(this.elements.slots['gacha_slot'], gachaSlot.item);

        // Update Inventory Slots
        inventorySlots.forEach(slotData => {
            const slotElement = this.elements.slots[slotData.id];
            if (slotElement) {
                this.updateSlot(slotElement, slotData.item);
                if (slotData.item) {
                    totalLuck += slotData.item.luckValue;
                }
            }
        });

        this.elements.totalLuckDisplay.textContent = `Total Luck: ${totalLuck}`;
    }

    updateSlot(slotElement, item) {
        slotElement.textContent = item ? item.emoji : '';
        slotElement.style.setProperty('--rarity-color', item ? item.rarity.color : '#555');
        this.updateLuckDisplay(slotElement, item);
    }

    updateLuckDisplay(slotElement, item) {
        const wrapper = slotElement.parentElement;
        const luckDisplay = wrapper.querySelector('.luck-display');
        if (!luckDisplay) return;

        if (item) {
            luckDisplay.textContent = item.luckValue;
            luckDisplay.style.display = 'flex';
        } else {
            luckDisplay.style.display = 'none';
        }
    }

    updateDisplayCircle(selectedItem) {
        if (selectedItem) {
            this.elements.displayEmoji.textContent = selectedItem.emoji;
            let infoText = `Name: ${selectedItem.name}\n`;
            infoText += `Luck: ${selectedItem.luckValue}\n`;
            infoText += `Rarity: ${selectedItem.rarity.name}`;
            this.elements.displayInfo.textContent = infoText;
        } else {
            this.elements.displayEmoji.textContent = '';
            this.elements.displayInfo.textContent = '';
        }
    }

    updateSeverityVisuals(handleTop) {
        const barRect = this.elements.severity.bar.getBoundingClientRect();
        const sectionHeight = barRect.height / 5;
        const middleOfHandle = handleTop + (this.elements.severity.handle.offsetHeight / 2);
        let sectionIndex = Math.floor(middleOfHandle / sectionHeight);
        sectionIndex = Math.max(0, Math.min(4, sectionIndex));

        const currentSection = this.elements.severity.sections[sectionIndex];
        const color = window.getComputedStyle(currentSection).backgroundColor;

        this.elements.severity.handle.style.setProperty('--handle-color', color);
        const fillHeightPercentage = (sectionIndex + 1) * 20;
        this.elements.severity.fill.style.height = `${fillHeightPercentage}%`;
        this.elements.severity.fill.style.backgroundColor = color;
    }
}