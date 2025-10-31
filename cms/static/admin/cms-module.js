// This handles CRUD operations for a specific module
class CMSModule {
    constructor(moduleName) {
        this.moduleName = moduleName;
        this.schema = null;
        this.apiBase = '/api';
    }

    async setModule(moduleName) {
        this.moduleName = moduleName;
        this.schema = window.cmsManager.getSchema(this.moduleName);
        
        if (!this.schema) {
            console.error(`Schema not found for module: ${this.moduleName}`);
            return;
        }

        await this.loadModule();
    }

    async fetchWithAuth(url, options = {}) {
        return window.cmsManager.fetchWithAuth(url, options);
    }

    async loadModule() {
        const response = await this.fetchWithAuth(`${this.schema.base_path.toLowerCase()}`);
        const data = await response.json();
        this.renderTable(data);
    }

    renderTable(data) {
        const content = document.getElementById('content-area');
        if (!content) return;

        if (data.success === false) {
            content.innerHTML = `
                <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
                    <strong class="font-bold">Error:</strong>
                    <span class="block sm:inline">${data.message}</span>
                </div>
            `;
            return;
        }

        const isEmpty = !data.data || data.data.length === 0;
        
        content.innerHTML = `
            <div class="bg-white rounded-lg shadow-sm">
                <div class="flex justify-between items-center p-6 border-b border-gray-200">
                    <div>
                        <h3 class="text-lg font-semibold text-gray-900">${this.schema.name}</h3>
                        <p class="text-sm text-gray-600 mt-1">${data.data.length} record(s) found</p>
                    </div> 
                    <button onclick="currentModule.showCreateForm()"  style="cursor: pointer;"
                            class="bg-blue-700 text-white px-4 py-2 rounded-lg hover:bg-blue-800 transition-colors flex items-center gap-2">
                        <i class="fas fa-plus"></i>
                        <span>Create New</span>
                    </button>
                </div>
                
                ${isEmpty ? `
                    <div class="p-12 text-center">
                        <i class="fas fa-inbox text-6xl text-gray-300 mb-4"></i>
                        <p class="text-gray-600 mb-4">No records found</p>
                        <button onclick="currentModule.showCreateForm()" 
                                class="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors">
                            Create First Record
                        </button>
                    </div>
                ` : `
                    <div class="overflow-x-auto">
                        <table class="min-w-full divide-y divide-gray-200">
                            <thead class="bg-gray-50">
                                <tr>
                                    ${this.schema.fields.map(f => `
                                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            ${f.name}
                                        </th>
                                    `).join('')}
                                    <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Actions
                                    </th>
                                </tr>
                            </thead>
                            <tbody class="bg-white divide-y divide-gray-200">
                                ${data.data.map(row => this.renderTableRow(row)).join('')}
                            </tbody>
                        </table>
                    </div>
                `}
            </div>
        `;
    }

    renderTableRow(row) {
        return `
            <tr class="hover:bg-gray-50 transition-colors">
                ${this.schema.fields.map(f => {
                    let value = row[f.json_tag];
                    
                    if (typeof value === 'boolean') {
                        value = value 
                            ? '<span class="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800"><i class="fas fa-check"></i> Yes</span>'
                            : '<span class="px-2 py-1 text-xs rounded-full bg-gray-100 text-gray-800"><i class="fas fa-times"></i> No</span>';
                    } else if (typeof value === 'object' && value !== null) {
                        value = `<code class="text-xs bg-gray-100 px-2 py-1 rounded">${JSON.stringify(value).substring(0, 50)}...</code>`;
                    } else if (value === null || value === undefined || value === '') {
                        value = '<span class="text-gray-400 italic">-</span>';
                    }
                    
                    return `<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${value}</td>`;
                }).join('')}
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button onclick="currentModule.edit('${row.id}')" 
                            class="text-blue-600 hover:text-blue-900 mr-4 transition-colors" style="cursor: pointer;"
                            title="Edit">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button onclick="currentModule.delete('${row.id}')" 
                            class="text-red-600 hover:text-red-900 transition-colors" style="cursor: pointer;"
                            title="Delete">
                        <i class="fas fa-trash"></i>
                    </button>
                </td>
            </tr>
        `;
    }

