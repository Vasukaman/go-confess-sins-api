document.addEventListener('DOMContentLoaded', () => {
    // --- CONFIGURATION ---
    const sinApiUrl = 'https://go-confess-sins-api-production.up.railway.app'; // Your live API URL
    const leaderboardApiUrl = '/api/leaderboard';

    // --- DOM ELEMENTS ---
    const sinsList = document.getElementById('sins-list');
    const confessForm = document.getElementById('confess-form');
    const getKeyButton = document.getElementById('get-key-button');
    const newKeyDisplay = document.getElementById('new-key-display');
    const apiKeyInput = document.getElementById('api-key-input');
    const leaderboardList = document.getElementById('leaderboard-list');

     const forgivenOverlay = document.getElementById('forgiven-overlay');
     const descriptionTextarea = document.getElementById('sin-description');
    const charCounter = document.getElementById('char-counter');
    const maxLength = 500;

    // --- FUNCTIONS ---

     descriptionTextarea.addEventListener('input', () => {
        const remaining = maxLength - descriptionTextarea.value.length;
        charCounter.textContent = remaining;
    });

    const showForgivenAnimation = () => {
       
           const elementsToShake = document.querySelectorAll('.sin-card, .leaderboard, header, api-key-section');
    
        setTimeout(() => {

    elementsToShake.forEach(el => el.classList.add('shaking'));
        forgivenOverlay.classList.add('visible');
        setTimeout(() => {
            forgivenOverlay.classList.remove('visible');
             elementsToShake.forEach(el => el.classList.remove('shaking'));
        }, 9000)}
        
        , 3000);
        
    };


   

    const playConfessionAudio = (text) => {
        const bellSound = new Audio('/static/submit.wav'); // Make sure you have this file
        bellSound.play();

        // When the bell sound finishes, play the TTS audio
     
            const encodedText = encodeURIComponent(text);
            // This URL points to our new proxy endpoint
        //    const speechUrl = `/api/speech?text=${encodedText}`;
            
        //    const speechAudio = new Audio(speechUrl);
         //   speechAudio.play();
        
    };

    // Function to fetch and display sins
    const fetchSins = async () => {
    try {
        const response = await fetch(`${sinApiUrl}/sins`);
        const sins = await response.json();
        
        sinsList.innerHTML = ''; // Clear the list
        sins.forEach(sin => {
            const sinCard = document.createElement('div');
            sinCard.className = 'sin-card';

            // Start building the inner HTML for the meta section
            let metaHTML = `<span class="count">Confessed: ${sin.count} times</span>`;

            // Conditionally add severity if it exists
            if (sin.severity != null) {
                metaHTML = `<span class="severity">Severity: ${sin.severity}</span>` + metaHTML;
            }

            // Conditionally add tags if they exist and are not empty
            if (sin.tags && sin.tags.length > 0) {
                const tagsHTML = sin.tags.map(tag => `<span class="tag">${tag}</span>`).join('');
                metaHTML = `<div class="tags-container">${tagsHTML}</div>` + metaHTML;
            }
            
            // Assemble the final card
            sinCard.innerHTML = `
                <p class="description">"${sin.description}"</p>
                <div class="meta">${metaHTML}</div>
            `;
            sinsList.appendChild(sinCard);
        });
    } catch (error) {
        sinsList.innerHTML = '<p>Could not load sins. Is the API running?</p>';
        console.error('Fetch sins error:', error);
    }
};
    // Function to handle confessing a sin
    const handleConfess = async (event) => {
        event.preventDefault();
    
        const description = document.getElementById('sin-description').value;
        const tags = document.getElementById('sin-tags').value;
        const severity = document.getElementById('sin-severity').value;
        const payload = { description };
        if (tags) {
            payload.tags = tags.split(',').map(tag => tag.trim());
        }
        if (severity) {
            payload.severity = parseInt(severity, 10);
        }
        
        const apiUrl = '/api/confess'; 

        try {
            const response = await fetch(apiUrl, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            if (!response.ok) throw new Error('API returned an error.');
            confessForm.reset(); // Clear the form
      
          showForgivenAnimation();
          playConfessionAudio(description);
            


            fetchSins(); // Refresh the list
        } catch (error) {
            alert('Failed to confess sin.');
        }
    };
    
    // Function to get a new API key
    const getNewKey = async () => {
        try {
            const response = await fetch(`${sinApiUrl}/keys`, { method: 'POST' });
            const data = await response.json();
            
            newKeyDisplay.innerHTML = `Your new key: <code>${data.api_key}</code>`;
        } catch (error) {
            alert('Failed to get new API key.');
            console.error('Get key error:', error);
        }
    };


    // Function to get leaderboard data
  const fetchLeaderboard = async () => {
    try {
        const response = await fetch(leaderboardApiUrl);
        console.log("Leaderboard data received json:", response);
        const topSins = await response.json();

        // THE DEBUGGING LINE IS HERE:
        console.log("Leaderboard data received:", topSins);

        leaderboardList.innerHTML = '';
        topSins.forEach(sin => {
            const item = document.createElement('div');
            item.className = 'leaderboard-item';
            item.innerHTML = `${sin.description} ${sin.count}`;
            leaderboardList.appendChild(item);
        });
    } catch (error) {
        console.error('Failed to fetch leaderboard:', error);
    }
};

    // Function to setup websocket
    const setupWebSocket = () => {
        // Construct the secure WebSocket URL from the page's location
        const socketUrl = `wss://${window.location.host}/ws`;
        const socket = new WebSocket(socketUrl);

        socket.onopen = () => {
            console.log("WebSocket connection established.");
        };

        // This function runs when a message is pushed from the server
        socket.onmessage = (event) => {

            console.log("Recieved socket message. "+event);

        if (typeof event.data === 'string') {
                const data = JSON.parse(event.data);
            if (data.type === 'update') {
                console.log("Update received from server. Refreshing lists...");
                // When an update is received, just re-run the fetch functions
                fetchSins();
                fetchLeaderboard();
            }
        }
            else if (event.data instanceof Blob) {
                   console.log("Receiving TTS from the server");
            
                // The playSound function needs an ArrayBuffer, so we convert the Blob.
                // The arrayBuffer() method is asynchronous and returns a Promise.
                event.data.arrayBuffer().then(arrayBuffer => {
                playSound(arrayBuffer);
        });
    }
        };

        socket.onclose = () => {
            console.log("WebSocket connection closed. Reconnecting in 3 seconds...");
            setTimeout(setupWebSocket, 3000); // Simple reconnect logic
        };

        socket.onerror = (error) => {
            console.error("WebSocket error:", error);
            socket.close();
        };
    };

     const playSound = (audioData) => {
        // Create an audio context
        const audioContext = new (window.AudioContext || window.webkitAudioContext)();
        // Decode the raw array buffer into playable audio
        audioContext.decodeAudioData(audioData, (buffer) => {
            const source = audioContext.createBufferSource();
            source.buffer = buffer;
            source.connect(audioContext.destination);
            source.start(0);
        });
    };
    // --- EVENT LISTENERS ---
    confessForm.addEventListener('submit', handleConfess);
    getKeyButton.addEventListener('click', getNewKey);

    // --- INITIAL LOAD ---
    fetchSins(); // Fetch sins when the page first loads
    fetchLeaderboard();
    setupWebSocket();
});