<script>
	import '../app.css';
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { Home, Calendar, FileText, History, Facebook, LayoutDashboard, ChevronRight, LogOut, User, ChevronDown, Users, Clock } from 'lucide-svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';
	
	let showUserMenu = false;
	
	function toggleUserMenu() {
		showUserMenu = !showUserMenu;
	}
	
	function closeUserMenu() {
		showUserMenu = false;
	}
	
	// Accept SvelteKit props (even if unused)
	export let data = undefined;
	export let params = undefined;
	
	const navItems = [
		{ href: '/', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/posts/new', label: 'Tạo bài mới', icon: FileText },
		{ href: '/schedule', label: 'Lịch đăng bài', icon: Calendar },
		{ href: '/logs', label: 'Lịch sử', icon: History },
		{ href: '/pages', label: 'Quản lý Pages', icon: Facebook },
		{ href: '/accounts', label: 'Quản lý Nick', icon: Users },
		{ href: '/timeslots', label: 'Khung giờ', icon: Clock }
	];
	
	$: currentPath = $page.url.pathname;
	$: isLoginPage = currentPath === '/login';
	
	onMount(() => {
		auth.init();
	});
	
	// Redirect to login only after auth is initialized
	$: if (browser && $auth.initialized && !$auth.token && !isLoginPage) {
		goto('/login');
	}
	
	// Redirect to home if already logged in and on login page
	$: if (browser && $auth.initialized && $auth.token && isLoginPage) {
		goto('/');
	}
	
	function handleLogout() {
		auth.logout();
		goto('/login');
	}
</script>

{#if isLoginPage}
	<slot />
{:else}
<div class="min-h-screen bg-gray-50">
	<!-- Sidebar - Minimal -->
	<aside class="fixed left-0 top-0 h-full w-64 bg-white border-r border-gray-200 z-50">
		<!-- Logo & Brand -->
		<div class="px-4 py-4 border-b border-gray-100">
			<div class="flex items-center gap-2.5">
				<div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
					<Facebook size={18} class="text-white" />
				</div>
				<div>
					<h1 class="text-base font-semibold text-gray-900">FB Scheduler</h1>
					<p class="text-xs text-gray-500">Đăng bài tự động</p>
				</div>
			</div>
		</div>
		
		<!-- Navigation -->
		<nav class="p-3 space-y-0.5">
			{#each navItems as item}
				{@const isActive = currentPath === item.href}
				<a
					href={item.href}
					class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors
						{isActive 
							? 'bg-blue-50 text-blue-700 font-medium' 
							: 'text-gray-700 hover:bg-gray-50'}"
				>
					<svelte:component 
						this={item.icon} 
						size={18} 
						class="{isActive ? 'text-blue-600' : 'text-gray-500'}"
					/>
					<span class="flex-1">{item.label}</span>
				</a>
			{/each}
		</nav>
		
		<!-- Bottom section - Minimal -->
		<div class="absolute bottom-0 left-0 right-0 p-3 border-t border-gray-100">
			<div class="bg-blue-50 rounded-lg p-3 border border-blue-100">
				<p class="text-xs font-medium text-gray-900 mb-1">Nâng cấp Pro</p>
				<p class="text-xs text-gray-600 mb-2">Mở khóa tính năng cao cấp</p>
				<button class="w-full px-3 py-1.5 bg-blue-600 text-white text-xs font-medium rounded-md hover:bg-blue-700 transition-colors">
					Nâng cấp
				</button>
			</div>
		</div>
	</aside>
	
	<!-- Main content -->
	<main class="ml-64 min-h-screen">
		<!-- Top bar - Minimal -->
		<div class="sticky top-0 z-40 bg-white border-b border-gray-200">
			<div class="px-6 py-3 flex items-center justify-between">
				<!-- Breadcrumb -->
				<div class="flex items-center gap-1.5 text-xs text-gray-500">
					<Home size={14} />
					<ChevronRight size={12} class="text-gray-400" />
					<span class="font-medium text-gray-900">
						{navItems.find(item => item.href === currentPath)?.label || 'Dashboard'}
					</span>
				</div>
				
				<!-- Notifications & User menu -->
				<div class="flex items-center gap-2">
					<NotificationBell />
				</div>
				
				<!-- User menu -->
				<div class="relative">
					<button 
						on:click={toggleUserMenu}
						class="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 rounded-lg transition-colors"
					>
						<div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center">
							<User size={16} class="text-white" />
						</div>
						<div class="text-left">
							<p class="text-sm font-medium text-gray-900">{$auth.username || 'Admin'}</p>
							<p class="text-xs text-gray-500">Quản trị viên</p>
						</div>
						<ChevronDown size={16} class="text-gray-400 transition-transform {showUserMenu ? 'rotate-180' : ''}" />
					</button>
					
					{#if showUserMenu}
						<!-- svelte-ignore a11y-click-events-have-key-events -->
						<!-- svelte-ignore a11y-no-static-element-interactions -->
						<div 
							class="fixed inset-0 z-40" 
							on:click={closeUserMenu}
						></div>
						
						<div class="absolute right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 py-2 z-50">
							<!-- User info -->
							<div class="px-4 py-3 border-b border-gray-100">
								<div class="flex items-center gap-3">
									<div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center">
										<User size={20} class="text-white" />
									</div>
									<div>
										<p class="text-sm font-semibold text-gray-900">{$auth.username || 'Admin'}</p>
										<p class="text-xs text-gray-500">Quản trị viên</p>
									</div>
								</div>
							</div>
							
							<!-- Menu items -->
							<div class="py-1">
								<button
									on:click={handleLogout}
									class="w-full flex items-center gap-3 px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors"
								>
									<LogOut size={16} />
									<span>Đăng xuất</span>
								</button>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
		
		<!-- Page content -->
		<div class="p-6">
			<slot />
		</div>
	</main>
</div>
{/if}

<style>
	:global(body) {
		overflow-x: hidden;
	}
</style>
