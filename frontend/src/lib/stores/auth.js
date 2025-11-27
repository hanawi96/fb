import { writable } from 'svelte/store';
import { browser } from '$app/environment';

function createAuthStore() {
	const { subscribe, set, update } = writable({ token: null, username: null, initialized: false });

	return {
		subscribe,
		init: () => {
			if (browser) {
				const token = localStorage.getItem('token');
				const username = localStorage.getItem('username');
				set({ token, username, initialized: true });
			}
		},
		login: (token, username) => {
			if (browser) {
				localStorage.setItem('token', token);
				localStorage.setItem('username', username);
			}
			set({ token, username, initialized: true });
		},
		logout: () => {
			if (browser) {
				localStorage.removeItem('token');
				localStorage.removeItem('username');
			}
			set({ token: null, username: null, initialized: true });
		},
		getToken: () => {
			if (browser) {
				return localStorage.getItem('token');
			}
			return null;
		}
	};
}

export const auth = createAuthStore();
