import { ApiClient } from './apiClient.js';
import { UIManager } from './uiManager.js';
// You would also create and import a WebSocketManager class

class App {
    constructor() {
          this.apiClient = new ApiClient(
            'https://go-confess-sins-api-production.up.railway.app',
            '' // The webApiUrl is relative, so we leave it blank
        );
        this.uiManager = new UIManager();
        
        // --- Form Elements ---
        this.confessForm = document.getElementById('confess-form');
        this.getKeyButton = document.getElementById('get-key-button');
        this.descriptionTextarea = document.getElementById('sin-description');
    }

    // --- Logic Methods ---
    async refreshData() {
        try {
            const sins = await this.apiClient.fetchSins();
            this.uiManager.renderSins(sins);
        } catch (error) {
            console.error('Fetch sins error:', error);
        }
        try {
            const leaderboard = await this.apiClient.fetchLeaderboard();
            this.uiManager.renderLeaderboard(leaderboard);
        } catch (error) {
            console.error('Failed to fetch leaderboard:', error);
        }
    }

        playSound(audioData) {
        const audioContext = new (window.AudioContext || window.webkitAudioContext)();
        audioContext.decodeAudioData(audioData, (buffer) => {
            const source = audioContext.createBufferSource();
            source.buffer = buffer;
            source.connect(audioContext.destination);
            source.start(0);
        });
    }
    
    setupWebSocket() {
        const socketUrl = `wss://${window.location.host}/ws`;
        const socket = new WebSocket(socketUrl);

        socket.onopen = () => console.log("WebSocket connection established.");
        socket.onmessage = (event) => {
            if (typeof event.data === 'string') {
                const data = JSON.parse(event.data);
                if (data.type === 'update') {
                    console.log("Update received from server. Refreshing lists...");
                    this.refreshData();
                }
            } else if (event.data instanceof Blob) {
                console.log("Receiving TTS from the server");
                event.data.arrayBuffer().then(arrayBuffer => this.playSound(arrayBuffer));
            }
        };
        socket.onclose = () => {
            console.log("WebSocket connection closed. Reconnecting in 3 seconds...");
            setTimeout(() => this.setupWebSocket(), 3000);
        };
        socket.onerror = (error) => {
            console.error("WebSocket error:", error);
            socket.close();
        };
    }

    setupEventListeners() {
        this.confessForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const description = this.descriptionTextarea.value;
            const tags = document.getElementById('sin-tags').value;
            const severity = document.getElementById('sin-severity').value;
            
            const payload = { description };
            if (tags) payload.tags = tags.split(',').map(tag => tag.trim());
            if (severity) payload.severity = parseInt(severity, 10);
            
            try {
                await this.apiClient.confessSin(payload);
                this.confessForm.reset();
                this.uiManager.updateCharCounter(500);
                
                // Play bell sound for the user who submitted
                new Audio('/static/submit.wav').play(); 
                
                // Show the animation after a short delay
                setTimeout(() => {
                    this.uiManager.showForgivenAnimation();
                }, 500); // 500ms pause

            } catch (error) {
                alert('Failed to confess sin.');
            }
        });

        this.getKeyButton.addEventListener('click', async () => {
            // ... (this logic is fine, but it can use the uiManager)
            try {
                const data = await this.apiClient.getNewKey();
                this.uiManager.displayNewKey(data.api_key);
            } catch (error) {
                alert('Failed to get new API key.');
            }
        });
        
        this.descriptionTextarea.addEventListener('input', () => {
            const maxLength = 500;
            const remaining = maxLength - this.descriptionTextarea.value.length;
            this.uiManager.updateCharCounter(remaining);
        });
    }

    init() {
        this.setupEventListeners();
        this.refreshData();
        this.setupWebSocket();
    }
}

// Start the application
document.addEventListener('DOMContentLoaded', () => {
    const app = new App();
    app.init();
});