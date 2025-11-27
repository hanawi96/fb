<script>
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { toast } from '$lib/stores/toast';
	import Toast from '$lib/components/Toast.svelte';
	import { LogIn } from 'lucide-svelte';

	// Accept SvelteKit props
	export let data = undefined;
	export let params = undefined;

	let username = '';
	let password = '';
	let loading = false;

	async function handleLogin() {
		if (!username || !password) {
			toast.show('Vui lòng nhập đầy đủ thông tin', 'warning');
			return;
		}

		loading = true;
		try {
			const response = await fetch('http://localhost:8080/api/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ username, password })
			});

			if (!response.ok) {
				throw new Error('Sai tên đăng nhập hoặc mật khẩu');
			}

			const data = await response.json();
			auth.login(data.token, username);
			toast.show('Đăng nhập thành công!', 'success');
			setTimeout(() => goto('/'), 100);
		} catch (error) {
			toast.show(error.message, 'error');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Đăng nhập - FB Scheduler</title>
</svelte:head>

{#if $toast}
	<Toast 
		message={$toast.message} 
		type={$toast.type} 
		duration={$toast.duration || 3000}
		onClose={() => toast.hide()} 
	/>
{/if}

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4">
	<div class="w-full max-w-md">
		<div class="bg-white rounded-2xl shadow-xl p-8">
			<div class="text-center mb-8">
				<div class="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center mx-auto mb-4">
					<LogIn size={32} class="text-white" />
				</div>
				<h1 class="text-2xl font-bold text-gray-900">Đăng nhập</h1>
				<p class="text-sm text-gray-500 mt-2">FB Scheduler Dashboard</p>
			</div>

			<form on:submit|preventDefault={handleLogin} class="space-y-5">
				<div>
					<label for="username" class="block text-sm font-medium text-gray-700 mb-2">
						Tên đăng nhập
					</label>
					<input
						id="username"
						type="text"
						bind:value={username}
						disabled={loading}
						class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed"
						placeholder="Nhập tên đăng nhập"
						autocomplete="username"
					/>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 mb-2">
						Mật khẩu
					</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						disabled={loading}
						class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed"
						placeholder="Nhập mật khẩu"
						autocomplete="current-password"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{#if loading}
						<div class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
						<span>Đang đăng nhập...</span>
					{:else}
						<LogIn size={20} />
						<span>Đăng nhập</span>
					{/if}
				</button>
			</form>

			<div class="mt-6 pt-6 border-t border-gray-200 text-center text-sm text-gray-500">
				<p>Tài khoản mặc định: <span class="font-medium text-gray-700">admin</span></p>
				<p class="mt-1">Mật khẩu: <span class="font-medium text-gray-700">admin123</span></p>
			</div>
		</div>
	</div>
</div>
