<script>
	export let scheduleType = 'scheduled'; // 'scheduled', 'later', 'daily', 'draft'
	export let scheduledDate = '';
	export let scheduledTime = '';
</script>

<div class="border-t border-gray-100">
	<!-- Lịch đăng bài -->
	<div class="border-t border-gray-100 p-3">
		<div class="flex items-center justify-between mb-3">
			<div class="flex items-center gap-2">
				<div class="w-1.5 h-1.5 rounded-full bg-blue-600"></div>
				<span class="text-sm font-medium text-gray-900">Lịch đăng bài</span>
			</div>
			
			<div class="flex items-center gap-2">
				<label class="flex items-center gap-1.5 cursor-pointer group">
					<input 
						type="radio" 
						name="scheduleType"
						value="scheduled"
						bind:group={scheduleType}
						class="w-4 h-4 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500"
					/>
					<span class="text-sm text-gray-700 group-hover:text-gray-900">Đăng ngay</span>
				</label>
				
				<label class="flex items-center gap-1.5 cursor-pointer group">
					<input 
						type="radio" 
						name="scheduleType"
						value="later"
						bind:group={scheduleType}
						class="w-4 h-4 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500"
					/>
					<span class="text-sm text-gray-700 group-hover:text-gray-900">Hẹn giờ</span>
				</label>
				
				<label class="flex items-center gap-1.5 cursor-pointer group">
					<input 
						type="radio" 
						name="scheduleType"
						value="daily"
						bind:group={scheduleType}
						class="w-4 h-4 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500"
					/>
					<span class="text-sm text-gray-700 group-hover:text-gray-900">Lịch tự động</span>
				</label>
				
				<label class="flex items-center gap-1.5 cursor-pointer group">
					<input 
						type="radio" 
						name="scheduleType"
						value="draft"
						bind:group={scheduleType}
						class="w-4 h-4 text-blue-600 border-gray-300 focus:ring-2 focus:ring-blue-500"
					/>
					<span class="text-sm text-gray-700 group-hover:text-gray-900">Nháp</span>
				</label>
			</div>
		</div>
		
		<!-- Content based on selected option -->
		{#if scheduleType === 'later'}
			<div class="flex items-center gap-2 mt-2">
				<input 
					type="date" 
					bind:value={scheduledDate}
					class="flex-1 px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					placeholder="Ngày đăng"
				/>
				<input 
					type="time" 
					bind:value={scheduledTime}
					class="flex-1 px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					placeholder="Giờ đăng"
				/>
			</div>
		{:else if scheduleType === 'daily'}
			<div class="flex items-center gap-2 mt-2">
				<span class="text-xs text-gray-600">Giờ đăng hàng ngày:</span>
				<input 
					type="time" 
					bind:value={scheduledTime}
					class="px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
			</div>
		{/if}
	</div>
</div>