    getFieldType(field) {
        const type = field.type.toLowerCase();
        
        if (type.includes('bool')) return 'checkbox';
        if (type.includes('int') || type.includes('float') || type.includes('number')) return 'number';
        if (type.includes('time') || field.json_tag.includes('date')) return 'datetime-local';
        if (type.includes('[]') || type.includes('map') || field.json_tag.includes('json')) return 'json';
        
        return 'text';
    }

    renderFormField(field, value = '') {
        const fieldType = this.getFieldType(field);
        const required = field.required ? 'required' : '';
        const fieldId = `field-${field.json_tag}`;

        switch (fieldType) {
            case 'checkbox':
                return `
                    <div class="mb-4">
                        <label class="flex items-center cursor-pointer">
                            <input type="checkbox" 
                                   id="${fieldId}"
                                   name="${field.json_tag}" 
                                   ${value ? 'checked' : ''}
                                   class="mr-2 h-4 w-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500">
                            <span class="text-sm font-medium text-gray-700">${field.name}</span>
                        </label>
                    </div>
                `;
            
            case 'json':
                const jsonValue = typeof value === 'object' ? JSON.stringify(value, null, 2) : value;
                return `
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            ${field.name} ${field.required ? '<span class="text-red-500">*</span>' : ''}
                        </label>
                        <textarea id="${fieldId}"
                                  name="${field.json_tag}" 
                                  rows="6"
                                  ${required}
                                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                                  placeholder='{"key": "value"}'>${jsonValue || ''}</textarea>
                        <p class="mt-1 text-xs text-gray-500"><i class="fas fa-info-circle"></i> Enter valid JSON format</p>
                    </div>
                `;
            
            case 'number':
                return `
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            ${field.name} ${field.required ? '<span class="text-red-500">*</span>' : ''}
                        </label>
                        <input type="number" 
                               id="${fieldId}"
                               name="${field.json_tag}" 
                               value="${value || ''}"
                               step="any"
                               ${required}
                               class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
                    </div>
                `;
            
            case 'datetime-local':
                let dateValue = '';
                if (value) {
                    try {
                        dateValue = new Date(value).toISOString().slice(0, 16);
                    } catch (e) {
                        dateValue = '';
                    }
                }
                return `
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            ${field.name} ${field.required ? '<span class="text-red-500">*</span>' : ''}
                        </label>
                        <input type="datetime-local" 
                               id="${fieldId}"
                               name="${field.json_tag}" 
                               value="${dateValue}"
                               ${required}
                               class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
                    </div>
                `;
            
            default:
                return `
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            ${field.name} ${field.required ? '<span class="text-red-500">*</span>' : ''}
                        </label>
                        <input type="text" 
                               id="${fieldId}"
                               name="${field.json_tag}" 
                               value="${value || ''}"
                               ${required}
                               class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                               placeholder="Enter ${field.name.toLowerCase()}">
                    </div>
                `;
        }
    }

    showCreateForm() {
        this.showEditForm(null);
    }

    async edit(id) {
        try {
            const response = await this.fetchWithAuth(`${this.schema.base_path.toLowerCase()}/${id}`);
            const result = await response.json();
            
            if (result.success) {
                this.showEditForm(result.data);
            } else {
                alert('Failed to load item: ' + result.message);
            }
        } catch (error) {
            alert('Error loading item: ' + error.message);
        }
    }

