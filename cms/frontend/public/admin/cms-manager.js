// This handles navigation and shared functionality
class CMSManager {
    constructor() {
        this.schemas = [];
        this.init();
    }

    async init() {
        try {
            await this.checkAuth();
            await this.loadSchemas();
            this.renderNavigation();
            this.setupLogout();
        } catch (error) {
            console.error('Initialization error:', error);
        }
    }

    async checkAuth() {
        try {
            const response = await this.fetchWithAuth('/cms/api/me');
            const result = await response.json();
            
            if (result.success) {
                const usernameDisplay = document.getElementById('username-display');
                if (usernameDisplay) {
                    usernameDisplay.textContent = result.data.username;
                }
            }
        } catch (error) {
            console.error('Auth check failed:', error);
        }
    }

    setupLogout() {
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', async () => {
                try {
                    await this.fetchWithAuth('/cms/logout', { method: 'POST' });
                    window.location.href = '/cms/login';
                } catch (error) {
                    console.error('Logout error:', error);
                    window.location.href = '/cms/login';
                }
            });
        }
    }

    async loadSchemas() {
        const response = await this.fetchWithAuth('/cms/api/schemas');
        this.schemas = await response.json();
    }

    renderNavigation() {
        const nav = document.getElementById('sidebar-nav');
        if (!nav) return;

        nav.innerHTML = this.schemas.map(schema => `
            <a href="#" 
               hx-get="/cms/api/content/module/${schema.name}" 
               hx-target="#main-content" 
               hx-swap="innerHTML"
               class="nav-link block px-4 py-2 hover:bg-gray-700 transition-colors">
                <i class="fas fa-table mr-2"></i> ${schema.name}
            </a>
        `).join('');
        
        // Process htmx attributes
        if (typeof htmx !== 'undefined') {
            htmx.process(nav);
        }
    }

    async fetchWithAuth(url, options = {}) {
        const response = await fetch(url, { 
            ...options, 
            credentials: 'include'
        });

        if (response.status === 401) {
            window.location.href = '/cms/login';
            throw new Error('Unauthorized');
        }
        
        return response;
    }

    getSchemas() {
        return this.schemas;
    }

    getSchema(name) {
        return this.schemas.find(s => s.name === name);
    }
}

// Initialize global manager
window.cmsManager = new CMSManager();