<script>
	import { onMount } from 'svelte';
	import { Search, X } from 'lucide-svelte';
	import { emojiCategories, saveRecentEmoji, getRecentEmojis, searchEmojis } from '$lib/utils/emojiData';
	
	export let show = false;
	export let onSelect = (emoji) => {};
	export let onClose = () => {};
	export let buttonElement = null;
	
	let searchQuery = '';
	let activeCategory = 'recent';
	let recentEmojis = [];
	let searchResults = [];
	let pickerElement;
	
	// Calculate position
	function updatePosition() {
		if (!pickerElement || !buttonElement) return;
		
		const rect = buttonElement.getBoundingClientRect();
		const pickerHeight = 400; // Approximate height
		const spaceBelow = window.innerHeight - rect.bottom;
		const spaceAbove = rect.top;
		
		// Position below or above button
		if (spaceBelow >= pickerHeight || spaceBelow > spaceAbove) {
			// Show below
			pickerElement.style.top = `${rect.bottom + 8}px`;
			pickerElement.style.bottom = 'auto';
		} else {
			// Show above
			pickerElement.style.bottom = `${window.innerHeight - rect.top + 8}px`;
			pickerElement.style.top = 'auto';
		}
		
		// Horizontal position
		pickerElement.style.left = `${rect.left}px`;
	}
	
	// Load recent emojis
	onMount(() => {
		recentEmojis = getRecentEmojis();
		if (recentEmojis.length > 0) {
			emojiCategories[0].emojis = recentEmojis;
		} else {
			activeCategory = 'smileys'; // Nếu chưa có recent, chọn smileys
		}
		
		// Update position when shown
		if (show) {
			setTimeout(updatePosition, 0);
		}
	});
	
	// Update position when show changes
	$: if (show && pickerElement) {
		setTimeout(updatePosition, 0);
		window.addEventListener('scroll', updatePosition);
		window.addEventListener('resize', updatePosition);
	} else {
		window.removeEventListener('scroll', updatePosition);
		window.removeEventListener('resize', updatePosition);
	}
	
	// Handle emoji click
	function handleEmojiClick(emoji) {
		onSelect(emoji);
		saveRecentEmoji(emoji);
		recentEmojis = getRecentEmojis();
		emojiCategories[0].emojis = recentEmojis;
		
		// Không đóng picker để có thể chọn nhiều emoji liên tiếp
		// onClose();
	}
	
	// Search với debounce
	let searchTimeout;
	$: {
		clearTimeout(searchTimeout);
		if (searchQuery.trim()) {
			searchTimeout = setTimeout(() => {
				searchResults = searchEmojis(searchQuery);
			}, 150);
		} else {
			searchResults = [];
		}
	}
	
	// Get current emojis to display
	$: currentEmojis = searchQuery.trim() 
		? searchResults 
		: emojiCategories.find(c => c.id === activeCategory)?.emojis || [];
</script>

{#if show}
	<!-- Backdrop -->
	<div 
		class="fixed inset-0 bg-black/5 backdrop-blur-[1px]"
		style="z-index: 9998;"
		on:click={onClose}
	></div>
	
	<!-- Picker -->
	<div 
		bind:this={pickerElement}
		class="fixed w-80 bg-white rounded-xl shadow-2xl border border-gray-200 emoji-picker"
		style="z-index: 9999;"
	>
		<!-- Search Bar -->
		<div class="p-3 border-b border-gray-100">
			<div class="relative">
				<Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Tìm emoji..."
					class="w-full pl-9 pr-8 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					autocomplete="off"
				/>
				{#if searchQuery}
					<button
						on:click={() => searchQuery = ''}
						class="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-gray-100 rounded"
					>
						<X size={14} class="text-gray-400" />
					</button>
				{/if}
			</div>
		</div>
		
		<!-- Categories -->
		{#if !searchQuery}
			<div class="flex gap-1 px-3 py-2 border-b border-gray-100 overflow-x-auto scrollbar-hide">
				{#each emojiCategories as category}
					<button
						on:click={() => activeCategory = category.id}
						class="flex-shrink-0 w-8 h-8 flex items-center justify-center rounded-lg transition-colors text-lg hover:bg-gray-100 {activeCategory === category.id ? 'bg-blue-50 ring-2 ring-blue-500' : ''}"
						title={category.name}
					>
						{category.icon}
					</button>
				{/each}
			</div>
		{/if}
		
		<!-- Emoji Grid -->
		<div class="p-2 h-64 overflow-y-auto emoji-grid">
			{#if currentEmojis.length === 0}
				<div class="flex items-center justify-center h-full text-sm text-gray-400">
					{searchQuery ? 'Không tìm thấy emoji' : 'Chưa có emoji gần đây'}
				</div>
			{:else}
				<div class="grid grid-cols-8 gap-1">
					{#each currentEmojis as emoji}
						<button
							on:click={() => handleEmojiClick(emoji)}
							class="w-9 h-9 flex items-center justify-center text-2xl rounded-lg hover:bg-gray-100 active:scale-95 transition-all emoji-btn"
							title={emoji}
						>
							{emoji}
						</button>
					{/each}
				</div>
			{/if}
		</div>
		
		<!-- Footer -->
		<div class="px-3 py-2 border-t border-gray-100 flex items-center justify-between text-xs text-gray-500">
			<span>{currentEmojis.length} emojis</span>
			<button
				on:click={onClose}
				class="text-blue-600 hover:text-blue-700 font-medium"
			>
				Đóng
			</button>
		</div>
	</div>
{/if}

<style>
	.emoji-picker {
		animation: slideUp 0.15s ease-out;
	}
	
	@keyframes slideUp {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	
	.emoji-grid {
		scrollbar-width: thin;
		scrollbar-color: #cbd5e1 transparent;
	}
	
	.emoji-grid::-webkit-scrollbar {
		width: 6px;
	}
	
	.emoji-grid::-webkit-scrollbar-track {
		background: transparent;
	}
	
	.emoji-grid::-webkit-scrollbar-thumb {
		background: #cbd5e1;
		border-radius: 3px;
	}
	
	.emoji-grid::-webkit-scrollbar-thumb:hover {
		background: #94a3b8;
	}
	
	.emoji-btn {
		user-select: none;
		-webkit-tap-highlight-color: transparent;
	}
	
	.scrollbar-hide {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
	
	.scrollbar-hide::-webkit-scrollbar {
		display: none;
	}
</style>
