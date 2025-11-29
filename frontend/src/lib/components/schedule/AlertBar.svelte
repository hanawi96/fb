<script>
	import { createEventDispatcher } from 'svelte';
	import { AlertTriangle, ArrowRight, X } from 'lucide-svelte';
	
	export let failedCount = 0;
	export let dismissible = true;
	
	const dispatch = createEventDispatcher();
	
	let dismissed = false;
	
	function handleViewFailed() {
		dispatch('viewFailed');
	}
	
	function dismiss() {
		dismissed = true;
	}
	
	$: visible = failedCount > 0 && !dismissed;
</script>

{#if visible}
	<div class="flex items-center justify-between px-3 py-2 bg-red-50 border border-red-200 rounded-lg mb-3">
		<div class="flex items-center gap-2">
			<AlertTriangle size={16} class="text-red-600" />
			<p class="text-sm text-red-800">
				<span class="font-medium">{failedCount}</span> bài thất bại cần xử lý
			</p>
		</div>
		
		<div class="flex items-center gap-2">
			<button
				type="button"
				on:click={handleViewFailed}
				class="flex items-center gap-1 px-2.5 py-1 bg-red-600 text-white text-xs font-medium rounded hover:bg-red-700 transition-colors"
			>
				Xem ngay
				<ArrowRight size={12} />
			</button>
			
			{#if dismissible}
				<button
					type="button"
					on:click={dismiss}
					class="p-1 text-red-400 hover:text-red-600 transition-colors"
				>
					<X size={14} />
				</button>
			{/if}
		</div>
	</div>
{/if}
