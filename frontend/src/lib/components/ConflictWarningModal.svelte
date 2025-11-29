<script>
	import { createEventDispatcher } from 'svelte';
	import { AlertTriangle, X, Clock, Zap } from 'lucide-svelte';
	
	export let conflictPages = [];
	export let scheduledTime = '';
	export let show = false;
	
	const dispatch = createEventDispatcher();
	
	function handleConfirm() {
		dispatch('confirm');
	}
	
	function handleCancel() {
		dispatch('cancel');
	}
	
	function handleAutoSchedule() {
		dispatch('autoSchedule');
	}
	
	function formatDateTime(dateTimeStr) {
		if (!dateTimeStr) return '';
		const date = new Date(dateTimeStr);
		return date.toLocaleString('vi-VN', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

{#if show}
	<!-- Backdrop -->
	<div class="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
		<!-- Modal -->
		<div class="bg-white rounded-lg shadow-xl max-w-2xl w-full animate-scale-in">
			<!-- Header -->
			<div class="flex items-center gap-4 p-6 border-b border-gray-200">
				<div class="w-12 h-12 rounded-full bg-yellow-100 flex items-center justify-center flex-shrink-0">
					<AlertTriangle size={24} class="text-yellow-600" />
				</div>
				<div class="flex-1">
					<h3 class="text-xl font-semibold text-gray-900">Xung đột thời gian</h3>
					<p class="text-base text-gray-500">Khung giờ đã có bài đăng</p>
				</div>
				<button
					on:click={handleCancel}
					class="text-gray-400 hover:text-gray-600 transition-colors"
				>
					<X size={20} />
				</button>
			</div>
			
			<!-- Content -->
			<div class="p-8 space-y-5">
				<!-- Main Conflict Message -->
				<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-5">
					<p class="text-lg text-gray-900 leading-relaxed">
						Khung giờ <span class="font-semibold text-yellow-800">{formatDateTime(scheduledTime)}</span>
						{#if conflictPages.length === 1}
							page <span class="font-semibold text-yellow-800">{conflictPages[0].page_name}</span> đã có 1 bài đăng được lên lịch trước rồi.
						{:else}
							đã có bài đăng được lên lịch trước rồi cho các page:
						{/if}
					</p>
					
					{#if conflictPages.length > 1}
						<ul class="mt-4 space-y-2">
							{#each conflictPages as page}
								<li class="flex items-center gap-3 px-4 py-3 bg-white rounded border border-yellow-200">
									<div class="w-2.5 h-2.5 rounded-full bg-red-500 flex-shrink-0"></div>
									<span class="text-base text-gray-900 font-semibold">{page.page_name}</span>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
				
				<!-- Instructions -->
				<div class="text-base text-gray-600 bg-gray-50 rounded-lg p-5">
					<p class="font-medium text-gray-700 mb-3">Bạn có thể:</p>
					<ul class="space-y-2">
						<li>• Chọn thời gian khác</li>
						<li>• Bỏ chọn các page trên</li>
						<li>• Cho phép đăng trùng giờ</li>
						<li>• Để hệ thống tự động xếp lịch</li>
					</ul>
				</div>
			</div>
			
			<!-- Actions -->
			<div class="flex items-center gap-3 p-6 border-t border-gray-200 bg-gray-50">
				<button
					on:click={handleCancel}
					class="px-5 py-3 text-base font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors whitespace-nowrap"
				>
					Chọn lại
				</button>
				<button
					on:click={handleAutoSchedule}
					class="flex-1 px-3 py-3 text-sm font-medium text-blue-700 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 transition-colors flex items-center justify-center gap-1.5 whitespace-nowrap"
				>
					<Zap size={14} />
					Chuyển vào lịch tự động
				</button>
				<button
					on:click={handleConfirm}
					class="px-5 py-3 text-base font-medium text-white bg-yellow-600 rounded-lg hover:bg-yellow-700 transition-colors whitespace-nowrap"
				>
					Đăng Luôn
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes scale-in {
		from {
			opacity: 0;
			transform: scale(0.95);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}
	
	.animate-scale-in {
		animation: scale-in 0.2s ease-out;
	}
</style>
