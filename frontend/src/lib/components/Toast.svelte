<script>
	import { onMount, onDestroy } from 'svelte';
	import { fly } from 'svelte/transition';
	import { CheckCircle, XCircle, AlertCircle, X } from 'lucide-svelte';
	
	export let message = '';
	export let type = 'success'; // success, error, warning
	export let onClose = () => {};
	export let duration = 3000; // Thời gian hiển thị (ms)
	
	const icons = {
		success: CheckCircle,
		error: XCircle,
		warning: AlertCircle
	};
	
	const colors = {
		success: 'bg-gradient-to-r from-green-600 to-emerald-600 border-green-700 text-white',
		error: 'bg-gradient-to-r from-red-600 to-rose-600 border-red-700 text-white',
		warning: 'bg-gradient-to-r from-yellow-500 to-amber-500 border-yellow-600 text-white'
	};
	
	const iconColors = {
		success: 'text-white',
		error: 'text-white',
		warning: 'text-white'
	};
	
	const progressColors = {
		success: 'bg-white bg-opacity-30',
		error: 'bg-white bg-opacity-30',
		warning: 'bg-white bg-opacity-30'
	};
	
	let progress = 100;
	let isPaused = false;
	let timeoutId = null;
	let intervalId = null;
	let remainingTime = duration;
	let lastTime = Date.now();
	
	function startTimer() {
		if (duration <= 0) return;
		
		lastTime = Date.now();
		
		// Update progress bar mỗi 50ms
		intervalId = setInterval(() => {
			if (isPaused) return;
			
			const now = Date.now();
			const elapsed = now - lastTime;
			remainingTime = Math.max(0, remainingTime - elapsed);
			lastTime = now;
			
			progress = (remainingTime / duration) * 100;
			
			if (remainingTime <= 0) {
				clearTimer();
				onClose();
			}
		}, 50);
	}
	
	function clearTimer() {
		if (intervalId) {
			clearInterval(intervalId);
			intervalId = null;
		}
		if (timeoutId) {
			clearTimeout(timeoutId);
			timeoutId = null;
		}
	}
	
	function pauseTimer() {
		isPaused = true;
	}
	
	function resumeTimer() {
		if (isPaused) {
			isPaused = false;
			lastTime = Date.now();
		}
	}
	
	onMount(() => {
		startTimer();
	});
	
	onDestroy(() => {
		clearTimer();
	});
</script>

<div
	class="fixed bottom-6 right-6 z-[100] max-w-sm"
	transition:fly={{ y: 50, duration: 400 }}
	on:mouseenter={pauseTimer}
	on:mouseleave={resumeTimer}
	role="alert"
>
	<div class="relative overflow-hidden rounded-lg border shadow-xl {colors[type]}">
		<div class="flex items-center gap-2.5 px-3.5 py-2.5">
			<div class="flex-shrink-0 {iconColors[type]}">
				<svelte:component this={icons[type]} size={18} strokeWidth={2.5} />
			</div>
			<p class="flex-1 text-sm font-semibold">{message}</p>
			<button 
				on:click={onClose} 
				class="flex-shrink-0 p-0.5 rounded hover:bg-white hover:bg-opacity-20 transition-all duration-200"
				aria-label="Đóng thông báo"
			>
				<X size={16} strokeWidth={2.5} />
			</button>
		</div>
		
		{#if duration > 0}
			<div class="absolute bottom-0 left-0 right-0 h-1 bg-black bg-opacity-10">
				<div 
					class="h-full {progressColors[type]}"
					style="width: {progress}%; transition: width 50ms linear;"
				/>
			</div>
		{/if}
	</div>
</div>
