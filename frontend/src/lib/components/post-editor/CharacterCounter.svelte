<script>
	export let count = 0;
	const maxLength = 63206; // Facebook post limit
	
	$: percentage = (count / maxLength) * 100;
	$: isWarning = percentage > 80;
	$: isDanger = percentage > 95;
</script>

<div class="flex items-center gap-2">
	<div class="text-xs {isDanger ? 'text-red-600 font-semibold' : isWarning ? 'text-orange-600 font-medium' : 'text-gray-500'}">
		{count.toLocaleString()} / {maxLength.toLocaleString()}
	</div>
	
	{#if percentage > 50}
		<div class="w-16 h-1.5 bg-gray-200 rounded-full overflow-hidden">
			<div 
				class="h-full transition-all duration-300 {isDanger ? 'bg-red-500' : isWarning ? 'bg-orange-500' : 'bg-blue-500'}"
				style="width: {Math.min(percentage, 100)}%"
			></div>
		</div>
	{/if}
</div>
