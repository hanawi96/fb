<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import Toast from '$lib/components/Toast.svelte';
	import { Trash2, Power, PowerOff, Plus, Facebook, CheckCircle2, XCircle } from 'lucide-svelte';
	
	// Accept SvelteKit props
	export let data = undefined;
	export let params = undefined;
	
	let pages = [];
	let loading = true;
	let connectingFacebook = false;
	let showPageSelectionModal = false;
	let loadingPages = false; // Loading state cho việc fetch pages từ Facebook
	let availablePages = [];
	let selectedPageIds = new Set();
	
	onMount(async () => {
		await loadPages();
	});
	
	async function loadPages() {
		try {
			const loadedPages = await api.getPages();
			// Update pages and loading in the same tick to avoid flickering
			pages = loadedPages;
			loading = false;
		} catch (error) {
			console.error('Error loading pages:', error);
			toast.show('Không thể tải danh sách pages', 'error');
			loading = false;
		}
	}
	
	async function connectFacebook() {
		try {
			connectingFacebook = true;
			
			const { url } = await api.getFacebookAuthURL();
			
			const width = 600;
			const height = 700;
			const left = (screen.width - width) / 2;
			const top = (screen.height - height) / 2;
			
			const popup = window.open(
				url,
				'Facebook Login',
				`width=${width},height=${height},left=${left},top=${top}`
			);
			
			// Kiểm tra nếu popup bị block
			if (!popup) {
				toast.show('Vui lòng cho phép popup để kết nối Facebook', 'error');
				connectingFacebook = false;
				return;
			}
			
			const handleMessage = async (event) => {
				console.log('📨 Received message:', event.data);
				
				if (event.data.type === 'facebook-callback') {
					console.log('✅ Facebook callback received with code:', event.data.code?.substring(0, 20) + '...');
					
					// Đảm bảo popup đã đóng
					popup?.close();
					window.removeEventListener('message', handleMessage);
					
					// Hiển thị loading modal NGAY LẬP TỨC
					showPageSelectionModal = true;
					loadingPages = true;
					connectingFacebook = false;
					
					try {
						console.log('🔄 Calling API with code...');
						// Gọi API để lấy danh sách pages
						const result = await api.facebookCallback(event.data.code);
						console.log('📊 API result:', result);
						
						// Chuẩn bị dữ liệu cho modal
						availablePages = result.pages || [];
						selectedPageIds = new Set(pages.map(p => p.page_id));
						
						console.log('📋 Available pages:', availablePages.length);
						console.log('✅ Selected page IDs:', selectedPageIds);
						
					} catch (error) {
						console.error('❌ Callback error:', error);
						toast.show('Lỗi kết nối: ' + error.message, 'error');
						showPageSelectionModal = false;
					} finally {
						loadingPages = false;
					}
				}
			};
			
			window.addEventListener('message', handleMessage);
			
			// Cleanup nếu popup bị đóng mà không có callback
			const checkPopup = setInterval(() => {
				if (popup.closed) {
					clearInterval(checkPopup);
					window.removeEventListener('message', handleMessage);
					connectingFacebook = false;
				}
			}, 500);
			
		} catch (error) {
			toast.show('Không thể kết nối Facebook', 'error');
			connectingFacebook = false;
		}
	}
	
	function togglePageSelection(pageId) {
		// Tối ưu: tạo Set mới thay vì reassign
		const newSet = new Set(selectedPageIds);
		if (newSet.has(pageId)) {
			newSet.delete(pageId);
		} else {
			newSet.add(pageId);
		}
		selectedPageIds = newSet;
	}
	
	async function saveSelectedPages() {
		try {
			const selectedPages = availablePages
				.filter(p => selectedPageIds.has(p.page_id))
				.map(p => ({
					page_id: p.page_id,
					page_name: p.page_name,
					access_token: p.access_token,
					category: p.category,
					profile_picture_url: p.profile_picture_url
				}));
			
			// Đóng modal ngay để UX mượt hơn
			showPageSelectionModal = false;
			toast.show(`Đang lưu ${selectedPages.length} pages...`, 'info');
			
			await api.saveSelectedPages(selectedPages);
			
			toast.show(`Đã lưu ${selectedPages.length} pages!`, 'success');
			await loadPages();
		} catch (error) {
			toast.show('Lỗi khi lưu pages: ' + error.message, 'error');
		}
	}
	
	async function togglePage(id) {
		// Optimistic update - cập nhật UI ngay lập tức
		const pageIndex = pages.findIndex(p => p.id === id);
		if (pageIndex === -1) return;
		
		const originalState = pages[pageIndex].is_active;
		pages[pageIndex].is_active = !originalState;
		pages = pages; // Trigger reactivity
		
		try {
			await api.togglePage(id);
			toast.show('Cập nhật trạng thái thành công', 'success');
		} catch (error) {
			// Rollback nếu lỗi
			pages[pageIndex].is_active = originalState;
			pages = pages;
			toast.show('Không thể cập nhật trạng thái', 'error');
		}
	}
	
	async function deletePage(id) {
		if (!confirm('Bạn có chắc muốn xóa page này?')) return;
		
		// Optimistic update - xóa khỏi UI ngay
		const originalPages = [...pages];
		pages = pages.filter(p => p.id !== id);
		
		try {
			await api.deletePage(id);
			toast.show('Đã xóa page', 'success');
		} catch (error) {
			// Rollback nếu lỗi
			pages = originalPages;
			toast.show('Không thể xóa page', 'error');
		}
	}
