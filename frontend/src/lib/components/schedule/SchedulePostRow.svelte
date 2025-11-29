<script>
	import { createEventDispatcher } from 'svelte';
	import { Clock, CheckCircle, XCircle, Loader, Play, RefreshCw, Trash2, MoreVertical, Eye } from 'lucide-svelte';
	
	export let post;
	export let selected = false;
	export let selectable = true;
	
	const dispatch = createEventDispatcher();
	
	let showMenu = false;
	let menuRef;
	
	// Generate consistent color for each page based on ID
	function getPageColor(pageId) {
		const colors = [
			'bg-blue-500', 'bg-green-500', 'bg-purple-500', 'bg-pink-500',
			'bg-yellow-500', 'bg-red-500', 'bg-indigo-500', 'bg-teal-500',
			'bg-orange-500', 'bg-cyan-500'
		];
		let hash = 0;
		const str = String(pageId);
		for (let i = 0; i < str.length; i++) {
			hash = str.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}
	
	function getStatusBadge(status) {
		switch (status) {
			case 'pending': 
				return { color: 'bg-yellow-100 text-yellow-700', icon: Clock, text: 'Chờ đăng' };
			case 'processing': 
				return { color: 'bg-blue-100 text-blue-700', icon: Loader, text: 'Đang đăng' };
			case 'completed': 
				return { color: 'bg-green-100 text-green-700', icon: CheckCircle, text: 'Thành công' };
			case 'failed': 
				return { color: 'bg-red-100 text-red-700', icon: XCircle, text: 'Thất bại' };
			default: 
				return { color: 'bg-gray-100 text-gray-700', icon: Clock, text: status };
		}
	}
	
	import { formatTimeVN } from '$lib/utils/datetime';
	
	function formatDateTime(dateStr) {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const day = String(date.getDate()).padStart(2, '0');
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const hours = String(date.getHours()).padStart(2, '0');
		const minutes = String(date.getMinutes()).padStart(2, '0');
		return `${day}/${month} ${hours}:${minutes}`;
	}
	
	function toggleSelect() {
		if (selectable) {
			dispatch('select', { id: post.id, selected: !selected });
		}
	}
	
	function handleAction(action) {
		showMenu = false;
		dispatch(action, { id: post.id });
	}
	
	function handleClickOutside(event) {
		if (menuRef && !menuRef.contains(event.target)) {
			showMenu = false;
		}
	}
	
	$: badge = getStatusBadge(post.status);
	$: pageColor = getPageColor(post.page_id);
</script>

<svelte:window on:click={handleClickOutside} />

<div class="flex items-center px-3 py-2.5 hover:bg-gray-50 transition-colors border-b border-gray-100 last:border-b-0">
	<!-- Checkbox -->
	<div class="w-6 flex-shrink-0">
		{#if selectable}
			<input
				type="checkbox"
				checked={selected}
				on:change={toggleSelect}
				class="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
			/>
		{/if}
	</div>
	
	<!-- Time with Date -->
	<div class="w-24 flex-shrink-0 text-xs font-medium text-gray-900">
		{formatDateTime(post.scheduled_time)}
	</div>
	
	<!-- Page Info - 35% of remaining space -->
	<div class="flex-[35] min-w-0 flex items-center gap-2 px-2">
		{#if post.page?.profile_picture_url}
			<img 
				src={post.page.profile_picture_url} 
				alt="" 
				class="w-6 h-6 rounded-full flex-shrink-0"
			/>
		{:else}
			<div class="w-6 h-6 rounded-full {pageColor} flex-shrink-0"></div>
		{/if}
		<span class="text-sm text-gray-700 truncate">
			{post.page?.page_name || post.page_id}
		</span>
	</div>
	
	<!-- Content Preview - 45% of remaining space -->
	<div class="flex-[45] min-w-0 px-2">
		<p class="text-sm text-gray-600 truncate">
			{post.post?.content || post.post_id}
		</p>
	</div>
	
	<!-- Account (Người đăng) - 20% of remaining space -->
	<div class="flex-[20] min-w-0 px-1">
		{#if post.account?.fb_user_name}
			<div class="flex items-center gap-1.5 justify-center">
				{#if post.account?.profile_picture_url}
					<img 
						src={post.account.profile_picture_url} 
						alt={post.account.fb_user_name}
						class="w-5 h-5 rounded-full flex-shrink-0"
						title={post.account.fb_user_name}
					/>
				{:else}
					<div class="w-5 h-5 rounded-full bg-blue-500 flex items-center justify-center text-white text-[10px] font-medium flex-shrink-0">
						{post.account.fb_user_name.charAt(0).toUpperCase()}
					</div>
				{/if}
				<span class="text-xs text-gray-700 truncate" title={post.account.fb_user_name}>
					{post.account.fb_user_name}
				</span>
			</div>
		{:else}
			<div class="text-xs text-gray-400 text-center">--</div>
		{/if}
	</div>
	
	<!-- Status Badge -->
	<div class="w-20 flex-shrink-0 flex justify-end">
		<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium whitespace-nowrap {badge.color}">
			<svelte:component this={badge.icon} size={10} />
			{badge.text}
		</span>
	</div>
	
	<!-- Actions Menu -->
	<div class="relative flex-shrink-0" bind:this={menuRef}>
		<button
			type="button"
			on:click|stopPropagation={() => showMenu = !showMenu}
			class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
		>
			<MoreVertical size={16} />
		</button>
		
		{#if showMenu}
			<div class="absolute right-0 top-full mt-1 w-40 bg-white border border-gray-200 rounded-lg shadow-lg z-50 py-1">
				<button
					type="button"
					on:click={() => handleAction('view')}
					class="w-full flex items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
				>
					<Eye size={14} />
					Xem chi tiết
				</button>
				
				{#if post.status === 'pending'}
					<button
						type="button"
						on:click={() => handleAction('test')}
						class="w-full flex items-center gap-2 px-3 py-2 text-sm text-green-600 hover:bg-green-50"
					>
						<Play size={14} />
						Đăng ngay
					</button>
				{/if}
				
				{#if post.status === 'failed'}
					<button
						type="button"
						on:click={() => handleAction('retry')}
						class="w-full flex items-center gap-2 px-3 py-2 text-sm text-blue-600 hover:bg-blue-50"
					>
						<RefreshCw size={14} />
						Thử lại
					</button>
				{/if}
				
				<button
					type="button"
					on:click={() => handleAction('delete')}
					class="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-50"
				>
					<Trash2 size={14} />
					Xóa
				</button>
			</div>
		{/if}
	</div>
</div>
