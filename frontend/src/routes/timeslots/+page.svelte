<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import Toast from '$lib/components/Toast.svelte';
	import { Clock, Plus, Trash2, X, Copy, Zap, Calendar, RotateCcw } from 'lucide-svelte';

	let pages = [];
	let loading = true;
	let slotsMap = {}; // pageId -> slots[]

	// Unified Modal Edit (dùng chung cho single và bulk)
	let showEditModal = false;
	let editMode = 'single'; // 'single' hoặc 'bulk'
	let editingPage = null;
	let saving = false;

	// Slots data cho modal
	let dailySlots = [];
	let customSlots = [];
	let scheduleMode = 'daily';

	// Unique ID counter
	let slotIdCounter = 0;

	// Unified Copy Modal (dùng chung cho tất cả)
	let showCopyModal = false;
	let copyMode = 'toEdit'; // 'toEdit' | 'toSingle' | 'toBulk'
	let copyTargetPage = null; // Dùng khi copyMode = 'toSingle'
	let copyFromPageId = '';
	let copyProcessing = false;

	// Bulk action
	let selectedPageIds = [];

	// Reactive
	$: isAllSelected = pages.length > 0 && selectedPageIds.length === pages.length;
	$: isIndeterminate = selectedPageIds.length > 0 && selectedPageIds.length < pages.length;
	$: currentSlots = scheduleMode === 'daily' ? dailySlots : customSlots;
	$: slotsByDay = dayNumbers.reduce((acc, day) => {
		acc[day] = customSlots.filter((slot) => slot.days_of_week?.includes(day));
		return acc;
	}, {});

	const dayLabels = ['T2', 'T3', 'T4', 'T5', 'T6', 'T7', 'CN'];
	const dayNumbers = [1, 2, 3, 4, 5, 6, 7];
	const hours = Array.from({ length: 24 }, (_, i) => i);

	const presets = [
		{
			name: 'Giờ vàng Facebook',
			icon: '🔥',
			description: '7h, 12h, 18h, 21h',
			slots: [
				{ start_hour: 7, end_hour: 8 },
				{ start_hour: 12, end_hour: 13 },
				{ start_hour: 18, end_hour: 19 },
				{ start_hour: 21, end_hour: 22 }
			]
		},
		{
			name: 'Giờ hành chính',
			icon: '💼',
			description: '9h, 12h, 17h',
			slots: [
				{ start_hour: 9, end_hour: 10 },
				{ start_hour: 12, end_hour: 13 },
				{ start_hour: 17, end_hour: 18 }
			]
		},
		{
			name: 'Buổi tối',
			icon: '🌙',
			description: '19h, 21h',
			slots: [
				{ start_hour: 19, end_hour: 20 },
				{ start_hour: 21, end_hour: 22 }
			]
		},
		{
			name: 'Cả ngày',
			icon: '☀️',
			description: '8h, 11h, 14h, 17h, 20h',
			slots: [
				{ start_hour: 8, end_hour: 9 },
				{ start_hour: 11, end_hour: 12 },
				{ start_hour: 14, end_hour: 15 },
				{ start_hour: 17, end_hour: 18 },
				{ start_hour: 20, end_hour: 21 }
			]
		}
	];

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		try {
			const result = await api.getPages();
			pages = result || [];
			for (const page of pages) {
				const slots = await api.getPageTimeSlots(page.id);
				slotsMap[page.id] = slots || [];
			}
			slotsMap = slotsMap;
		} catch (e) {
			pages = [];
		} finally {
			loading = false;
		}
	}

	$: slotCountsByPage = pages.reduce((acc, page) => {
		acc[page.id] = {};
		const slots = slotsMap[page.id] || [];
		for (const day of dayNumbers) {
			acc[page.id][day] = slots.filter((s) => s.days_of_week?.includes(day)).length;
		}
		return acc;
	}, {});

	function detectScheduleMode(slots) {
		if (slots.length === 0) return 'daily';
		const allSlotsHaveAllDays = slots.every(
			(s) => s.days_of_week?.length === 7 && dayNumbers.every((d) => s.days_of_week.includes(d))
		);
		return allSlotsHaveAllDays ? 'daily' : 'custom';
	}

	function openEditSingle(page) {
		editMode = 'single';
		editingPage = page;
		slotIdCounter = 0;

		const existingSlots = slotsMap[page.id] || [];
		const detectedMode = detectScheduleMode(existingSlots);

		// Convert start_time/end_time to start_hour/end_hour
		const convertSlot = (s) => ({
			...JSON.parse(JSON.stringify(s)),
			_uid: ++slotIdCounter,
			start_hour: parseTimeToHour(s.start_time),
			end_hour: parseTimeToHour(s.end_time),
			slot_capacity: s.slot_capacity || 10
		});

		if (detectedMode === 'daily') {
			dailySlots = existingSlots.map(convertSlot);
			customSlots = [];
		} else {
			customSlots = existingSlots.map(convertSlot);
			dailySlots = [];
		}

		scheduleMode = detectedMode;
		showEditModal = true;
	}
	
	// Parse "09:00:00" or "09:00" to 9
	function parseTimeToHour(timeStr) {
		if (!timeStr) return 9;
		const parts = timeStr.split(':');
		return parseInt(parts[0], 10) || 9;
	}

	function openEditBulk() {
		editMode = 'bulk';
		editingPage = null;
		slotIdCounter = 0;
		dailySlots = [];
		customSlots = [];
		scheduleMode = 'daily';
		showEditModal = true;
	}

	function closeEditModal() {
		showEditModal = false;
		editingPage = null;
		dailySlots = [];
		customSlots = [];
	}

	function addNewSlot() {
		dailySlots = [
			...dailySlots,
			{ _uid: ++slotIdCounter, id: null, start_hour: 9, end_hour: 10, days_of_week: [1, 2, 3, 4, 5, 6, 7], slot_capacity: 10, isNew: true }
		];
	}

	function addSlotForDay(day) {
		customSlots = [
			...customSlots,
			{ _uid: ++slotIdCounter, id: null, start_hour: 9, end_hour: 10, days_of_week: [day], slot_capacity: 10, isNew: true }
		];
	}

	function removeSlot(uid) {
		if (scheduleMode === 'daily') {
			dailySlots = dailySlots.filter((s) => s._uid !== uid);
		} else {
			customSlots = customSlots.filter((s) => s._uid !== uid);
		}
	}

	function switchMode(mode) {
		scheduleMode = mode;
	}

	function applyPreset(preset) {
		if (scheduleMode === 'daily') {
			dailySlots = preset.slots.map((s) => ({
				_uid: ++slotIdCounter, id: null, start_hour: s.start_hour, end_hour: s.end_hour,
				days_of_week: [1, 2, 3, 4, 5, 6, 7], slot_capacity: 10, isNew: true
			}));
		} else {
			const newSlots = [];
			for (const day of dayNumbers) {
				for (const s of preset.slots) {
					newSlots.push({
						_uid: ++slotIdCounter, id: null, start_hour: s.start_hour, end_hour: s.end_hour,
						days_of_week: [day], slot_capacity: 10, isNew: true
					});
				}
			}
			customSlots = newSlots;
		}
	}

	function resetSlots() {
		if (scheduleMode === 'daily') dailySlots = [];
		else customSlots = [];
	}

	// ========== UNIFIED COPY MODAL ==========
	// Copy vào modal edit (chưa lưu DB)
	function openCopyToEdit() {
		copyMode = 'toEdit';
		copyTargetPage = null;
		copyFromPageId = '';
		showCopyModal = true;
	}

	// Copy trực tiếp vào 1 page (lưu DB ngay)
	function openCopyToSingle(page) {
		copyMode = 'toSingle';
		copyTargetPage = page;
		copyFromPageId = '';
		showCopyModal = true;
	}

	// Copy vào nhiều page đã chọn (lưu DB ngay)
	function openCopyToBulk() {
		copyMode = 'toBulk';
		copyTargetPage = null;
		copyFromPageId = '';
		showCopyModal = true;
	}

	function closeCopyModal() {
		showCopyModal = false;
		copyTargetPage = null;
		copyFromPageId = '';
	}

	async function confirmCopy() {
		if (!copyFromPageId) return;

		const sourceSlots = slotsMap[copyFromPageId] || [];
		if (sourceSlots.length === 0) {
			toast.show('Page nguồn chưa có khung giờ nào', 'error');
			return;
		}

		if (copyMode === 'toEdit') {
			// Copy vào modal edit (chưa lưu)
			const copiedSlots = sourceSlots.map((s) => ({
				_uid: ++slotIdCounter, id: null,
				start_hour: getSlotHour(s, 'start'),
				end_hour: getSlotHour(s, 'end'),
				days_of_week: [...(s.days_of_week || [1, 2, 3, 4, 5, 6, 7])],
				slot_capacity: s.slot_capacity || 10,
				isNew: true
			}));

			if (scheduleMode === 'daily') dailySlots = copiedSlots;
			else customSlots = copiedSlots;

			closeCopyModal();
			toast.show(`Đã copy ${copiedSlots.length} khung giờ`, 'success');
		} else if (copyMode === 'toSingle') {
			// Copy trực tiếp vào 1 page
			copyProcessing = true;
			try {
				const targetPageId = copyTargetPage.id;
				await copySlotsToDB(sourceSlots, [targetPageId]);
				closeCopyModal();
				toast.show(`Đã copy ${sourceSlots.length} khung giờ`, 'success');
			} catch (e) {
				toast.show('Lỗi copy: ' + e.message, 'error');
			} finally {
				copyProcessing = false;
			}
		} else if (copyMode === 'toBulk') {
			// Copy vào nhiều page
			copyProcessing = true;
			try {
				const targetIds = selectedPageIds.filter((id) => id !== copyFromPageId);
				await copySlotsToDB(sourceSlots, targetIds);
				closeCopyModal();
				selectedPageIds = [];
				toast.show(`Đã copy khung giờ cho ${targetIds.length} page`, 'success');
			} catch (e) {
				toast.show('Lỗi: ' + e.message, 'error');
			} finally {
				copyProcessing = false;
			}
		}
	}

	async function copySlotsToDB(sourceSlots, targetPageIds) {
		for (const pageId of targetPageIds) {
			const oldSlots = slotsMap[pageId] || [];
			for (const old of oldSlots) {
				await api.deleteTimeSlot(old.id);
			}

			for (const slot of sourceSlots) {
				const startTime = `${String(getSlotHour(slot, 'start')).padStart(2, '0')}:00`;
				const endTime = `${String(getSlotHour(slot, 'end')).padStart(2, '0')}:00`;

				await api.createTimeSlot(pageId, {
					start_time: startTime,
					end_time: endTime,
					days_of_week: [...(slot.days_of_week || [1, 2, 3, 4, 5, 6, 7])],
					slot_capacity: slot.slot_capacity || 10,
					slot_name: ''
				});
			}

			const freshSlots = await api.getPageTimeSlots(pageId);
			slotsMap[pageId] = freshSlots || [];
		}
		slotsMap = { ...slotsMap };
	}

	// ========== SAVE CHANGES ==========
	async function saveChanges() {
		saving = true;
		const slotsToSave = scheduleMode === 'daily' ? dailySlots : customSlots;

		try {
			if (editMode === 'single') {
				const pageId = editingPage.id;
				const oldSlots = slotsMap[pageId] || [];

				for (const old of oldSlots) await api.deleteTimeSlot(old.id);

				for (const slot of slotsToSave) {
					const startTime = `${String(slot.start_hour ?? parseInt(slot.start_time)).padStart(2, '0')}:00`;
					const endTime = `${String(slot.end_hour ?? parseInt(slot.end_time)).padStart(2, '0')}:00`;

					await api.createTimeSlot(pageId, {
						start_time: startTime, end_time: endTime,
						days_of_week: slot.days_of_week, 
						slot_capacity: slot.slot_capacity || 10,
						slot_name: ''
					});
				}

				const freshSlots = await api.getPageTimeSlots(pageId);
				slotsMap[pageId] = freshSlots || [];
				slotsMap = { ...slotsMap };
				toast.show('Đã lưu khung giờ', 'success');
			} else {
				const pageCount = selectedPageIds.length;

				for (const pageId of selectedPageIds) {
					const oldSlots = slotsMap[pageId] || [];
					for (const old of oldSlots) await api.deleteTimeSlot(old.id);

					for (const slot of slotsToSave) {
						const startTime = `${String(slot.start_hour).padStart(2, '0')}:00`;
						const endTime = `${String(slot.end_hour).padStart(2, '0')}:00`;

						await api.createTimeSlot(pageId, {
							start_time: startTime, end_time: endTime,
							days_of_week: slot.days_of_week, 
							slot_capacity: slot.slot_capacity || 10,
							slot_name: ''
						});
					}

					const freshSlots = await api.getPageTimeSlots(pageId);
					slotsMap[pageId] = freshSlots || [];
				}

				slotsMap = { ...slotsMap };
				selectedPageIds = [];
				toast.show(`Đã cập nhật khung giờ cho ${pageCount} page`, 'success');
			}

			closeEditModal();
		} catch (e) {
			toast.show('Lỗi lưu: ' + e.message, 'error');
		} finally {
			saving = false;
		}
	}

	// ========== BULK ACTIONS ==========
	function toggleSelectAll() {
		if (isAllSelected) selectedPageIds = [];
		else selectedPageIds = pages.map((p) => p.id);
	}

	function toggleSelectPage(pageId) {
		if (selectedPageIds.includes(pageId)) {
			selectedPageIds = selectedPageIds.filter((id) => id !== pageId);
		} else {
			selectedPageIds = [...selectedPageIds, pageId];
		}
	}

	async function bulkDeleteSlots() {
		if (selectedPageIds.length === 0) return;
		if (!confirm(`Xóa tất cả khung giờ của ${selectedPageIds.length} page đã chọn?`)) return;

		copyProcessing = true;
		try {
			for (const pageId of selectedPageIds) {
				const oldSlots = slotsMap[pageId] || [];
				for (const old of oldSlots) await api.deleteTimeSlot(old.id);
				slotsMap[pageId] = [];
			}

			slotsMap = { ...slotsMap };
			selectedPageIds = [];
			toast.show('Đã xóa khung giờ', 'success');
		} catch (e) {
			toast.show('Lỗi: ' + e.message, 'error');
		} finally {
			copyProcessing = false;
		}
	}

	function getSlotHour(slot, type) {
		if (type === 'start') return slot.start_hour ?? parseTimeToHour(slot.start_time);
		return slot.end_hour ?? parseTimeToHour(slot.end_time);
	}

	// Copy modal helpers
	$: copyModalTitle = copyMode === 'toEdit' ? 'Copy khung giờ' :
		copyMode === 'toSingle' ? 'Copy khung giờ' : 'Copy khung giờ hàng loạt';

	$: copyModalSubtitle = copyMode === 'toEdit' ? 'Chọn page nguồn để copy cấu hình' :
		copyMode === 'toSingle' ? `Copy vào ${copyTargetPage?.page_name || ''}` :
		`Áp dụng cho ${selectedPageIds.length} page đã chọn`;

	$: copyButtonText = copyMode === 'toBulk' ? `Copy cho ${selectedPageIds.length} page` : 'Copy';

	$: excludePageId = copyMode === 'toEdit' ? (editMode === 'single' ? editingPage?.id : null) :
		copyMode === 'toSingle' ? copyTargetPage?.id : null;
