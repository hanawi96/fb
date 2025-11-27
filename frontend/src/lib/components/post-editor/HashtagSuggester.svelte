<script>
	import { onMount } from 'svelte';
	import { Hash, Search, X, Check } from 'lucide-svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	
	export let show = false;
	export let onSelect = (hashtags) => {};
	export let onClose = () => {};
	export let buttonElement = null;
	
	let suggestElement;
	let activeTab = 'new'; // 'new' or 'saved'
	let searchQuery = '';
	let searchResults = [];
	let savedHashtags = [];
	let selectedHashtags = [];
	let searching = false;
	let loading = false;
	
	// For saved tab - create new hashtag set
	let newSetName = '';
	let newSetHashtags = '';
	let newSetTags = [];
	let randomCount = 0;
	
	// Calculate position
	function updatePosition() {
		if (!suggestElement || !buttonElement) return;
		
		const rect = buttonElement.getBoundingClientRect();
		const panelHeight = 600;
		const spaceBelow = window.innerHeight - rect.bottom;
		const spaceAbove = rect.top;
		
		if (spaceBelow >= panelHeight || spaceBelow > spaceAbove) {
			suggestElement.style.top = `${rect.bottom + 8}px`;
			suggestElement.style.bottom = 'auto';
		} else {
			suggestElement.style.bottom = `${window.innerHeight - rect.top + 8}px`;
			suggestElement.style.top = 'auto';
		}
		
		suggestElement.style.left = `${rect.left}px`;
	}
	
	$: if (show && suggestElement) {
		setTimeout(updatePosition, 0);
	}
	
	// Load saved hashtags on mount
	onMount(async () => {
		if (show) {
			await loadSavedHashtags();
		}
	});
	
	async function loadSavedHashtags() {
		try {
			loading = true;
			const result = await api.getSavedHashtags();
			savedHashtags = result.map(h => ({
				id: h.id,
				name: h.name,
				media_count: h.media_count
			}));
		} catch (error) {
			console.error('Failed to load saved hashtags:', error);
		} finally {
			loading = false;
		}
	}
	
	// Search hashtags
	let searchTimeout;
	async function handleSearch() {
		if (!searchQuery.trim()) {
			searchResults = [];
			return;
		}
		
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(async () => {
			try {
				searching = true;
				const result = await api.searchHashtags(searchQuery);
				searchResults = result.data || [];
			} catch (error) {
				console.error('Search failed:', error);
				toast.show('Lỗi tìm kiếm hashtag', 'error');
			} finally {
				searching = false;
			}
		}, 500);
	}
	
	// Toggle hashtag selection
	function toggleHashtag(hashtag) {
		const index = selectedHashtags.findIndex(h => h.name === hashtag.name);
		if (index >= 0) {
			selectedHashtags = selectedHashtags.filter((_, i) => i !== index);
		} else {
			selectedHashtags = [...selectedHashtags, hashtag];
		}
	}
	
	// Check if hashtag is selected
	function isSelected(hashtag) {
		return selectedHashtags.some(h => h.name === hashtag.name);
	}
	
	// Remove selected hashtag
	function removeSelected(hashtag) {
		selectedHashtags = selectedHashtags.filter(h => h.name !== hashtag.name);
	}
	
	// Save hashtags
	async function handleSaveHashtags() {
		if (selectedHashtags.length === 0) {
			toast.show('Chọn ít nhất 1 hashtag', 'warning');
			return;
		}
		
		try {
			await api.saveHashtags(selectedHashtags);
			toast.show('Đã lưu hashtags', 'success');
			await loadSavedHashtags();
		} catch (error) {
			toast.show('Lỗi lưu hashtags', 'error');
		}
	}
	
	// Confirm and insert hashtags
	function handleConfirm() {
		if (selectedHashtags.length === 0) {
			toast.show('Chọn ít nhất 1 hashtag', 'warning');
			return;
		}
		
		const hashtagsText = selectedHashtags.map(h => h.name).join(' ');
		onSelect(hashtagsText);
		onClose();
	}
	
	// Format number
	function formatNumber(num) {
		if (num >= 1000000) {
			return (num / 1000000).toFixed(1) + 'M';
		} else if (num >= 1000) {
			return (num / 1000).toFixed(1) + 'K';
		}
		return num.toString();
	}
	
	// Handle hashtag input in saved tab
	function handleHashtagInput(e) {
		if (e.key === 'Enter') {
			e.preventDefault();
			const value = newSetHashtags.trim();
			if (value) {
				// Split by spaces and filter out empty strings
				const tags = value.split(/\s+/).filter(t => t.length > 0);
				tags.forEach(tag => {
					// Remove # if present and add to tags
					const cleanTag = tag.startsWith('#') ? tag.slice(1) : tag;
					if (cleanTag && !newSetTags.includes(cleanTag)) {
						newSetTags = [...newSetTags, cleanTag];
					}
				});
				newSetHashtags = '';
			}
		}
	}
	
	// Remove tag from new set
	function removeNewTag(tag) {
		newSetTags = newSetTags.filter(t => t !== tag);
	}
	
	// Save new hashtag set and apply
	async function handleSaveNewSet() {
		// Collect all hashtags to save and apply
		let allHashtags = [...selectedHashtags];
		
		// Add new tags if any
		if (newSetTags.length > 0) {
			const newTags = newSetTags.map(tag => ({
				name: tag,
				media_count: 0
			}));
			allHashtags = [...allHashtags, ...newTags];
			
			// Save new tags to database
			try {
				await api.saveHashtags(newTags);
			} catch (error) {
				console.error('Failed to save new tags:', error);
			}
		}
		
		if (allHashtags.length === 0) {
			toast.show('Vui lòng chọn hoặc tạo hashtag', 'warning');
			return;
		}
		
		// Apply hashtags (with random selection if specified)
		let tagsToApply = allHashtags;
		if (randomCount > 0 && randomCount < allHashtags.length) {
			// Randomly select tags
			const shuffled = [...allHashtags].sort(() => Math.random() - 0.5);
			tagsToApply = shuffled.slice(0, randomCount);
		}
		
		const hashtagsText = tagsToApply.map(h => '#' + h.name).join(' ');
		onSelect(hashtagsText);
		
		// Reset form
		newSetName = '';
		newSetHashtags = '';
		newSetTags = [];
		randomCount = 0;
		selectedHashtags = [];
		
		onClose();
	}
	
	$: displayHashtags = activeTab === 'new' ? searchResults : savedHashtags;
