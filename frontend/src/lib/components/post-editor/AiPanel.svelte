<script>
	import { Sparkles } from 'lucide-svelte';
	
	export let show = false;
	export let prompt = '';
	export let generating = false;
	export let onGenerate = () => {};
	export let onClose = () => {};
</script>

{#if show}
	<div class="mb-4 bg-gradient-to-r from-purple-50 to-blue-50 border border-purple-200 rounded-xl p-4">
		<div class="flex items-start gap-3">
			<div class="w-10 h-10 bg-gradient-to-br from-purple-500 to-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
				<Sparkles size={20} class="text-white" />
			</div>
			<div class="flex-1">
				<h3 class="text-sm font-semibold text-gray-900 mb-2">Trợ lý AI viết bài</h3>
				<textarea
					bind:value={prompt}
					placeholder="Ví dụ: Viết bài giới thiệu sản phẩm mới, tone vui vẻ, có emoji..."
					rows="3"
					class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent resize-none"
				></textarea>
				<div class="flex gap-2 mt-3">
					<button
						on:click={onGenerate}
						disabled={generating}
						class="px-4 py-2 bg-gradient-to-r from-purple-600 to-blue-600 text-white text-sm font-medium rounded-lg hover:from-purple-700 hover:to-blue-700 transition-all disabled:opacity-50"
					>
						{#if generating}
							<div class="flex items-center gap-2">
								<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
								<span>Đang tạo...</span>
							</div>
						{:else}
							Tạo nội dung
						{/if}
					</button>
					<button
						on:click={onClose}
						class="px-4 py-2 text-sm text-gray-700 hover:bg-white rounded-lg transition-colors"
					>
						Hủy
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
