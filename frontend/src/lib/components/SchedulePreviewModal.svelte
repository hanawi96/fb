<script>
	import { createEventDispatcher } from 'svelte';
	import { api } from '$lib/api';
	import { X, Clock, User, AlertTriangle, CheckCircle, Calendar } from 'lucide-svelte';

	export let postId = '';
	export let pageIds = [];
	export let preferredDate = '';
	export let show = false;

	const dispatch = createEventDispatcher();

	let preview = null;
	let loading = false;
	let confirming = false;
	let error = '';

	$: if (show && pageIds.length > 0) {
		loadPreview();
	}

	async function loadPreview() {
		loading = true;
		error = '';
		try {
			preview = await api.previewSchedule(postId, pageIds, preferredDate);
		} catch (err) {
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function confirmSchedule() {
		confirming = true;
		try {
			const result = await api.scheduleWithPreview(postId, pageIds, preferredDate, true);
			dispatch('confirmed', result);
			close();
		} catch (err) {
			error = err.message;
		} finally {
			confirming = false;
		}
	}

	function close() {
		show = false;
		preview = null;
		error = '';
		dispatch('close');
	}

	function formatTime(dateStr) {
		return new Date(dateStr).toLocaleString('vi-VN', {
			hour: '2-digit',
			minute: '2-digit',
			day: '2-digit',
			month: '2-digit'
		});
	}
</script>

{#if show}
	<!-- svelte-ignore a11y-click-events-have-key-events -->
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<div class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" on:click={close}>
		<div class="bg-white rounded-xl shadow-xl max-w-2xl w-full max-h-[85vh] overflow-hidden" on:click|stopPropagation>
			<!-- Header -->
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
				<div>
					<h2 class="text-lg font-semibold text-gray-900">Xem trước lịch đăng</h2>
					<p class="text-sm text-gray-500 mt-0.5">Kiểm tra thời gian đăng trước khi xác nhận</p>
				</div>
				<button on:click={close} class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg">
					<X size={20} />
				</button>
			</div>

			<!-- Content -->
			<div class="p-6 overflow-y-auto max-h-[calc(85vh-180px)]">
				{#if loading}
					<div class="flex flex-col items-center justify-center py-12">
						<div class="w-10 h-10 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin mb-4"></div>
						<p class="text-sm text-gray-600">Đang tính toán thời gian...</p>
					</div>
				{:else if error}
					<div class="bg-red-50 text-red-700 p-4 rounded-lg text-sm">
						{error}
					</div>
				{:else if preview}
					<!-- Stats -->
					<div class="grid grid-cols-4 gap-3 mb-6">
						<div class="bg-gray-50 rounded-lg p-3 text-center">
							<div class="text-xl font-semibold text-gray-900">{preview.total_pages}</div>
							<div class="text-xs text-gray-500">Tổng pages</div>
						</div>
						<div class="bg-green-50 rounded-lg p-3 text-center">
							<div class="text-xl font-semibold text-green-600">{preview.success_count}</div>
							<div class="text-xs text-gray-500">Thành công</div>
						</div>
						<div class="bg-yellow-50 rounded-lg p-3 text-center">
							<div class="text-xl font-semibold text-yellow-600">{preview.warning_count}</div>
							<div class="text-xs text-gray-500">Cảnh báo</div>
						</div>
						<div class="bg-orange-50 rounded-lg p-3 text-center">
							<div class="text-xl font-semibold text-orange-600">{preview.next_day_count}</div>
							<div class="text-xs text-gray-500">Sang ngày mai</div>
						</div>
					</div>

					<!-- Results List -->
					<div class="space-y-2">
						{#each preview.results as result}
							<div class="flex items-center gap-3 p-3 rounded-lg border {result.warning ? 'border-yellow-200 bg-yellow-50/50' : 'border-gray-200'}">
								<!-- Status Icon -->
								<div class="flex-shrink-0">
									{#if result.error}
										<div class="w-8 h-8 bg-red-100 rounded-full flex items-center justify-center">
											<X size={16} class="text-red-600" />
										</div>
									{:else if result.warning}
										<div class="w-8 h-8 bg-yellow-100 rounded-full flex items-center justify-center">
											<AlertTriangle size={16} class="text-yellow-600" />
										</div>
									{:else}
										<div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
											<CheckCircle size={16} class="text-green-600" />
										</div>
									{/if}
								</div>

								<!-- Page Info -->
								<div class="flex-1 min-w-0">
									<div class="font-medium text-sm text-gray-900 truncate">{result.page_name}</div>
									{#if result.account_name}
										<div class="flex items-center gap-1 text-xs text-gray-500 mt-0.5">
											<User size={12} />
											<span>{result.account_name}</span>
										</div>
									{/if}
								</div>

								<!-- Time -->
								<div class="flex-shrink-0 text-right">
									<div class="flex items-center gap-1 text-sm font-medium text-gray-900">
										<Clock size={14} class="text-gray-400" />
										{formatTime(result.scheduled_time)}
									</div>
									{#if result.warning}
										<div class="text-xs text-yellow-600 mt-0.5">{result.warning}</div>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Footer -->
			{#if !loading && preview}
				<div class="border-t border-gray-100 px-6 py-4 bg-gray-50/50 flex gap-3">
					<button
						on:click={close}
						class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
					>
						Hủy
					</button>
					<button
						on:click={confirmSchedule}
						disabled={confirming || preview.success_count === 0}
						class="flex-1 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
					>
						{#if confirming}
							<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
							<span>Đang xử lý...</span>
						{:else}
							<Calendar size={16} />
							<span>Xác nhận đăng ({preview.success_count} pages)</span>
						{/if}
					</button>
				</div>
			{/if}
		</div>
	</div>
{/if}
