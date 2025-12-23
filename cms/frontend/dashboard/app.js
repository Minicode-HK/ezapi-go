// --- API Helper ---
const api = {
    token: () => localStorage.getItem('cms_token'),
    async call(url, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token()}`,
            ...options.headers
        }
        const res = await fetch(url, { ...options, headers })
        if (res.status === 401) window.dispatchEvent(new CustomEvent('auth-error'))
        if (!res.ok) throw new Error(await res.text())
        return res.json()
    }
}

document.addEventListener('alpine:init', () => {
    
    // --- Global Store ---
    Alpine.store('cms', {
        resources: [],
        currentResource: null,
        activeView: 'dashboard', // 'dashboard' | 'snapshots' | 'resource'

        async loadSchema() {
            try {
                const res = await api.call('/cms/api/schema')
                this.resources = res.data || []
            } catch (e) { 
                console.error('Schema load failed', e) 
            }
        },

        selectResource(res) {
            this.currentResource = res
            this.activeView = 'resource'
        },

        selectSystemView(viewName) {
            this.currentResource = null
            this.activeView = viewName
        }
    })

    // --- Global Controller (Shell) ---
    Alpine.data('globalApp', () => ({
        token: api.token(),
        authRequired: true,
        sidebarOpen: false,
        toast: { show: false, message: '', type: 'success' },
        loginForm: { username: '', password: '' },
        loading: false,
        error: '',

        async init() {
            window.addEventListener('auth-error', () => {
                if (this.authRequired) this.logout()
            })
            window.addEventListener('show-toast', (e) => {
                this.toast = { show: true, message: e.detail.msg, type: e.detail.type || 'success' }
                setTimeout(() => this.toast.show = false, 3000)
            })

            // Check if auth is required
            await this.checkAuthRequired()

            // Load schema if logged in or no auth required
            if (this.token || !this.authRequired) {
                this.$store.cms.loadSchema()
            }
        },

        async checkAuthRequired() {
            try {
                const res = await fetch('/cms/api/auth-status')
                const data = await res.json()
                this.authRequired = data.data?.auth_required !== false
                if (!this.authRequired) {
                    this.token = 'no-auth-mode' // Set a dummy token
                }
            } catch (e) {
                // Default to requiring auth if check fails
                this.authRequired = true
            }
        },

        async login() {
            this.loading = true; this.error = '';
            try {
                const res = await fetch('/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(this.loginForm)
                })
                const data = await res.json()
                if (data.data?.token) {
                    this.token = data.data.token
                    localStorage.setItem('cms_token', data.data.token)
                    await this.$store.cms.loadSchema()
                } else {
                    this.error = 'Invalid credentials'
                }
            } catch (e) { this.error = e.message } 
            finally { this.loading = false }
        },

        logout() {
            this.token = ''; localStorage.removeItem('cms_token');
            window.location.reload();
        }
    }))

    // --- Resources Controller ---
    Alpine.data('resourceController', () => ({
        items: [],
        formData: {},
        jsonViewData: '',
        isEditing: false,
        loading: false,
        pendingDeleteItem: null,

        async init() {
            // If no resource selected (e.g. reload), try to select first one or wait
            if (!this.$store.cms.currentResource && this.$store.cms.resources.length > 0) {
                this.$store.cms.selectResource(this.$store.cms.resources[0])
            }
            
            if (this.$store.cms.currentResource) {
                await this.fetchItems()
            }

            // Watch for sidebar clicks changing the resource while staying on this view
            this.$watch('$store.cms.currentResource', async (val) => {
                if (val) await this.fetchItems()
            })
        },

        async fetchItems() {
            this.loading = true
            try {
                const endpoint = this.$store.cms.currentResource.endpoint
                const data = await api.call(endpoint)
                this.items = data.data || []
            } catch (e) { this.items = [] }
            finally { this.loading = false }
        },

        visibleFields() {
            const res = this.$store.cms.currentResource
            if (!res) return []
            return res.fields.filter(f => f.key !== 'password')
        },

        // CRUD
        openCreateModal() {
            this.isEditing = false; this.formData = {};
            this.$store.cms.currentResource.fields.forEach(f => {
                if (f.type === 'checkbox') this.formData[f.key] = false
            })
            document.getElementById('res_editor_modal').showModal()
        },

        viewJson(data) {
            this.jsonViewData = JSON.stringify(data, null, 2);
            document.getElementById('json_view_modal').showModal();
        },

        editItem(item) {
            this.isEditing = true; this.formData = { ...item };
            // Stringify objects for editing
            this.$store.cms.currentResource.fields.forEach(f => {
                const val = this.formData[f.key];
                if (typeof val === 'object' && val !== null) {
                    this.formData[f.key] = JSON.stringify(val, null, 2);
                }
            });
            document.getElementById('res_editor_modal').showModal()
        },

        async saveItem() {
            this.loading = true
            try {
                const res = this.$store.cms.currentResource
                const method = this.isEditing ? 'PUT' : 'POST'
                const url = this.isEditing ? `${res.endpoint}/${this.formData.id}` : res.endpoint
                
                const payload = { ...this.formData };

                res.fields.forEach(f => {
                    let val = payload[f.key];
                    if(f.type === 'number' && val) 
                        payload[f.key] = Number(val)
                    
                    // Try to parse JSON strings
                    if (typeof val === 'string' && (val.trim().startsWith('{') || val.trim().startsWith('['))) {
                        try {
                            payload[f.key] = JSON.parse(val);
                        } catch (e) {
                            // Ignore parse error, send as string
                        }
                    }
                })

                await api.call(url, { method, body: JSON.stringify(payload) })
                document.getElementById('res_editor_modal').close()
                this.notify('Saved successfully')
                await this.fetchItems()
            } catch (e) { this.notify(e.message, 'error') }
            finally { this.loading = false }
        },

        openDeleteConfirm(item) {
            this.pendingDeleteItem = item
            document.getElementById('res_delete_modal').showModal()
        },

        async confirmDelete() {
            if (!this.pendingDeleteItem) return
            try {
                const res = this.$store.cms.currentResource
                await api.call(`${res.endpoint}/${this.pendingDeleteItem.id}`, { method: 'DELETE' })
                document.getElementById('res_delete_modal').close()
                this.notify('Deleted successfully')
                await this.fetchItems()
            } catch (e) { this.notify('Failed to delete', 'error') }
        },

        notify(msg, type) {
            window.dispatchEvent(new CustomEvent('show-toast', { detail: { msg, type } }))
        }
    }))

    // --- Snapshots Controller ---
    Alpine.data('snapshotController', () => ({
        snapshots: [],
        loading: false,
        viewData: null,
        selectedModule: null,

        async init() { await this.fetchSnapshots() },

        getHeaders() {
            return {
                'Authorization': 'Bearer ' + localStorage.getItem('cms_token'),
                'Content-Type': 'application/json'
            };
        },

        async fetchSnapshots() {
            this.loading = true
            try {
                const res = await api.call('/cms/api/snapshots')
                this.snapshots = res.data || []
            } catch (e) { this.notify('Failed to load snapshots', 'error') }
            finally { this.loading = false }
        },

        async createSnapshot() {
            this.loading = true
            try {
                await api.call('/cms/api/snapshots', { method: 'POST' })
                this.notify('Snapshot created')
                await this.fetchSnapshots()
            } catch (e) { this.notify(e.message, 'error') }
            finally { this.loading = false }
        },

        async restoreSnapshot(snap) {
            if (!confirm(`Restore ${snap.name}?`)) return;
            this.loading = true
            try {
                await api.call(`/cms/api/snapshots/${snap.name}/restore`, { method: 'POST' })
                this.notify('System restored')
            } catch (e) { this.notify(e.message, 'error') }
            finally { this.loading = false }
        },

        async deleteSnapshot(snap) {
            if (!confirm(`Delete ${snap.name}?`)) return;
            this.loading = true
            try {
                await api.call(`/cms/api/snapshots/${snap.name}`, { method: 'DELETE' })
                this.notify('Snapshot deleted')
                await this.fetchSnapshots()
            } catch (e) { this.notify(e.message, 'error') }
            finally { this.loading = false }
        },

        async viewSnapshot(snap) {
            this.loading = true;
            this.viewData = null;
            try {
                const res = await fetch(`/cms/api/snapshots/${snap.id}/content`, { headers: this.getHeaders() });
                if (res.ok) {
                    this.viewData = await res.json();
                    document.getElementById('snapshot_modal').showModal();
                } else {
                    alert("Failed to load snapshot content");
                }
            } catch (e) {
                console.error(e);
                alert("Error loading content");
            } finally {
                this.loading = false;
            }
        },

        formatBytes(bytes) {
            if (!+bytes) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
        },

        notify(msg, type) {
            window.dispatchEvent(new CustomEvent('show-toast', { detail: { msg, type } }))
        }
    }))

    Alpine.data('systemStats', () => ({
        stats: { alloc: 0, total_alloc: 0, sys: 0, num_gc: 0, goroutines: 0, timestamp: '--:--:--' },
        controller: null,
        buffer: '',

        init() {
            this.startStream();
        },

        destroy() {
            if (this.controller) {
                this.controller.abort();
            }
        },

        async startStream() {
            this.controller = new AbortController();
            const token = localStorage.getItem('cms_token');

            try {
                const response = await fetch('/cms/api/stream/stats', {
                    headers: { 
                        'Authorization': 'Bearer ' + token,
                        'Accept': 'text/event-stream'
                    },
                    signal: this.controller.signal
                });

                if (!response.ok) throw new Error(response.statusText);

                const reader = response.body.getReader();
                const decoder = new TextDecoder();

                while (true) {
                    const { value, done } = await reader.read();
                    if (done) break;
                    
                    const chunk = decoder.decode(value, { stream: true });
                    this.processChunk(chunk);
                }
            } catch (e) {
                if (e.name !== 'AbortError') {
                    console.error("Stream error:", e);
                    // Retry after 5 seconds on error
                    setTimeout(() => this.startStream(), 5000);
                }
            }
        },

        processChunk(chunk) {
            this.buffer += chunk;
            
            // SSE messages are separated by double newlines
            const parts = this.buffer.split('\n\n');
            this.buffer = parts.pop(); // Keep the last incomplete part

            for (const part of parts) {
                if (!part.trim()) continue;
                
                const lines = part.split('\n');
                let eventType = 'message';
                let data = '';

                for (const line of lines) {
                    if (line.startsWith('event:')) eventType = line.substring(6).trim();
                    if (line.startsWith('data:')) data = line.substring(5).trim();
                }

                if (eventType === 'stats' && data) {
                    try {
                        this.stats = JSON.parse(data);
                    } catch (e) {
                        console.error("Failed to parse stats JSON", e);
                    }
                }
            }
        },

        formatBytes(bytes, decimals = 2) {
            if (!+bytes) return '0 B';
            const k = 1024;
            const dm = decimals < 0 ? 0 : decimals;
            const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
        }
    }));
})