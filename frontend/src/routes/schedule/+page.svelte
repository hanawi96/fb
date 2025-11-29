<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import Toast from '$lib/components/Toast.svelte';
	import { RefreshCw, Plus, Search, X } from 'lucide-svelte';
	
	// Import schedule components
	import PageSelector from '$lib/components/schedule/PageSelector.svelte';
	import TimeRangeSelector from '$lib/components/schedule/TimeRangeSelector.svelte';
	import StatusFilter from '$lib/components/schedule/StatusFilter.svelte';
	import SchedulePostRow from '$lib/components/schedule/SchedulePostRow.svelte';
	import AlertBar from '$lib/components/schedule/AlertBar.svelte';
	import BulkActionsBar from '$lib/components/schedule/BulkActionsBar.svelte';

	// Data
	let scheduledPosts = [];
	let pages = [];
	let loading = true;
	let loadingMore = false;
	
	// Filters
	let searchQuery = '';
	let selectedPageIds = new Set();
	let timeRange = 'all';
	let timeRangeCustomStart = '';
	let timeRangeCustomEnd = '';
	let statusFilter = '';
	
	// Selection
	let selectedPostIds = new Set();
	
	// Pagination
	let offset = 0;
	let limit = 200;
	let hasMore = true;

	onMount(async () => {
		await Promise.all([loadPages(), loadScheduledPosts()]);
	});

	async function loadPages() {
		try {
			pages = await api.getPages();
		} catch (e) {
			console.error('Error loading pages:', e);
		}
	}

	async function loadScheduledPosts(append = false) {
		if (append) {
			loadingMore = true;
		} else {
			loading = true;
			offset = 0;
		}
		
		try {
			// Always load all posts, filter by status in frontend
			const newPosts = await api.getScheduledPosts('', limit, offset);
			
			if (append) {
				scheduledPosts = [...scheduledPosts, ...newPosts];
			} else {
				scheduledPosts = newPosts;
			}
			
			hasMore = newPosts.length === limit;
		} catch (e) {
			toast.show('Lỗi tải dữ liệu: ' + e.message, 'error');
		} finally {
			loading = false;
			loadingMore = false;
		}
	}

	function loadMore() {
		offset += limit;
		loadScheduledPosts(true);
	}

	// Filter handlers
	function handlePageChange(event) {
		selectedPageIds = event.detail.selectedPageIds;
		selectedPostIds = new Set();
	}

	function handleTimeRangeChange(event) {
		timeRange = event.detail.value;
		timeRangeCustomStart = event.detail.customStart || '';
		timeRangeCustomEnd = event.detail.customEnd || '';
		selectedPostIds = new Set();
	}

	function handleStatusChange(event) {
		statusFilter = event.detail.value;
		selectedPostIds = new Set();
	}

	function clearFilters() {
		searchQuery = '';
		selectedPageIds = new Set();
		timeRange = 'all';
		timeRangeCustomStart = '';
		timeRangeCustomEnd = '';
		statusFilter = '';
		selectedPostIds = new Set();
		loadScheduledPosts();
	}

	// Selection handlers
	function handlePostSelect(event) {
		const { id, selected } = event.detail;
		const newSet = new Set(selectedPostIds);
		if (selected) {
			newSet.add(id);
		} else {
			newSet.delete(id);
		}
		selectedPostIds = newSet;
	}

	function selectAllPosts() {
		selectedPostIds = new Set(filteredPosts.map(p => p.id));
	}

	function clearSelection() {
		selectedPostIds = new Set();
	}

	// Post actions
	async function handleTestPost(event) {
		const { id } = event.detail;
		try {
			const response = await fetch(`http://localhost:8080/api/schedule/${id}/test`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' }
			});
			const data = await response.json();
			toast.show(data.message, 'success');
			await loadScheduledPosts();
		} catch (e) {
			toast.show('Lỗi: ' + e.message, 'error');
		}
	}

	async function handleRetryPost(event) {
		const { id } = event.detail;
		try {
			await api.retryScheduledPost(id);
			toast.show('Đã đưa vào hàng đợi retry', 'success');
			await loadScheduledPosts();
		} catch (e) {
			toast.show('Lỗi: ' + e.message, 'error');
		}
	}

	async function handleDeletePost(event) {
		const { id } = event.detail;
		if (!confirm('Xóa bài đăng này?')) return;
		try {
			await api.deleteScheduledPost(id);
			toast.show('Đã xóa', 'success');
			selectedPostIds.delete(id);
			selectedPostIds = selectedPostIds;
			await loadScheduledPosts();
		} catch (e) {
			toast.show('Lỗi: ' + e.message, 'error');
		}
	}

	async function handleBulkDelete() {
		if (selectedPostIds.size === 0) return;
		if (!confirm(`Xóa ${selectedPostIds.size} bài đăng đã chọn?`)) return;
		
		try {
			const deletePromises = Array.from(selectedPostIds).map(id => 
				api.deleteScheduledPost(id)
			);
			await Promise.all(deletePromises);
			toast.show(`Đã xóa ${selectedPostIds.size} bài`, 'success');
			selectedPostIds = new Set();
			await loadScheduledPosts();
		} catch (e) {
			toast.show('Lỗi khi xóa: ' + e.message, 'error');
		}
	}

	function handleViewFailed() {
		statusFilter = 'failed';
		loadScheduledPosts();
	}

	// Computed values
	$: filteredPosts = scheduledPosts.filter(post => {
		// Filter by status
		if (statusFilter && post.status !== statusFilter) {
			return false;
		}
		
		// Filter by search query
		if (searchQuery) {
			const query = searchQuery.toLowerCase();
			const content = post.post?.content?.toLowerCase() || '';
			const pageName = post.page?.page_name?.toLowerCase() || '';
			if (!content.includes(query) && !pageName.includes(query)) {
				return false;
			}
		}
		
		// Filter by selected pages - check both page.id and page_id for compatibility
		if (selectedPageIds.size > 0) {
			const pageId = post.page?.id || post.page_id;
			if (!selectedPageIds.has(pageId)) {
				return false;
			}
		}
		
		// Filter by time range
		if (timeRange !== 'all' && post.scheduled_time) {
			const postDate = new Date(post.scheduled_time);
			const today = new Date();
			today.setHours(0, 0, 0, 0);
			const tomorrow = new Date(today);
			tomorrow.setDate(tomorrow.getDate() + 1);
			const nextWeek = new Date(today);
			nextWeek.setDate(nextWeek.getDate() + 7);
			const nextMonth = new Date(today);
			nextMonth.setDate(nextMonth.getDate() + 30);
			
			switch (timeRange) {
				case 'today':
					if (postDate < today || postDate >= tomorrow) return false;
					break;
				case 'tomorrow':
					const dayAfterTomorrow = new Date(tomorrow);
					dayAfterTomorrow.setDate(dayAfterTomorrow.getDate() + 1);
					if (postDate < tomorrow || postDate >= dayAfterTomorrow) return false;
					break;
				case 'week':
					if (postDate < today || postDate >= nextWeek) return false;
					break;
				case 'month':
					if (postDate < today || postDate >= nextMonth) return false;
					break;
				case 'past':
					if (postDate >= today) return false;
					break;
				case 'custom':
					if (timeRangeCustomStart && postDate < new Date(timeRangeCustomStart)) return false;
					if (timeRangeCustomEnd) {
						const endDate = new Date(timeRangeCustomEnd);
						endDate.setDate(endDate.getDate() + 1);
						if (postDate >= endDate) return false;
					}
					break;
			}
		}
		
		return true;
	});

	// Group posts by date
	$: groupedPosts = groupPostsByDate(filteredPosts);

	function groupPostsByDate(posts) {
		const groups = {};
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const tomorrow = new Date(today);
		tomorrow.setDate(tomorrow.getDate() + 1);
		
		posts.forEach(post => {
			if (!post.scheduled_time) return;
			
			const postDate = new Date(post.scheduled_time);
			postDate.setHours(0, 0, 0, 0);
			
			let label;
			if (postDate.getTime() === today.getTime()) {
				label = 'Hôm nay';
			} else if (postDate.getTime() === tomorrow.getTime()) {
				label = 'Ngày mai';
			} else if (postDate < today) {
				label = 'Đã qua';
			} else {
				label = postDate.toLocaleDateString('vi-VN', { 
					weekday: 'long', 
					day: 'numeric', 
					month: 'numeric',
					year: 'numeric'
				});
			}
			
			if (!groups[label]) {
				groups[label] = {
					label,
					date: postDate,
					posts: []
				};
			}
			groups[label].posts.push(post);
		});
		
		// Sort posts within each group by time
		Object.values(groups).forEach(group => {
			group.posts.sort((a, b) => new Date(a.scheduled_time) - new Date(b.scheduled_time));
		});
		
		// Sort groups by date
		return Object.values(groups).sort((a, b) => {
			if (a.label === 'Hôm nay') return -1;
			if (b.label === 'Hôm nay') return 1;
			if (a.label === 'Ngày mai') return -1;
			if (b.label === 'Ngày mai') return 1;
			if (a.label === 'Đã qua') return 1;
			if (b.label === 'Đã qua') return -1;
			return a.date - b.date;
		});
	}

	$: failedCount = scheduledPosts.filter(p => p.status === 'failed').length;
	$: hasActiveFilters = searchQuery || selectedPageIds.size > 0 || timeRange !== 'all' || statusFilter;
	
	// Reactive counts for StatusFilter
	$: statusCounts = {
		all: scheduledPosts.length,
		pending: scheduledPosts.filter(p => p.status === 'pending').length,
		completed: scheduledPosts.filter(p => p.status === 'completed').length,
		failed: scheduledPosts.filter(p => p.status === 'failed').length
	};

	// Selected pages for display
	$: selectedPagesDisplay = pages
		.filter(p => selectedPageIds.has(p.id))
		.sort((a, b) => a.page_name.localeCompare(b.page_name, 'vi'));

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

	function removeSelectedPage(pageId) {
		const newSet = new Set(selectedPageIds);
		newSet.delete(pageId);
		selectedPageIds = newSet;
	}

	// Today stats
	$: todayStats = (() => {
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const tomorrow = new Date(today);
		tomorrow.setDate(tomorrow.getDate() + 1);
		
		const todayPosts = scheduledPosts.filter(p => {
			if (!p.scheduled_time) return false;
			const postDate = new Date(p.scheduled_time);
			return postDate >= today && postDate < tomorrow;
		});
		
		return {
			completed: todayPosts.filter(p => p.status === 'completed').length,
			pending: todayPosts.filter(p => p.status === 'pending').length,
			failed: todayPosts.filter(p => p.status === 'failed').length
		};
	})();
