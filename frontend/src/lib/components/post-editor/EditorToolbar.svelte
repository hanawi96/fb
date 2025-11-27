<script>
	import { Sparkles, Image, Video, Smile, MapPin, Tag } from 'lucide-svelte';
	
	export let onAiClick = () => {};
</script>

<div class="border-b border-gray-100 p-3 flex items-center gap-2">
	<button
		on:click={onAiClick}
		class="flex items-center gap-2 px-3 py-2 text-sm font-medium text-purple-700 bg-purple-50 hover:bg-purple-100 rounded-lg transition-colors"
	>
		<Sparkles size={16} />
		<span>Viết với AI</span>
	</button>
	
	<div class="w-px h-6 bg-gray-200"></div>
	
	<button class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" title="Thêm ảnh">
		<Image size={20} />
	</button>
	<button class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" title="Thêm video">
		<Video size={20} />
	</button>
	<button class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" title="Thêm emoji">
		<Smile size={20} />
	</button>
	<button class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" title="Thêm vị trí">
		<MapPin size={20} />
	</button>
	<button class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" title="Gắn thẻ">
		<Tag size={20} />
	</button>
</div>
