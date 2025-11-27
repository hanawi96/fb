<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import Button from '$lib/components/Button.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import { Calendar, Trash2, RefreshCw } from 'lucide-svelte';
	
	// Accept SvelteKit props
	export let data = undefined;
	export let params = undefined;
	
	let posts = [];
	let pages = [];
	let scheduledPosts = [];
	let loading = true;
	let showScheduleModal = false;
	
	let selectedPost = null;
	let selectedPages = [];
	let scheduledTime = '';
	let scheduling = false;
	
	onMount(async () => {
		await Promise.all([loadPosts(), loadPages(), loadScheduled()]);
		loading = false;
	});
	
	async function loadPosts() {
		try {
			posts = await api.getPosts(100, 0);
		} catch (error) {
			toast.show('Không thể tải bài viết', 'error');
		}
	}
	
	async function loadPages() {
		try {
			const allPages = await api.getPages();
			pages = allPages.filter(p => p.is_active);
		} catch (error) {
			toast.show('Không thể tải pages', 'error');
		}
	}
	
	async function loadScheduled() {
		try {
			scheduledPosts = await api.getScheduledPosts('', 100, 0);
		} catch (error) {
			toast.show('Không thể tải lịch đăng', 'error');
		}
	}
	
	function openScheduleModal(post) {
		selectedPost = post;
		selectedPages = [];
		
		// Set default time to 1 hour from now
		const now = new Date();
		now.setHours(now.getHours() + 1);
		scheduledTime = now.toISOString().slice(0, 16);
		
		showScheduleModal = true;
	}
	
	function togglePage(pageId) {
		if (selectedPages.includes(pageId)) {
			selectedPages = selectedPages.filter(id => id !== pageId);
		} else {
			selectedPages = [...selectedPages, pageId];
		}
	}
	
	async function schedulePost() {
		if (selectedPages.length === 0) {
			toast.show('Vui lòng chọn ít nhất 1 page', 'warning');
			return;
		}
		
		if (!scheduledTime) {
			toast.show('Vui lòng chọn thời gian', 'warning');
			return;
		}
		
		scheduling = true;
		
		try {
			await api.schedulePost({
				post_id: selectedPost.id,
				page_ids: selectedPages,
				scheduled_time: new Date(scheduledTime).toISOString()
			});
			
			toast.show('Đã hẹn giờ đăng bài', 'success');
			showScheduleModal = false;
			await loadScheduled();
		} catch (error) {
			toast.show('Lỗi: ' + error.message, 'error');
		} finally {
			scheduling = false;
		}
	}
	
	async function deleteScheduled(id) {
		if (!confirm('Hủy lịch đăng này?')) return;
		
		try {
			await api.deleteScheduledPost(id);
			toast.show('Đã hủy lịch đăng', 'success');
			await loadScheduled();
		} catch (error) {
			toast.show('Không thể hủy', 'error');
		}
	}
	
	async function retryScheduled(id) {
		try {
			await api.retryScheduledPost(id);
			toast.show('Đã thêm vào hàng đợi', 'success');
			await loadScheduled();
		} catch (error) {
			toast.show('Không thể retry', 'error');
		}
	}
	
	function formatDate(dateString) {
		return new Date(dateString).toLocaleString('vi-VN');
	}
	
	function getStatusBadge(status) {
		const badges = {
			pending: 'badge-warning',
			processing: 'badge-info',
			completed: 'badge-success',
			failed: 'badge-error'
		};
		return badges[status] || 'badge-info';
	}
	
	function getStatusText(status) {
		const texts = {
			pending: 'Chờ đăng',
			processing: 'Đang đăng',
			completed: 'Thành công',
			failed: 'Thất bại'
		};
		return texts[status] || status;
	}
</script>

<svelte:head>
	<title>Lịch đăng bài - FB Scheduler</title>
</svelte:head>

