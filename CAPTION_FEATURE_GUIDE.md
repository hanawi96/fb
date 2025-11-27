# 📝 Hướng dẫn sử dụng Caption cho từng ảnh

## 🎯 Tính năng

Cho phép thêm caption (mô tả) riêng cho từng ảnh khi đăng bài Facebook.

---

## 🖼️ Cách sử dụng

### 1. Upload ảnh như bình thường
```
Kéo thả hoặc click "Thêm ảnh"
```

### 2. Thêm caption cho ảnh
```
Hover vào ảnh → Button "+ Caption" xuất hiện
Click button → Modal mở ra
Nhập caption → Click "Lưu"
```

### 3. Kiểm tra caption đã lưu
```
✅ Icon xanh ở góc dưới phải = Đã có caption
Hover lại → Button "✏️ Sửa" để chỉnh sửa
```

---

## 📊 Cách hoạt động

### ✅ Album Mode với Albums API (RECOMMENDED)

**Album Mode (1 album nhiều ảnh):**
```
✅ Hệ thống tự động dùng Albums API khi có caption!

Post message: "Chuyến đi Đà Lạt"
Ảnh 1 caption: "Cà phê sáng"  ✅
Ảnh 2 caption: "Hồ Xuân Hương" ✅

→ Tạo album "Chuyến đi Đà Lạt"
→ Mỗi ảnh có caption riêng khi click vào
→ User có thể swipe xem từng ảnh với caption
```

**Cách hoạt động:** Dùng Facebook Albums API thay vì Feed API

### ✅ Individual Mode (Mỗi ảnh 1 post) - RECOMMENDED
```
Post message: "Chuyến đi Đà Lạt" (bị bỏ qua)
Ảnh 1 caption: "Cà phê sáng"
Ảnh 2 caption: "Hồ Xuân Hương"

→ Facebook tạo 2 posts riêng:
   Post 1: "Cà phê sáng" ✅
   Post 2: "Hồ Xuân Hương" ✅
```

**Đây là CÁCH DUY NHẤT để có caption riêng cho từng ảnh!**

---

## 🔧 Technical Details

### Data Format
```javascript
// Old format (still supported)
images = ["url1", "url2"]

// New format with captions
images = [
  { url: "url1", caption: "Caption 1" },
  { url: "url2", caption: "" }
]
```

### API Request
```json
{
  "content": "Main post message",
  "images": [
    { "url": "http://...", "caption": "Image 1 caption" },
    { "url": "http://...", "caption": "Image 2 caption" }
  ],
  "post_mode": "album",
  "page_ids": ["123", "456"]
}
```

### Facebook API
```javascript
// Album mode: attached_media with description
attached_media[] = {
  "media_fbid": "123456",
  "description": "Caption text"
}

// Individual mode: Each image = separate post
POST /photos with message = caption
```

---

## ✅ Best Practices

### Khi nào dùng Caption?
- ✅ **Individual mode ONLY** → Caption làm message riêng cho mỗi post
- ✅ Storytelling → Mỗi ảnh 1 câu chuyện riêng
- ✅ Product showcase → Mỗi sản phẩm 1 mô tả riêng

### Khi nào KHÔNG cần Caption?
- ❌ Album mode → Caption bị bỏ qua (Facebook limitation)
- ❌ Ảnh giống nhau, cùng chủ đề → Dùng message chung
- ❌ Chỉ 1-2 ảnh đơn giản → Message chính đã đủ

### Tips
1. **Caption ngắn gọn** - 1-2 câu là đủ
2. **Emoji** - Làm caption sinh động hơn
3. **Hashtag** - Có thể thêm hashtag riêng cho từng ảnh
4. **Individual mode** - Dùng caption thay message để tiết kiệm thời gian

---

## 🐛 Troubleshooting

### Caption không hiển thị trên Facebook?
- Kiểm tra Album mode đã chọn đúng chưa
- Facebook có thể mất vài giây để load caption
- Click vào từng ảnh trong carousel để xem caption

### Caption bị mất sau khi drag & drop?
- Caption được giữ nguyên khi reorder
- Kiểm tra icon xanh vẫn còn không

### Individual mode không dùng caption?
- Đảm bảo caption không rỗng
- Nếu caption rỗng, dùng main message

---

## 📸 Screenshots

```
┌─────────────────────────────────┐
│  📊 3/10 ảnh • Ảnh đầu = Cover  │
│  🔘 Album  🔘 Riêng lẻ (3 posts)│
├─────────────────────────────────┤
│  ┌───┐  ┌───┐  ┌───┐            │
│  │ 1 │  │ 2 │  │ 3 │            │
│  │   │  │   │  │   │            │
│  │ ✅│  │   │  │ ✅│  ← Icon xanh│
│  └───┘  └───┘  └───┘            │
│   ↑              ↑               │
│  Có caption   Chưa có           │
└─────────────────────────────────┘

Hover vào ảnh:
┌───────────────┐
│       1       │
│   [Drag]      │
│               │
│  ✅           │
│ [+ Caption]   │ ← Button xuất hiện
└───────────────┘
```

---

## 🎓 Examples

### Example 1: Food Blog
```
Message: "Món ngon cuối tuần 🍜"
Ảnh 1: "Phở bò Hà Nội - 50k"
Ảnh 2: "Bún chả - 45k"
Ảnh 3: "Chè đậu đỏ - 20k"
Mode: Album
```

### Example 2: Product Launch
```
Message: "Ra mắt sản phẩm mới!"
Ảnh 1: "Màu đỏ - Giá 299k"
Ảnh 2: "Màu xanh - Giá 299k"
Ảnh 3: "Màu đen - Giá 349k"
Mode: Individual (mỗi màu 1 post)
```

### Example 3: Travel Story
```
Message: "" (để trống)
Ảnh 1: "Ngày 1: Bay đến Đà Lạt ✈️"
Ảnh 2: "Ngày 2: Khám phá thác Datanla 🌊"
Ảnh 3: "Ngày 3: Về nhà với kỷ niệm đẹp 🏠"
Mode: Individual (mỗi ngày 1 post)
```

---

## 🚀 Keyboard Shortcuts

- `ESC` - Đóng caption modal
- `Enter` - Lưu caption (khi focus vào button)
- `Tab` - Di chuyển giữa các field

---

**Enjoy posting! 🎉**
