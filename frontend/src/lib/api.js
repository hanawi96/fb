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
	
	try {
		const response = await fetch(url, config);
		
		if (!response.ok) {
			const error = await response.json().catch(() => ({ error: 'Request failed' }));
			throw new Error(error.error || `HTTP ${response.status}: ${response.statusText}`);
		}
		
		return response.json();
	} catch (error) {
		if (error.name === 'TypeError' && error.message === 'Failed to fetch') {
			throw new Error('Không thể kết nối đến server. Vui lòng kiểm tra lại.');
		}
		throw error;
	}
}

export const api = {
	// Auth
	getFacebookAuthURL: () => request('/api/auth/facebook/url'),
	facebookCallback: (code) => request('/api/auth/facebook/callback', {
		method: 'POST',
		body: JSON.stringify({ code })
	}),
	saveSelectedPages: (pages, accountId) => request('/api/auth/pages/save', {
		method: 'POST',
		body: JSON.stringify({ pages, account_id: accountId })
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
	publishPost: (post) => request('/api/posts/publish', {
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
	},
	
	// Hashtags
	searchHashtags: (query) => request(`/api/hashtags/search?q=${encodeURIComponent(query)}`),
	getSavedHashtags: () => request('/api/hashtags/saved'),
	saveHashtags: (data) => request('/api/hashtags/saved', {
		method: 'POST',
		body: JSON.stringify(data)
	}),
	deleteSavedHashtag: (id) => request(`/api/hashtags/saved?id=${id}`, { method: 'DELETE' }),

	// Facebook Accounts (Multi-Account)
	getAccounts: () => request('/api/accounts'),
	getAccount: (id) => request(`/api/accounts/${id}`),
	createAccount: (data) => request('/api/accounts', {
		method: 'POST',
		body: JSON.stringify(data)
	}),
	updateAccount: (id, data) => request(`/api/accounts/${id}`, {
		method: 'PUT',
		body: JSON.stringify(data)
	}),
	deleteAccount: (id) => request(`/api/accounts/${id}`, { method: 'DELETE' }),
	getAccountPages: (id) => request(`/api/accounts/${id}/pages`),
	refreshAccountToken: (id, token) => request(`/api/accounts/${id}/refresh`, {
		method: 'POST',
		body: JSON.stringify({ access_token: token })
	}),

	// Page Assignments
	getPageAssignments: (pageId) => request(`/api/pages/${pageId}/assignments`),
	assignPageToAccount: (pageId, accountId, isPrimary = true) => request(`/api/pages/${pageId}/assign`, {
		method: 'POST',
		body: JSON.stringify({ account_id: accountId, is_primary: isPrimary })
	}),
	unassignPage: (pageId, accountId) => request(`/api/pages/${pageId}/assign/${accountId}`, { method: 'DELETE' }),
	setPrimaryAccount: (pageId, accountId) => request(`/api/pages/${pageId}/primary`, {
		method: 'PUT',
		body: JSON.stringify({ account_id: accountId })
	}),
	getUnassignedPages: () => request('/api/pages/unassigned'),

	// Time Slots
	getPageTimeSlots: (pageId) => request(`/api/pages/${pageId}/timeslots`),
	createTimeSlot: (pageId, data) => request(`/api/pages/${pageId}/timeslots`, {
		method: 'POST',
		body: JSON.stringify(data)
	}),
	updateTimeSlot: (id, data) => request(`/api/timeslots/${id}`, {
		method: 'PUT',
		body: JSON.stringify(data)
	}),
	deleteTimeSlot: (id) => request(`/api/timeslots/${id}`, { method: 'DELETE' }),

	// Schedule Preview
	previewSchedule: (postId, pageIds, preferredDate) => request('/api/schedule/preview', {
		method: 'POST',
		body: JSON.stringify({ post_id: postId, page_ids: pageIds, preferred_date: preferredDate })
	}),
	scheduleWithPreview: (postId, pageIds, preferredDate, confirm = false) => request('/api/schedule/smart', {
		method: 'POST',
		body: JSON.stringify({ post_id: postId, page_ids: pageIds, preferred_date: preferredDate, confirm })
	}),
	getScheduleStats: (date) => request(`/api/schedule/stats?date=${date || ''}`),

	// Notifications
	getNotifications: (unreadOnly = false) => request(`/api/notifications?unread=${unreadOnly}`),
	getUnreadCount: () => request('/api/notifications/count'),
	markNotificationRead: (id) => request(`/api/notifications/${id}/read`, { method: 'PUT' }),
	markAllNotificationsRead: () => request('/api/notifications/read-all', { method: 'PUT' }),
	deleteNotification: (id) => request(`/api/notifications/${id}`, { method: 'DELETE' })
};
