<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	
	let authLog = [];
	let dbLog = [];
	let debugLog = [];
	let dbPages = [];
	let accessToken = '';
	
	function log(logArray, message, type = 'info') {
		const timestamp = new Date().toLocaleTimeString();
		logArray.push({ timestamp, message, type });
		logArray = logArray; // Trigger reactivity
	}
	
	async function connectFacebook() {
		try {
			authLog = [];
			log(authLog, '🔄 Đang lấy Facebook Auth URL...', 'info');
			
			const { url } = await api.getFacebookAuthURL();
			
			log(authLog, `✅ Nhận được Auth URL: ${url.substring(0, 100)}...`, 'success');
			
			// Open popup
			const width = 600;
			const height = 700;
			const left = (screen.width - width) / 2;
			const top = (screen.height - height) / 2;
			
			const popup = window.open(
				url,
				'Facebook Login',
				`width=${width},height=${height},left=${left},top=${top}`
			);
			
			log(authLog, '🪟 Đã mở popup Facebook', 'info');
			
			// Listen for callback
			const handleMessage = async (event) => {
				if (event.data.type === 'facebook-callback') {
					log(authLog, `📥 Nhận được callback code: ${event.data.code.substring(0, 20)}...`, 'info');
					popup?.close();
					window.removeEventListener('message', handleMessage);
					
					try {
						log(authLog, '🔄 Đang gửi code đến backend...', 'info');
						
						const result = await api.facebookCallback(event.data.code);
						
						log(authLog, `✅ Thành công! Đã kết nối ${result.count} pages`, 'success');
						result.pages.forEach((page, i) => {
							log(authLog, `  📄 Page ${i+1}: ${page.page_name} (ID: ${page.page_id})`, 'success');
						});
						
						// Auto reload DB pages
						await loadPagesFromDB();
					} catch (error) {
						log(authLog, `❌ Lỗi khi gọi callback: ${error.message}`, 'error');
					}
				}
			};
			
			window.addEventListener('message', handleMessage);
			
		} catch (error) {
			log(authLog, `❌ Lỗi: ${error.message}`, 'error');
		}
	}
	
	async function loadPagesFromDB() {
		try {
			dbLog = [];
			log(dbLog, '🔄 Đang tải pages từ database...', 'info');
			
			dbPages = await api.getPages();
			
			log(dbLog, `✅ Tìm thấy ${dbPages.length} pages trong database`, 'success');
			
			dbPages.forEach((page, i) => {
				log(dbLog, `  📄 Page ${i+1}: ${page.page_name} (ID: ${page.page_id})`, 'info');
			});
			
		} catch (error) {
			log(dbLog, `❌ Lỗi: ${error.message}`, 'error');
		}
	}
	
	async function debugFacebookAPI() {
		if (!accessToken) {
			debugLog = [];
			log(debugLog, '⚠️ Vui lòng nhập access token', 'warning');
			return;
		}
		
		try {
			debugLog = [];
			log(debugLog, `🔄 Đang test với token: ${accessToken.substring(0, 20)}...`, 'info');
			
			const response = await fetch('http://localhost:8080/api/auth/debug/pages', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ access_token: accessToken })
			});
			
			const result = await response.json();
			
			if (response.ok) {
				log(debugLog, `✅ Facebook API trả về ${result.count} pages`, 'success');
				result.pages.forEach((page, i) => {
					log(debugLog, `  📄 Page ${i+1}: ${page.name} (ID: ${page.id})`, 'success');
				});
			} else {
				log(debugLog, `❌ Lỗi: ${result.error}`, 'error');
			}
			
		} catch (error) {
			log(debugLog, `❌ Lỗi: ${error.message}`, 'error');
		}
	}
	
	onMount(() => {
		loadPagesFromDB();
	});
</script>