</script>

<svelte:head>
	<title>Lịch đăng bài - FB Scheduler</title>
</svelte:head>

{#if $toast}
	<Toast message={$toast.message} type={$toast.type} onClose={() => toast.hide()} />
{/if}

<div class="w-full">
	<!-- Header -->
	<div class="mb-4 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-gray-900">Lịch đăng bài</h1>
			<p class="text-sm text-gray-500 mt-1">Quản lý các bài viết đã hẹn giờ</p>
		</div>
		<div class="flex items-center gap-2">
			<button 
				on:click={() => loadScheduledPosts()} 
				class="flex items-center gap-2 px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
			>
				<RefreshCw size={16} class={loading ? 'animate-spin' : ''} />
				Làm mới
			</button>
			<a 
				href="/posts/new"
				class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				<Plus size={16} />
				Tạo bài mới
			</a>
		</div>
	</div>

	<!-- Quick Stats -->
	{#if !loading && scheduledPosts.length > 0}
		<div class="mb-4 px-4 py-3 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
			<div class="flex items-center gap-6 text-sm">
				<div class="flex items-center gap-2">
					<span class="text-gray-600">Hôm nay:</span>
					<span class="font-semibold text-green-600">{todayStats.completed} đã đăng</span>
					<span class="text-gray-400">•</span>
					<span class="font-semibold text-yellow-600">{todayStats.pending} chờ đăng</span>
					{#if todayStats.failed > 0}
						<span class="text-gray-400">•</span>
						<span class="font-semibold text-red-600">{todayStats.failed} thất bại</span>
					{/if}
				</div>
				<div class="flex-1"></div>
				<div class="text-gray-500">
					Tổng: <span class="font-medium text-gray-900">{scheduledPosts.length} bài</span>
				</div>
			</div>
		</div>
	{/if}

	<!-- Filters Bar - Compact -->
	<div class="mb-4 flex flex-wrap items-center gap-3">
		<!-- Search -->
		<div class="relative flex-1 min-w-[200px] max-w-md">
			<Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Tìm kiếm..."
				class="w-full pl-9 pr-8 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:border-blue-500"
			/>
			{#if searchQuery}
				<button
					type="button"
					on:click={() => searchQuery = ''}
					class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
				>
					<X size={14} />
				</button>
			{/if}
		</div>
		
		<PageSelector 
			{pages} 
			bind:selectedPageIds 
			on:change={handlePageChange}
			placeholder="Chọn page"
		/>
		
		<TimeRangeSelector 
			bind:value={timeRange}
			bind:customStart={timeRangeCustomStart}
			bind:customEnd={timeRangeCustomEnd}
			on:change={handleTimeRangeChange}
		/>
		
		<!-- Status Filter Inline -->
		<StatusFilter 
			bind:value={statusFilter} 
			counts={statusCounts}
			on:change={handleStatusChange}
		/>
		
		{#if hasActiveFilters}
			<button
				type="button"
				on:click={clearFilters}
				class="flex items-center gap-1 px-2 py-2 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
				title="Xóa bộ lọc"
			>
				<X size={14} />
			</button>
		{/if}
	</div>

	<!-- Selected Pages Chips -->
	{#if selectedPagesDisplay.length > 0}
		<div class="flex flex-wrap items-center gap-2 mb-3">
			<span class="text-xs text-gray-500">Đang lọc:</span>
			{#if selectedPagesDisplay.length <= 5}
				{#each selectedPagesDisplay as page (page.id)}
					<span class="inline-flex items-center gap-1.5 px-2 py-1 bg-gray-100 rounded-full text-xs text-gray-700">
						<span class="w-2 h-2 rounded-full {getPageColor(page.id)}"></span>
						<span class="max-w-[120px] truncate">{page.page_name}</span>
						<button
							type="button"
							on:click={() => removeSelectedPage(page.id)}
							class="text-gray-400 hover:text-gray-600"
						>
							<X size={12} />
						</button>
					</span>
				{/each}
			{:else}
				{#each selectedPagesDisplay.slice(0, 3) as page (page.id)}
					<span class="inline-flex items-center gap-1.5 px-2 py-1 bg-gray-100 rounded-full text-xs text-gray-700">
						<span class="w-2 h-2 rounded-full {getPageColor(page.id)}"></span>
						<span class="max-w-[100px] truncate">{page.page_name}</span>
						<button
							type="button"
							on:click={() => removeSelectedPage(page.id)}
							class="text-gray-400 hover:text-gray-600"
						>
							<X size={12} />
						</button>
					</span>
				{/each}
				<span class="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-700 rounded-full text-xs">
					+{selectedPagesDisplay.length - 3} pages khác
				</span>
				<button
					type="button"
					on:click={() => selectedPageIds = new Set()}
					class="text-xs text-gray-500 hover:text-gray-700"
				>
					Xóa hết
				</button>
			{/if}
		</div>
	{/if}

	<!-- Alert Bar for Failed Posts -->
	<AlertBar 
		{failedCount} 
		on:viewFailed={handleViewFailed}
	/>

	<!-- Bulk Actions Bar -->
	<BulkActionsBar 
		selectedCount={selectedPostIds.size}
		totalCount={filteredPosts.length}
		on:selectAll={selectAllPosts}
		on:clearSelection={clearSelection}
		on:bulkDelete={handleBulkDelete}
	/>

	<!-- Content -->
	{#if loading}
		<div class="flex items-center justify-center py-16">
			<div class="w-10 h-10 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
		</div>
	{:else if filteredPosts.length === 0}
		<div class="bg-gray-50 rounded-xl p-12 text-center border border-gray-200">
			<div class="w-16 h-16 bg-white rounded-full flex items-center justify-center mx-auto mb-4 border border-gray-200">
				<Search size={28} class="text-gray-400" />
			</div>
			{#if hasActiveFilters}
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Không tìm thấy bài viết</h3>
				<p class="text-sm text-gray-500 mb-4">Thử thay đổi bộ lọc để xem thêm kết quả</p>
				<button
					on:click={clearFilters}
					class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
				>
					Xóa bộ lọc
				</button>
			{:else}
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Chưa có bài viết nào</h3>
				<p class="text-sm text-gray-500 mb-4">Tạo bài viết mới và hẹn giờ đăng</p>
				<a
					href="/posts/new"
					class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
				>
					<Plus size={16} />
					Tạo bài mới
				</a>
			{/if}
		</div>
	{:else}
		<!-- Table Header -->
		<div class="bg-gray-50 border border-gray-200 rounded-t-lg px-3 py-2 flex items-center text-xs font-medium text-gray-500 uppercase tracking-wide">
			<div class="w-6 flex-shrink-0">
				<input
					type="checkbox"
					checked={selectedPostIds.size > 0 && selectedPostIds.size === filteredPosts.length}
					indeterminate={selectedPostIds.size > 0 && selectedPostIds.size < filteredPosts.length}
					on:change={(e) => {
						if (e.target.checked) {
							selectAllPosts();
						} else {
							clearSelection();
						}
					}}
					class="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
					title={selectedPostIds.size > 0 ? 'Bỏ chọn tất cả' : 'Chọn tất cả'}
				/>
			</div>
			<div class="w-24 flex-shrink-0">Ngày/Giờ</div>
			<div class="flex-[35] min-w-0 px-2">Page</div>
			<div class="flex-[45] min-w-0 px-2">Nội dung</div>
			<div class="flex-[20] min-w-0 text-center px-1">Người đăng</div>
			<div class="w-20 text-right">Trạng thái</div>
			<div class="w-8"></div>
		</div>

		<!-- Posts List -->
		<div class="bg-white rounded-lg border border-gray-200 border-t-0 overflow-hidden">
			{#each filteredPosts as post (post.id)}
				<SchedulePostRow 
					{post}
					selected={selectedPostIds.has(post.id)}
					on:select={handlePostSelect}
					on:test={handleTestPost}
					on:retry={handleRetryPost}
					on:delete={handleDeletePost}
				/>
			{/each}
		</div>

		<!-- Load More -->
		{#if hasMore}
			<div class="text-center py-4">
				<button
					on:click={loadMore}
					disabled={loadingMore}
					class="px-6 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
				>
					{#if loadingMore}
						<span class="flex items-center gap-2">
							<RefreshCw size={14} class="animate-spin" />
							Đang tải...
						</span>
					{:else}
						Tải thêm
					{/if}
				</button>
			</div>
		{/if}
	{/if}
</div>