</script>

<svelte:head>
	<title>Khung giờ đăng bài - FB Scheduler</title>
</svelte:head>

{#if $toast}
	<Toast message={$toast.message} type={$toast.type} onClose={() => toast.hide()} />
{/if}

<div class="w-full">
	<div class="mb-6">
		<h1 class="text-2xl font-semibold text-gray-900">Khung giờ đăng bài</h1>
		<p class="text-sm text-gray-500 mt-1">Cài đặt khung giờ đăng bài theo ngày cho từng page</p>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-8 h-8 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
		</div>
	{:else if pages.length === 0}
		<div class="bg-gray-50 rounded-lg p-8 text-center border border-gray-200">
			<Clock size={32} class="text-gray-400 mx-auto mb-3" />
			<p class="text-gray-600">Chưa có page nào. Hãy kết nối Facebook trước.</p>
		</div>
	{:else}
		{#if selectedPageIds.length > 0}
			<div class="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-xl flex items-center justify-between">
				<div class="flex items-center gap-3">
					<span class="text-sm font-medium text-blue-700">Đã chọn {selectedPageIds.length} page</span>
					<button on:click={() => (selectedPageIds = [])} class="text-xs text-blue-600 hover:text-blue-800 underline">Bỏ chọn</button>
				</div>
				<div class="flex items-center gap-2">
					<button on:click={openEditBulk} disabled={copyProcessing}
						class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50 transition-colors">
						<Clock size={14} /><span>Chỉnh sửa khung giờ</span>
					</button>
					<button on:click={openCopyToBulk} disabled={copyProcessing}
						class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors">
						<Copy size={14} /><span>Copy khung giờ</span>
					</button>
					<button on:click={bulkDeleteSlots} disabled={copyProcessing}
						class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50 transition-colors">
						<Trash2 size={14} /><span>Xóa khung giờ</span>
					</button>
				</div>
			</div>
		{/if}

		<div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b border-gray-200">
					<tr>
						<th class="text-center px-3 py-3 w-12">
							<input type="checkbox" checked={isAllSelected} indeterminate={isIndeterminate} on:change={toggleSelectAll}
								class="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500" />
						</th>
						<th class="text-left px-4 py-3 font-medium text-gray-600" style="width: 35%">Tên page</th>
						{#each dayLabels as day}
							<th class="text-center px-3 py-3 font-medium text-gray-500 text-xs">{day}</th>
						{/each}
						<th class="text-center px-4 py-3 font-medium text-gray-600">Thao tác</th>
					</tr>
				</thead>
				<tbody>
					{#each pages as page}
						<tr class="border-b border-gray-100 hover:bg-gray-50 {selectedPageIds.includes(page.id) ? 'bg-blue-50' : ''}">
							<td class="text-center px-3 py-3">
								<input type="checkbox" checked={selectedPageIds.includes(page.id)} on:change={() => toggleSelectPage(page.id)}
									class="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500" />
							</td>
							<td class="px-4 py-3">
								<div class="flex items-center gap-3">
									<img src={page.profile_picture_url || 'https://via.placeholder.com/36'} alt="" class="w-9 h-9 rounded-full" />
									<div>
										<div class="font-medium text-gray-900 text-sm">{page.page_name}</div>
										<div class="text-xs text-gray-500">{page.category || ''}</div>
									</div>
								</div>
							</td>
							{#each dayNumbers as day}
								<td class="text-center px-3 py-3">
									<span class="inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-medium
										{(slotCountsByPage[page.id]?.[day] || 0) > 0 ? 'bg-blue-100 text-blue-700' : 'text-gray-400'}">
										{slotCountsByPage[page.id]?.[day] || 0}
									</span>
								</td>
							{/each}
							<td class="text-center px-4 py-3">
								<div class="flex items-center justify-center gap-2">
									<button on:click={() => openEditSingle(page)}
										class="px-3 py-1.5 text-xs font-medium text-blue-600 border border-blue-200 rounded-lg hover:bg-blue-50 transition-colors">
										Chỉnh sửa
									</button>
									{#if pages.length > 1}
										<button on:click={() => openCopyToSingle(page)}
											class="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
											title="Copy khung giờ từ page khác">
											<Copy size={16} />
										</button>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>


<!-- Unified Edit Modal -->
{#if showEditModal}
	<!-- svelte-ignore a11y-click-events-have-key-events -->
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" on:click={closeEditModal}>
		<div class="bg-white rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden" on:click|stopPropagation>
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between
				{editMode === 'single' ? 'bg-gradient-to-r from-blue-50 to-indigo-50' : 'bg-gradient-to-r from-green-50 to-emerald-50'}">
				{#if editMode === 'single' && editingPage}
					<div class="flex items-center gap-3">
						<img src={editingPage.profile_picture_url || 'https://via.placeholder.com/44'} alt="" class="w-11 h-11 rounded-full ring-2 ring-white shadow" />
						<div>
							<h2 class="font-semibold text-gray-900">{editingPage.page_name}</h2>
							<p class="text-xs text-gray-500">Thiết lập khung giờ đăng bài</p>
						</div>
					</div>
				{:else}
					<div>
						<h2 class="text-lg font-semibold text-gray-900">Chỉnh sửa khung giờ hàng loạt</h2>
						<p class="text-sm text-gray-500 mt-1">Áp dụng cho <span class="font-medium text-green-600">{selectedPageIds.length} page</span> đã chọn</p>
					</div>
				{/if}
				<button on:click={closeEditModal} class="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-white/50"><X size={20} /></button>
			</div>

			<div class="p-6 overflow-y-auto max-h-[calc(90vh-180px)]">
				<div class="flex items-center gap-2 p-1 bg-gray-100 rounded-xl mb-5">
					<button on:click={() => switchMode('daily')}
						class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all
							{scheduleMode === 'daily' ? (editMode === 'single' ? 'bg-white text-blue-600 shadow-sm' : 'bg-white text-green-600 shadow-sm') : 'text-gray-600 hover:text-gray-900'}">
						<RotateCcw size={16} /><span>Hằng ngày</span>
					</button>
					<button on:click={() => switchMode('custom')}
						class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all
							{scheduleMode === 'custom' ? (editMode === 'single' ? 'bg-white text-blue-600 shadow-sm' : 'bg-white text-green-600 shadow-sm') : 'text-gray-600 hover:text-gray-900'}">
						<Calendar size={16} /><span>Tùy chỉnh theo ngày</span>
					</button>
				</div>

				<div class="flex items-center gap-2 mb-5">
					{#if pages.length > 1}
						<button on:click={openCopyToEdit} class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors">
							<Copy size={14} /><span>Copy từ page khác</span>
						</button>
					{/if}
					{#if currentSlots.length > 0}
						<button on:click={resetSlots} class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 transition-colors">
							<Trash2 size={14} /><span>Xóa tất cả</span>
						</button>
					{/if}
				</div>

				<div class="mb-5">
					<div class="flex items-center gap-2 mb-3">
						<Zap size={16} class="text-amber-500" />
						<span class="text-sm font-medium text-gray-700">Mẫu nhanh</span>
					</div>
					<div class="grid grid-cols-2 gap-2">
						{#each presets as preset}
							<button on:click={() => applyPreset(preset)}
								class="flex items-center gap-3 p-3 text-left bg-gray-50 border border-gray-200 rounded-xl transition-all group
									{editMode === 'single' ? 'hover:bg-blue-50 hover:border-blue-200' : 'hover:bg-green-50 hover:border-green-200'}">
								<span class="text-xl">{preset.icon}</span>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-medium text-gray-900">{preset.name}</div>
									<div class="text-xs text-gray-500">{preset.description}</div>
								</div>
							</button>
						{/each}
					</div>
				</div>

				{#if scheduleMode === 'daily'}
					<div class="space-y-3">
						{#if dailySlots.length === 0}
							<div class="text-center py-8 text-gray-500">
								<Clock size={32} class="mx-auto mb-2 text-gray-300" />
								<p class="text-sm">Chưa có khung giờ nào</p>
								<p class="text-xs text-gray-400 mt-1">Chọn mẫu nhanh hoặc thêm thủ công</p>
							</div>
						{:else}
							{#each dailySlots as slot (slot._uid)}
								<div class="flex items-center gap-3 p-3 bg-white border border-gray-200 rounded-xl">
									<div class="flex items-center gap-2">
										<select bind:value={slot.start_hour} on:change={() => (slot.start_hour = parseInt(slot.start_hour))}
											class="px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-medium focus:outline-none focus:ring-2 focus:ring-blue-500">
											{#each hours as h}<option value={h}>{String(h).padStart(2, '0')}:00</option>{/each}
										</select>
										<span class="text-gray-400 font-medium">→</span>
										<select bind:value={slot.end_hour} on:change={() => (slot.end_hour = parseInt(slot.end_hour))}
											class="px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-medium focus:outline-none focus:ring-2 focus:ring-blue-500">
											{#each hours as h}<option value={h}>{String(h).padStart(2, '0')}:00</option>{/each}
										</select>
									</div>
									<div class="flex items-center gap-2 ml-2">
										<label class="text-xs text-gray-500 font-medium whitespace-nowrap">Số bài:</label>
										<input 
											type="number" 
											bind:value={slot.slot_capacity}
											min="1" 
											max="100"
											class="w-16 px-2 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-medium text-center focus:outline-none focus:ring-2 focus:ring-blue-500"
										/>
										<span class="text-xs text-gray-500">bài</span>
									</div>
									<div class="flex-1"></div>
									<button on:click={() => removeSlot(slot._uid)} class="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors">
										<Trash2 size={18} />
									</button>
								</div>
							{/each}
						{/if}
					</div>
					<button on:click={addNewSlot}
						class="w-full mt-4 flex items-center justify-center gap-2 px-4 py-3 border-2 border-dashed border-gray-300 text-gray-600 rounded-xl transition-all
							{editMode === 'single' ? 'hover:border-blue-400 hover:text-blue-600 hover:bg-blue-50/50' : 'hover:border-green-400 hover:text-green-600 hover:bg-green-50/50'}">
						<Plus size={18} /><span class="font-medium">Thêm khung giờ</span>
					</button>
				{:else}
					<div class="space-y-4">
						{#each dayNumbers as day (day)}
							<div class="border border-gray-200 rounded-xl overflow-hidden">
								<div class="px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
									<div class="flex items-center gap-2">
										<span class="w-8 h-8 flex items-center justify-center text-white text-xs font-bold rounded-lg {editMode === 'single' ? 'bg-blue-600' : 'bg-green-600'}">
											{dayLabels[day - 1]}
										</span>
										<span class="text-sm font-medium text-gray-700">{day === 7 ? 'Chủ nhật' : `Thứ ${day + 1}`}</span>
									</div>
									<span class="text-xs text-gray-500">{(slotsByDay[day] || []).length} khung giờ</span>
								</div>
								<div class="p-3 space-y-2">
									{#if !slotsByDay[day] || slotsByDay[day].length === 0}
										<div class="text-center py-4 text-gray-400 text-sm">Chưa có khung giờ</div>
									{:else}
										{#each slotsByDay[day] as slot (slot._uid)}
											<div class="flex items-center gap-2 p-2 bg-gray-50 rounded-lg">
												<select bind:value={slot.start_hour} on:change={() => (slot.start_hour = parseInt(slot.start_hour))}
													class="px-2 py-1.5 bg-white border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
													{#each hours as h}<option value={h}>{String(h).padStart(2, '0')}:00</option>{/each}
												</select>
												<span class="text-gray-400 text-sm">→</span>
												<select bind:value={slot.end_hour} on:change={() => (slot.end_hour = parseInt(slot.end_hour))}
													class="px-2 py-1.5 bg-white border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
													{#each hours as h}<option value={h}>{String(h).padStart(2, '0')}:00</option>{/each}
												</select>
												<div class="flex items-center gap-1.5 ml-1">
													<label class="text-xs text-gray-500 whitespace-nowrap">Số bài:</label>
													<input 
														type="number" 
														bind:value={slot.slot_capacity}
														min="1" 
														max="100"
														class="w-14 px-1.5 py-1.5 bg-white border border-gray-200 rounded-lg text-xs text-center focus:outline-none focus:ring-2 focus:ring-blue-500"
													/>
												</div>
												<div class="flex-1"></div>
												<button on:click={() => removeSlot(slot._uid)} class="p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded transition-colors">
													<Trash2 size={16} />
												</button>
											</div>
										{/each}
									{/if}
									<button on:click={() => addSlotForDay(day)}
										class="w-full flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium border border-dashed rounded-lg transition-colors
											{editMode === 'single' ? 'text-blue-600 border-blue-300 hover:bg-blue-50' : 'text-green-600 border-green-300 hover:bg-green-50'}">
										<Plus size={14} /><span>Thêm giờ</span>
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<div class="px-6 py-4 border-t border-gray-100 bg-gray-50 flex items-center gap-3">
				<button on:click={closeEditModal} class="px-5 py-2.5 text-sm font-medium text-gray-600 hover:bg-gray-200 rounded-xl transition-colors">Hủy</button>
				<button on:click={saveChanges} disabled={saving}
					class="flex-1 px-5 py-2.5 text-white text-sm font-medium rounded-xl disabled:opacity-50 transition-colors flex items-center justify-center gap-2
						{editMode === 'single' ? 'bg-blue-600 hover:bg-blue-700' : 'bg-green-600 hover:bg-green-700'}">
					{#if saving}<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>{/if}
					<span>{editMode === 'single' ? 'Lưu thay đổi' : `Lưu cho ${selectedPageIds.length} page`}</span>
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Unified Copy Modal -->
{#if showCopyModal}
	<!-- svelte-ignore a11y-click-events-have-key-events -->
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-[60] p-4" on:click={closeCopyModal}>
		<div class="bg-white rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden" on:click|stopPropagation>
			<div class="px-6 py-5 border-b border-gray-100 bg-gradient-to-r from-blue-50 to-indigo-50">
				<h3 class="text-lg font-semibold text-gray-900">{copyModalTitle}</h3>
				<p class="text-sm text-gray-500 mt-1">{copyModalSubtitle}</p>
			</div>
			<div class="p-6">
				<div class="text-sm text-gray-600 mb-4">Chọn page nguồn:</div>
				<div class="space-y-3 max-h-80 overflow-y-auto">
					{#each pages.filter((p) => p.id !== excludePageId) as page}
						<label class="flex items-center gap-4 p-4 border border-gray-200 rounded-xl cursor-pointer hover:bg-gray-50 transition-colors
							{copyFromPageId === page.id ? 'border-blue-500 bg-blue-50 ring-2 ring-blue-200' : ''}">
							<input type="radio" name="copyFrom" value={page.id} bind:group={copyFromPageId} class="w-4 h-4 text-blue-600" />
							<img src={page.profile_picture_url || 'https://via.placeholder.com/40'} alt="" class="w-10 h-10 rounded-full" />
							<div class="flex-1 min-w-0">
								<div class="text-sm font-medium text-gray-900">{page.page_name}</div>
								<div class="text-xs text-gray-500 mt-0.5">{(slotsMap[page.id] || []).length} khung giờ</div>
							</div>
							{#if copyMode === 'toBulk' && selectedPageIds.includes(page.id)}
								<span class="text-xs text-blue-600 bg-blue-100 px-2 py-0.5 rounded">Đã chọn</span>
							{/if}
						</label>
					{/each}
				</div>
			</div>
			<div class="px-6 py-4 border-t border-gray-100 bg-gray-50 flex gap-3">
				<button on:click={closeCopyModal} class="flex-1 px-5 py-2.5 text-sm font-medium text-gray-600 hover:bg-gray-200 rounded-xl transition-colors">Hủy</button>
				<button on:click={confirmCopy} disabled={!copyFromPageId || copyProcessing}
					class="flex-1 px-5 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-xl hover:bg-blue-700 disabled:opacity-50 transition-colors flex items-center justify-center gap-2">
					{#if copyProcessing}<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>{/if}
					<span>{copyButtonText}</span>
				</button>
			</div>
		</div>
	</div>
{/if}

