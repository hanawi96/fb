<script>
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import Button from '$lib/components/Button.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import { Upload, X, Image as ImageIcon } from 'lucide-svelte';
	
	// Accept SvelteKit props
	export let data = undefined;
	export let params = undefined;
	
	let content = '';
	let images = [];
	let uploading = false;
	let saving = false;
	
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
				toast.show('Lỗi upload ảnh: ' + error.message, 'error');
			}
		}
		
		uploading = false;
	}
	
	function removeImage(index) {
		images = images.filter((_, i) => i !== index);
	}
	
	async function savePost() {
		if (!content.trim()) {
			toast.show('Vui lòng nhập nội dung', 'warning');
			return;
		}
		
		saving = true;
		
		try {
			await api.createPost({
				content,
				media_urls: images,
				media_type: images.length > 0 ? 'photo' : 'text',
				status: 'draft'
			});
			
			toast.show('Đã lưu bài viết', 'success');
			setTimeout(() => goto('/schedule'), 1000);
		} catch (error) {
			toast.show('Lỗi lưu bài: ' + error.message, 'error');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Tạo bài mới - FB Scheduler</title>
</svelte:head>

{#if $toast}
	<Toast 
		message={$toast.message} 
		type={$toast.type} 
		duration={$toast.duration || 3000}
		onClose={() => toast.hide()} 
	/>
{/if}

<div class="max-w-3xl">
	<h1 class="text-3xl font-bold mb-2">Tạo bài viết mới</h1>
	<p class="text-gray-600 mb-8">Viết nội dung và thêm hình ảnh cho bài viết</p>
	
	<div class="card">
		<div class="mb-6">
			<label class="block text-sm font-medium text-gray-700 mb-2">
				Nội dung bài viết
			</label>
			<textarea
				bind:value={content}
				placeholder="Nhập nội dung bài viết..."
				rows="8"
				class="input resize-none"
			></textarea>
			<p class="text-sm text-gray-500 mt-1">{content.length} ký tự</p>
		</div>
		
		<div class="mb-6">
			<label class="block text-sm font-medium text-gray-700 mb-2">
				Hình ảnh (tối đa 10 ảnh)
			</label>
			
			<div class="grid grid-cols-2 md:grid-cols-3 gap-4 mb-4">
				{#each images as image, index}
					<div class="relative group">
						<img src={image} alt="Preview" class="w-full h-32 object-cover rounded-lg" />
						<button
							on:click={() => removeImage(index)}
							class="absolute top-2 right-2 p-1 bg-red-600 text-white rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
						>
							<X size={16} />
						</button>
					</div>
				{/each}
				
				{#if images.length < 10}
					<label class="flex flex-col items-center justify-center h-32 border-2 border-dashed border-gray-300 rounded-lg cursor-pointer hover:border-primary-500 hover:bg-primary-50 transition-colors">
						<input
							type="file"
							accept="image/*"
							multiple
							on:change={handleImageUpload}
							class="hidden"
							disabled={uploading}
						/>
						{#if uploading}
							<div class="animate-spin rounded-full h-8 w-8 border-4 border-primary-600 border-t-transparent"></div>
						{:else}
							<Upload size={24} class="text-gray-400 mb-2" />
							<span class="text-sm text-gray-600">Thêm ảnh</span>
						{/if}
					</label>
				{/if}
			</div>
		</div>
		
		<div class="flex gap-3">
			<Button on:click={savePost} loading={saving} class="flex-1">
				Lưu bài viết
			</Button>
			<Button variant="secondary" on:click={() => goto('/')}>
				Hủy
			</Button>
		</div>
	</div>
</div>
