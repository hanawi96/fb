<script>
	import { createEventDispatcher } from 'svelte';
	import { Facebook, Search, X, Check, ChevronDown } from 'lucide-svelte';
	
	export let pages = [];
	export let selectedPageIds = new Set();
	export let placeholder = 'Chọn page...';
	
	const dispatch = createEventDispatcher();
	
	let isOpen = false;
	let searchQuery = '';
	let dropdownRef;
	
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
	
	// Filter and sort pages by search query (A-Z)
	$: filteredPages = pages
		.filter(page => page.page_name.toLowerCase().includes(searchQuery.toLowerCase()))
		.sort((a, b) => a.page_name.localeCompare(b.page_name, 'vi'));
	
	// Selected pages for chips display (sorted)
	$: selectedPages = pages
		.filter(p => selectedPageIds.has(p.id))
		.sort((a, b) => a.page_name.localeCompare(b.page_name, 'vi'));
	
	function togglePage(pageId) {
		const newSet = new Set(selectedPageIds);
		if (newSet.has(pageId)) {
			newSet.delete(pageId);
		} else {
			newSet.add(pageId);
		}
		selectedPageIds = newSet;
		dispatch('change', { selectedPageIds: newSet });
	}
	
	function selectAll() {
		const newSet = new Set(filteredPages.map(p => p.id));
		selectedPageIds = newSet;
		dispatch('change', { selectedPageIds: newSet });
	}
	
	function clearAll() {
		selectedPageIds = new Set();
		dispatch('change', { selectedPageIds: new Set() });
	}
	
	function removeSelected(pageId) {
		const newSet = new Set(selectedPageIds);
		newSet.delete(pageId);
		selectedPageIds = newSet;
		dispatch('change', { selectedPageIds: newSet });
	}
	
	function handleClickOutside(event) {
		if (dropdownRef && !dropdownRef.contains(event.target)) {
			isOpen = false;
		}
	}
</script>

<svelte:window on:click={handleClickOutside} />

<div class="relative" bind:this={dropdownRef}>
	<!-- Trigger Button -->
	<button
		type="button"
		on:click|stopPropagation={() => isOpen = !isOpen}
		class="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 rounded-lg text-sm hover:border-gray-300 transition-colors min-w-[160px]"
	>
		<Facebook size={16} class="text-gray-400" />
		<span class="flex-1 text-left text-gray-700">
			{#if selectedPageIds.size === 0}
				{placeholder}
			{:else if selectedPageIds.size === 1}
				{selectedPages[0]?.page_name || '1 page'}
			{:else}
				{selectedPageIds.size} pages
			{/if}
		</span>
		<ChevronDown size={16} class="text-gray-400 transition-transform {isOpen ? 'rotate-180' : ''}" />
	</button>
	
	<!-- Dropdown -->
	{#if isOpen}
		<div 
			class="absolute top-full left-0 mt-1 w-80 bg-white border border-gray-200 rounded-lg shadow-lg z-50 overflow-hidden"
			on:click|stopPropagation
		>
			<!-- Search -->
			<div class="p-2 border-b border-gray-100">
				<div class="relative">
					<Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Tìm page..."
						class="w-full pl-9 pr-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:border-blue-500"
					/>
				</div>
			</div>
			
			<!-- Actions -->
			<div class="px-3 py-2 border-b border-gray-100 flex items-center justify-between">
				<span class="text-xs text-gray-500">
					{selectedPageIds.size}/{pages.length} đã chọn
				</span>
				<div class="flex gap-2">
					<button
						type="button"
						on:click={selectAll}
						class="text-xs text-blue-600 hover:text-blue-700"
					>
						Chọn tất cả
					</button>
					<button
						type="button"
						on:click={clearAll}
						class="text-xs text-gray-500 hover:text-gray-700"
					>
						Bỏ chọn
					</button>
				</div>
			</div>
			
			<!-- Pages List -->
			<div class="max-h-64 overflow-y-auto">
				{#if filteredPages.length === 0}
					<div class="p-4 text-center text-sm text-gray-500">
						Không tìm thấy page nào
					</div>
				{:else}
					{#each filteredPages as page (page.id)}
						<button
							type="button"
							on:click={() => togglePage(page.id)}
							class="w-full flex items-center gap-3 px-3 py-2 hover:bg-gray-50 transition-colors text-left"
						>
							<div class="relative flex-shrink-0">
								{#if page.profile_picture_url}
									<img 
										src={page.profile_picture_url} 
										alt="" 
										class="w-8 h-8 rounded-full"
									/>
								{:else}
									<div class="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center">
										<Facebook size={14} class="text-gray-400" />
									</div>
								{/if}
								<div class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full {getPageColor(page.id)}"></div>
							</div>
							<div class="flex-1 min-w-0">
								<div class="text-sm text-gray-900 truncate">{page.page_name}</div>
								<div class="text-xs text-gray-500 truncate">{page.category || ''}</div>
							</div>
							{#if selectedPageIds.has(page.id)}
								<Check size={16} class="text-blue-600 flex-shrink-0" />
							{/if}
						</button>
					{/each}
				{/if}
			</div>
		</div>
	{/if}
</div>


