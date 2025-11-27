<script>
	import { Globe, Lock, Users } from 'lucide-svelte';
	
	export let privacy = 'public'; // 'public', 'private', 'friends'
	
	const privacyOptions = [
		{ value: 'public', label: 'Công khai', icon: Globe, desc: 'Ai cũng có thể xem' },
		{ value: 'private', label: 'Riêng tư', icon: Lock, desc: 'Chỉ mình tôi' }
	];
</script>

<div class="border-t border-gray-100 p-3">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-2">
			<div class="w-1.5 h-1.5 rounded-full bg-blue-600"></div>
			<span class="text-sm font-medium text-gray-900">Quyền riêng tư</span>
		</div>
		
		<div class="flex items-center gap-3">
			{#each privacyOptions as option}
				<label class="flex items-center gap-2 cursor-pointer group">
					<input 
						type="radio" 
						name="privacy"
						value={option.value}
						bind:group={privacy}
						class="w-4 h-4 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500"
					/>
					<div class="flex items-center gap-1.5">
						<svelte:component this={option.icon} size={14} class="text-gray-600 group-hover:text-gray-900" />
						<span class="text-sm text-gray-700 group-hover:text-gray-900">{option.label}</span>
					</div>
				</label>
			{/each}
		</div>
	</div>
</div>
