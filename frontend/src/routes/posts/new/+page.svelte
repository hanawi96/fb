<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import Toast from '$lib/components/Toast.svelte';
	import AiPanel from '$lib/components/post-editor/AiPanel.svelte';
	import EditorToolbar from '$lib/components/post-editor/EditorToolbar.svelte';
	import InfoPanel from '$lib/components/post-editor/InfoPanel.svelte';
	import PostOptions from '$lib/components/post-editor/PostOptions.svelte';
	import PageSelector from '$lib/components/post-editor/PageSelector.svelte';
	import ImageUploader from '$lib/components/post-editor/ImageUploader.svelte';
	import VideoUploader from '$lib/components/post-editor/VideoUploader.svelte';
	import LinkPreview from '$lib/components/post-editor/LinkPreview.svelte';
	import PostPreviewModal from '$lib/components/post-editor/PostPreviewModal.svelte';
	import PrivacySettings from '$lib/components/post-editor/PrivacySettings.svelte';
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
	
	// Auto detect links in content
	$: {
		if (content && !linkPreview && !fetchingLink) {
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
		if (!content.trim() && images.length === 0 && videos.length === 0) {
			toast.show('Vui lòng nhập nội dung hoặc thêm media', 'warning');
			return;
		}
		
		if (selectedPages.length === 0) {
			toast.show('Vui lòng chọn ít nhất 1 trang', 'warning');
			return;
		}
		
		posting = true;
		try {
			const postData = {
				content,
				media_urls: videos.length > 0 ? videos : images,
				media_type: videos.length > 0 ? 'video' : (images.length > 0 ? 'photo' : 'text'),
				page_ids: selectedPages.map(p => p.id),
				privacy: privacy,
				post_mode: postMode
			};
			
			// Add link if exists
			if (linkPreview) {
				postData.link = linkPreview.url;
			}
			
			if (scheduleType === 'draft') {
				// Lưu nháp
				postData.status = 'draft';
				await api.createPost(postData);
				toast.show('Đã lưu nháp', 'success');
			} else if (scheduleType === 'scheduled') {
				// Đăng ngay lập tức
				if (images.length > 5) {
					toast.show('Đang xử lý nhiều ảnh, vui lòng đợi...', 'info');
				}
				
				const result = await api.publishPost(postData);
				
				// Kiểm tra kết quả
				const failedPages = result.results?.filter(r => r.status === 'failed') || [];
				
				if (failedPages.length === 0) {
					toast.show(`Đã đăng bài thành công lên ${result.results.length} trang!`, 'success');
				} else if (failedPages.length === result.results.length) {
					toast.show('Đăng bài thất bại trên tất cả các trang', 'error');
				} else {
					toast.show(`Đăng thành công ${result.results.length - failedPages.length}/${result.results.length} trang`, 'warning');
				}
			} else {
				// Lên lịch đăng sau
				postData.status = 'draft';
				const post = await api.createPost(postData);
				
				// Tạo schedule
				const scheduleData = {
					post_id: post.id,
					page_ids: selectedPages.map(p => p.id),
					scheduled_time: scheduleType === 'later' && scheduledDate && scheduledTime 
						? `${scheduledDate}T${scheduledTime}:00Z`
						: new Date(Date.now() + 60000).toISOString() // 1 phút sau
				};
				
				await api.schedulePost(scheduleData);
				toast.show('Đã lên lịch đăng bài!', 'success');
			}
			
			// Reset form
			content = '';
			images = [];
			videos = [];
			linkPreview = null;
			selectedPages = [];
		} catch (error) {
			console.error('Publish error:', error);
			toast.show('Lỗi: ' + error.message, 'error');
		} finally {
			posting = false;
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
					class="px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
				>
					Xem trước
				</button>
				<button 
					on:click={saveDraft}
					class="px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
				>
					Lưu nháp
				</button>
				<button 
					on:click={publishPost}
					disabled={posting}
					class="px-4 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
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
				<EditorToolbar onAiClick={() => showAiPanel = !showAiPanel} />
				
				<!-- Content Editor -->
				<div class="p-3">
					<textarea
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
				<span>{content.length} ký tự</span>
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


