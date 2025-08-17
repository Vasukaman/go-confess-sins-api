document.addEventListener('DOMContentLoaded', () => {
    // --- CONFIGURATION ---
    const sinApiUrl = 'https://go-confess-sins-api-production.up.railway.app'; // Your live API URL
    const websiteApiKey = 'YOUR_WEBSITE_API_KEY'; // Your hardcoded key for the website

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
            
            sinsList.innerHTML = ''; // Clear the list before adding new items
            sins.forEach(sin => {
                const sinCard = document.createElement('div');
                sinCard.className = 'sin-card';
                sinCard.innerHTML = `<p class="description">"${sin.description}"</p>`;
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
        
        // Play a sound on button click
        new Audio('/static/submit.wav').play(); // Make sure you have this sound file

        const description = document.getElementById('sin-description').value;
        const apiKey = apiKeyInput.value;

        try {
            const response = await fetch(`${sinApiUrl}/sins`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${apiKey}`,
                },
                body: JSON.stringify({ description: description }),
            });

            if (!response.ok) throw new Error('API returned an error.');

            confessForm.reset(); // Clear the form
            fetchSins(); // Refresh the list of sins
        } catch (error) {
            alert('Failed to confess sin.');
            console.error('Confess error:', error);
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

    // --- EVENT LISTENERS ---
    confessForm.addEventListener('submit', handleConfess);
    getKeyButton.addEventListener('click', getNewKey);

    // --- INITIAL LOAD ---
    fetchSins(); // Fetch sins when the page first loads
});