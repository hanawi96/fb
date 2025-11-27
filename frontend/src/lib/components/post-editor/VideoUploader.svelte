<script>
	import { X, Video, Upload } from 'lucide-svelte';
	
	export let videos = [];
	export let uploading = false;
	export let onUpload = () => {};
	export let onRemove = () => {};
	
	let dragOver = false;
	
	function handleDrop(e) {
		dragOver = false;
		const files = Array.from(e.dataTransfer.files).filter(f => f.type.startsWith('video/'));
		if (files.length > 0) {
			onUpload({ target: { files } });
		}
	}
</script>

{#if videos.length > 0}
	<div class="mt-2 flex flex-wrap gap-1">
		{#each videos as video, i}
			<div class="relative group w-32 h-24 rounded overflow-hidden bg-gray-900">
				<video 
					src={video} 
					class="w-full h-full object-cover"
				/>
				<div class="absolute inset-0 bg-black/30 flex items-center justify-center">
					<div class="w-8 h-8 bg-white/90 rounded-full flex items-center justify-center">
						<svg class="w-4 h-4 text-gray-900 ml-0.5" fill="currentColor" viewBox="0 0 24 24">
							<path d="M8 5v14l11-7z"/>
						</svg>
					</div>
				</div>
				<button
					on:click={() => onRemove(i)}
					class="absolute top-1 right-1 p-0.5 bg-gray-900/80 text-white rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
				>
					<X size={12} />
				</button>
			</div>
		{/each}
		
		{#if videos.length < 1}
			<label class="w-32 h-24 border-2 border-dashed border-gray-300 rounded flex flex-col items-center justify-center cursor-pointer hover:border-blue-500 hover:bg-blue-50 transition-colors">
				<input
					type="file"
					accept="video/*"
					class="hidden"
					on:change={onUpload}
					disabled={uploading}
				/>
				{#if uploading}
					<div class="w-5 h-5 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
				{:else}
					<Upload size={18} class="text-gray-400" />
				{/if}
			</label>
		{/if}
	</div>
{:else}
	<div
		class="mt-3 border-2 border-dashed rounded-lg transition-colors {dragOver ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'}"
		on:dragover|preventDefault={() => dragOver = true}
		on:dragleave={() => dragOver = false}
		on:drop|preventDefault={handleDrop}
	>
		<label class="flex flex-col items-center justify-center py-6 cursor-pointer">
			<div class="w-10 h-10 bg-gray-100 rounded-full flex items-center justify-center mb-2">
				{#if uploading}
					<div class="w-5 h-5 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
				{:else}
					<Video size={20} class="text-gray-600" />
				{/if}
			</div>
			<p class="text-sm font-medium text-gray-900 mb-0.5">
				{uploading ? 'Đang tải video...' : 'Thêm video'}
			</p>
			<p class="text-xs text-gray-500">Kéo thả hoặc click</p>
			<input
				type="file"
				accept="video/*"
				class="hidden"
				on:change={onUpload}
				disabled={uploading}
			/>
		</label>
	</div>
{/if}
