<script>
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api';
	import { Bell, X, Check, AlertTriangle, Clock, XCircle } from 'lucide-svelte';

	let notifications = [];
	let unreadCount = 0;
	let showDropdown = false;
	let loading = false;
	let interval;

	onMount(() => {
		loadNotifications();
		// Poll every 30 seconds
		interval = setInterval(loadUnreadCount, 30000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
	});

	async function loadNotifications() {
		loading = true;
		try {
			const result = await api.getNotifications();
			notifications = result.notifications || [];
			unreadCount = result.unread_count || 0;
		} catch (error) {
			console.error('Failed to load notifications:', error);
		} finally {
			loading = false;
		}
	}

	async function loadUnreadCount() {
		try {
			const result = await api.getUnreadCount();
			unreadCount = result.count || 0;
		} catch (error) {
			console.error('Failed to load unread count:', error);
		}
	}

	async function markAsRead(id) {
		try {
			await api.markNotificationRead(id);
			notifications = notifications.map(n => 
				n.id === id ? { ...n, is_read: true } : n
			);
			unreadCount = Math.max(0, unreadCount - 1);
		} catch (error) {
			console.error('Failed to mark as read:', error);
		}
	}

	async function markAllAsRead() {
		try {
			await api.markAllNotificationsRead();
			notifications = notifications.map(n => ({ ...n, is_read: true }));
			unreadCount = 0;
		} catch (error) {
			console.error('Failed to mark all as read:', error);
		}
	}

	async function deleteNotification(id) {
		const original = [...notifications];
		notifications = notifications.filter(n => n.id !== id);
		
		try {
			await api.deleteNotification(id);
		} catch (error) {
			notifications = original;
		}
	}

	function toggleDropdown() {
		showDropdown = !showDropdown;
		if (showDropdown) {
			loadNotifications();
		}
	}

	function closeDropdown() {
		showDropdown = false;
	}

	function getIcon(type) {
		switch (type) {
			case 'rate_limit': return AlertTriangle;
			case 'token_expiring': return Clock;
			case 'post_failed': return XCircle;
			case 'daily_limit': return AlertTriangle;
			default: return Bell;
		}
	}

	function getIconColor(type) {
		switch (type) {
			case 'rate_limit': return 'text-yellow-500';
			case 'token_expiring': return 'text-orange-500';
			case 'post_failed': return 'text-red-500';
			case 'daily_limit': return 'text-yellow-500';
			default: return 'text-blue-500';
		}
	}

	function formatTime(dateStr) {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now - date;
		
		if (diff < 60000) return 'Vừa xong';
		if (diff < 3600000) return `${Math.floor(diff / 60000)} phút trước`;
		if (diff < 86400000) return `${Math.floor(diff / 3600000)} giờ trước`;
		return date.toLocaleDateString('vi-VN');
	}
</script>

<div class="relative">
	<button
		on:click={toggleDropdown}
		class="relative p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
	>
		<Bell size={20} />
		{#if unreadCount > 0}
			<span class="absolute -top-0.5 -right-0.5 w-5 h-5 bg-red-500 text-white text-xs font-medium rounded-full flex items-center justify-center">
				{unreadCount > 9 ? '9+' : unreadCount}
			</span>
		{/if}
	</button>

	{#if showDropdown}
		<!-- Backdrop -->
		<!-- svelte-ignore a11y-click-events-have-key-events -->
		<!-- svelte-ignore a11y-no-static-element-interactions -->
		<div class="fixed inset-0 z-40" on:click={closeDropdown}></div>

		<!-- Dropdown -->
		<div class="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-lg border border-gray-200 z-50 overflow-hidden">
			<!-- Header -->
			<div class="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
				<h3 class="font-semibold text-gray-900">Thông báo</h3>
				{#if unreadCount > 0}
					<button
						on:click={markAllAsRead}
						class="text-xs text-blue-600 hover:text-blue-700 font-medium"
					>
						Đánh dấu tất cả đã đọc
					</button>
				{/if}
			</div>

			<!-- Notifications List -->
			<div class="max-h-96 overflow-y-auto">
				{#if loading}
					<div class="p-8 text-center">
						<div class="w-6 h-6 border-2 border-blue-200 border-t-blue-600 rounded-full animate-spin mx-auto"></div>
					</div>
				{:else if notifications.length === 0}
					<div class="p-8 text-center">
						<Bell size={32} class="text-gray-300 mx-auto mb-2" />
						<p class="text-sm text-gray-500">Không có thông báo</p>
					</div>
				{:else}
					{#each notifications as notification}
						<div 
							class="px-4 py-3 border-b border-gray-50 hover:bg-gray-50 transition-colors
								{notification.is_read ? 'bg-white' : 'bg-blue-50/50'}"
						>
							<div class="flex gap-3">
								<div class="flex-shrink-0 mt-0.5">
									<svelte:component this={getIcon(notification.type)} size={18} class={getIconColor(notification.type)} />
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-900">{notification.title}</p>
									<p class="text-xs text-gray-500 mt-0.5 line-clamp-2">{notification.message}</p>
									<p class="text-xs text-gray-400 mt-1">{formatTime(notification.created_at)}</p>
								</div>
								<div class="flex-shrink-0 flex gap-1">
									{#if !notification.is_read}
										<button
											on:click={() => markAsRead(notification.id)}
											class="p-1 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded transition-colors"
											title="Đánh dấu đã đọc"
										>
											<Check size={14} />
										</button>
									{/if}
									<button
										on:click={() => deleteNotification(notification.id)}
										class="p-1 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
										title="Xóa"
									>
										<X size={14} />
									</button>
								</div>
							</div>
						</div>
					{/each}
				{/if}
			</div>
		</div>
	{/if}
</div>
