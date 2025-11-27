// Text formatting utilities - siêu nhẹ, không dependencies

// Insert text vào vị trí cursor
export function insertAtCursor(textarea, text) {
	const start = textarea.selectionStart;
	const end = textarea.selectionEnd;
	const value = textarea.value;
	
	// Insert text
	const newValue = value.substring(0, start) + text + value.substring(end);
	textarea.value = newValue;
	
	// Set cursor position sau text vừa insert
	const newPos = start + text.length;
	textarea.setSelectionRange(newPos, newPos);
	
	// Trigger input event để Svelte update
	textarea.dispatchEvent(new Event('input', { bubbles: true }));
	textarea.focus();
}

// Wrap selected text với prefix/suffix
export function wrapSelection(textarea, prefix, suffix = prefix) {
	const start = textarea.selectionStart;
	const end = textarea.selectionEnd;
	const value = textarea.value;
	const selectedText = value.substring(start, end);
	
	if (!selectedText) {
		// Không có text được chọn, insert prefix+suffix và đặt cursor ở giữa
		insertAtCursor(textarea, prefix + suffix);
		const newPos = start + prefix.length;
		textarea.setSelectionRange(newPos, newPos);
		return;
	}
	
	// Wrap selected text
	const newValue = value.substring(0, start) + prefix + selectedText + suffix + value.substring(end);
	textarea.value = newValue;
	
	// Select wrapped text
	textarea.setSelectionRange(start, end + prefix.length + suffix.length);
	textarea.dispatchEvent(new Event('input', { bubbles: true }));
	textarea.focus();
}

// Format actions
export const formatActions = {
	bold: (textarea) => wrapSelection(textarea, '**'),
	italic: (textarea) => wrapSelection(textarea, '*'),
	strikethrough: (textarea) => wrapSelection(textarea, '~~'),
	code: (textarea) => wrapSelection(textarea, '`'),
	
	// List formatting
	bulletList: (textarea) => {
		const start = textarea.selectionStart;
		const value = textarea.value;
		const lineStart = value.lastIndexOf('\n', start - 1) + 1;
		insertAtCursor(textarea, '\n• ');
	},
	
	numberedList: (textarea) => {
		const start = textarea.selectionStart;
		const value = textarea.value;
		const lineStart = value.lastIndexOf('\n', start - 1) + 1;
		insertAtCursor(textarea, '\n1. ');
	},
	
	// Hashtag
	hashtag: (textarea) => {
		const start = textarea.selectionStart;
		const value = textarea.value;
		const charBefore = value[start - 1];
		
		// Thêm space nếu cần
		const prefix = (!charBefore || charBefore === '\n') ? '#' : ' #';
		insertAtCursor(textarea, prefix);
	}
};

// Keyboard shortcuts
export function handleKeyboardShortcut(event, textarea) {
	if (!event.ctrlKey && !event.metaKey) return false;
	
	switch (event.key.toLowerCase()) {
		case 'b':
			event.preventDefault();
			formatActions.bold(textarea);
			return true;
		case 'i':
			event.preventDefault();
			formatActions.italic(textarea);
			return true;
		case 'u':
			event.preventDefault();
			formatActions.strikethrough(textarea);
			return true;
		default:
			return false;
	}
}

// Popular hashtags cho suggestions
export const popularHashtags = [
	'#marketing', '#business', '#sale', '#khuyenmai', '#giamgia',
	'#vietnam', '#hanoi', '#saigon', '#danang',
	'#food', '#travel', '#fashion', '#beauty', '#tech',
	'#motivation', '#success', '#entrepreneur', '#startup',
	'#love', '#happy', '#life', '#instagood', '#photooftheday'
];

// Debounce helper
export function debounce(func, wait) {
	let timeout;
	return function executedFunction(...args) {
		const later = () => {
			clearTimeout(timeout);
			func(...args);
		};
		clearTimeout(timeout);
		timeout = setTimeout(later, wait);
	};
}
