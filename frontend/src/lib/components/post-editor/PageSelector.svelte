<script>
	export let pages = [];
	export let selectedPages = [];
	
	function togglePage(page) {
		const index = selectedPages.findIndex(p => p.id === page.id);
		if (index > -1) {
			selectedPages = selectedPages.filter(p => p.id !== page.id);
		} else {
			selectedPages = [...selectedPages, page];
		}
	}
	
	function toggleAll() {
		const isAllSelected = pages.length > 0 && selectedPages.length === pages.length;
		if (isAllSelected) {
			selectedPages = [];
		} else {
			selectedPages = [...pages];
		}
	}
	
	// Tạo reactive Set để trigger re-render khi selectedPages thay đổi
	$: selectedPageIds = new Set(selectedPages.map(p => p.id));
	
	// Reactive: tính toán trạng thái checkbox "Chọn tất cả"
	$: allSelected = pages.length > 0 && selectedPages.length === pages.length;
	$: indeterminate = selectedPages.length > 0 && selectedPages.length < pages.length;
	
	// Ref cho checkbox để set indeterminate
	let checkAllRef;
	
	// Update indeterminate state
	$: if (checkAllRef) {
		checkAllRef.indeterminate = indeterminate;
	}
</script>

<div class="bg-white border border-gray-200 rounded-lg flex flex-col h-full">
	<div class="px-3 py-3 border-b border-gray-100">
		<div class="flex items-center justify-between mb-2">
			<h2 class="text-xs font-semibold text-gray-700 uppercase">Chọn tài khoản</h2>
			<span class="text-xs text-blue-600 font-medium">{selectedPages.length}/{pages.length}</span>
		</div>
		
		{#if pages.length > 0}
			<div class="flex items-center gap-2">
				<input 
					bind:this={checkAllRef}
					type="checkbox" 
					checked={allSelected}
					on:change={toggleAll}
					class="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-2 focus:ring-blue-500 cursor-pointer"
				/>
				<label class="text-xs text-gray-600 cursor-pointer" on:click={() => checkAllRef?.click()}>
					Chọn tất cả
				</label>
			</div>
		{/if}
	</div>
	
	<div class="flex-1 overflow-y-auto p-2">
		{#if pages.length === 0}
			<div class="p-4 text-center">
				<div class="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-3">
					<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
					</svg>
				</div>
				<p class="text-sm font-medium text-gray-900 mb-1">Đang tải pages...</p>
				<p class="text-xs text-gray-500">Vui lòng đợi</p>
			</div>
		{:else}
			{#each pages as page (page.id)}
				{@const isPageSelected = selectedPageIds.has(page.id)}
				<div
					class="w-full flex items-center gap-2 p-2 mb-1 rounded-lg hover:bg-gray-50 transition-colors {isPageSelected ? 'bg-blue-50 border border-blue-200' : 'border border-transparent'}"
				>
					<label class="flex items-center gap-2 flex-1 min-w-0 cursor-pointer">
						<input 
							type="checkbox" 
							checked={isPageSelected}
							on:change={() => togglePage(page)}
							class="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-2 focus:ring-blue-500 flex-shrink-0"
						/>
						<img 
							src={page.profile_picture_url || 'https://via.placeholder.com/40'} 
							alt={page.page_name}
							class="w-9 h-9 rounded-full flex-shrink-0 border border-gray-200"
						/>
						<div class="flex-1 text-left min-w-0">
							<div class="text-sm font-medium text-gray-900 truncate">{page.page_name}</div>
							<div class="text-xs text-gray-500 truncate">{page.category || 'Facebook'}</div>
						</div>
					</label>
					<a
						href="https://facebook.com/{page.page_id}"
						target="_blank"
						rel="noopener noreferrer"
						class="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors flex-shrink-0"
						title="Xem trang"
						on:click|stopPropagation|preventDefault={() => window.open(`https://facebook.com/${page.page_id}`, '_blank')}
					>
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
						</svg>
					</a>
				</div>
			{/each}
		{/if}
	</div>
</div>
