document.addEventListener('DOMContentLoaded', () => {
    // --- CONFIGURATION ---
    // This assumes your web-frontend server is running and proxying requests
    const searchApiUrl = '/api/search'; 

    // --- DOM ELEMENTS ---
    const searchForm = document.getElementById('search-form');
    const resultsContainer = document.getElementById('search-results');
    const paginationContainer = document.getElementById('pagination-controls');

    // --- STATE ---
    let currentPage = 1;
    let currentParams = new URLSearchParams(); // Stores the current search filters

    // --- FUNCTIONS ---

    // The main function to perform a search
    const performSearch = async () => {
        currentParams.set('page', currentPage);
        currentParams.set('limit', 25); // 25 results per page

        try {
            const response = await fetch(`${searchApiUrl}?${currentParams.toString()}`);
            if (!response.ok) throw new Error('Search request failed');
            
            const sins = await response.json();
            renderResults(sins);
            renderPagination(sins.length);
        } catch (error) {
            resultsContainer.innerHTML = `<p>Error: ${error.message}</p>`;
            console.error('Search error:', error);
        }
    };

    // Renders the list of sins
    const renderResults = (sins) => {
        resultsContainer.innerHTML = '';
        if (!sins || sins.length === 0) {
            resultsContainer.innerHTML = '<p>No sins found for this query.</p>';
            return;
        }
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
            resultsContainer.appendChild(sinCard);
        });
    };

    // Renders the "Previous" and "Next" buttons
    const renderPagination = (resultsCount) => {
        paginationContainer.innerHTML = '';

        if (currentPage > 1) {
            const prevButton = document.createElement('button');
            prevButton.textContent = 'Previous Page';
            prevButton.onclick = () => {
                currentPage--;
                performSearch();
            };
            paginationContainer.appendChild(prevButton);
        }

        if (resultsCount === 25) { // If we got a full page, there might be a next one
            const nextButton = document.createElement('button');
            nextButton.textContent = 'Next Page';
            nextButton.onclick = () => {
                currentPage++;
                performSearch();
            };
            paginationContainer.appendChild(nextButton);
        }
    };

    // --- EVENT LISTENERS ---
    searchForm.addEventListener('submit', (event) => {
        event.preventDefault();
        currentPage = 1; // Reset to page 1 for a new search

        // Build the query parameters from the form
        const tags = document.getElementById('tags-input').value;
        const sortBy = document.getElementById('sort-by-select').value;
        const order = document.getElementById('order-select').value;

        currentParams = new URLSearchParams();
        if (tags) currentParams.set('tags', tags);
        if (sortBy) currentParams.set('sortBy', sortBy);
        if (order) currentParams.set('order', order);
        
        performSearch();
    });

    // --- INITIAL LOAD ---
    performSearch(); // Perform an initial search when the page loads
});