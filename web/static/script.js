document.addEventListener('DOMContentLoaded', () => {
    // --- CONFIGURATION ---
    const sinApiUrl = 'https://go-confess-sins-api-production.up.railway.app'; // Your live API URL

    // --- DOM ELEMENTS ---
    const sinsList = document.getElementById('sins-list');
    const confessForm = document.getElementById('confess-form');
    const getKeyButton = document.getElementById('get-key-button');
    const newKeyDisplay = document.getElementById('new-key-display');
    const apiKeyInput = document.getElementById('api-key-input');

    // --- FUNCTIONS ---

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
              new Audio('/static/submit.wav').play(); 
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
            apiKeyInput.value = data.api_key; // Auto-fill the key into the confess form
        } catch (error) {
            alert('Failed to get new API key.');
            console.error('Get key error:', error);
        }
    };

    const leaderboardList = document.getElementById('leaderboard-list');
    const leaderboardApiUrl = 'https://appealing-reverence.railway.internal/leaderboard';
    // Function to get leaderboard data
  const fetchLeaderboard = async () => {
    try {
        const response = await fetch(leaderboardApiUrl);
        console.log("Leaderboard data receivedjson:", response);
        const topSins = await response.json();

        // THE DEBUGGING LINE IS HERE:
        console.log("Leaderboard data received:", topSins);

        leaderboardList.innerHTML = '';
        topSins.forEach(sin => {
            const item = document.createElement('div');
            item.className = 'leaderboard-item';
            item.innerHTML = `<span>${sin.Description}</span><span>${sin.Count}</span>`;
            leaderboardList.appendChild(item);
        });
    } catch (error) {
        console.error('Failed to fetch leaderboard:', error);
    }
};

    // --- EVENT LISTENERS ---
    confessForm.addEventListener('submit', handleConfess);
    getKeyButton.addEventListener('click', getNewKey);

    // --- INITIAL LOAD ---
    fetchSins(); // Fetch sins when the page first loads
    fetchLeaderboard();
});