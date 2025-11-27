<script>
	import { X, ExternalLink } from 'lucide-svelte';
	
	export let preview = null;
	export let onRemove = () => {};
	
	$: hasPreview = preview && (preview.title || preview.image);
</script>

{#if hasPreview}
	<div class="mt-3 border border-gray-200 rounded-lg overflow-hidden hover:border-gray-300 transition-colors">
		<div class="relative group">
			{#if preview.image}
				<img 
					src={preview.image} 
					alt={preview.title}
					class="w-full h-48 object-cover"
				/>
			{/if}
			<button
				on:click={onRemove}
				class="absolute top-2 right-2 p-1.5 bg-gray-900/80 hover:bg-gray-900 text-white rounded-full transition-colors opacity-0 group-hover:opacity-100"
			>
				<X size={16} />
			</button>
		</div>
		<div class="p-3 bg-gray-50">
			<div class="flex items-start gap-2">
				<div class="flex-1 min-w-0">
					{#if preview.title}
						<h4 class="text-sm font-semibold text-gray-900 line-clamp-2 mb-1">
							{preview.title}
						</h4>
					{/if}
					{#if preview.description}
						<p class="text-xs text-gray-600 line-clamp-2 mb-2">
							{preview.description}
						</p>
					{/if}
					{#if preview.url}
						<a 
							href={preview.url} 
							target="_blank" 
							rel="noopener noreferrer"
							class="text-xs text-blue-600 hover:text-blue-700 flex items-center gap-1"
						>
							<ExternalLink size={10} />
							<span class="truncate">{new URL(preview.url).hostname}</span>
						</a>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}
