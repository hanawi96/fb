<svelte:head>
	<title>Đang xử lý...</title>
	<!-- Xử lý ngay khi page load, trước cả Svelte -->
	<script>
		(function() {
			const urlParams = new URLSearchParams(window.location.search);
			const code = urlParams.get('code');
			
			if (code && window.opener) {
				// Gửi message ngay lập tức
				window.opener.postMessage({ type: 'facebook-callback', code: code }, '*');
				// Đóng popup ngay
				window.close();
			} else if (!window.opener) {
				// Nếu không có opener, redirect về pages
				window.location.href = '/pages';
			}
		})();
	</script>
</svelte:head>

<script>
	// Accept SvelteKit props
	export let data = undefined;
	export let params = undefined;
</script>

<!-- Blank page - không render gì -->
<div style="background: #fff; position: fixed; inset: 0;"></div>
