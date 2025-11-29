<script>
	import { createEventDispatcher } from 'svelte';
	import { Trash2, Calendar, X, CheckSquare } from 'lucide-svelte';
	
	export let selectedCount = 0;
	export let totalCount = 0;
	
	const dispatch = createEventDispatcher();
	
	function handleSelectAll() {
		dispatch('selectAll');
	}
	
	function handleClearSelection() {
		dispatch('clearSelection');
	}
	
	function handleDelete() {
		dispatch('bulkDelete');
	}
	
	function handleReschedule() {
		dispatch('bulkReschedule');
	}
</script>

{#if selectedCount > 0}
	<div class="flex items-center justify-between px-3 py-2 bg-blue-50 border border-blue-200 rounded-lg mb-3">
		<div class="flex items-center gap-3">
			<div class="flex items-center gap-2">
				<CheckSquare size={16} class="text-blue-600" />
				<span class="text-sm font-medium text-blue-800">
					{selectedCount} đã chọn
				</span>
			</div>
			
			{#if selectedCount < totalCount}
				<button
					type="button"
					on:click={handleSelectAll}
					class="text-xs text-blue-600 hover:text-blue-700 hover:underline"
				>
					Chọn tất cả ({totalCount})
				</button>
			{/if}
		</div>
		
		<div class="flex items-center gap-2">
			<button
				type="button"
				on:click={handleReschedule}
				class="flex items-center gap-1 px-2.5 py-1 bg-white border border-gray-200 text-gray-700 text-xs font-medium rounded hover:bg-gray-50 transition-colors"
			>
				<Calendar size={12} />
				Đổi lịch
			</button>
			
			<button
				type="button"
				on:click={handleDelete}
				class="flex items-center gap-1 px-2.5 py-1 bg-red-600 text-white text-xs font-medium rounded hover:bg-red-700 transition-colors"
			>
				<Trash2 size={12} />
				Xóa
			</button>
			
			<button
				type="button"
				on:click={handleClearSelection}
				class="p-1 text-gray-400 hover:text-gray-600 transition-colors"
				title="Bỏ chọn"
			>
				<X size={14} />
			</button>
		</div>
	</div>
{/if}
