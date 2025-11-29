<script>
	import { Image as ImageIcon, X, Upload, GripVertical } from 'lucide-svelte';
	
	export let images = []; // Array of image URLs (strings)
	export let uploading = false;
	export let onUpload = () => {};
	export let onRemove = () => {};
	export let postMode = 'album'; // 'album' | 'individual'
	
	const MAX_IMAGES = 10;
	
	// Drag & drop state
	let draggedIndex = null;
	
	function handleDragStart(index) {
		draggedIndex = index;
	}
	
	function handleDragOver(e, index) {
		e.preventDefault();
		if (draggedIndex === null || draggedIndex === index) return;
		
		// Reorder images
		const newImages = [...images];
		const draggedItem = newImages[draggedIndex];
		newImages.splice(draggedIndex, 1);
		newImages.splice(index, 0, draggedItem);
		images = newImages;
		draggedIndex = index;
	}
	
	function handleDragEnd() {
		draggedIndex = null;
	}
</script>

{#if images.length > 0 || uploading}
	<!-- Counter + Post Mode -->
	<div class="mt-2 flex items-center justify-between mb-2">
		<span class="text-xs font-medium text-gray-600">
			📊 {images.length}/{MAX_IMAGES} ảnh
			{#if images.length > 0}
				<span class="text-blue-600">• Ảnh đầu = Cover</span>
			{/if}
		</span>
		
		{#if images.length >= 2}
			<div class="flex flex-col gap-1">
				<div class="flex gap-3 text-xs">
					<label class="flex items-center gap-1.5 cursor-pointer group">
						<input 
							type="radio" 
							bind:group={postMode} 
							value="album"
							class="w-3.5 h-3.5 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500"
						/>
						<span class="text-gray-700 group-hover:text-gray-900 font-medium">
							Album (1 post)
						</span>
					</label>
					<label class="flex items-center gap-1.5 cursor-pointer group">
						<input 
							type="radio" 
							bind:group={postMode} 
							value="individual"
							class="w-3.5 h-3.5 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500"
						/>
						<span class="text-gray-700 group-hover:text-gray-900 font-medium">
							Riêng lẻ ({images.length} posts)
						</span>
					</label>
				</div>

			</div>
		{/if}
	</div>
	
	<!-- Draggable Image Grid -->
	<div class="flex flex-wrap gap-1">
		{#each images as image, index (image)}
			<div 
				class="relative group w-20 h-20 cursor-move"
				draggable="true"
				on:dragstart={() => handleDragStart(index)}
				on:dragover={(e) => handleDragOver(e, index)}
				on:dragend={handleDragEnd}
			>
				<!-- Order Badge -->
				<div class="absolute top-1 left-1 w-5 h-5 bg-blue-600 text-white text-xs rounded-full flex items-center justify-center font-bold z-10 shadow-sm">
					{index + 1}
				</div>
				
				<!-- Drag Handle -->
				<div class="absolute top-1 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity z-10">
					<GripVertical size={16} class="text-white drop-shadow-md" />
				</div>
				
				<img src={image} alt="Preview {index + 1}" class="w-full h-full object-cover rounded" />
				
				<!-- Remove Button -->
				<button
					on:click={() => onRemove(index)}
					class="absolute top-0.5 right-0.5 p-0.5 bg-red-600 text-white rounded-full opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-700"
				>
					<X size={12} />
				</button>
			</div>
		{/each}
		
		{#if images.length < MAX_IMAGES}
			<label class="w-20 h-20 border-2 border-dashed border-gray-300 rounded flex items-center justify-center cursor-pointer hover:border-blue-500 hover:bg-blue-50 transition-colors">
				<input
					type="file"
					accept="image/*"
					multiple
					on:change={onUpload}
					class="hidden"
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
	<label class="mt-3 border-2 border-dashed border-gray-300 rounded-lg p-6 flex flex-col items-center justify-center cursor-pointer hover:border-blue-500 hover:bg-blue-50 transition-colors">
		<input
			type="file"
			accept="image/*"
			multiple
			on:change={onUpload}
			class="hidden"
		/>
		<div class="w-10 h-10 bg-gray-100 rounded-full flex items-center justify-center mb-2">
			<ImageIcon size={20} class="text-gray-400" />
		</div>
		<p class="text-sm font-medium text-gray-900 mb-0.5">Thêm ảnh</p>
		<p class="text-xs text-gray-500">hoặc kéo thả</p>
	</label>
{/if}
