<script>
	import { X, ThumbsUp, MessageCircle, Share2 } from 'lucide-svelte';
	
	export let show = false;
	export let content = '';
	export let images = [];
	export let videos = [];
	export let linkPreview = null;
	export let selectedPages = [];
	
	function close() {
		show = false;
	}
	
	function formatContent(text) {
		return text.replace(/\n/g, '<br>');
	}
</script>

{#if show}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50" on:click={close}>
		<div class="bg-white rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-hidden" on:click|stopPropagation>
			<!-- Header -->
			<div class="flex items-center justify-between px-4 py-3 border-b border-gray-200">
				<h3 class="text-lg font-semibold text-gray-900">Xem trước bài viết</h3>
				<button on:click={close} class="p-1 hover:bg-gray-100 rounded-full transition-colors">
					<X size={20} class="text-gray-600" />
				</button>
			</div>
			
			<!-- Content -->
			<div class="overflow-y-auto max-h-[calc(90vh-120px)]">
				{#each selectedPages as page}
					<div class="p-4 border-b border-gray-100">
						<!-- Post Header -->
						<div class="flex items-center gap-3 mb-3">
							<img 
								src={page.profile_picture_url || 'https://via.placeholder.com/40'} 
								alt={page.page_name}
								class="w-10 h-10 rounded-full"
							/>
							<div>
								<h4 class="text-sm font-semibold text-gray-900">{page.page_name}</h4>
								<p class="text-xs text-gray-500">Vừa xong · 🌍</p>
							</div>
						</div>
						
						<!-- Post Content -->
						{#if content}
							<div class="text-sm text-gray-900 mb-3 whitespace-pre-wrap">
								{@html formatContent(content)}
							</div>
						{/if}
						
						<!-- Media -->
						{#if images.length > 0}
							<div class="grid gap-1 mb-3 {images.length === 1 ? 'grid-cols-1' : images.length === 2 ? 'grid-cols-2' : images.length === 3 ? 'grid-cols-3' : 'grid-cols-2'}">
								{#each images.slice(0, 4) as image, i}
									<div class="relative aspect-square bg-gray-100 rounded overflow-hidden">
										<img src={image} alt="" class="w-full h-full object-cover" />
										{#if i === 3 && images.length > 4}
											<div class="absolute inset-0 bg-black/60 flex items-center justify-center">
												<span class="text-white text-2xl font-bold">+{images.length - 4}</span>
											</div>
										{/if}
									</div>
								{/each}
							</div>
						{/if}
						
						{#if videos.length > 0}
							<div class="mb-3">
								<video src={videos[0]} controls class="w-full rounded-lg" />
							</div>
						{/if}
						
						{#if linkPreview}
							<div class="border border-gray-200 rounded-lg overflow-hidden mb-3">
								{#if linkPreview.image}
									<img src={linkPreview.image} alt="" class="w-full h-48 object-cover" />
								{/if}
								<div class="p-3 bg-gray-50">
									{#if linkPreview.title}
										<h5 class="text-sm font-semibold text-gray-900 mb-1">{linkPreview.title}</h5>
									{/if}
									{#if linkPreview.description}
										<p class="text-xs text-gray-600 line-clamp-2">{linkPreview.description}</p>
									{/if}
								</div>
							</div>
						{/if}
						
						<!-- Engagement -->
						<div class="flex items-center justify-between py-2 border-t border-gray-200">
							<button class="flex items-center gap-2 px-4 py-2 hover:bg-gray-100 rounded-lg transition-colors">
								<ThumbsUp size={18} class="text-gray-600" />
								<span class="text-sm font-medium text-gray-700">Thích</span>
							</button>
							<button class="flex items-center gap-2 px-4 py-2 hover:bg-gray-100 rounded-lg transition-colors">
								<MessageCircle size={18} class="text-gray-600" />
								<span class="text-sm font-medium text-gray-700">Bình luận</span>
							</button>
							<button class="flex items-center gap-2 px-4 py-2 hover:bg-gray-100 rounded-lg transition-colors">
								<Share2 size={18} class="text-gray-600" />
								<span class="text-sm font-medium text-gray-700">Chia sẻ</span>
							</button>
						</div>
					</div>
				{/each}
			</div>
			
			<!-- Footer -->
			<div class="px-4 py-3 border-t border-gray-200 bg-gray-50">
				<button 
					on:click={close}
					class="w-full px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
				>
					Đóng
				</button>
			</div>
		</div>
	</div>
{/if}
