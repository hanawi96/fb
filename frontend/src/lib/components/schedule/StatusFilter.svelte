<script>
	import { createEventDispatcher } from 'svelte';
	import { Clock, CheckCircle, XCircle } from 'lucide-svelte';
	
	export let value = '';
	export let counts = { all: 0, pending: 0, completed: 0, failed: 0 };
	
	const dispatch = createEventDispatcher();
	
	const statuses = [
		{ value: '', key: 'all', label: 'Tất cả', icon: null, color: 'gray' },
		{ value: 'pending', key: 'pending', label: 'Chờ đăng', icon: Clock, color: 'yellow' },
		{ value: 'completed', key: 'completed', label: 'Thành công', icon: CheckCircle, color: 'green' },
		{ value: 'failed', key: 'failed', label: 'Thất bại', icon: XCircle, color: 'red' }
	];
	
	function selectStatus(statusValue) {
		value = statusValue;
		dispatch('change', { value: statusValue });
	}
	
	function getButtonClass(status, isActive) {
		if (isActive) {
			switch (status.color) {
				case 'yellow': return 'bg-yellow-500 text-white border-yellow-500';
				case 'green': return 'bg-green-600 text-white border-green-600';
				case 'red': return 'bg-red-600 text-white border-red-600';
				default: return 'bg-blue-600 text-white border-blue-600';
			}
		}
		return 'bg-white text-gray-700 border-gray-200 hover:bg-gray-50';
	}
</script>

<div class="flex gap-2">
	{#each statuses as status}
		<button
			type="button"
			on:click={() => selectStatus(status.value)}
			class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg border transition-colors {getButtonClass(status, value === status.value)}"
		>
			{#if status.icon}
				<svelte:component this={status.icon} size={14} />
			{/if}
			<span>{status.label}</span>
			<span class="text-xs opacity-75">({counts[status.key] || 0})</span>
		</button>
	{/each}
</div>