{#if $toast}
	<Toast 
		message={$toast.message} 
		type={$toast.type} 
		duration={$toast.duration || 3000}
		onClose={() => toast.hide()} 
	/>
{/if}

<div>
	<h1 class="text-3xl font-bold mb-2">Lịch đăng bài</h1>
	<p class="text-gray-600 mb-8">Hẹn giờ đăng bài lên các pages</p>
	
	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-4 border-primary-600 border-t-transparent"></div>
		</div>
	{:else}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Available Posts -->
			<div class="card">
				<h2 class="text-xl font-semibold mb-4">Bài viết có sẵn</h2>
				<div class="space-y-3 max-h-[600px] overflow-y-auto">
					{#each posts as post}
						<div class="p-4 border rounded-lg hover:border-primary-500 transition-colors">
							<p class="text-sm text-gray-700 line-clamp-2 mb-2">{post.content}</p>
							{#if post.media_urls && post.media_urls.length > 0}
								<p class="text-xs text-gray-500 mb-2">📷 {post.media_urls.length} ảnh</p>
							{/if}
							<Button on:click={() => openScheduleModal(post)} class="w-full text-sm">
								<Calendar size={16} class="mr-2" />
								Hẹn giờ đăng
							</Button>
						</div>
					{/each}
				</div>
			</div>
			
			<!-- Scheduled Posts -->
			<div class="card">
				<h2 class="text-xl font-semibold mb-4">Lịch đã hẹn</h2>
				<div class="space-y-3 max-h-[600px] overflow-y-auto">
					{#each scheduledPosts as scheduled}
						<div class="p-4 border rounded-lg">
							<div class="flex items-start justify-between mb-2">
								<div class="flex-1">
									<p class="text-sm font-medium">{scheduled.page?.page_name}</p>
									<p class="text-xs text-gray-600 mt-1">{formatDate(scheduled.scheduled_time)}</p>
								</div>
								<span class="badge {getStatusBadge(scheduled.status)}">
									{getStatusText(scheduled.status)}
								</span>
							</div>
							<p class="text-sm text-gray-700 line-clamp-2 mb-3">{scheduled.post?.content}</p>
							<div class="flex gap-2">
								{#if scheduled.status === 'failed'}
									<button
										on:click={() => retryScheduled(scheduled.id)}
										class="flex-1 flex items-center justify-center gap-2 px-3 py-1.5 text-sm rounded-lg border hover:bg-gray-50"
									>
										<RefreshCw size={14} />
										Thử lại
									</button>
								{/if}
								{#if scheduled.status === 'pending'}
									<button
										on:click={() => deleteScheduled(scheduled.id)}
										class="flex-1 flex items-center justify-center gap-2 px-3 py-1.5 text-sm rounded-lg border border-red-200 text-red-600 hover:bg-red-50"
									>
										<Trash2 size={14} />
										Hủy
									</button>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Schedule Modal -->
{#if showScheduleModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
		<div class="bg-white rounded-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
			<div class="p-6">
				<h2 class="text-2xl font-bold mb-4">Hẹn giờ đăng bài</h2>
				
				<div class="mb-6">
					<label class="block text-sm font-medium text-gray-700 mb-2">
						Thời gian đăng
					</label>
					<input
						type="datetime-local"
						bind:value={scheduledTime}
						class="input"
					/>
				</div>
				
				<div class="mb-6">
					<label class="block text-sm font-medium text-gray-700 mb-2">
						Chọn Pages ({selectedPages.length} đã chọn)
					</label>
					<div class="grid grid-cols-1 gap-2 max-h-60 overflow-y-auto">
						{#each pages as page}
							<label class="flex items-center gap-3 p-3 border rounded-lg cursor-pointer hover:bg-gray-50">
								<input
									type="checkbox"
									checked={selectedPages.includes(page.id)}
									on:change={() => togglePage(page.id)}
									class="w-4 h-4"
								/>
								<img src={page.profile_picture_url} alt="" class="w-10 h-10 rounded-full" />
								<span class="flex-1 font-medium">{page.page_name}</span>
							</label>
						{/each}
					</div>
				</div>
				
				<div class="flex gap-3">
					<Button on:click={schedulePost} loading={scheduling} class="flex-1">
						Xác nhận
					</Button>
					<Button variant="secondary" on:click={() => showScheduleModal = false}>
						Hủy
					</Button>
				</div>
			</div>
		</div>
	</div>
{/if}
