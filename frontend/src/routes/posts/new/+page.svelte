<script>
	import { onMount, tick } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import Toast from '$lib/components/Toast.svelte';
	import AiPanel from '$lib/components/post-editor/AiPanel.svelte';
	import EditorToolbar from '$lib/components/post-editor/EditorToolbar.svelte';
	import InfoPanel from '$lib/components/post-editor/InfoPanel.svelte';
	import CharacterCounter from '$lib/components/post-editor/CharacterCounter.svelte';
	import { insertAtCursor } from '$lib/utils/textFormatting';
	import PostOptions from '$lib/components/post-editor/PostOptions.svelte';
	import PageSelector from '$lib/components/post-editor/PageSelector.svelte';
	import ImageUploader from '$lib/components/post-editor/ImageUploader.svelte';
	import VideoUploader from '$lib/components/post-editor/VideoUploader.svelte';
	import LinkPreview from '$lib/components/post-editor/LinkPreview.svelte';
	import PostPreviewModal from '$lib/components/post-editor/PostPreviewModal.svelte';
	import PrivacySettings from '$lib/components/post-editor/PrivacySettings.svelte';
	import ConflictWarningModal from '$lib/components/ConflictWarningModal.svelte';
	import { Clock } from 'lucide-svelte';
	
	let pages = [];
	let selectedPages = [];
	let content = '';
	let images = [];
	let videos = [];
	let linkPreview = null;
	let uploading = false;
	let showAiPanel = false;
	let showPreview = false;
	let aiPrompt = '';
	let generatingAi = false;
	let scheduleType = 'scheduled';
	let scheduledDate = '';
	let scheduledTime = '';
	let posting = false;
	let privacy = 'public';
	let mediaType = 'image'; // 'image' or 'video'
	let fetchingLink = false;
	let postMode = 'album'; // 'album' | 'individual'
	let textareaElement = null;
	
	// Conflict handling
	let showConflictModal = false;
	let conflictData = null;
	let pendingScheduleData = null;
	
	// Handle emoji selection
	function handleEmojiSelect(emoji) {
		if (textareaElement) {
			insertAtCursor(textareaElement, emoji);
			content = textareaElement.value;
		}
	}
	
	// Handle hashtag selection
	function handleHashtagSelect(hashtagsText) {
		if (textareaElement) {
			const start = textareaElement.selectionStart;
			const value = textareaElement.value;
			const charBefore = value[start - 1];
			
			// Thêm space nếu cần
			const prefix = (!charBefore || charBefore === '\n' || charBefore === ' ') ? '' : '\n';
			insertAtCursor(textareaElement, prefix + hashtagsText + ' ');
			content = textareaElement.value;
		}
	}
	
	// Auto detect links in content
	$: {
		if (content && content.length > 10 && !linkPreview && !fetchingLink) {
			const urlRegex = /(https?:\/\/[^\s]+)/g;
			const urls = content.match(urlRegex);
			if (urls && urls.length > 0) {
				fetchLinkPreview(urls[0]);
			}
		}
	}
	
	onMount(async () => {
		try {
			pages = await api.getPages();
			selectedPages = [];
		} catch (error) {
			console.error('Error loading pages:', error);
		}
	});
	
	async function handleImageUpload(event) {
		const files = Array.from(event.target.files);
		if (images.length + files.length > 10) {
			toast.show('Tối đa 10 ảnh', 'warning');
			return;
		}
		
		uploading = true;
		for (const file of files) {
			try {
				const { url } = await api.uploadImage(file);
				images = [...images, url];
			} catch (error) {
				toast.show('Lỗi upload: ' + error.message, 'error');
			}
		}
		uploading = false;
	}
	
	function removeImage(index) {
		images = images.filter((_, i) => i !== index);
	}
	
	async function handleVideoUpload(event) {
		const files = Array.from(event.target.files);
		if (videos.length + files.length > 1) {
			toast.show('Chỉ được upload 1 video', 'warning');
			return;
		}
		
		uploading = true;
		for (const file of files) {
			try {
				const { url } = await api.uploadImage(file); // Reuse same endpoint
				videos = [...videos, url];
			} catch (error) {
				toast.show('Lỗi upload video: ' + error.message, 'error');
			}
		}
		uploading = false;
	}
	
	function removeVideo(index) {
		videos = videos.filter((_, i) => i !== index);
	}
	
	async function fetchLinkPreview(url) {
		fetchingLink = true;
		try {
			// Mock API call - replace with real endpoint
			const response = await fetch(`https://api.linkpreview.net/?key=YOUR_KEY&q=${encodeURIComponent(url)}`);
			const data = await response.json();
			linkPreview = {
				url: url,
				title: data.title,
				description: data.description,
				image: data.image
			};
		} catch (error) {
			// Fallback: basic preview
			linkPreview = {
				url: url,
				title: url,
				description: '',
				image: null
			};
		} finally {
			fetchingLink = false;
		}
	}
	
	function removeLinkPreview() {
		linkPreview = null;
	}
	
	function generateWithAi() {
		if (!aiPrompt.trim()) {
			toast.show('Nhập yêu cầu cho AI', 'warning');
			return;
		}
		
		generatingAi = true;
		// Demo: Simulate AI generation
		setTimeout(() => {
			content = `🎉 ${aiPrompt}\n\nĐây là nội dung được tạo bởi AI dựa trên yêu cầu của bạn. Bạn có thể chỉnh sửa nội dung này trước khi đăng.\n\n#AI #ContentCreation #SocialMedia`;
			generatingAi = false;
			showAiPanel = false;
			aiPrompt = '';
			toast.show('Đã tạo nội dung bằng AI!', 'success');
		}, 1500);
	}
	
	// Hàm schedule post (tách riêng để tái sử dụng)
	async function schedulePost(postData, localDateTime, allowConflict = false) {
		postData.status = 'draft';
		const post = await api.createPost(postData);
		
		// Nếu cho phép conflict, thêm random offset
		let scheduledTime = localDateTime;
		if (allowConflict && conflictData && conflictData.conflict_pages.length > 0) {
			// Thêm random 5-30 giây cho pages có xung đột
			const randomSeconds = Math.floor(Math.random() * 26) + 5;
			scheduledTime = new Date(localDateTime.getTime() + randomSeconds * 1000);
		}
		
		// Schedule với thời gian chính xác
		await api.schedulePost({
			post_id: post.id,
			page_ids: selectedPages.map(p => p.id),
			scheduled_time: scheduledTime.toISOString()
		});
		
		// Hiển thị thông báo
		const formattedTime = localDateTime.toLocaleString('vi-VN', {
			day: '2-digit', month: '2-digit', year: 'numeric',
			hour: '2-digit', minute: '2-digit',
			timeZone: 'Asia/Ho_Chi_Minh'
		});
		toast.show(`Đã hẹn đăng bài lúc ${formattedTime}`, 'success');
		
		// KHÔNG reset form ở đây - sẽ được xử lý trong finally block của publishPost
	}
	
	// Xử lý khi user chọn "Có, đăng luôn"
	async function handleConflictConfirm() {
		showConflictModal = false;
		posting = true;
		const startTime = Date.now();
		
		try {
			await schedulePost(pendingScheduleData.postData, pendingScheduleData.localDateTime, true);
		} catch (error) {
			console.error('Schedule error:', error);
			toast.show('Lỗi: ' + error.message, 'error');
		} finally {
			// Đảm bảo button hiển thị loading ít nhất 300ms
			const elapsedTime = Date.now() - startTime;
			const minDelay = 300;
			const remainingDelay = Math.max(0, minDelay - elapsedTime);
			
			if (remainingDelay > 0) {
				await new Promise(resolve => setTimeout(resolve, remainingDelay));
			}
			
			posting = false;
			await tick();
			
			// Reset form sau khi button về trạng thái bình thường
			setTimeout(() => {
				resetForm();
			}, 150);
			
			pendingScheduleData = null;
			conflictData = null;
		}
	}
	
	// Xử lý khi user chọn "Không"
	function handleConflictCancel() {
		showConflictModal = false;
		posting = false;
		pendingScheduleData = null;
		conflictData = null;
		// Giữ nguyên form để user chọn lại
	}
	
	// Xử lý khi user chọn "Lịch tự động"
	async function handleConflictAutoSchedule() {
		showConflictModal = false;
		posting = true;
		const startTime = Date.now();
		
		try {
			const postData = pendingScheduleData.postData;
			postData.status = 'draft';
			const post = await api.createPost(postData);
			const pageIds = selectedPages.map(p => p.id);
			const preferredDate = pendingScheduleData.localDateTime.toISOString().split('T')[0];
			const result = await api.scheduleWithPreview(
				post.id,
				pageIds,
				preferredDate,
				true
			);
			
			if (result.success_count > 0) {
				toast.show(`Đã tự động lên lịch cho ${result.success_count} trang!`, 'success');
			} else {
				toast.show('Không thể lên lịch tự động', 'warning');
			}
		} catch (error) {
			console.error('Auto schedule error:', error);
			toast.show('Lỗi: ' + error.message, 'error');
		} finally {
			// Đảm bảo button hiển thị loading ít nhất 300ms
			const elapsedTime = Date.now() - startTime;
			const minDelay = 300;
			const remainingDelay = Math.max(0, minDelay - elapsedTime);
			
			if (remainingDelay > 0) {
				await new Promise(resolve => setTimeout(resolve, remainingDelay));
			}
			
			posting = false;
			await tick();
			
			// Reset form sau khi button về trạng thái bình thường
			setTimeout(() => {
				resetForm();
			}, 150);
			
			pendingScheduleData = null;
			conflictData = null;
		}
	}
	
	// Reset form
	function resetForm() {
		console.log('[DEBUG] resetForm called, posting state before:', posting);
		console.log('[DEBUG] Resetting content...');
		content = '';
		console.log('[DEBUG] Resetting images...');
		images = [];
		console.log('[DEBUG] Resetting videos...');
		videos = [];
		console.log('[DEBUG] Resetting linkPreview...');
		linkPreview = null;
		// KHÔNG reset selectedPages để tránh flicker
		// selectedPages = [];
		console.log('[DEBUG] Resetting scheduledDate...');
		scheduledDate = '';
		console.log('[DEBUG] Resetting scheduledTime...');
		scheduledTime = '';
		// KHÔNG reset scheduleType để tránh re-render PostOptions component
		// Giữ nguyên chế độ đã chọn để user có thể đăng bài tiếp theo với cùng chế độ
		// console.log('[DEBUG] Resetting scheduleType...');
		// scheduleType = 'scheduled';
		console.log('[DEBUG] resetForm done, posting state after:', posting);
	}
	
	async function saveDraft() {
		if (!content.trim()) {
			toast.show('Vui lòng nhập nội dung', 'warning');
			return;
		}
		
		try {
			await api.createPost({
				content,
				media_urls: images,
				media_type: images.length > 0 ? 'photo' : 'text',
				status: 'draft'
			});
			toast.show('Đã lưu nháp', 'success');
		} catch (error) {
			toast.show('Lỗi: ' + error.message, 'error');
		}
	}
	
	async function publishPost() {
		console.log('[DEBUG] publishPost called, posting state:', posting);
		
		if (!content.trim() && images.length === 0 && videos.length === 0) {
			toast.show('Vui lòng nhập nội dung hoặc thêm media', 'warning');
			return;
		}
		
		if (selectedPages.length === 0) {
			toast.show('Vui lòng chọn ít nhất 1 trang', 'warning');
			return;
		}
		
		// Tạo postData
		const postData = {
			content,
			media_urls: videos.length > 0 ? videos : images,
			media_type: videos.length > 0 ? 'video' : (images.length > 0 ? 'photo' : 'text'),
			page_ids: selectedPages.map(p => p.id),
			privacy: privacy,
			post_mode: postMode
		};
		
		if (linkPreview) {
			postData.link = linkPreview.url;
		}
		
		console.log('[DEBUG] Setting posting = true, scheduleType:', scheduleType);
		posting = true;
		// Đợi DOM update xong trước khi tiếp tục
		await tick();
		console.log('[DEBUG] After tick(), DOM updated');
		
		// Lưu thời điểm bắt đầu để đảm bảo minimum delay
		const startTime = Date.now();
		
		try {
			if (scheduleType === 'draft') {
				console.log('[DEBUG] Draft mode');
				// Lưu nháp
				postData.status = 'draft';
				await api.createPost(postData);
				toast.show('Đã lưu nháp', 'success');
				console.log('[DEBUG] Success, will reset form after posting = false');
				
			} else if (scheduleType === 'scheduled') {
				console.log('[DEBUG] Scheduled mode (publish now)');
				// Đăng ngay
				if (images.length > 5) {
					toast.show('Đang xử lý nhiều ảnh, vui lòng đợi...', 'info');
				}
				const result = await api.publishPost(postData);
				const failedPages = result.results?.filter(r => r.status === 'failed') || [];
				
				if (failedPages.length === 0) {
					toast.show(`Đã đăng bài thành công lên ${result.results.length} trang!`, 'success');
				} else if (failedPages.length === result.results.length) {
					toast.show('Đăng bài thất bại trên tất cả các trang', 'error');
				} else {
					toast.show(`Đăng thành công ${result.results.length - failedPages.length}/${result.results.length} trang`, 'warning');
				}
				console.log('[DEBUG] Success, will reset form after posting = false');
				
			} else if (scheduleType === 'auto') {
				console.log('[DEBUG] Auto schedule mode - START');
				// Lịch tự động
				postData.status = 'draft';
				console.log('[DEBUG] Creating post...');
				const post = await api.createPost(postData);
				console.log('[DEBUG] Post created:', post.id);
				
				const pageIds = selectedPages.map(p => p.id);
				const preferredDate = scheduledDate || new Date().toISOString().split('T')[0];
				console.log('[DEBUG] Calling scheduleWithPreview, pageIds:', pageIds, 'preferredDate:', preferredDate);
				
				const result = await api.scheduleWithPreview(
					post.id,
					pageIds,
					preferredDate,
					true
				);
				console.log('[DEBUG] scheduleWithPreview result:', result);
				
				if (result.success_count > 0) {
					toast.show(`Đã lên lịch tự động cho ${result.success_count} trang!`, 'success');
				} else {
					toast.show('Không thể lên lịch tự động. Vui lòng kiểm tra cấu hình khung giờ.', 'warning');
				}
				console.log('[DEBUG] Success, will reset form after posting = false');
				
			} else if (scheduleType === 'later') {
				console.log('[DEBUG] Later mode');
				// Hẹn giờ cụ thể - Check conflict trước
				if (!scheduledDate || !scheduledTime) {
					toast.show('Vui lòng chọn ngày và giờ đăng', 'warning');
					console.log('[DEBUG] Missing date/time, setting posting = false');
					posting = false;
					return;
				}
				
				// Tạo datetime từ local time
				const localDateTime = new Date(`${scheduledDate}T${scheduledTime}:00`);
				const now = new Date();
				
				// Validate: không cho phép đặt lịch trong quá khứ
				if (localDateTime <= now) {
					toast.show('Thời gian đăng phải sau thời điểm hiện tại', 'warning');
					console.log('[DEBUG] Past time, setting posting = false');
					posting = false;
					return;
				}
				
				// Validate: không cho phép đặt lịch quá xa (30 ngày)
				const maxDate = new Date();
				maxDate.setDate(maxDate.getDate() + 30);
				if (localDateTime > maxDate) {
					toast.show('Chỉ có thể hẹn giờ trong vòng 30 ngày', 'warning');
					console.log('[DEBUG] Too far, setting posting = false');
					posting = false;
					return;
				}
				
				// Check conflict
				const conflictResult = await api.checkScheduleConflict(
					selectedPages.map(p => p.id),
					localDateTime.toISOString()
				);
				
				if (conflictResult.has_conflict) {
					// Có xung đột - hiện modal
					conflictData = conflictResult;
					pendingScheduleData = { postData, localDateTime };
					showConflictModal = true;
					console.log('[DEBUG] Conflict detected, setting posting = false');
					posting = false;
					return;
				}
				
				// Không xung đột - đăng bình thường
				console.log('[DEBUG] No conflict, scheduling post');
				await schedulePost(postData, localDateTime, false);
				console.log('[DEBUG] Schedule post completed, will reset form in finally');
			}
		} catch (error) {
			console.error('[DEBUG] Publish error:', error);
			toast.show('Lỗi: ' + error.message, 'error');
		} finally {
			console.log('[DEBUG] Finally block, setting posting = false');
			
			// Đảm bảo button hiển thị loading ít nhất 300ms để tránh flicker
			const elapsedTime = Date.now() - startTime;
			const minDelay = 300;
			const remainingDelay = Math.max(0, minDelay - elapsedTime);
			
			console.log('[DEBUG] Elapsed time:', elapsedTime, 'ms, remaining delay:', remainingDelay, 'ms');
			
			if (remainingDelay > 0) {
				await new Promise(resolve => setTimeout(resolve, remainingDelay));
			}
			
			posting = false;
			// Đợi DOM update xong sau khi posting = false
			await tick();
			console.log('[DEBUG] After posting = false tick, scheduling reset form');
			// Reset form SAU KHI button đã về trạng thái bình thường
			// Sử dụng setTimeout với delay nhỏ để đảm bảo browser đã paint xong
			if (scheduleType === 'auto' || scheduleType === 'scheduled' || scheduleType === 'draft' || scheduleType === 'later') {
				setTimeout(() => {
					console.log('[DEBUG] setTimeout - now reset form');
					resetForm();
				}, 150);
			}
			console.log('[DEBUG] Finally block done');
		}
	}
