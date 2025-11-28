<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast';
	import Toast from '$lib/components/Toast.svelte';
	import { User, Plus, Trash2, RefreshCw, AlertTriangle, CheckCircle, Clock, ChevronDown, ChevronUp, Settings } from 'lucide-svelte';

	let accounts = [];
	let loading = true;
	let expandedAccount = null;

	onMount(async () => {
		await loadAccounts();
	});

	async function loadAccounts() {
		try {
			const result = await api.getAccounts();
			accounts = result || [];
		} catch (error) {
			accounts = [];
			toast.show('Không thể tải danh sách nick', 'error');
		} finally {
			loading = false;
		}
	}

	async function deleteAccount(id) {
		if (!confirm('Bạn có chắc muốn xóa nick này? Các pages được gán sẽ bị bỏ gán.')) return;
		
		const original = [...accounts];
		accounts = accounts.filter(a => a.id !== id);
		
		try {
			await api.deleteAccount(id);
			toast.show('Đã xóa nick', 'success');
		} catch (error) {
			accounts = original;
			toast.show('Không thể xóa nick', 'error');
		}
	}

	function toggleExpand(id) {
		expandedAccount = expandedAccount === id ? null : id;
	}

	function getStatusColor(status) {
		switch (status) {
			case 'active': return 'bg-green-100 text-green-700';
			case 'rate_limited': return 'bg-yellow-100 text-yellow-700';
			case 'disabled': return 'bg-red-100 text-red-700';
			case 'token_expired': return 'bg-gray-100 text-gray-700';
			default: return 'bg-gray-100 text-gray-600';
		}
	}

	function getStatusText(status) {
		switch (status) {
			case 'active': return 'Hoạt động';
			case 'rate_limited': return 'Rate Limited';
			case 'disabled': return 'Đã tắt';
			case 'token_expired': return 'Token hết hạn';
			default: return status;
		}
	}

	function getProgressColor(current, max) {
		const percent = (current / max) * 100;
		if (percent >= 100) return 'bg-red-500';
		if (percent >= 80) return 'bg-yellow-500';
		return 'bg-blue-500';
	}
</script>

<svelte:head>
	<title>Quản lý Nick Facebook - FB Scheduler</title>
</svelte:head>

