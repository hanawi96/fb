import { auth } from './stores/auth';
import { get } from 'svelte/store';

const API_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

async function request(endpoint, options = {}) {
	const url = `${API_URL}${endpoint}`;
	const authState = get(auth);
	const token = authState?.token || authState;
	
	const config = {
		...options,
		headers: {
			'Content-Type': 'application/json',
			...(token && { 'Authorization': `Bearer ${token}` }),
			...options.headers
		}
	};
	
	const response = await fetch(url, config);
	
	if (!response.ok) {
		const error = await response.json().catch(() => ({ error: 'Unknown error' }));
		throw new Error(error.error || 'Request failed');
	}
	
	return response.json();
}

export const api = {
	// Auth
	getFacebookAuthURL: () => request('/api/auth/facebook/url'),
	facebookCallback: (code) => request('/api/auth/facebook/callback', {
		method: 'POST',
		body: JSON.stringify({ code })
	}),
	saveSelectedPages: (pages) => request('/api/auth/pages/save', {
		method: 'POST',
		body: JSON.stringify({ pages })
	}),
	
	// Pages
	getPages: () => request('/api/pages'),
	deletePage: (id) => request(`/api/pages/${id}`, { method: 'DELETE' }),
	togglePage: (id) => request(`/api/pages/${id}/toggle`, { method: 'PATCH' }),
	
	// Posts
	createPost: (post) => request('/api/posts', {
		method: 'POST',
		body: JSON.stringify(post)
	}),
	getPosts: (limit = 20, offset = 0) => request(`/api/posts?limit=${limit}&offset=${offset}`),
	getPost: (id) => request(`/api/posts/${id}`),
	updatePost: (id, post) => request(`/api/posts/${id}`, {
		method: 'PUT',
		body: JSON.stringify(post)
	}),
	deletePost: (id) => request(`/api/posts/${id}`, { method: 'DELETE' }),
	
	// Schedule
	schedulePost: (data) => request('/api/schedule', {
		method: 'POST',
		body: JSON.stringify(data)
	}),
	getScheduledPosts: (status = '', limit = 50, offset = 0) => 
		request(`/api/schedule?status=${status}&limit=${limit}&offset=${offset}`),
	deleteScheduledPost: (id) => request(`/api/schedule/${id}`, { method: 'DELETE' }),
	retryScheduledPost: (id) => request(`/api/schedule/${id}/retry`, { method: 'POST' }),
	
	// Logs
	getLogs: (limit = 50, offset = 0) => request(`/api/logs?limit=${limit}&offset=${offset}`),
	
	// Upload
	uploadImage: async (file) => {
		const formData = new FormData();
		formData.append('image', file);
		
		const response = await fetch(`${API_URL}/api/upload`, {
			method: 'POST',
			body: formData
		});
		
		if (!response.ok) {
			throw new Error('Upload failed');
		}
		
		return response.json();
	}
};