</script>

<svelte:head>
	<title>Quản lý Pages - FB Scheduler</title>
</svelte:head>

{#if $toast}
	<Toast 
		message={$toast.message} 
		type={$toast.type} 
		duration={$toast.duration || 3000}
		onClose={() => toast.hide()} 
	/>
{/if}

<div class="max-w-6xl mx-auto">
	<!-- Header - Clean -->
	<div class="mb-6">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Quản lý Facebook Pages</h1>
				<p class="text-sm text-gray-500 mt-1">Kết nối và quản lý các trang Facebook của bạn</p>
			</div>
			<button
				on:click={connectFacebook}
				disabled={connectingFacebook || loading}
				class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{#if connectingFacebook}
					<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
					<span>Đang kết nối...</span>
				{:else}
					<Facebook size={18} />
					<span>Kết nối Facebook</span>
				{/if}
			</button>
		</div>
	</div>
	
	<!-- Content -->
	{#if loading}
		<div class="flex items-center justify-center min-h-[400px]">
			<div class="w-10 h-10 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
		</div>
	{:else if pages.length === 0}
		<div class="bg-gray-50 rounded-xl p-12 text-center border border-gray-200">
			<div class="w-16 h-16 bg-white rounded-full flex items-center justify-center mx-auto mb-4 border border-gray-200">
				<Facebook size={32} class="text-gray-400" />
			</div>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Chưa có page nào được kết nối</h3>
			<p class="text-sm text-gray-500 mb-6 max-w-md mx-auto">
				Kết nối Facebook Pages của bạn để bắt đầu đăng bài tự động và quản lý nội dung dễ dàng
			</p>
			<button
				on:click={connectFacebook}
				class="inline-flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				<Facebook size={18} />
				<span>Kết nối ngay</span>
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each pages as page}
				<div class="bg-white rounded-lg p-4 border border-gray-200 hover:border-gray-300 transition-colors">
					<div class="flex items-start gap-3 mb-3">
						<div class="relative flex-shrink-0">
							<img 
								src={page.profile_picture_url || 'https://via.placeholder.com/60'} 
								alt={page.page_name}
								class="w-12 h-12 rounded-full"
							/>
							{#if page.is_active}
								<div class="absolute -bottom-0.5 -right-0.5 w-4 h-4 bg-green-500 rounded-full border-2 border-white"></div>
							{:else}
								<div class="absolute -bottom-0.5 -right-0.5 w-4 h-4 bg-gray-400 rounded-full border-2 border-white"></div>
							{/if}
						</div>
						<div class="flex-1 min-w-0">
							<h3 class="font-medium text-sm text-gray-900 truncate">{page.page_name}</h3>
							<p class="text-xs text-gray-500 truncate mt-0.5">{page.category || 'Không rõ'}</p>
						</div>
					</div>
					
					<div class="mb-3">
						{#if page.is_active}
							<span class="inline-flex items-center gap-1.5 px-2 py-1 bg-green-50 text-green-700 rounded text-xs font-medium">
								<div class="w-1.5 h-1.5 bg-green-500 rounded-full"></div>
								Hoạt động
							</span>
						{:else}
							<span class="inline-flex items-center gap-1.5 px-2 py-1 bg-gray-100 text-gray-600 rounded text-xs font-medium">
								<div class="w-1.5 h-1.5 bg-gray-400 rounded-full"></div>
								Tạm dừng
							</span>
						{/if}
					</div>
					
					<div class="flex gap-2">
						<button
							on:click={() => togglePage(page.id)}
							class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg border text-xs font-medium transition-colors
								{page.is_active 
									? 'border-gray-200 text-gray-700 hover:bg-gray-50' 
									: 'border-blue-200 text-blue-700 hover:bg-blue-50'}"
						>
							{#if page.is_active}
								<PowerOff size={14} />
								<span>Tắt</span>
							{:else}
								<Power size={14} />
								<span>Bật</span>
							{/if}
						</button>
						<button
							on:click={() => deletePage(page.id)}
							class="px-3 py-2 rounded-lg border border-red-200 text-red-600 hover:bg-red-50 transition-colors"
							title="Xóa page"
						>
							<Trash2 size={14} />
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Modal chọn pages - Minimalist Design -->
{#if showPageSelectionModal}
	<!-- svelte-ignore a11y-click-events-have-key-events -->
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<div class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4 animate-fade-in" on:click={() => !loadingPages && (showPageSelectionModal = false)}>
		<div class="bg-white rounded-2xl shadow-xl max-w-2xl w-full max-h-[85vh] overflow-hidden animate-scale-in" on:click|stopPropagation>
			<!-- Header - Clean & Simple -->
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">
					{loadingPages ? 'Đang tải danh sách Pages...' : 'Chọn Pages muốn kết nối'}
				</h2>
				<p class="text-sm text-gray-500 mt-1">
					{loadingPages ? 'Vui lòng đợi trong giây lát' : 'Chọn các pages bạn muốn quản lý trong ứng dụng'}
				</p>
			</div>
			
			<div class="p-6 overflow-y-auto max-h-[calc(85vh-180px)]">
				{#if loadingPages}
					<!-- Loading state -->
					<div class="flex flex-col items-center justify-center py-16">
						<div class="w-12 h-12 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin mb-4"></div>
						<p class="text-sm text-gray-600 font-medium">Đang lấy danh sách Pages từ Facebook...</p>
						<p class="text-xs text-gray-500 mt-2">Thường mất 2-3 giây</p>
					</div>
				{:else if availablePages.length === 0}
					<div class="text-center py-12">
						<div class="w-12 h-12 bg-gray-50 rounded-full flex items-center justify-center mx-auto mb-3">
							<Facebook size={24} class="text-gray-400" />
						</div>
						<p class="text-sm text-gray-500">Không tìm thấy page nào</p>
					</div>
				{:else}
					<!-- Stats Bar - Minimal -->
					<div class="flex items-center justify-between mb-5 pb-4 border-b border-gray-100">
						<div class="flex items-center gap-2">
							<span class="text-sm text-gray-600">
								<span class="font-semibold text-gray-900">{selectedPageIds.size}</span> / {availablePages.length} đã chọn
							</span>
						</div>
						<div class="flex gap-2">
							<button 
								on:click={() => selectedPageIds = new Set(availablePages.map(p => p.page_id))}
								class="px-3 py-1.5 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
							>
								Chọn tất cả
							</button>
							<button 
								on:click={() => selectedPageIds = new Set()}
								class="px-3 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-50 rounded-lg transition-colors"
							>
								Bỏ chọn
							</button>
						</div>
					</div>
					
					<!-- Pages list - Clean Cards -->
					<div class="space-y-2">
						{#each availablePages as page}
							<label 
								class="flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-all duration-150
									{selectedPageIds.has(page.page_id)
										? 'border-blue-200 bg-blue-50/50'
										: 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'}"
							>
								<input 
									type="checkbox" 
									checked={selectedPageIds.has(page.page_id)}
									on:change={() => togglePageSelection(page.page_id)}
									class="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-2 focus:ring-blue-500 focus:ring-offset-0"
								/>
								<img 
									src={page.profile_picture_url || 'https://via.placeholder.com/50'} 
									alt={page.page_name}
									class="w-10 h-10 rounded-full"
								/>
								<div class="flex-1 min-w-0">
									<div class="font-medium text-sm text-gray-900 truncate">{page.page_name}</div>
									<div class="text-xs text-gray-500 truncate">{page.category || 'Không rõ'}</div>
								</div>
								{#if selectedPageIds.has(page.page_id)}
									<CheckCircle2 size={18} class="text-blue-600 flex-shrink-0" />
								{/if}
							</label>
						{/each}
					</div>
				{/if}
			</div>
			
			<!-- Footer - Minimal -->
			{#if !loadingPages}
				<div class="border-t border-gray-100 px-6 py-4 bg-gray-50/50 flex gap-3">
					<button 
						on:click={() => showPageSelectionModal = false}
						class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
					>
						Hủy
					</button>
					<button 
						on:click={saveSelectedPages}
						disabled={selectedPageIds.size === 0}
						class="flex-1 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Lưu ({selectedPageIds.size})
					</button>
				</div>
			{/if}
		</div>
	</div>
{/if}