{#if $toast}
	<Toast message={$toast.message} type={$toast.type} onClose={() => toast.hide()} />
{/if}

<div class="max-w-6xl mx-auto">
	<!-- Header -->
	<div class="mb-6">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Quản lý Nick Facebook</h1>
				<p class="text-sm text-gray-500 mt-1">Quản lý các tài khoản Facebook để đăng bài</p>
			</div>
			<a
				href="/accounts/new"
				class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				<Plus size={18} />
				<span>Thêm Nick</span>
			</a>
		</div>
	</div>

	<!-- Stats Overview -->
	<div class="grid grid-cols-4 gap-4 mb-6">
		<div class="bg-white rounded-lg p-4 border border-gray-200">
			<div class="text-2xl font-semibold text-gray-900">{accounts.length}</div>
			<div class="text-sm text-gray-500">Tổng nick</div>
		</div>
		<div class="bg-white rounded-lg p-4 border border-gray-200">
			<div class="text-2xl font-semibold text-green-600">{accounts.filter(a => a.status === 'active').length}</div>
			<div class="text-sm text-gray-500">Đang hoạt động</div>
		</div>
		<div class="bg-white rounded-lg p-4 border border-gray-200">
			<div class="text-2xl font-semibold text-yellow-600">{accounts.filter(a => a.is_warning).length}</div>
			<div class="text-sm text-gray-500">Sắp đạt giới hạn</div>
		</div>
		<div class="bg-white rounded-lg p-4 border border-gray-200">
			<div class="text-2xl font-semibold text-red-600">{accounts.filter(a => a.status === 'rate_limited').length}</div>
			<div class="text-sm text-gray-500">Bị rate limit</div>
		</div>
	</div>

	<!-- Accounts List -->
	{#if loading}
		<div class="flex items-center justify-center min-h-[300px]">
			<div class="w-10 h-10 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
		</div>
	{:else if accounts.length === 0}
		<div class="bg-gray-50 rounded-xl p-12 text-center border border-gray-200">
			<div class="w-16 h-16 bg-white rounded-full flex items-center justify-center mx-auto mb-4 border border-gray-200">
				<User size={32} class="text-gray-400" />
			</div>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Chưa có nick nào</h3>
			<p class="text-sm text-gray-500 mb-6">Thêm nick Facebook để bắt đầu quản lý pages</p>
			<a href="/accounts/new" class="inline-flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700">
				<Plus size={18} />
				<span>Thêm Nick đầu tiên</span>
			</a>
		</div>
	{:else}
		<div class="space-y-3">
			{#each accounts as account}
				<div class="bg-white rounded-lg border border-gray-200 overflow-hidden">
					<!-- Main Row -->
					<div class="p-4">
						<div class="flex items-center gap-4">
							<!-- Avatar -->
							{#if account.profile_picture_url}
								<img src={account.profile_picture_url} alt={account.fb_user_name} class="w-12 h-12 rounded-full flex-shrink-0" />
							{:else}
								<div class="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center flex-shrink-0">
									<User size={24} class="text-white" />
								</div>
							{/if}

							<!-- Info -->
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<h3 class="font-medium text-gray-900 truncate">{account.fb_user_name || 'Nick ' + account.fb_user_id.slice(-4)}</h3>
									<span class="px-2 py-0.5 text-xs font-medium rounded-full {getStatusColor(account.status)}">
										{getStatusText(account.status)}
									</span>
								</div>
								<div class="flex items-center gap-4 mt-1 text-sm text-gray-500">
									<span>{account.pages_count}/{account.max_pages} pages</span>
									<span>•</span>
									<span>{account.posts_today}/{account.max_posts_per_day} bài hôm nay</span>
									{#if account.token_days_left !== undefined}
										<span>•</span>
										<span class="{account.token_days_left <= 7 ? 'text-red-600' : ''}">
											Token: {account.token_days_left} ngày
										</span>
									{/if}
								</div>
							</div>

							<!-- Progress Bar -->
							<div class="w-32 flex-shrink-0">
								<div class="flex justify-between text-xs text-gray-500 mb-1">
									<span>Bài/ngày</span>
									<span>{account.posts_today}/{account.max_posts_per_day}</span>
								</div>
								<div class="h-2 bg-gray-100 rounded-full overflow-hidden">
									<div 
										class="h-full rounded-full transition-all {getProgressColor(account.posts_today, account.max_posts_per_day)}"
										style="width: {Math.min((account.posts_today / account.max_posts_per_day) * 100, 100)}%"
									></div>
								</div>
							</div>

							<!-- Actions -->
							<div class="flex items-center gap-2">
								<button
									on:click={() => toggleExpand(account.id)}
									class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									title="Chi tiết"
								>
									{#if expandedAccount === account.id}
										<ChevronUp size={18} />
									{:else}
										<ChevronDown size={18} />
									{/if}
								</button>
								<a
									href="/accounts/{account.id}"
									class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									title="Cài đặt"
								>
									<Settings size={18} />
								</a>
								<button
									on:click={() => deleteAccount(account.id)}
									class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
									title="Xóa"
								>
									<Trash2 size={18} />
								</button>
							</div>
						</div>
					</div>

					<!-- Expanded Details -->
					{#if expandedAccount === account.id}
						<div class="px-4 pb-4 pt-2 border-t border-gray-100 bg-gray-50">
							<div class="grid grid-cols-3 gap-4 text-sm">
								<div>
									<div class="text-gray-500 mb-1">FB User ID</div>
									<div class="font-mono text-gray-900">{account.fb_user_id}</div>
								</div>
								<div>
									<div class="text-gray-500 mb-1">Lần đăng cuối</div>
									<div class="text-gray-900">
										{account.last_post_at ? new Date(account.last_post_at).toLocaleString('vi-VN') : 'Chưa có'}
									</div>
								</div>
								<div>
									<div class="text-gray-500 mb-1">Lỗi liên tiếp</div>
									<div class="text-gray-900">{account.consecutive_failures || 0} lần</div>
								</div>
							</div>
							{#if account.notes}
								<div class="mt-3 pt-3 border-t border-gray-200">
									<div class="text-gray-500 text-sm mb-1">Ghi chú</div>
									<div class="text-gray-900 text-sm">{account.notes}</div>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
