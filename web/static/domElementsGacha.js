// in /static/domElements.js

// This module finds all necessary DOM elements once and exports them.
export const elements = {
    gachaContainer: document.getElementById('gacha-game-container'),
    rollButton: document.getElementById('gacha-roll-button'),
    rollButtonVisual: document.getElementById('gacha-roll-button-visual'),
    
    // Reel & Viewport
    viewport: document.getElementById('gacha-viewport'),
    reel: document.getElementById('gacha-reel'),
    
    // Slots
    slots: {
        'gacha_slot': document.getElementById('gacha_slot'),
        'inventory_slot_0': document.getElementById('inventory_slot_0'),
        'inventory_slot_1': document.getElementById('inventory_slot_1'),
        'inventory_slot_2': document.getElementById('inventory_slot_2'),
    },

    prizeInfo:{
        prizeNameDisplay: document.getElementById('gacha-prize-name'),
       prizeChanceDisplay: document.getElementById('gacha-prize-chance'),
    },
    
    // Item Display
    displayCircleContainer: document.getElementById('item-display-circle-container'),
    displayEmoji: document.getElementById('item-display-emoji'),
    displayInfo: document.getElementById('item-display-info'),
    totalLuckDisplay: document.getElementById('total-luck-display'),

    // Severity Selector
    severity: {
        bar: document.getElementById('severity-selector-bar'),
        handle: document.getElementById('severity-handle-wrapper'),
        grip: document.getElementById('severity-handle-grip'),
        fill: document.getElementById('gacha-bar-fill'),
        sections: document.querySelectorAll('.severity-section'),
    },

    codex:{  
          collectionGrid : document.getElementById('collection-grid'),
    collectionSeveritySpan: document.getElementById('collection-severity')}
};