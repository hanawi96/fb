<script>
	import { Sparkles, Smile, Hash } from 'lucide-svelte';
	import EmojiPicker from './EmojiPicker.svelte';
	import HashtagSuggester from './HashtagSuggester.svelte';
	
	export let onAiClick = () => {};
	export let onEmojiSelect = (emoji) => {};
	export let onHashtagSelect = (hashtag) => {};
	
	let showEmojiPicker = false;
	let showHashtagSuggester = false;
	let emojiButtonElement;
	let hashtagButtonElement;
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
	
	<!-- Emoji Picker -->
	<button 
		bind:this={emojiButtonElement}
		on:click={() => {
			showEmojiPicker = !showEmojiPicker;
			if (showEmojiPicker) showHashtagSuggester = false;
		}}
		class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors {showEmojiPicker ? 'bg-blue-50 text-blue-600' : ''}" 
		title="Thêm emoji"
	>
		<Smile size={20} />
	</button>
	
	<!-- Hashtag -->
	<button 
		bind:this={hashtagButtonElement}
		on:click={() => {
			showHashtagSuggester = !showHashtagSuggester;
			if (showHashtagSuggester) showEmojiPicker = false;
		}}
		class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors {showHashtagSuggester ? 'bg-blue-50 text-blue-600' : ''}" 
		title="Thêm hashtag"
	>
		<Hash size={20} />
	</button>
</div>

<!-- Emoji Picker Popup -->
{#if showEmojiPicker}
	<EmojiPicker 
		show={showEmojiPicker}
		buttonElement={emojiButtonElement}
		onSelect={onEmojiSelect}
		onClose={() => showEmojiPicker = false}
	/>
{/if}

<!-- Hashtag Suggester Popup -->
{#if showHashtagSuggester}
	<HashtagSuggester 
		show={showHashtagSuggester}
		buttonElement={hashtagButtonElement}
		onSelect={onHashtagSelect}
		onClose={() => showHashtagSuggester = false}
	/>
{/if}
