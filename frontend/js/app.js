// API Configuration
const API_BASE_URL = 'http://localhost:3000/api';

// DOM Elements
const addTodoForm = document.getElementById('addTodoForm');
const editTodoForm = document.getElementById('editTodoForm');
const todoItemsContainer = document.getElementById('todoItems');
const editModal = document.getElementById('editModal');
const closeModal = document.querySelector('.close');

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    loadTodos();

    addTodoForm.addEventListener('submit', handleAddTodo);
    editTodoForm.addEventListener('submit', handleEditTodo);
    closeModal.addEventListener('click', () => {
        editModal.style.display = 'none';
    });

    window.addEventListener('click', (e) => {
        if (e.target === editModal) {
            editModal.style.display = 'none';
        }
    });
});

// Load all todos
async function loadTodos() {
    try {
        const response = await fetch(`${API_BASE_URL}/todos`);
        if (!response.ok) throw new Error('Failed to fetch todos');

        const todos = await response.json();
        renderTodos(todos);
    } catch (error) {
        console.error('Error loading todos:', error);
        todoItemsContainer.innerHTML = '<div class="empty-state">Failed to load todos. Make sure the backend is running.</div>';
    }
}

// Render todos
function renderTodos(todos) {
    if (todos.length === 0) {
        todoItemsContainer.innerHTML = '<div class="empty-state">No todos yet. Add one above!</div>';
        return;
    }

    todoItemsContainer.innerHTML = todos.map(todo => `
        <div class="todo-item ${todo.completed ? 'completed' : ''}">
            <h3>${escapeHtml(todo.title)}</h3>
            <p>${escapeHtml(todo.content || 'No content')}</p>
            <div class="todo-meta">
                Created: ${new Date(todo.created_at).toLocaleString()}
                ${todo.completed ? ' • ✓ Completed' : ''}
            </div>
            <div class="todo-actions">
                <button class="btn-edit" onclick="openEditModal('${todo.id}')">Edit</button>
                <button class="btn-delete" onclick="deleteTodo('${todo.id}')">Delete</button>
            </div>
        </div>
    `).join('');
}

// Handle add todo
async function handleAddTodo(e) {
    e.preventDefault();

    const formData = new FormData(e.target);
    const title = formData.get('title');
    const content = formData.get('content');

    try {
        const response = await fetch(`${API_BASE_URL}/todos`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, content }),
        });

        if (!response.ok) throw new Error('Failed to create todo');

        e.target.reset();
        await loadTodos();
    } catch (error) {
        console.error('Error creating todo:', error);
        alert('Failed to create todo');
    }
}

// Open edit modal
async function openEditModal(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/todos/${id}`);
        if (!response.ok) throw new Error('Failed to fetch todo');

        const todo = await response.json();

        document.getElementById('editId').value = todo.id;
        document.getElementById('editTitle').value = todo.title;
        document.getElementById('editContent').value = todo.content || '';
        document.getElementById('editCompleted').checked = todo.completed;

        editModal.style.display = 'block';
    } catch (error) {
        console.error('Error fetching todo:', error);
        alert('Failed to load todo');
    }
}

// Handle edit todo
async function handleEditTodo(e) {
    e.preventDefault();

    const id = document.getElementById('editId').value;
    const formData = new FormData(e.target);
    const title = formData.get('title');
    const content = formData.get('content');
    const completed = document.getElementById('editCompleted').checked;

    try {
        const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, content, completed }),
        });

        if (!response.ok) throw new Error('Failed to update todo');

        editModal.style.display = 'none';
        await loadTodos();
    } catch (error) {
        console.error('Error updating todo:', error);
        alert('Failed to update todo');
    }
}

// Delete todo
async function deleteTodo(id) {
    if (!confirm('Are you sure you want to delete this todo?')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
            method: 'DELETE',
        });

        if (!response.ok) throw new Error('Failed to delete todo');

        await loadTodos();
    } catch (error) {
        console.error('Error deleting todo:', error);
        alert('Failed to delete todo');
    }
}

// Utility: Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
