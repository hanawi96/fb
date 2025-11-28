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
	let savedHashtagSets = [];
	let selectedHashtags = [];
	let topHashtags = [];
	let searching = false;
	let loading = false;
	
	// For saved tab - create new hashtag set
	let showCreateForm = false;
	let selectedSetId = '';
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
			savedHashtagSets = result || [];
			
			// Extract top 10 most used hashtags
			const hashtagCount = {};
			savedHashtagSets.forEach(set => {
				const tags = set.hashtags.split(' ').filter(h => h);
				tags.forEach(tag => {
					const cleanTag = tag.replace('#', '');
					hashtagCount[cleanTag] = (hashtagCount[cleanTag] || 0) + 1;
				});
			});
			
			topHashtags = Object.entries(hashtagCount)
				.sort((a, b) => b[1] - a[1])
				.slice(0, 10)
				.map(([name]) => ({ name, media_count: 0 }));
		} catch (error) {
			console.error('Failed to load hashtag sets:', error);
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
	
	// Save hashtags from new tab - switch to saved tab
	function handleSaveFromNew() {
		if (selectedHashtags.length === 0) {
			toast.show('Chọn ít nhất 1 hashtag', 'warning');
			return;
		}
		
		// Convert selected hashtags to tags array
		newSetTags = selectedHashtags.map(h => h.name);
		
		// Switch to saved tab
		activeTab = 'saved';
		selectedSetId = 'new';
		
		// Clear new tab state
		selectedHashtags = [];
		searchQuery = '';
		searchResults = [];
		
		toast.show('Đã chuyển sang tab "Hashtag đã lưu"', 'info');
	}
	
	// Confirm and insert hashtags (for new tab)
	function handleConfirm() {
		if (selectedHashtags.length === 0) {
			toast.show('Chọn ít nhất 1 hashtag', 'warning');
			return;
		}
		
		const hashtagsText = selectedHashtags.map(h => '#' + h.name).join(' ');
		onSelect(hashtagsText);
		
		// Reset
		selectedHashtags = [];
		searchQuery = '';
		searchResults = [];
		
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
	
	// Select existing set (from preset button)
	function selectExistingSet(set) {
		selectedSetId = set.id.toString();
		newSetName = set.name;
		newSetTags = set.hashtags.split(' ').map(h => h.replace('#', '')).filter(h => h);
		showCreateForm = true; // Show form to allow editing
	}
	
	// Reset form
	function resetForm() {
		selectedSetId = '';
		newSetName = '';
		newSetHashtags = '';
		newSetTags = [];
		randomCount = 0;
	}
	
	// Apply hashtags with random selection
	function applyHashtags() {
		if (newSetTags.length === 0) {
			toast.show('Không có hashtag để áp dụng', 'warning');
			return;
		}
		
		let tagsToApply = [...newSetTags];
		if (randomCount > 0 && randomCount < newSetTags.length) {
			const shuffled = [...newSetTags].sort(() => Math.random() - 0.5);
			tagsToApply = shuffled.slice(0, randomCount);
		}
		
		const hashtagsText = tagsToApply.map(t => '#' + t).join(' ');
		onSelect(hashtagsText);
		
		// Reset
		selectedSetId = '';
		newSetName = '';
		newSetHashtags = '';
		newSetTags = [];
		randomCount = 0;
		
		onClose();
	}
	
	// Save new hashtag set and apply
	async function handleSaveNewSet() {
		if (!newSetName.trim()) {
			toast.show('Vui lòng nhập tên', 'warning');
			return;
		}
		
		if (newSetTags.length === 0) {
			toast.show('Vui lòng thêm ít nhất 1 hashtag', 'warning');
			return;
		}
		
		// Save hashtag set to database
		try {
			const hashtagsString = newSetTags.map(t => '#' + t).join(' ');
			await api.saveHashtags({
				name: newSetName,
				hashtags: hashtagsString
			});
			toast.show('Đã lưu hashtag set', 'success');
			await loadSavedHashtags();
		} catch (error) {
			toast.show('Lỗi lưu hashtag set', 'error');
			return;
		}
		
		// Apply hashtags
		applyHashtags();
	}
</script>

{#if show}
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
					Hashtag đã lưu ({savedHashtagSets.length})
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
				{:else if searchResults.length === 0 && searchQuery}
					<div class="p-8 text-center text-gray-400">
						Không tìm thấy hashtag
					</div>
				{:else if searchResults.length === 0 && !searchQuery}
					<div class="p-8 text-center text-gray-400">
						Nhập từ khóa để tìm kiếm
					</div>
					
					<!-- Top Hashtags (below empty state) -->
					{#if topHashtags.length > 0}
						<div class="px-4 pb-4">
							<div class="text-xs text-gray-600 mb-2 font-medium">Hashtag dùng nhiều nhất</div>
							<div class="flex flex-wrap gap-1.5">
								{#each topHashtags as hashtag}
									<button
										on:click={() => toggleHashtag(hashtag)}
										class="px-2 py-1 text-xs rounded-full transition-colors {selectedHashtags.some(h => h.name === hashtag.name) ? 'bg-blue-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-100 border border-gray-200'}"
									>
										#{hashtag.name}
									</button>
								{/each}
							</div>
						</div>
					{/if}
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
								<div class="w-5 h-5 rounded border-2 flex items-center justify-center transition-colors {selectedHashtags.some(h => h.name === hashtag.name) ? 'bg-blue-600 border-blue-600' : 'border-gray-300'}">
									{#if selectedHashtags.some(h => h.name === hashtag.name)}
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
				{#if !showCreateForm}
					<!-- Preset Buttons -->
					<div class="flex items-center justify-between mb-3">
						<div class="text-sm text-gray-700 font-medium">Chọn hashtag set</div>
						<button on:click={() => showCreateForm = true} class="px-2 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded font-medium">+ Thêm mới</button>
					</div>
					{#if loading}
						<div class="py-8 text-center"><div class="w-6 h-6 border-2 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto"></div></div>
					{:else if savedHashtagSets.length === 0}
						<div class="py-8 text-center text-gray-400 text-sm">Chưa có hashtag set</div>
					{:else}
						<div class="space-y-2 max-h-80 overflow-y-auto">
							{#each savedHashtagSets as set}
								<button on:click={() => selectExistingSet(set)} class="w-full px-3 py-2.5 bg-gray-50 hover:bg-blue-50 border border-gray-200 hover:border-blue-300 rounded-lg transition-all text-left group">
									<div class="flex items-center justify-between gap-2">
										<div class="flex-1 min-w-0">
											<div class="text-sm font-medium text-gray-900 mb-0.5">{set.name}</div>
											<div class="text-xs text-gray-500 truncate">{set.hashtags}</div>
										</div>
										<svg class="w-4 h-4 text-blue-600 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
									</div>
								</button>
							{/each}
						</div>
					{/if}
				{:else}
					<!-- Create/Edit Form -->
					<div class="flex items-center justify-between mb-3">
						<div class="text-sm text-gray-700 font-medium">{selectedSetId ? 'Chỉnh sửa' : 'Tạo mới'}</div>
						<button on:click={() => { showCreateForm = false; resetForm(); }} class="text-xs text-gray-600 hover:text-gray-700">← Quay lại</button>
					</div>
					<input type="text" bind:value={newSetName} placeholder="Tên hashtag set" disabled={selectedSetId !== ''} class="w-full px-3 py-2 text-sm bg-gray-50 border border-gray-300 rounded-lg mb-3 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-60 disabled:cursor-not-allowed" />
					<div class="w-full min-h-[80px] px-3 py-2 text-sm bg-gray-50 border border-gray-300 rounded-lg focus-within:ring-2 focus-within:ring-blue-500 mb-3">
						<div class="flex flex-wrap gap-1.5 items-center">
							{#each newSetTags as tag}
								<div class="inline-flex items-center gap-1 px-2 py-0.5 bg-gray-200 rounded-full text-xs">
									<span class="text-gray-700">#{tag}</span>
									<button on:click={() => removeNewTag(tag)} class="text-gray-500 hover:text-gray-700"><X size={12} /></button>
								</div>
							{/each}
							<input type="text" bind:value={newSetHashtags} on:keydown={handleHashtagInput} placeholder={newSetTags.length === 0 ? "Nhập hashtag và Enter" : ""} class="flex-1 min-w-[100px] bg-transparent border-none outline-none text-sm" />
						</div>
					</div>
					<div class="flex items-center gap-2 text-xs text-gray-600">
						<span>Ngẫu nhiên chọn</span>
						<input type="number" bind:value={randomCount} min="0" max={newSetTags.length} class="w-16 px-2 py-1 text-xs text-center bg-gray-50 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500" />
						<span>hashtag/bài</span>
					</div>
				{/if}
				

			</div>
		{/if}
		
		<!-- Selected Hashtags Pills -->
		{#if selectedHashtags.length > 0}
			<div class="px-3 py-2 border-t border-gray-100 bg-gray-50">
				<div class="flex flex-wrap gap-1.5">
					{#each selectedHashtags as hashtag}
						<div class="inline-flex items-center gap-1 px-2.5 py-1 bg-white border border-gray-300 rounded-full text-xs">
							<span class="text-blue-600">#{hashtag.name}</span>
							<button
								on:click={() => removeSelected(hashtag)}
								class="text-gray-400 hover:text-gray-600"
							>
								<X size={12} />
							</button>
						</div>
					{/each}
				</div>
			</div>
		{/if}
		
		<!-- Footer Actions -->
		<div class="p-3 border-t border-gray-200">
			{#if activeTab === 'new'}
				<div class="flex gap-2">
					<button
						on:click={handleSaveFromNew}
						disabled={selectedHashtags.length === 0}
						class="flex-1 px-4 py-2 text-sm text-gray-700 bg-gray-100 font-semibold rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Lưu hashtag
					</button>
					<button
						on:click={handleConfirm}
						disabled={selectedHashtags.length === 0}
						class="flex-1 px-4 py-2 text-sm bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Áp dụng
					</button>
				</div>
			{:else}
				{#if showCreateForm}
					{#if selectedSetId && selectedSetId !== ''}
						<!-- Editing existing set -->
						<button
							on:click={applyHashtags}
							disabled={newSetTags.length === 0}
							class="w-full px-4 py-2 text-sm bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
						>
							Áp dụng
						</button>
					{:else}
						<!-- Creating new set -->
						<button
							on:click={handleSaveNewSet}
							disabled={!newSetName.trim() || newSetTags.length === 0}
							class="w-full px-4 py-2 text-sm bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
						>
							Lưu & Áp dụng
						</button>
					{/if}
				{/if}
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