</script>

<svelte:head>
	<title>Đăng bài - FB Scheduler</title>
</svelte:head>

{#if $toast}
	<Toast 
		message={$toast.message} 
		type={$toast.type} 
		duration={$toast.duration || 3000}
		onClose={() => toast.hide()} 
	/>
{/if}

<!-- Conflict Warning Modal -->
<ConflictWarningModal
	bind:show={showConflictModal}
	conflictPages={conflictData?.conflict_pages || []}
	scheduledTime={pendingScheduleData?.localDateTime?.toISOString() || ''}
	on:confirm={handleConflictConfirm}
	on:cancel={handleConflictCancel}
	on:autoSchedule={handleConflictAutoSchedule}
/>

<!-- Sử dụng layout chính, chỉ custom nội dung -->
<div class="flex gap-4 h-full">
	<!-- Left: Pages List -->
	<div class="w-80 flex-shrink-0">
		<PageSelector bind:pages bind:selectedPages />
	</div>
	
	<!-- Center: Editor -->
	<div class="flex-1 flex flex-col min-w-0">
		<!-- Top Bar -->
		<div class="bg-white border border-gray-200 rounded-lg px-4 py-2.5 flex items-center justify-between mb-4">
			<h1 class="text-base font-semibold text-gray-900">Đăng bài</h1>
			<div class="flex items-center gap-2">
				<button 
					on:click={() => showPreview = true}
					disabled={posting}
					class="px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
				>
					Xem trước
				</button>
				<button 
					on:click={saveDraft}
					disabled={posting}
					class="px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
				>
					Lưu nháp
				</button>
				<button 
					on:click={publishPost}
					disabled={posting}
					class="px-4 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 min-w-[110px] justify-center"
				>
					{#if posting}
						<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
						<span>Đang đăng...</span>
					{:else}
						<span>Đăng bài</span>
					{/if}
				</button>
			</div>
		</div>
		
		<!-- Editor Area -->
		<div class="flex-1 overflow-y-auto">
			<!-- AI Assistant Panel -->
			<AiPanel 
				bind:show={showAiPanel}
				bind:prompt={aiPrompt}
				bind:generating={generatingAi}
				onGenerate={generateWithAi}
				onClose={() => showAiPanel = false}
			/>
			
			<!-- Main Editor Card -->
			<div class="bg-white rounded-lg border border-gray-200 shadow-sm">
				<!-- Toolbar -->
				<EditorToolbar 
					onAiClick={() => showAiPanel = !showAiPanel}
					onEmojiSelect={handleEmojiSelect}
					onHashtagSelect={handleHashtagSelect}
				/>
				
				<!-- Content Editor -->
				<div class="p-3">
					<textarea
						bind:this={textareaElement}
						bind:value={content}
						placeholder="Bạn đang nghĩ gì?"
						rows="6"
						class="w-full text-gray-900 placeholder-gray-400 focus:outline-none resize-none text-base"
					></textarea>
					
					<!-- Media Upload Area -->
					<div class="flex gap-2 mb-2">
						<button
							on:click={() => mediaType = 'image'}
							class="px-3 py-1 text-xs rounded-md transition-colors {mediaType === 'image' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
						>
							Ảnh
						</button>
						<button
							on:click={() => mediaType = 'video'}
							class="px-3 py-1 text-xs rounded-md transition-colors {mediaType === 'video' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
						>
							Video
						</button>
					</div>
					
					{#if mediaType === 'image'}
						<ImageUploader 
							bind:images
							bind:uploading
							bind:postMode
							onUpload={handleImageUpload}
							onRemove={removeImage}
						/>
					{:else}
						<VideoUploader 
							bind:videos
							bind:uploading
							onUpload={handleVideoUpload}
							onRemove={removeVideo}
						/>
					{/if}
					
					<!-- Link Preview -->
					<LinkPreview 
						preview={linkPreview}
						onRemove={removeLinkPreview}
					/>
				</div>
				
				<!-- Privacy Settings -->
				<PrivacySettings bind:privacy />
				
				<!-- Options Sections -->
				<PostOptions 
					bind:scheduleType
					bind:scheduledDate
					bind:scheduledTime
				/>
			</div>
			
			<!-- Character Count -->
			<div class="mt-2 flex items-center justify-between text-xs text-gray-500">
				<div class="flex items-center gap-2">
					<Clock size={12} />
					<span>Lưu lần cuối: Chưa lưu</span>
				</div>
				<CharacterCounter count={content.length} />
			</div>
		</div>
	</div>
	
	<!-- Right: Info Panel -->
	<div class="w-72 flex-shrink-0">
		<InfoPanel selectedPages={selectedPages} />
	</div>
</div>

<!-- Post Preview Modal -->
<PostPreviewModal 
	bind:show={showPreview}
	{content}
	{images}
	{videos}
	linkPreview={linkPreview}
	{selectedPages}
/>


