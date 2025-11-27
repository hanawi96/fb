import { writable } from 'svelte/store';

function createToastStore() {
	const { subscribe, set } = writable(null);
	
	return {
		subscribe,
		show: (message, type = 'success', duration = 3000) => {
			set({ message, type, duration });
		},
		hide: () => {
			set(null);
		}
	};
}

export const toast = createToastStore();
