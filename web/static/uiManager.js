class UIManager {
    constructor() {
        this.sinsList = document.getElementById('sins-list');
        this.leaderboardList = document.getElementById('leaderboard-list');
        this.newKeyDisplay = document.getElementById('new-key-display');
        this.apiKeyInput = document.getElementById('api-key-input');
        this.forgivenOverlay = document.getElementById('forgiven-overlay');
        this.charCounter = document.getElementById('char-counter');
    }

    // --- Render Methods ---
    renderSins(sins) {
        this.sinsList.innerHTML = ''; // Clear the list
        sins.forEach(sin => {
            const sinCard = document.createElement('div');
            sinCard.className = 'sin-card';
            
            let metaHTML = `<span class="count">Confessed: ${sin.Count} times</span>`;
            if (sin.Severity != null) {
                metaHTML = `<span class="severity">Severity: ${sin.Severity}</span>` + metaHTML;
            }
            if (sin.Tags && sin.Tags.length > 0) {
                const tagsHTML = sin.Tags.map(tag => `<span class="tag">${tag}</span>`).join('');
                metaHTML = `<div class="tags-container">${tagsHTML}</div>` + metaHTML;
            }

            sinCard.innerHTML = `<p class="description">"${sin.Description}"</p><div class="meta">${metaHTML}</div>`;
            this.sinsList.appendChild(sinCard);
        });
    }

    renderLeaderboard(topSins) {
        this.leaderboardList.innerHTML = '';
        topSins.forEach(sin => {
            const item = document.createElement('div');
            item.className = 'leaderboard-item';
            item.innerHTML = `<span>${sin.Description}</span><span>${sin.Count}</span>`;
            this.leaderboardList.appendChild(item);
        });
    }

    // --- UI Update Methods ---
    displayNewKey(apiKey) {
        this.newKeyDisplay.innerHTML = `Your new key: <code>${apiKey}</code>`;
        this.apiKeyInput.value = apiKey;
    }

    updateCharCounter(remaining) {
        this.charCounter.textContent = remaining;
    }
    
    // --- Animation Method ---
    showForgivenAnimation() {
        const elementsToShake = document.querySelectorAll('.sin-card, .leaderboard, header, .api-key-section, .confess-form');
        
        // This function now just handles the visual effect
        elementsToShake.forEach(el => el.classList.add('shaking'));
        this.forgivenOverlay.classList.add('visible');

        setTimeout(() => {
            this.forgivenOverlay.classList.remove('visible');
            elementsToShake.forEach(el => el.classList.remove('shaking'));
        }, 5000); // Animation duration is 5s
    }
}