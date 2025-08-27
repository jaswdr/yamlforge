package ui

import (
	"strings"
)

func GetStaticFile(path string) ([]byte, string, bool) {
	files := getStaticFiles()

	if content, ok := files[path]; ok {
		contentType := "text/plain"

		if strings.HasSuffix(path, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(path, ".js") {
			contentType = "application/javascript"
		}

		return []byte(content), contentType, true
	}

	return nil, "", false
}

func getStaticFiles() map[string]string {
	return map[string]string{
		"css/style.css": getCSS(),
		"js/app.js":     getJS(),
	}
}

func getCSS() string {
	return `/* Reset and Base Styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

:root {
    --primary-gradient: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    --primary-color: #667eea;
    --primary-dark: #5a67d8;
    --secondary-color: #764ba2;
    --success-color: #10b981;
    --danger-color: #ef4444;
    --warning-color: #f59e0b;
    --info-color: #3b82f6;
    --gray-50: #f9fafb;
    --gray-100: #f3f4f6;
    --gray-200: #e5e7eb;
    --gray-300: #d1d5db;
    --gray-400: #9ca3af;
    --gray-500: #6b7280;
    --gray-600: #4b5563;
    --gray-700: #374151;
    --gray-800: #1f2937;
    --gray-900: #111827;
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
    --shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
    --shadow-2xl: 0 25px 50px -12px rgb(0 0 0 / 0.25);
    --radius: 12px;
    --radius-sm: 8px;
    --radius-lg: 16px;
}

body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    line-height: 1.6;
    color: var(--gray-800);
    background: linear-gradient(to bottom right, var(--gray-50), var(--gray-100));
    min-height: 100vh;
}

/* Layout */
.app-container {
    display: flex;
    min-height: 100vh;
    background: white;
}

.sidebar {
    width: 280px;
    background: var(--primary-gradient);
    color: white;
    padding: 0;
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-xl);
    position: relative;
    overflow: hidden;
}

.sidebar::before {
    content: '';
    position: absolute;
    width: 100%;
    height: 100%;
    background: radial-gradient(circle at top right, rgba(255,255,255,0.1) 0%, transparent 50%);
    pointer-events: none;
}

.sidebar .logo {
    padding: 2rem 1.5rem;
    border-bottom: 1px solid rgba(255,255,255,0.1);
    position: relative;
    z-index: 1;
}

.sidebar .logo h1 {
    font-size: 1.75rem;
    font-weight: 700;
    letter-spacing: -0.5px;
    margin: 0;
    text-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.nav-menu {
    list-style: none;
    padding: 1.5rem;
    flex: 1;
    position: relative;
    z-index: 1;
}

.nav-menu li {
    margin-bottom: 0.5rem;
}

.nav-menu a {
    color: rgba(255,255,255,0.9);
    text-decoration: none;
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    border-radius: var(--radius-sm);
    transition: all 0.3s;
    font-weight: 500;
    position: relative;
    overflow: hidden;
}

.nav-menu a::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    height: 100%;
    width: 3px;
    background: white;
    transform: translateX(-100%);
    transition: transform 0.3s;
}

.nav-menu a:hover, .nav-menu a.active {
    background-color: rgba(255,255,255,0.15);
    color: white;
    transform: translateX(4px);
}

.nav-menu a:hover::before, .nav-menu a.active::before {
    transform: translateX(0);
}

.main-content {
    flex: 1;
    padding: 2rem;
    background: var(--gray-50);
    overflow-y: auto;
}

/* Page Header */
.page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
}

.page-header h2 {
    font-size: 2rem;
    font-weight: 700;
    color: var(--gray-800);
    margin: 0;
}

.page-actions {
    display: flex;
    gap: 0.75rem;
}

/* Buttons */
.btn {
    padding: 0.625rem 1.25rem;
    border: none;
    border-radius: var(--radius-sm);
    cursor: pointer;
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    font-weight: 600;
    transition: all 0.3s;
    position: relative;
    overflow: hidden;
    box-shadow: var(--shadow-sm);
}

.btn::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 0;
    height: 0;
    border-radius: 50%;
    background: rgba(255,255,255,0.3);
    transform: translate(-50%, -50%);
    transition: width 0.6s, height 0.6s;
}

.btn:hover::before {
    width: 300px;
    height: 300px;
}

.btn-primary {
    background: var(--primary-gradient);
    color: white;
}

.btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
}

.btn-secondary {
    background: white;
    color: var(--gray-700);
    border: 2px solid var(--gray-200);
}

.btn-secondary:hover {
    background: var(--gray-50);
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
    border-color: var(--gray-300);
}

.btn-danger {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
    color: white;
}

.btn-danger:hover {
    transform: translateY(-2px);
    box-shadow: 0 10px 20px rgba(239, 68, 68, 0.3);
}

.btn-success {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    color: white;
}

.btn-success:hover {
    transform: translateY(-2px);
    box-shadow: 0 10px 20px rgba(16, 185, 129, 0.3);
}

/* Dashboard */
.dashboard {
    padding: 0;
}

.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
}

.stat-card {
    background: white;
    padding: 1.75rem;
    border-radius: var(--radius);
    box-shadow: var(--shadow-sm);
    transition: all 0.3s;
    position: relative;
    overflow: hidden;
}

.stat-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 4px;
    background: var(--primary-gradient);
    transform: translateX(-100%);
    transition: transform 0.3s;
}

.stat-card:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-lg);
}

.stat-card:hover::before {
    transform: translateX(0);
}

.stat-card h3 {
    margin-bottom: 1.25rem;
    color: var(--gray-800);
    font-size: 1.25rem;
    font-weight: 600;
}

.stat-actions {
    display: flex;
    gap: 0.75rem;
}

/* List View */
.list-container {
    background: white;
    padding: 1.5rem;
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
}

.search-bar {
    margin-bottom: 1.5rem;
}

.search-bar input {
    width: 100%;
    max-width: 400px;
    padding: 0.625rem 1rem;
    border: 2px solid var(--gray-200);
    border-radius: var(--radius-sm);
    font-size: 0.875rem;
    transition: all 0.3s;
    background: var(--gray-50);
}

.search-bar input:focus {
    outline: none;
    border-color: var(--primary-color);
    background: white;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

/* Table */
.data-table-container {
    overflow-x: auto;
    border-radius: var(--radius);
    border: 1px solid var(--gray-200);
}

.data-table {
    width: 100%;
    border-collapse: collapse;
}

.data-table th {
    text-align: left;
    padding: 1rem;
    background: var(--gray-50);
    border-bottom: 2px solid var(--gray-200);
    font-weight: 600;
    color: var(--gray-700);
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.data-table td {
    padding: 1rem;
    border-bottom: 1px solid var(--gray-100);
    color: var(--gray-700);
    font-size: 0.875rem;
}

.data-table tbody tr {
    transition: all 0.2s;
}

.data-table tbody tr:hover {
    background: linear-gradient(to right, rgba(102, 126, 234, 0.05), rgba(118, 75, 162, 0.05));
}

.data-table .actions {
    display: flex;
    gap: 0.5rem;
}

.btn-sm {
    padding: 0.375rem 0.75rem;
    font-size: 0.8125rem;
}

/* Forms */
.form-container {
    background: white;
    padding: 2rem;
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-md);
    max-width: 800px;
}

.model-form {
    max-width: 600px;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 600;
    color: var(--gray-700);
    font-size: 0.875rem;
}

.form-control {
    width: 100%;
    padding: 0.625rem 0.875rem;
    border: 2px solid var(--gray-200);
    border-radius: var(--radius-sm);
    font-size: 0.875rem;
    transition: all 0.3s;
    background: var(--gray-50);
}

.form-control:focus {
    outline: none;
    border-color: var(--primary-color);
    background: white;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-control:disabled {
    background: var(--gray-100);
    cursor: not-allowed;
    opacity: 0.6;
}

textarea.form-control {
    min-height: 120px;
    resize: vertical;
    font-family: inherit;
    line-height: 1.5;
}

select.form-control {
    appearance: none;
    background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
    background-position: right 0.75rem center;
    background-repeat: no-repeat;
    background-size: 1.25em 1.25em;
    padding-right: 2.5rem;
    cursor: pointer;
}

.checkbox-group {
    display: flex;
    align-items: center;
    cursor: pointer;
}

.checkbox-group input[type="checkbox"] {
    width: 1.25rem;
    height: 1.25rem;
    margin-right: 0.5rem;
    cursor: pointer;
    accent-color: var(--primary-color);
}

.form-actions {
    display: flex;
    gap: 0.75rem;
    margin-top: 2rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--gray-200);
}

/* View Page */
.view-container,
.detail-container {
    background: white;
    padding: 2rem;
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-md);
    max-width: 800px;
}

.field-view,
.detail-row {
    margin-bottom: 1.5rem;
    padding-bottom: 1.5rem;
    border-bottom: 1px solid var(--gray-100);
}

.field-view:last-child,
.detail-row:last-child {
    border-bottom: none;
    padding-bottom: 0;
}

.field-label,
.detail-label {
    font-weight: 600;
    color: var(--gray-600);
    margin-bottom: 0.5rem;
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.field-value,
.detail-value {
    color: var(--gray-800);
    font-size: 1rem;
    line-height: 1.6;
}

/* Pagination */
.pagination {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 0.5rem;
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--gray-200);
}

.pagination button {
    padding: 0.5rem 0.75rem;
    border: 2px solid var(--gray-200);
    background: white;
    cursor: pointer;
    border-radius: var(--radius-sm);
    font-weight: 500;
    font-size: 0.875rem;
    transition: all 0.3s;
    min-width: 2.5rem;
}

.pagination button:hover:not(:disabled) {
    background: var(--gray-50);
    border-color: var(--primary-color);
    transform: translateY(-2px);
}

.pagination button.active {
    background: var(--primary-gradient);
    color: white;
    border-color: transparent;
    box-shadow: var(--shadow-sm);
}

.pagination button:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}

/* Loading State */
.loading {
    text-align: center;
    padding: 3rem;
    color: var(--gray-500);
    font-style: italic;
}

.loading-spinner {
    border: 3px solid var(--gray-200);
    border-top: 3px solid transparent;
    border-right: 3px solid transparent;
    border-image: var(--primary-gradient) 1;
    border-radius: 50%;
    width: 48px;
    height: 48px;
    animation: spin 0.8s linear infinite;
    margin: 2rem auto;
}

@keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}

/* Empty State */
.empty-state {
    text-align: center;
    padding: 4rem;
    color: var(--gray-500);
}

.empty-state h3 {
    font-size: 1.5rem;
    margin-bottom: 0.75rem;
    color: var(--gray-700);
}

/* Error & Success Messages */
.error {
    color: var(--danger-color);
    font-size: 0.8125rem;
    margin-top: 0.25rem;
    display: flex;
    align-items: center;
    gap: 0.25rem;
}

.error::before {
    content: '⚠';
    font-size: 1rem;
}

.field-error {
    border-color: var(--danger-color) !important;
    background: rgba(239, 68, 68, 0.05) !important;
}

.error-message {
    background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
    color: var(--danger-color);
    padding: 1rem;
    border-radius: var(--radius);
    margin-bottom: 1.5rem;
    border-left: 4px solid var(--danger-color);
}

.success-message {
    background: linear-gradient(135deg, #dcfce7 0%, #bbf7d0 100%);
    color: var(--success-color);
    padding: 1rem;
    border-radius: var(--radius);
    margin-bottom: 1.5rem;
    border-left: 4px solid var(--success-color);
}

/* Animations */
@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(10px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.fade-in {
    animation: fadeIn 0.3s ease-out;
}

/* Custom Scrollbar */
::-webkit-scrollbar {
    width: 10px;
    height: 10px;
}

::-webkit-scrollbar-track {
    background: var(--gray-100);
}

::-webkit-scrollbar-thumb {
    background: var(--gray-400);
    border-radius: 5px;
}

::-webkit-scrollbar-thumb:hover {
    background: var(--gray-500);
}

/* Responsive */
@media (max-width: 768px) {
    .app-container {
        flex-direction: column;
    }
    
    .sidebar {
        width: 100%;
        min-height: auto;
    }
    
    .stats-grid {
        grid-template-columns: 1fr;
    }
    
    .page-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 1rem;
    }
    
    .form-container,
    .view-container,
    .detail-container {
        padding: 1.5rem;
    }
    
    .main-content {
        padding: 1rem;
    }
}`
}

func getJS() string {
	return `const API_BASE = '/api';

let currentPage = 1;
let currentSearch = '';
let currentSort = [];

async function loadList(modelName, columns, searchable, sortable, modelInfo) {
    const params = new URLSearchParams({
        page: currentPage,
        page_size: 20
    });

    if (currentSearch) {
        params.append('search', currentSearch);
    }

    if (currentSort.length > 0) {
        params.append('sort', currentSort.join(','));
    }

    try {
        const response = await fetch(` + "`${API_BASE}/${modelName}?${params}`" + `);
        const data = await response.json();

        if (data.success) {
            renderTable(data.data, columns, modelName, modelInfo || window.modelInfo);
            renderPagination(data.meta);
        } else {
            showError(data.error);
        }
    } catch (error) {
        showError('Failed to load data');
    }
}


function formatFieldValue(fieldName, value, modelInfo) {
    if (modelInfo && modelInfo.fields && modelInfo.fields[fieldName]) {
        const fieldInfo = modelInfo.fields[fieldName];
        
        if (fieldInfo.type === 'password') {
            return '********';
        }
        
        if (fieldInfo.options && fieldInfo.options[value] !== undefined) {
            return fieldInfo.options[value];
        }
    }
    
    if (typeof value === 'boolean') {
        return value ? 'Yes' : 'No';
    }
    
    if (fieldName.includes('date') || fieldName.includes('_at')) {
        if (!value) return '';
        const date = new Date(value);
        if (isNaN(date.getTime())) return value;
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
    }
    
    return value || '';
}


function renderTable(data, columns, modelName, modelInfo) {
    const tbody = document.getElementById('tableBody');
    tbody.innerHTML = '';

    if (data.length === 0) {
        tbody.innerHTML = '<tr><td colspan="' + (columns.length + 1) + '" class="empty-state">No records found</td></tr>';
        return;
    }

    data.forEach(record => {
        const row = document.createElement('tr');

        columns.forEach(col => {
            const cell = document.createElement('td');
            cell.textContent = formatFieldValue(col, record[col], modelInfo);
            row.appendChild(cell);
        });

        const actionsCell = document.createElement('td');
        actionsCell.className = 'actions';
        let actionsHTML = ` + "`" + `<a href="/${modelName}/${record.id}" class="btn btn-sm btn-secondary">View</a>` + "`" + `;
        
        if (typeof canWrite !== 'undefined' && canWrite) {
            actionsHTML += ` + "`" + ` <a href="/${modelName}/${record.id}/edit" class="btn btn-sm btn-primary">Edit</a>` + "`" + `;
            actionsHTML += ` + "`" + ` <button onclick="deleteRecord('${modelName}', '${record.id}')" class="btn btn-sm btn-danger">Delete</button>` + "`" + `;
        }
        
        actionsCell.innerHTML = actionsHTML;
        row.appendChild(actionsCell);

        tbody.appendChild(row);
    });
}


function renderPagination(meta) {
    const pagination = document.getElementById('pagination');
    if (!pagination || !meta) return;

    pagination.innerHTML = '';

    const prevBtn = document.createElement('button');
    prevBtn.textContent = 'Previous';
    prevBtn.disabled = meta.page === 1;
    prevBtn.onclick = () => {
        currentPage = meta.page - 1;
        if (typeof modelName !== 'undefined' && typeof columns !== 'undefined') {
            loadList(modelName, columns, searchable, sortable, window.modelInfo);
        }
    };
    pagination.appendChild(prevBtn);

    const totalPages = meta.total_pages || 1;
    const startPage = Math.max(1, meta.page - 2);
    const endPage = Math.min(totalPages, meta.page + 2);

    for (let i = startPage; i <= endPage; i++) {
        const pageBtn = document.createElement('button');
        pageBtn.textContent = i;
        pageBtn.className = i === meta.page ? 'active' : '';
        pageBtn.onclick = () => {
            currentPage = i;
            if (typeof modelName !== 'undefined' && typeof columns !== 'undefined') {
                loadList(modelName, columns, searchable, sortable, window.modelInfo);
            }
        };
        pagination.appendChild(pageBtn);
    }

    const nextBtn = document.createElement('button');
    nextBtn.textContent = 'Next';
    nextBtn.disabled = meta.page === totalPages;
    nextBtn.onclick = () => {
        currentPage = meta.page + 1;
        if (typeof modelName !== 'undefined' && typeof columns !== 'undefined') {
            loadList(modelName, columns, searchable, sortable, window.modelInfo);
        }
    };
    pagination.appendChild(nextBtn);
}


async function handleForm(modelName, action, recordId) {
    const form = document.getElementById('modelForm');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const formData = new FormData(form);
        const data = {};

        for (const [key, value] of formData.entries()) {
            if (form.elements[key].type === 'checkbox') {
                data[key] = form.elements[key].checked;
            } else if (form.elements[key].type === 'password') {
                if (value !== '') {
                    data[key] = value;
                }
                if (action === 'create') {
                    data[key] = value;
                }
            } else if (form.elements[key].type === 'number') {
                data[key] = value === '' ? null : (parseFloat(value) || 0);
            } else if (form.elements[key].type === 'date' && value === '') {
                data[key] = null;
            } else if (form.elements[key].type === 'datetime-local' && value === '') {
                data[key] = null;
            } else if (form.elements[key].type === 'time' && value === '') {
                data[key] = null;
            } else {
                data[key] = value;
            }
        }

        try {
            let response;
            if (action === 'create') {
                response = await fetch(` + "`${API_BASE}/${modelName}`" + `, {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });
            } else if (action === 'edit' && recordId) {
                response = await fetch(` + "`${API_BASE}/${modelName}/${recordId}`" + `, {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });
            } else {
                showError('Invalid action or missing record ID');
                return;
            }

            const result = await response.json();

            if (result.success) {
                window.location.href = ` + "`/${modelName}`" + `;
            } else {
                showError(result.error);
            }
        } catch (error) {
            showError('Failed to save data');
        }
    });
}


async function deleteRecord(modelName, recordId) {
    if (!confirm('Are you sure you want to delete this record?')) {
        return;
    }

    try {
        const response = await fetch(` + "`${API_BASE}/${modelName}/${recordId}`" + `, {
            method: 'DELETE'
        });

        const result = await response.json();

        if (result.success) {
            if (window.location.pathname.includes('/' + recordId)) {
                window.location.href = ` + "`/${modelName}`" + `;
            } else {
                window.location.reload();
            }
        } else {
            showError(result.error);
        }
    } catch (error) {
        showError('Failed to delete record');
    }
}


function showError(message) {
    alert('Error: ' + message);
}


document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search');
    if (searchInput) {
        let searchTimeout;
        searchInput.addEventListener('input', (e) => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                currentSearch = e.target.value;
                currentPage = 1;
                if (typeof modelName !== 'undefined' && typeof columns !== 'undefined') {
                    loadList(modelName, columns, searchable, sortable, window.modelInfo);
                }
            }, 300);
        });
    }
});`
}

