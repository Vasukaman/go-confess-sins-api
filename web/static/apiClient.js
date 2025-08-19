export class ApiClient {
    constructor(sinApiUrl, webApiUrl) {
        this.sinApiUrl = sinApiUrl;
        this.webApiUrl = webApiUrl;
    }

    async fetchSins() {
        const response = await fetch(`${this.sinApiUrl}/sins`);
        return response.json();
    }

    async fetchLeaderboard() {
        const response = await fetch(`${this.webApiUrl}/api/leaderboard`);
        return response.json();
    }

    async getNewKey() {
        const response = await fetch(`${this.sinApiUrl}/keys`, { method: 'POST' });
        return response.json();
    }

    async confessSin(payload) {
        const response = await fetch(`${this.webApiUrl}/api/confess`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        if (!response.ok) {
            throw new Error('API returned an error during confession.');
        }
        return response.json();
    }

    async fetchAllowedEmojis() {
        const response = await fetch(`${this.sinApiUrl}/emojis`);
        return response.json();
    }
}