</script>

{#if show}
	<!-- Backdrop -->
	<div 
		class="fixed inset-0 bg-black/5 backdrop-blur-[1px]"
		style="z-index: 9998;"
		on:click={onClose}
	></div>
	
	<!-- Panel -->
	<div 
		bind:this={suggestElement}
		class="fixed w-[600px] bg-white rounded-xl shadow-2xl border border-gray-200 hashtag-panel"
		style="z-index: 9999;"
	>
		<!-- Header with Tabs -->
		<div class="border-b border-gray-200">
			<div class="flex">
				<button
					on:click={() => activeTab = 'new'}
					class="flex-1 px-6 py-4 text-sm font-semibold transition-colors relative {activeTab === 'new' ? 'text-blue-600' : 'text-gray-600 hover:text-gray-900'}"
				>
					Hashtag mới
					{#if activeTab === 'new'}
						<div class="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
					{/if}
				</button>
				<button
					on:click={() => activeTab = 'saved'}
					class="flex-1 px-6 py-4 text-sm font-semibold transition-colors relative {activeTab === 'saved' ? 'text-blue-600' : 'text-gray-600 hover:text-gray-900'}"
				>
					Hashtag đã lưu ({savedHashtags.length})
					{#if activeTab === 'saved'}
						<div class="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
					{/if}
				</button>
			</div>
		</div>
		
		<!-- Content Area -->
		{#if activeTab === 'new'}
			<!-- Search Box for new tab -->
			<div class="p-3 border-b border-gray-100 bg-gray-50">
				<div class="text-xs text-gray-500 mb-1.5 flex items-center gap-1">
					<span>Nhập từ khóa để tìm kiếm hashtag</span>
					<div class="w-3.5 h-3.5 rounded-full border border-gray-400 flex items-center justify-center text-[10px] text-gray-400">?</div>
				</div>
				<div class="flex gap-2">
					<div class="flex-1 relative">
						<input
							type="text"
							bind:value={searchQuery}
							on:input={handleSearch}
							placeholder="Nhập từ khóa..."
							class="w-full px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						/>
					</div>
					<button
						on:click={handleSearch}
						disabled={!searchQuery.trim() || searching}
						class="px-4 py-1.5 text-sm bg-teal-500 text-white font-medium rounded-lg hover:bg-teal-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{searching ? 'Đang tìm...' : 'Tìm kiếm'}
					</button>
				</div>
			</div>
			
			<!-- Search Results List -->
			<div class="max-h-80 overflow-y-auto">
				{#if searching}
					<div class="p-8 text-center text-gray-500">
						<div class="w-8 h-8 border-2 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
						Đang tìm...
					</div>
				{:else if searchResults.length === 0}
					<div class="p-8 text-center text-gray-400">
						{searchQuery ? 'Không tìm thấy hashtag' : 'Nhập từ khóa để tìm kiếm'}
					</div>
				{:else}
					<div class="divide-y divide-gray-100">
						{#each searchResults as hashtag}
							<button
								on:click={() => toggleHashtag(hashtag)}
								class="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50 transition-colors text-left"
							>
								<div class="flex-1">
									<div class="text-blue-600 font-medium">#{hashtag.name}</div>
									<div class="text-xs text-gray-500 mt-0.5">
										({formatNumber(hashtag.media_count)} Lượt xem)
									</div>
								</div>
								<div class="w-5 h-5 rounded border-2 flex items-center justify-center transition-colors {isSelected(hashtag) ? 'bg-blue-600 border-blue-600' : 'border-gray-300'}">
									{#if isSelected(hashtag)}
										<Check size={14} class="text-white" />
									{/if}
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{:else}
			<!-- Saved Hashtags Tab -->
			<div class="p-4">
				<div class="text-sm text-gray-700 font-medium mb-3">
					Tạo mới hoặc chọn từ các hashtag đã lưu
				</div>
				
				<!-- Create New Hashtag Set -->
				<div class="mb-4">
					<div class="relative mb-3">
						<select class="w-full px-3 py-2 text-sm bg-gray-50 border border-gray-300 rounded-lg appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-700">
							<option>Tạo hashtag mới</option>
						</select>
						<div class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-gray-400">
							<svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
								<path d="M4 6l4 4 4-4" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round"/>
							</svg>
						</div>
					</div>
					
					<input
						type="text"
						bind:value={newSetName}
						placeholder="Tên"
						class="w-full px-3 py-2 text-sm bg-gray-50 border border-gray-300 rounded-lg mb-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
					
					<!-- Hashtag Input with Pills -->
					<div class="w-full min-h-[80px] px-3 py-2 text-sm bg-gray-50 border border-gray-300 rounded-lg focus-within:ring-2 focus-within:ring-blue-500">
						<div class="flex flex-wrap gap-1.5 items-center">
							{#each newSetTags as tag}
								<div class="inline-flex items-center gap-1 px-2 py-0.5 bg-gray-200 rounded-full text-xs">
									<span class="text-gray-700">#{tag}</span>
									<button
										on:click={() => removeNewTag(tag)}
										class="text-gray-500 hover:text-gray-700"
									>
										<X size={12} />
									</button>
								</div>
							{/each}
							<input
								type="text"
								bind:value={newSetHashtags}
								on:keydown={handleHashtagInput}
								placeholder={newSetTags.length === 0 ? "Hashtag" : ""}
								class="flex-1 min-w-[100px] bg-transparent border-none outline-none text-sm"
							/>
						</div>
					</div>
				</div>
				
				<!-- Random Selection -->
				<div class="flex items-center gap-3 mb-4">
					<span class="text-sm text-gray-600">Ngẫu nhiên chọn</span>
					<input
						type="number"
						bind:value={randomCount}
						min="0"
						max={newSetTags.length}
						class="w-20 px-2 py-1 text-sm text-center bg-gray-50 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
					<span class="text-sm text-gray-600">hashtag cho mỗi bài đăng</span>
				</div>
				
				<!-- Saved Hashtags List -->
				{#if loading}
					<div class="py-8 text-center text-gray-500">
						<div class="w-6 h-6 border-2 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
						<div class="text-xs">Đang tải...</div>
					</div>
				{:else if savedHashtags.length === 0}
					<div class="py-8 text-center text-gray-400 text-sm">
						Chưa có hashtag đã lưu
					</div>
				{:else}
					<div class="max-h-48 overflow-y-auto space-y-2">
						{#each savedHashtags as hashtag}
							<button
								on:click={() => toggleHashtag(hashtag)}
								class="w-full px-3 py-2 flex items-center justify-between bg-gray-50 hover:bg-gray-100 rounded-lg transition-colors text-left border border-gray-200"
							>
								<div class="flex-1">
									<div class="text-sm text-blue-600 font-medium">#{hashtag.name}</div>
									<div class="text-xs text-gray-500 mt-0.5">
										({formatNumber(hashtag.media_count)} Lượt xem)
									</div>
								</div>
								<div class="w-4 h-4 rounded border-2 flex items-center justify-center transition-colors {isSelected(hashtag) ? 'bg-blue-600 border-blue-600' : 'border-gray-300'}">
									{#if isSelected(hashtag)}
										<Check size={12} class="text-white" />
									{/if}
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
		
		<!-- Selected Hashtags Pills -->
		{#if selectedHashtags.length > 0}
			<div class="px-4 py-3 border-t border-gray-100 bg-gray-50">
				<div class="flex flex-wrap gap-2">
					{#each selectedHashtags as hashtag}
						<div class="inline-flex items-center gap-2 px-3 py-1.5 bg-white border border-gray-300 rounded-full text-sm">
							<span class="text-blue-600">#{hashtag.name}</span>
							<button
								on:click={() => removeSelected(hashtag)}
								class="text-gray-400 hover:text-gray-600"
							>
								<X size={14} />
							</button>
						</div>
					{/each}
				</div>
			</div>
		{/if}
		
		<!-- Footer Actions -->
		<div class="p-3 border-t border-gray-200">
			{#if activeTab === 'new'}
				<div class="flex items-center justify-between gap-2">
					<button
						on:click={handleSaveHashtags}
						disabled={selectedHashtags.length === 0}
						class="px-4 py-1.5 text-sm text-gray-700 bg-gray-100 font-medium rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Lưu hashtag
					</button>
					<button
						on:click={handleConfirm}
						disabled={selectedHashtags.length === 0}
						class="px-4 py-1.5 text-sm bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Xác nhận
					</button>
				</div>
			{:else}
				<button
					on:click={handleSaveNewSet}
					disabled={newSetTags.length === 0 && selectedHashtags.length === 0}
					class="w-full px-4 py-2 text-sm bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					Lưu & Áp dụng
				</button>
			{/if}
		</div>
	</div>
{/if}

<style>
	.hashtag-panel {
		animation: slideUp 0.2s ease-out;
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
</style>