<svelte:head>
	<title>🧪 Test Facebook Pages - FB Scheduler</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
	<h1 class="text-3xl font-bold mb-2">🧪 Test Facebook Pages Connection</h1>
	<p class="text-gray-600 mb-8">Debug tool để kiểm tra kết nối Facebook và tìm nguyên nhân chính xác</p>
	
	<!-- Bước 1 -->
	<div class="card mb-6">
		<h2 class="text-xl font-bold mb-2">Bước 1: Kết nối Facebook</h2>
		<p class="text-gray-600 mb-4">Bấm nút bên dưới để mở popup Facebook OAuth</p>
		<button class="btn btn-primary" on:click={connectFacebook}>
			Kết nối Facebook
		</button>
		
		{#if authLog.length > 0}
			<div class="mt-4 bg-gray-50 p-4 rounded-lg max-h-96 overflow-y-auto font-mono text-sm">
				{#each authLog as entry}
					<div class="mb-1" class:text-green-600={entry.type === 'success'} 
						class:text-red-600={entry.type === 'error'}
						class:text-blue-600={entry.type === 'info'}
						class:text-orange-600={entry.type === 'warning'}>
						[{entry.timestamp}] {entry.message}
					</div>
				{/each}
			</div>
		{/if}
	</div>
	
	<!-- Bước 2 -->
	<div class="card mb-6">
		<h2 class="text-xl font-bold mb-2">Bước 2: Xem Pages từ Database</h2>
		<p class="text-gray-600 mb-4">Xem các pages đã được lưu trong database</p>
		<button class="btn btn-secondary" on:click={loadPagesFromDB}>
			Tải Pages từ DB
		</button>
		
		{#if dbPages.length > 0}
			<div class="mt-4 grid grid-cols-1 md:grid-cols-2 gap-4">
				{#each dbPages as page}
					<div class="border rounded-lg p-4 flex items-center gap-3">
						<img src={page.profile_picture_url} alt={page.page_name} class="w-12 h-12 rounded-full">
						<div class="flex-1 min-w-0">
							<div class="font-semibold truncate">{page.page_name}</div>
							<div class="text-sm text-gray-600">ID: {page.page_id}</div>
							<div class="text-xs text-gray-500">{page.category}</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
		
		{#if dbLog.length > 0}
			<div class="mt-4 bg-gray-50 p-4 rounded-lg max-h-96 overflow-y-auto font-mono text-sm">
				{#each dbLog as entry}
					<div class="mb-1" class:text-green-600={entry.type === 'success'} 
						class:text-red-600={entry.type === 'error'}
						class:text-blue-600={entry.type === 'info'}
						class:text-orange-600={entry.type === 'warning'}>
						[{entry.timestamp}] {entry.message}
					</div>
				{/each}
			</div>
		{/if}
	</div>
	
	<!-- Bước 3 -->
	<div class="card mb-6">
		<h2 class="text-xl font-bold mb-2">Bước 3: Debug - Kiểm tra Facebook API</h2>
		<p class="text-gray-600 mb-4">Nhập access token để test trực tiếp với Facebook API</p>
		<input 
			type="text" 
			bind:value={accessToken}
			placeholder="Nhập access token (lấy từ backend log sau khi kết nối)..." 
			class="input w-full mb-3"
		/>
		<button class="btn btn-secondary" on:click={debugFacebookAPI}>
			Test Facebook API
		</button>
		
		{#if debugLog.length > 0}
			<div class="mt-4 bg-gray-50 p-4 rounded-lg max-h-96 overflow-y-auto font-mono text-sm">
				{#each debugLog as entry}
					<div class="mb-1" class:text-green-600={entry.type === 'success'} 
						class:text-red-600={entry.type === 'error'}
						class:text-blue-600={entry.type === 'info'}
						class:text-orange-600={entry.type === 'warning'}>
						[{entry.timestamp}] {entry.message}
					</div>
				{/each}
			</div>
		{/if}
	</div>
	
	<!-- Hướng dẫn -->
	<div class="card bg-blue-50 border-blue-200">
		<h3 class="font-bold mb-2">📝 Hướng dẫn sử dụng</h3>
		<ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
			<li>Bấm "Kết nối Facebook" và quan sát popup Facebook có hiện modal chọn pages không</li>
			<li>Kiểm tra backend terminal log để xem số pages Facebook trả về</li>
			<li>So sánh số pages trong DB với số pages bạn thực sự có trên Facebook</li>
			<li>Nếu thiếu pages, copy access token từ backend log và test ở Bước 3</li>
			<li>Kết quả sẽ cho biết chính xác nguyên nhân: Facebook không hiện modal, hoặc API không trả đủ pages</li>
		</ol>
	</div>
</div>

<style>
	.btn {
		@apply px-4 py-2 rounded-lg font-medium transition-colors;
	}
	.btn-primary {
		@apply bg-blue-600 text-white hover:bg-blue-700;
	}
	.btn-secondary {
		@apply bg-gray-600 text-white hover:bg-gray-700;
	}
</style>
