// API Configuration
const API_BASE_URL = 'http://localhost:3000/api';

// DOM Elements
const addBlogForm = document.getElementById('addBlogForm');
const editBlogForm = document.getElementById('editBlogForm');
const blogItemsContainer = document.getElementById('blogItems');
const editModal = document.getElementById('editModal');
const closeModal = document.querySelector('.close');

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    loadBlogs();

    addBlogForm.addEventListener('submit', handleAddBlog);
    editBlogForm.addEventListener('submit', handleEditBlog);
    closeModal.addEventListener('click', () => {
        editModal.style.display = 'none';
    });

    window.addEventListener('click', (e) => {
        if (e.target === editModal) {
            editModal.style.display = 'none';
        }
    });
});

// Load all blog posts
async function loadBlogs() {
    try {
        const response = await fetch(`${API_BASE_URL}/blogs`);
        if (!response.ok) throw new Error('Failed to fetch blogs');

        const blogs = await response.json();
        renderBlogs(blogs);
    } catch (error) {
        console.error('Error loading blogs:', error);
        blogItemsContainer.innerHTML = '<div class="empty-state">블로그 포스트를 불러오지 못했습니다. 백엔드 서버가 실행 중인지 확인하세요.</div>';
    }
}

// Render blog posts
function renderBlogs(blogs) {
    if (blogs.length === 0) {
        blogItemsContainer.innerHTML = '<div class="empty-state">아직 블로그 포스트가 없습니다. 첫 포스트를 작성해보세요!</div>';
        return;
    }

    blogItemsContainer.innerHTML = blogs.map(blog => `
        <div class="blog-item">
            <h3>${escapeHtml(blog.title)}</h3>
            ${blog.tags ? `<div class="blog-tags">${renderTags(blog.tags)}</div>` : ''}
            <div class="blog-content">${escapeHtml(blog.content)}</div>
            <div class="todo-meta">
                작성일: ${new Date(blog.created_at).toLocaleString('ko-KR')}
                ${blog.updated_at !== blog.created_at ? ` • 수정일: ${new Date(blog.updated_at).toLocaleString('ko-KR')}` : ''}
            </div>
            <div class="todo-actions">
                <button class="btn-edit" onclick="openEditModal('${blog.id}')">수정</button>
                <button class="btn-delete" onclick="deleteBlog('${blog.id}')">삭제</button>
            </div>
        </div>
    `).join('');
}

// Render tags
function renderTags(tagsString) {
    if (!tagsString || tagsString.trim() === '') return '';

    const tags = tagsString.split(',').map(tag => tag.trim()).filter(tag => tag);
    return tags.map(tag => `<span class="tag">${escapeHtml(tag)}</span>`).join('');
}

// Handle add blog
async function handleAddBlog(e) {
    e.preventDefault();

    const formData = new FormData(e.target);
    const title = formData.get('title');
    const content = formData.get('content');
    const tags = formData.get('tags') || '';

    try {
        const response = await fetch(`${API_BASE_URL}/blogs`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, content, tags }),
        });

        if (!response.ok) throw new Error('Failed to create blog post');

        e.target.reset();
        await loadBlogs();

        // Scroll to top to see the new post
        window.scrollTo({ top: 0, behavior: 'smooth' });
    } catch (error) {
        console.error('Error creating blog post:', error);
        alert('블로그 포스트 작성에 실패했습니다.');
    }
}

// Open edit modal
async function openEditModal(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/blogs/${id}`);
        if (!response.ok) throw new Error('Failed to fetch blog post');

        const blog = await response.json();

        document.getElementById('editId').value = blog.id;
        document.getElementById('editTitle').value = blog.title;
        document.getElementById('editContent').value = blog.content || '';
        document.getElementById('editTags').value = blog.tags || '';

        editModal.style.display = 'block';
    } catch (error) {
        console.error('Error fetching blog post:', error);
        alert('블로그 포스트를 불러오지 못했습니다.');
    }
}

// Handle edit blog
async function handleEditBlog(e) {
    e.preventDefault();

    const id = document.getElementById('editId').value;
    const formData = new FormData(e.target);
    const title = formData.get('title');
    const content = formData.get('content');
    const tags = formData.get('tags') || '';

    try {
        const response = await fetch(`${API_BASE_URL}/blogs/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, content, tags }),
        });

        if (!response.ok) throw new Error('Failed to update blog post');

        editModal.style.display = 'none';
        await loadBlogs();
    } catch (error) {
        console.error('Error updating blog post:', error);
        alert('블로그 포스트 수정에 실패했습니다.');
    }
}

// Delete blog
async function deleteBlog(id) {
    if (!confirm('이 블로그 포스트를 삭제하시겠습니까?')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/blogs/${id}`, {
            method: 'DELETE',
        });

        if (!response.ok) throw new Error('Failed to delete blog post');

        await loadBlogs();
    } catch (error) {
        console.error('Error deleting blog post:', error);
        alert('블로그 포스트 삭제에 실패했습니다.');
    }
}

// Utility: Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
