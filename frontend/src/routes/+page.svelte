<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { Users, FileText, Calendar, CheckCircle } from 'lucide-svelte';
	
	// Accept SvelteKit props
	export let data = undefined;
	export let params = undefined;
	
	let stats = {
		pages: 0,
		posts: 0,
		scheduled: 0,
		published: 0
	};
	
	let loading = true;
	
	onMount(async () => {
		try {
			const [pages, posts, scheduled] = await Promise.all([
				api.getPages(),
				api.getPosts(100, 0),
				api.getScheduledPosts('', 100, 0)
			]);
			
			stats.pages = pages.filter(p => p.is_active).length;
			stats.posts = posts.length;
			stats.scheduled = scheduled.filter(s => s.status === 'pending').length;
			stats.published = scheduled.filter(s => s.status === 'completed').length;
		} catch (error) {
			console.error('Failed to load stats:', error);
		} finally {
			loading = false;
		}
	});
	
	const statCards = [
		{ label: 'Pages đang hoạt động', value: stats.pages, icon: Users, color: 'blue' },
		{ label: 'Tổng bài viết', value: stats.posts, icon: FileText, color: 'purple' },
		{ label: 'Bài chờ đăng', value: stats.scheduled, icon: Calendar, color: 'yellow' },
		{ label: 'Đã đăng thành công', value: stats.published, icon: CheckCircle, color: 'green' }
	];
</script>

<svelte:head>
	<title>Dashboard - FB Scheduler</title>
</svelte:head>

<div>
	<h1 class="text-3xl font-bold mb-2">Dashboard</h1>
	<p class="text-gray-600 mb-8">Tổng quan hệ thống đăng bài Facebook</p>
	
	{#if loading}
		<div class="text-center py-12">
			<div class="inline-block animate-spin rounded-full h-8 w-8 border-4 border-primary-600 border-t-transparent"></div>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
			{#each statCards as stat}
				<div class="card">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm text-gray-600 mb-1">{stat.label}</p>
							<p class="text-3xl font-bold text-{stat.color}-600">{stat.value}</p>
						</div>
						<div class="p-3 bg-{stat.color}-100 rounded-lg">
							<svelte:component this={stat.icon} size={24} class="text-{stat.color}-600" />
						</div>
					</div>
				</div>
			{/each}
		</div>
		
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<div class="card">
				<h2 class="text-xl font-semibold mb-4">Bắt đầu nhanh</h2>
				<div class="space-y-3">
					<a href="/pages" class="block p-4 border border-gray-200 rounded-lg hover:border-primary-500 hover:bg-primary-50 transition-colors">
						<h3 class="font-medium text-gray-900">1. Kết nối Facebook Pages</h3>
						<p class="text-sm text-gray-600 mt-1">Đăng nhập và chọn các pages bạn muốn quản lý</p>
					</a>
					<a href="/posts/new" class="block p-4 border border-gray-200 rounded-lg hover:border-primary-500 hover:bg-primary-50 transition-colors">
						<h3 class="font-medium text-gray-900">2. Tạo bài viết mới</h3>
						<p class="text-sm text-gray-600 mt-1">Viết nội dung và upload hình ảnh</p>
					</a>
					<a href="/schedule" class="block p-4 border border-gray-200 rounded-lg hover:border-primary-500 hover:bg-primary-50 transition-colors">
						<h3 class="font-medium text-gray-900">3. Hẹn giờ đăng bài</h3>
						<p class="text-sm text-gray-600 mt-1">Chọn thời gian và pages để đăng</p>
					</a>
				</div>
			</div>
			
			<div class="card">
				<h2 class="text-xl font-semibold mb-4">Hướng dẫn sử dụng</h2>
				<div class="space-y-4 text-sm text-gray-600">
					<div>
						<h3 class="font-medium text-gray-900 mb-1">📱 Kết nối Pages</h3>
						<p>Vào mục "Quản lý Pages" và đăng nhập Facebook để kết nối các pages của bạn.</p>
					</div>
					<div>
						<h3 class="font-medium text-gray-900 mb-1">✍️ Tạo bài viết</h3>
						<p>Viết nội dung, thêm hình ảnh (tối đa 10 ảnh), và lưu bài viết.</p>
					</div>
					<div>
						<h3 class="font-medium text-gray-900 mb-1">⏰ Hẹn giờ đăng</h3>
						<p>Chọn bài viết, chọn pages, và đặt thời gian đăng. Hệ thống sẽ tự động đăng.</p>
					</div>
					<div>
						<h3 class="font-medium text-gray-900 mb-1">📊 Theo dõi</h3>
						<p>Xem lịch sử đăng bài và trạng thái trong mục "Lịch sử".</p>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
