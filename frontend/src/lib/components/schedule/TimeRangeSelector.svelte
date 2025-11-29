<script>
	import { createEventDispatcher } from 'svelte';
	import { Calendar, ChevronDown } from 'lucide-svelte';
	
	export let value = 'all'; // 'all', 'today', 'tomorrow', 'week', 'month', 'past', 'custom'
	export let customStart = '';
	export let customEnd = '';
	
	const dispatch = createEventDispatcher();
	
	let isOpen = false;
	let dropdownRef;
	
	const options = [
		{ value: 'all', label: 'Tất cả' },
		{ value: 'today', label: 'Hôm nay' },
		{ value: 'tomorrow', label: 'Ngày mai' },
		{ value: 'week', label: '7 ngày tới' },
		{ value: 'month', label: '30 ngày tới' },
		{ value: 'past', label: 'Đã qua' }
	];
	
	$: selectedLabel = options.find(o => o.value === value)?.label || 'Tất cả';
	
	function selectOption(optionValue) {
		value = optionValue;
		dispatch('change', { value: optionValue, customStart, customEnd });
		if (optionValue !== 'custom') {
			isOpen = false;
		}
	}
	
	function handleCustomChange() {
		if (customStart && customEnd) {
			value = 'custom';
			dispatch('change', { value: 'custom', customStart, customEnd });
		}
	}
	
	function handleClickOutside(event) {
		if (dropdownRef && !dropdownRef.contains(event.target)) {
			isOpen = false;
		}
	}
</script>

<svelte:window on:click={handleClickOutside} />

<div class="relative" bind:this={dropdownRef}>
	<!-- Trigger Button -->
	<button
		type="button"
		on:click|stopPropagation={() => isOpen = !isOpen}
		class="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 rounded-lg text-sm hover:border-gray-300 transition-colors min-w-[140px]"
	>
		<Calendar size={16} class="text-gray-400" />
		<span class="flex-1 text-left text-gray-700">{selectedLabel}</span>
		<ChevronDown size={16} class="text-gray-400 transition-transform {isOpen ? 'rotate-180' : ''}" />
	</button>
	
	<!-- Dropdown -->
	{#if isOpen}
		<div 
			class="absolute top-full left-0 mt-1 w-56 bg-white border border-gray-200 rounded-lg shadow-lg z-50 overflow-hidden"
			on:click|stopPropagation
		>
			<!-- Options -->
			<div class="py-1">
				{#each options as option}
					<button
						type="button"
						on:click={() => selectOption(option.value)}
						class="w-full flex items-center gap-2 px-3 py-2 text-sm text-left hover:bg-gray-50 transition-colors
							{value === option.value ? 'text-blue-600 bg-blue-50' : 'text-gray-700'}"
					>
						<span class="w-4 h-4 flex items-center justify-center">
							{#if value === option.value}
								<span class="w-2 h-2 rounded-full bg-blue-600"></span>
							{/if}
						</span>
						{option.label}
					</button>
				{/each}
			</div>
			
			<!-- Custom Range -->
			<div class="border-t border-gray-100 p-3">
				<div class="text-xs text-gray-500 mb-2">Tùy chọn khoảng thời gian</div>
				<div class="flex items-center gap-2">
					<input
						type="date"
						bind:value={customStart}
						on:change={handleCustomChange}
						class="flex-1 px-2 py-1.5 text-xs border border-gray-200 rounded focus:outline-none focus:border-blue-500"
					/>
					<span class="text-gray-400">→</span>
					<input
						type="date"
						bind:value={customEnd}
						on:change={handleCustomChange}
						class="flex-1 px-2 py-1.5 text-xs border border-gray-200 rounded focus:outline-none focus:border-blue-500"
					/>
				</div>
			</div>
		</div>
	{/if}
</div>