    showEditForm(data) {
        const isEdit = data !== null;
        const modalTitle = isEdit ? `Edit ${this.schema.name}` : `Create ${this.schema.name}`;
        const modalIcon = isEdit ? 'fa-edit' : 'fa-plus-circle';
        
        const modalHTML = `
            <div id="edit-modal" class="fixed inset-0 flex items-center justify-center z-50">
                <div class="fixed inset-0 bg-black opacity-50 z-10"  onclick="currentModule.closeModal()"></div>
                <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden z-20 mx-4">
                    <div class="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-gray-600 to-blue-700 text-white">
                        <div class="flex justify-between items-center">
                            <h3 class="text-xl font-semibold flex items-center gap-2">
                                <i class="fas ${modalIcon}"></i>
                                ${modalTitle}
                            </h3>
                            <button onclick="currentModule.closeModal()"  style="cursor: pointer;"
                                    class="text-white hover:text-gray-200 transition-colors">
                                <i class="fas fa-times text-xl"></i>
                            </button>
                        </div>
                    </div>
                    
                    <form id="edit-form" class="px-6 py-4 overflow-y-auto max-h-[calc(90vh-140px)]">
                        ${this.schema.fields.map(field => {
                            const value = data ? data[field.json_tag] : '';
                            return this.renderFormField(field, value);
                        }).join('')}
                    </form>
                    
                    <div class="px-6 py-4 border-t border-gray-200 bg-gray-50 flex justify-end gap-3">
                        <button type="button" style="cursor: pointer;"
                                onclick="currentModule.closeModal()" 
                                class="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
                            <i class="fas fa-times mr-2"></i>Cancel
                        </button>
                        <button type="button" style="cursor: pointer;"
                                onclick="currentModule.submitForm(${isEdit}, '${data?.id || ''}')" 
                                class="px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2">
                            <i class="fas fa-save"></i>
                            <span>${isEdit ? 'Update' : 'Create'}</span>
                        </button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.insertAdjacentHTML('beforeend', modalHTML);
    }

    closeModal() {
        const modal = document.getElementById('edit-modal');
        if (modal) modal.remove();
    }

    async submitForm(isEdit, id) {
        const form = document.getElementById('edit-form');
        const formData = new FormData(form);
        
        const data = {};
        for (const [key, value] of formData.entries()) {
            const field = this.schema.fields.find(f => f.json_tag === key);
            const fieldType = this.getFieldType(field);
            
            if (fieldType === 'number') {
                data[key] = parseFloat(value) || 0;
            } else if (fieldType === 'json') {
                try {
                    data[key] = value ? JSON.parse(value) : null;
                } catch (e) {
                    alert(`Invalid JSON in field "${field.name}". Please check the format.`);
                    return;
                }
            } else {
                data[key] = value;
            }
        }
        
        this.schema.fields.forEach(field => {
            if (this.getFieldType(field) === 'checkbox') {
                const checkbox = document.getElementById(`field-${field.json_tag}`);
                data[field.json_tag] = checkbox ? checkbox.checked : false;
            }
        });
        
        if (isEdit) data.id = id;
        
        const url = isEdit 
            ? `${this.schema.base_path.toLowerCase()}/${id}`
            : `${this.schema.base_path.toLowerCase()}`;
        
        const method = isEdit ? 'PUT' : 'POST';
        
        try {
            const response = await this.fetchWithAuth(url, {
                method: method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            
            const result = await response.json();
            
            if (result.success) {
                this.closeModal();
                this.loadModule();
                this.showNotification(`${isEdit ? 'Updated' : 'Created'} successfully`, 'success');
            } else {
                alert('Error: ' + (result.message || 'Unknown error'));
            }
        } catch (error) {
            alert('Network error: ' + error.message);
        }
    }

    async delete(id) {
        if (confirm('Are you sure you want to delete this item? This action cannot be undone.')) {
            try {
                const response = await this.fetchWithAuth(`${this.schema.base_path.toLowerCase()}/${id}`, {
                    method: 'DELETE'
                });
                
                const result = await response.json();
                
                if (result.success) {
                    this.loadModule();
                    this.showNotification('Deleted successfully', 'success');
                } else {
                    alert('Error: ' + result.message);
                }
            } catch (error) {
                alert('Error deleting item: ' + error.message);
            }
        }
    }

    showNotification(message, type = 'success') {
        const bgColor = type === 'success' ? 'bg-green-500' : 'bg-red-500';
        const icon = type === 'success' ? 'fa-check-circle' : 'fa-exclamation-circle';
        
        const notification = document.createElement('div');
        notification.className = `fixed top-4 right-4 ${bgColor} text-white px-6 py-3 rounded-lg shadow-lg z-50 flex items-center gap-2`;
        notification.innerHTML = `
            <i class="fas ${icon}"></i>
            <span>${message}</span>
        `;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.style.opacity = '0';
            notification.style.transition = 'opacity 0.3s';
            setTimeout(() => notification.remove(), 300);
        }, 3000);
    }
}