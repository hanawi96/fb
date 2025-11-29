/**
 * Datetime utilities for Vietnam timezone (UTC+7)
 */

const VN_TIMEZONE = 'Asia/Ho_Chi_Minh';

/**
 * Format time in HH:mm format (Vietnam timezone)
 * @param {string|Date} dateStr - ISO date string or Date object
 * @returns {string} Formatted time (e.g., "12:30")
 */
export function formatTimeVN(dateStr) {
	if (!dateStr) return '--:--';
	
	const date = new Date(dateStr);
	return date.toLocaleTimeString('vi-VN', {
		hour: '2-digit',
		minute: '2-digit',
		timeZone: VN_TIMEZONE
	});
}

/**
 * Format datetime in full format (Vietnam timezone)
 * @param {string|Date} dateStr - ISO date string or Date object
 * @returns {string} Formatted datetime (e.g., "29/11/2025, 12:30:45")
 */
export function formatDateTimeVN(dateStr) {
	if (!dateStr) return '';
	
	const date = new Date(dateStr);
	return date.toLocaleString('vi-VN', {
		timeZone: VN_TIMEZONE
	});
}

/**
 * Format date in dd/MM/yyyy format (Vietnam timezone)
 * @param {string|Date} dateStr - ISO date string or Date object
 * @returns {string} Formatted date (e.g., "29/11/2025")
 */
export function formatDateVN(dateStr) {
	if (!dateStr) return '';
	
	const date = new Date(dateStr);
	return date.toLocaleDateString('vi-VN', {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
		timeZone: VN_TIMEZONE
	});
}

/**
 * Format datetime with custom format (Vietnam timezone)
 * @param {string|Date} dateStr - ISO date string or Date object
 * @param {object} options - Intl.DateTimeFormat options
 * @returns {string} Formatted datetime
 */
export function formatCustomVN(dateStr, options = {}) {
	if (!dateStr) return '';
	
	const date = new Date(dateStr);
	return date.toLocaleString('vi-VN', {
		...options,
		timeZone: VN_TIMEZONE
	});
}

/**
 * Get current date in YYYY-MM-DD format (Vietnam timezone)
 * @returns {string} Date string (e.g., "2025-11-29")
 */
export function getTodayVN() {
	const now = new Date();
	return now.toLocaleDateString('en-CA', { timeZone: VN_TIMEZONE }); // en-CA gives YYYY-MM-DD
}

/**
 * Check if a date is today (Vietnam timezone)
 * @param {string|Date} dateStr - ISO date string or Date object
 * @returns {boolean}
 */
export function isTodayVN(dateStr) {
	if (!dateStr) return false;
	
	const date = new Date(dateStr);
	const dateString = date.toLocaleDateString('en-CA', { timeZone: VN_TIMEZONE });
	const todayString = getTodayVN();
	
	return dateString === todayString;
}
