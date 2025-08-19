import { ApiClient } from './apiClient.js';
import { UIManager } from './uiManager.js';

// --- The SearchApp Class ---
class SearchApp {
    constructor() {
        // 2. Initialize your components.
        this.apiClient = new ApiClient('https://go-confess-sins-api-production.up.railway.app', '');
        this.uiManager = new UIManager();
        
        // --- DOM ELEMENTS ---
        this.searchForm = document.getElementById('search-form');
        this.paginationContainer = document.getElementById('pagination-controls');
        this.resultsContainer = document.getElementById('search-results');
        
        // --- STATE ---
        this.currentPage = 1;
        this.currentParams = new URLSearchParams();
    }

    // The search function now uses the ApiClient.
    async performSearch() {
        this.currentParams.set('page', this.currentPage);
        this.currentParams.set('limit', 25);

        try {
            const sins = await this.apiClient.searchSins(this.currentParams.toString());
            // The UIManager handles the rendering.
             this.uiManager.renderSins(sins, this.resultsContainer);  
            this.renderPagination(sins.length);
        } catch (error) {
            console.error('Search error:', error);
        }
    }

    // Pagination logic remains here as it's specific to the search page.
    renderPagination = (resultsCount) => {
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

    setupEventListeners() {
        this.searchForm.addEventListener('submit', (event) => {
            event.preventDefault();
            this.currentPage = 1;

            const tags = document.getElementById('tags-input').value;
            const description = document.getElementById('description-input').value;
            const emoji = document.getElementById('emoji-input').value;
            const sortBy = document.getElementById('sort-by-select').value;
            const order = document.getElementById('order-select').value;

            this.currentParams = new URLSearchParams();
            if (tags) this.currentParams.set('tags', tags);
            if (description) this.currentParams.set('description', description);
            if (emoji) this.currentParams.set('emoji', emoji);
            if (sortBy) this.currentParams.set('sortBy', sortBy);
            if (order) this.currentParams.set('order', order);
            
            this.performSearch();
        });
    }

    async init() {
        this.setupEventListeners();
        
        // Fetch and render the emoji picker for the search form.
        try {
            const emojis = await this.apiClient.fetchAllowedEmojis();
            this.uiManager.renderEmojiPicker(emojis); // Reusing the UIManager method
        } catch (error) {
            console.error("Failed to load emoji picker:", error);
        }

        this.performSearch(); // Perform initial search
    }
}

// --- Start the Application ---
document.addEventListener('DOMContentLoaded', () => {
    const searchApp = new SearchApp();
    searchApp.init();
});