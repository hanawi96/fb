<script>
	import { Facebook } from 'lucide-svelte';
	
	export let selectedPages = [];
</script>

<div class="bg-white border border-gray-200 rounded-lg overflow-y-auto h-full">
	<div class="p-3">
		<h3 class="text-xs font-semibold text-gray-700 uppercase mb-3">Thông tin</h3>
		
		<!-- Selected Pages Info -->
		{#if selectedPages.length > 0}
			<div class="mb-4">
				<div class="text-xs text-gray-500 mb-2">ĐĂNG BÀI TỚI ({selectedPages.length})</div>
				<div class="space-y-2">
					{#each selectedPages as page}
						<div class="flex items-center gap-2 p-2 bg-gray-50 rounded-lg">
							<img 
								src={page.profile_picture_url || 'https://via.placeholder.com/40'} 
								alt={page.page_name}
								class="w-8 h-8 rounded-full"
							/>
							<div class="flex-1 min-w-0">
								<div class="text-xs font-medium text-gray-900 truncate">{page.page_name}</div>
								<div class="flex items-center gap-1 text-xs text-gray-500">
									<Facebook size={9} />
									<span>Facebook</span>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{:else}
			<div class="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
				<p class="text-xs text-yellow-800">⚠️ Chưa chọn page nào</p>
			</div>
		{/if}
		
		<!-- Stats -->
		<div class="mb-4">
			<div class="text-xs text-gray-500 mb-2">THỐNG KÊ</div>
			<div class="space-y-2">
				<div class="flex items-center justify-between">
					<span class="text-xs text-gray-700">Lượt thích</span>
					<span class="text-xs font-semibold text-gray-900">0</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-xs text-gray-700">Bình luận</span>
					<span class="text-xs font-semibold text-gray-900">0</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-xs text-gray-700">Chia sẻ</span>
					<span class="text-xs font-semibold text-gray-900">0</span>
				</div>
			</div>
		</div>
		
		<!-- Tips -->
		<div class="bg-blue-50 border border-blue-200 rounded-lg p-3">
			<h4 class="text-xs font-semibold text-blue-900 mb-2">💡 Mẹo viết bài</h4>
			<ul class="text-xs text-blue-800 space-y-1.5">
				<li>• Sử dụng emoji để thu hút</li>
				<li>• Đặt câu hỏi tăng tương tác</li>
				<li>• Thêm hashtag phù hợp</li>
				<li>• Đăng giờ vàng (18h-21h)</li>
			</ul>
		</div>
	</div>
</div>
