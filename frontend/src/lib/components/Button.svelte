<script>
	export let variant = 'primary'; // primary, secondary, danger
	export let disabled = false;
	export let loading = false;
	export let type = 'button';
	
	const variants = {
		primary: 'btn-primary',
		secondary: 'btn-secondary',
		danger: 'btn-danger'
	};
</script>

<button
	{type}
	class="btn {variants[variant]} {$$props.class || ''}"
	disabled={disabled || loading}
	on:click
>
	{#if loading}
		<span class="inline-block animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent mr-2"></span>
	{/if}
	<slot />
</button>
