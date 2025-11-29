<script>
	import { Zap, Calendar, Clock, Info } from 'lucide-svelte';
	
	export let scheduleType = 'scheduled'; // 'scheduled', 'later', 'auto', 'draft'
	export let scheduledDate = '';
	export let scheduledTime = '';
	
	// Tính min date (hôm nay) và max date (30 ngày sau)
	const today = new Date();
	const minDate = today.toISOString().split('T')[0];
	const maxDateObj = new Date();
	maxDateObj.setDate(maxDateObj.getDate() + 30);
	const maxDate = maxDateObj.toISOString().split('T')[0];
	
	// Set default date to today khi chọn hẹn giờ
	$: if (scheduleType === 'later' && !scheduledDate) {
		scheduledDate = minDate;
	}
	
	// Set default date for auto schedule
	$: if (scheduleType === 'auto' && !scheduledDate) {
		scheduledDate = minDate;
	}
	
	// Tính min time nếu chọn ngày hôm nay
	$: minTime = scheduledDate === minDate 
		? `${String(today.getHours()).padStart(2, '0')}:${String(today.getMinutes() + 5).padStart(2, '0')}`
		: '00:00';
</script>

<div class="border-t border-gray-100">
	<div class="p-3">
		<div class="flex items-center justify-between mb-3">
			<div class="flex items-center gap-2">
				<div class="w-1.5 h-1.5 rounded-full bg-blue-600"></div>
				<span class="text-sm font-medium text-gray-900">Lịch đăng bài</span>
			</div>
			
			<div class="flex items-center gap-2">
				<label class="flex items-center gap-1.5 cursor-pointer group">
					<input type="radio" name="scheduleType" value="scheduled" bind:group={scheduleType}
						class="w-4 h-4 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500" />
					<span class="text-sm text-gray-700 group-hover:text-gray-900">Đăng ngay</span>
				</label>
				
				<label class="flex items-center gap-1.5 cursor-pointer group">
					<input type="radio" name="scheduleType" value="later" bind:group={scheduleType}
						class="w-4 h-4 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500" />
					<span class="text-sm text-gray-700 group-hover:text-gray-900 flex items-center gap-1">
						<Clock size={12} class="text-blue-500" />
						Hẹn giờ
					</span>
				</label>
				
				<label class="flex items-center gap-1.5 cursor-pointer group">
					<input type="radio" name="scheduleType" value="auto" bind:group={scheduleType}
						class="w-4 h-4 text-green-600 border-gray-300 focus:ring-2 focus:ring-green-500" />
					<span class="text-sm text-gray-700 group-hover:text-green-600 flex items-center gap-1">
						<Zap size={12} class="text-green-500" />
						Lịch tự động
					</span>
				</label>
				
				<label class="flex items-center gap-1.5 cursor-pointer group">
					<input type="radio" name="scheduleType" value="draft" bind:group={scheduleType}
						class="w-4 h-4 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500" />
					<span class="text-sm text-gray-700 group-hover:text-gray-900">Nháp</span>
				</label>
			</div>
		</div>
		
		{#if scheduleType === 'later'}
			<div class="mt-2 p-3 bg-blue-50 border border-blue-200 rounded-lg">
				<div class="flex items-center gap-3">
					<div class="flex items-center gap-2 flex-1">
						<Calendar size={16} class="text-blue-600" />
						<input type="date" bind:value={scheduledDate} min={minDate} max={maxDate}
							class="flex-1 px-3 py-1.5 text-sm border border-blue-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white" />
					</div>
					<div class="flex items-center gap-2 flex-1">
						<Clock size={16} class="text-blue-600" />
						<input type="time" bind:value={scheduledTime}
							class="flex-1 px-3 py-1.5 text-sm border border-blue-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white" />
					</div>
				</div>
				<p class="text-xs text-blue-600 mt-2">Bài viết sẽ được đăng vào thời gian bạn chọn (tối đa 30 ngày)</p>
			</div>
		{:else if scheduleType === 'auto'}
			<div class="mt-2 p-3 bg-green-50 border border-green-200 rounded-lg">
				<div class="flex items-start gap-2">
					<Info size={16} class="text-green-600 mt-0.5 flex-shrink-0" />
					<div class="flex-1">
						<p class="text-sm text-green-800 font-medium">Lịch tự động thông minh</p>
						<p class="text-xs text-green-600 mt-1">
							Hệ thống sẽ tự động xếp bài vào khung giờ trống gần nhất dựa trên cấu hình timeslots của từng page.
						</p>
					</div>
				</div>
				<div class="flex items-center gap-2 mt-3">
					<Calendar size={14} class="text-green-600" />
					<span class="text-xs text-green-700">Ngày ưu tiên:</span>
					<input type="date" bind:value={scheduledDate} min={minDate} max={maxDate}
						class="px-2 py-1 text-sm border border-green-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 bg-white" />
				</div>
			</div>
		{/if}
	</div>
</div>
