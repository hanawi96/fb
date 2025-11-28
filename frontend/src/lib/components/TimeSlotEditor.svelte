<script>
	import { createEventDispatcher } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import { X, Plus, Trash2, Clock } from 'lucide-svelte';

	export let pageId = '';
	export let show = false;

	const dispatch = createEventDispatcher();

	let slots = [];
	let loading = false;
	let saving = false;

	// New slot form
	let newSlot = {
		slot_name: '',
		start_time: '09:00',
		end_time: '12:00',
		days_of_week: [1, 2, 3, 4, 5, 6, 7]
	};

	const dayNames = ['T2', 'T3', 'T4', 'T5', 'T6', 'T7', 'CN'];

	$: if (show && pageId) {
		loadSlots();
	}

	async function loadSlots() {
		loading = true;
		try {
			slots = await api.getPageTimeSlots(pageId);
		} catch (error) {
			toast.show('Không thể tải khung giờ', 'error');
		} finally {
			loading = false;
		}
	}

	async function addSlot() {
		if (!newSlot.start_time || !newSlot.end_time) {
			toast.show('Vui lòng chọn thời gian', 'error');
			return;
		}

		saving = true;
		try {
			const created = await api.createTimeSlot(pageId, newSlot);
			slots = [...slots, created];
			newSlot = {
				slot_name: '',
				start_time: '09:00',
				end_time: '12:00',
				days_of_week: [1, 2, 3, 4, 5, 6, 7]
			};
			toast.show('Đã thêm khung giờ', 'success');
		} catch (error) {
			toast.show('Không thể thêm khung giờ', 'error');
		} finally {
			saving = false;
		}
	}

	async function deleteSlot(id) {
		const original = [...slots];
		slots = slots.filter(s => s.id !== id);

		try {
			await api.deleteTimeSlot(id);
			toast.show('Đã xóa khung giờ', 'success');
		} catch (error) {
			slots = original;
			toast.show('Không thể xóa khung giờ', 'error');
		}
	}

	async function toggleSlotActive(slot) {
		const original = [...slots];
		const index = slots.findIndex(s => s.id === slot.id);
		slots[index] = { ...slot, is_active: !slot.is_active };
		slots = slots;

		try {
			await api.updateTimeSlot(slot.id, { is_active: !slot.is_active });
		} catch (error) {
			slots = original;
			toast.show('Không thể cập nhật', 'error');
		}
	}

	function toggleDay(day) {
		if (newSlot.days_of_week.includes(day)) {
			newSlot.days_of_week = newSlot.days_of_week.filter(d => d !== day);
		} else {
			newSlot.days_of_week = [...newSlot.days_of_week, day].sort();
		}
	}

	function close() {
		show = false;
		dispatch('close');
	}

	function formatTime(timeStr) {
		return timeStr ? timeStr.slice(0, 5) : '';
	}

	function getDaysText(days) {
		if (!days || days.length === 0) return 'Không có';
		if (days.length === 7) return 'Hàng ngày';
		if (days.length === 5 && !days.includes(6) && !days.includes(7)) return 'Thứ 2-6';
		if (days.length === 2 && days.includes(6) && days.includes(7)) return 'Cuối tuần';
		return days.map(d => dayNames[d - 1]).join(', ');
	}
</script>

{#if show}
	<!-- svelte-ignore a11y-click-events-have-key-events -->
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<div class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" on:click={close}>
		<div class="bg-white rounded-xl shadow-xl max-w-lg w-full max-h-[85vh] overflow-hidden" on:click|stopPropagation>
			<!-- Header -->
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
				<div>
					<h2 class="text-lg font-semibold text-gray-900">Khung giờ đăng bài</h2>
					<p class="text-sm text-gray-500 mt-0.5">Cấu hình thời gian đăng bài tự động</p>
				</div>
				<button on:click={close} class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg">
					<X size={20} />
				</button>
			</div>

			<!-- Content -->
			<div class="p-6 overflow-y-auto max-h-[calc(85vh-200px)]">
				{#if loading}
					<div class="flex items-center justify-center py-8">
						<div class="w-8 h-8 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
					</div>
				{:else}
					<!-- Existing Slots -->
					{#if slots.length > 0}
						<div class="space-y-2 mb-6">
							{#each slots as slot}
								<div class="flex items-center gap-3 p-3 rounded-lg border border-gray-200 {slot.is_active ? '' : 'opacity-50'}">
									<div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center flex-shrink-0">
										<Clock size={18} class="text-blue-600" />
									</div>
									<div class="flex-1 min-w-0">
										<div class="font-medium text-sm text-gray-900">
											{formatTime(slot.start_time)} - {formatTime(slot.end_time)}
										</div>
										<div class="text-xs text-gray-500 mt-0.5">
											{slot.slot_name || getDaysText(slot.days_of_week)}
										</div>
									</div>
									<button
										on:click={() => toggleSlotActive(slot)}
										class="px-2 py-1 text-xs font-medium rounded {slot.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}"
									>
										{slot.is_active ? 'Bật' : 'Tắt'}
									</button>
									<button
										on:click={() => deleteSlot(slot.id)}
										class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
									>
										<Trash2 size={16} />
									</button>
								</div>
							{/each}
						</div>
					{/if}

					<!-- Add New Slot -->
					<div class="border border-dashed border-gray-300 rounded-lg p-4">
						<h4 class="text-sm font-medium text-gray-900 mb-3">Thêm khung giờ mới</h4>
						
						<!-- Time Range -->
						<div class="grid grid-cols-2 gap-3 mb-3">
							<div>
								<label class="block text-xs text-gray-500 mb-1">Bắt đầu</label>
								<input
									type="time"
									bind:value={newSlot.start_time}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
								/>
							</div>
							<div>
								<label class="block text-xs text-gray-500 mb-1">Kết thúc</label>
								<input
									type="time"
									bind:value={newSlot.end_time}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
								/>
							</div>
						</div>

						<!-- Days of Week -->
						<div class="mb-3">
							<label class="block text-xs text-gray-500 mb-2">Ngày trong tuần</label>
							<div class="flex gap-1">
								{#each [1, 2, 3, 4, 5, 6, 7] as day}
									<button
										type="button"
										on:click={() => toggleDay(day)}
										class="w-9 h-9 text-xs font-medium rounded-lg transition-colors
											{newSlot.days_of_week.includes(day) 
												? 'bg-blue-600 text-white' 
												: 'bg-gray-100 text-gray-600 hover:bg-gray-200'}"
									>
										{dayNames[day - 1]}
									</button>
								{/each}
							</div>
						</div>

						<!-- Slot Name (optional) -->
						<div class="mb-3">
							<label class="block text-xs text-gray-500 mb-1">Tên (tùy chọn)</label>
							<input
								type="text"
								bind:value={newSlot.slot_name}
								placeholder="VD: Buổi sáng, Giờ vàng..."
								class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
							/>
						</div>

						<button
							on:click={addSlot}
							disabled={saving}
							class="w-full flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
						>
							{#if saving}
								<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
							{:else}
								<Plus size={16} />
							{/if}
							<span>Thêm khung giờ</span>
						</button>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="border-t border-gray-100 px-6 py-4 bg-gray-50/50">
				<button
					on:click={close}
					class="w-full px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
				>
					Đóng
				</button>
			</div>
		</div>
	</div>
{/if}
