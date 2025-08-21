export class UIManager {
    constructor() {
        this.sinsList = document.getElementById('sins-list');
        this.leaderboardList = document.getElementById('leaderboard-list');
        this.newKeyDisplay = document.getElementById('new-key-display');
        this.apiKeyInput = document.getElementById('api-key-input');
        this.forgivenOverlay = document.getElementById('forgiven-overlay');
        this.charCounter = document.getElementById('char-counter');
        this.emojiPicker = document.getElementById('emoji-picker');
        this.selectedEmoji = null; 
    }

    // --- Render Methods ---
   renderSins(sins, container) {
            if (!container) return;
                container.innerHTML = ''; 
        sins.forEach(sin => {
            const sinCard = document.createElement('div');
            sinCard.className = 'sin-card';
            
            let metaHTML = `<span class="count">Confessed: ${sin.count} times</span>`;
            if (sin.Severity != null) {
                metaHTML = `<span class="severity">Severity: ${sin.severity}</span>` + metaHTML;
            }
            if (sin.Tags && sin.Tags.length > 0) {
                const tagsHTML = sin.Tags.map(tag => `<span class="tag">${tag}</span>`).join('');
                metaHTML = `<div class="tags-container">${tagsHTML}</div>` + metaHTML;
            }

             let emojiHTML = '';
            if (sin.emoji) {
                emojiHTML = `<div class="sin-emoji">${sin.emoji}</div>`;
            }
            

              sinCard.innerHTML = `
                ${emojiHTML}
                <div class="sin-content">
                    <p class="description">"${sin.description}"</p>
                    <div class="meta">${metaHTML}</div>
                </div>
            `;
            container.appendChild(sinCard);
        });
    }

    renderLeaderboard(topSins) {
        this.leaderboardList.innerHTML = '';
        topSins.forEach(sin => {
            const item = document.createElement('div');
            item.className = 'leaderboard-item';

            
     let emojiSpan = '';
        if (sin.emoji) {
            emojiSpan = `<span class="leaderboard-emoji">${sin.emoji}</span>`;
        }



        item.innerHTML = `
            ${emojiSpan}
            <span class="leaderboard-text">${sin.description}</span>
            <span class="leaderboard-count">${sin.count}</span>
        `;

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


     renderEmojiPicker(emojis) {
        this.emojiPicker.innerHTML = '';
        emojis.forEach(emoji => {
            const emojiEl = document.createElement('div');
            emojiEl.className = 'emoji';
            emojiEl.textContent = emoji;
            emojiEl.dataset.emoji = emoji;

            emojiEl.addEventListener('click', () => {
                // Deselect any currently selected emoji
                this.emojiPicker.querySelector('.emoji.selected')?.classList.remove('selected');
                // Select the new one
                emojiEl.classList.add('selected');
                this.selectedEmoji = emoji;
            });

            this.emojiPicker.appendChild(emojiEl);
        });
    }
}