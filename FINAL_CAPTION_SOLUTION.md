# 🎯 Giải pháp cuối cùng: Caption cho ảnh

## ❌ Những gì KHÔNG hoạt động

### 1. Feed API với description
```
attached_media[] = {"media_fbid": "123", "description": "caption"}
→ Error: Invalid keys "description"
```

### 2. Albums API
```
POST /albums → Create album
POST /album/photos → Upload with captions
→ Error: (#3) Application does not have the capability
→ Cần quyền "pages_manage_albums" (phức tạp)
```

### 3. Unpublished photos với caption
```
POST /photos (published=false, caption="...")
→ Caption bị bỏ qua
```

---

## ✅ Giải pháp HOẠT ĐỘNG

### Individual Mode (Mỗi ảnh 1 post)

**Đây là CÁCH DUY NHẤT để có caption riêng!**

```
Ảnh 1 → POST /photos (published=true, message="Caption 1")
Ảnh 2 → POST /photos (published=true, message="Caption 2")
Ảnh 3 → POST /photos (published=true, message="Caption 3")

→ 3 posts riêng, mỗi post 1 ảnh với caption riêng
```

**Ưu điểm:**
- ✅ Caption riêng cho từng ảnh
- ✅ Mỗi post có engagement riêng
- ✅ Không cần quyền đặc biệt

**Nhược điểm:**
- ❌ Nhiều posts trên timeline
- ❌ Không có carousel

---

## 🎯 Recommendation cho User

### Khi nào dùng Album mode?
```
✅ Ảnh cùng chủ đề
✅ Không cần caption riêng
✅ Muốn carousel đẹp
✅ Tiết kiệm space trên timeline
```

### Khi nào dùng Individual mode?
```
✅ Cần caption riêng cho từng ảnh
✅ Ảnh khác chủ đề
✅ Product showcase
✅ Storytelling (mỗi ảnh 1 câu chuyện)
```

---

## 💡 Tại sao đối thủ làm được?

Có thể họ:

1. **Có quyền Albums** - App của họ đã được approve quyền `pages_manage_albums`
2. **Dùng Individual mode** - Giống như solution của chúng ta
3. **Manual post** - Không dùng API, post thủ công trên Facebook
4. **Dùng Facebook Creator Studio** - Tool chính thức của Facebook

---

## 🚀 Implementation hiện tại

### Album Mode
```go
// Upload ảnh unpublished
mediaIDs := uploadPhotos(images)

// Tạo post với carousel
POST /feed {
    "message": "Main message",
    "attached_media": [
        {"media_fbid": "123"},
        {"media_fbid": "456"}
    ]
}

→ 1 post với carousel, 1 message chung
→ Caption riêng KHÔNG được hỗ trợ
```

### Individual Mode
```go
// Upload từng ảnh published với caption
for each image {
    POST /photos {
        "message": image.caption,
        "published": true,
        "source": image.data
    }
}

→ N posts riêng, mỗi post có caption riêng
→ Caption riêng HOẠT ĐỘNG ✅
```

---

## 📊 Performance

### Album Mode (2 ảnh)
```
Upload 2 photos (concurrent): 2s
Create post: 0.5s
Total: ~2.5s
```

### Individual Mode (2 ảnh)
```
Post 2 photos (concurrent): 2s
Total: ~2s
```

→ Individual mode thậm chí nhanh hơn!

---

## ✅ Kết luận

**Caption riêng cho từng ảnh CHỈ hoạt động ở Individual mode.**

Đây là **GIỚI HẠN CỦA FACEBOOK GRAPH API**, không phải bug của app.

Để có caption riêng trong carousel, cần:
- Quyền `pages_manage_albums` (phức tạp, cần Facebook review)
- Hoặc dùng Facebook Creator Studio (không phải API)

**Recommendation:** Dùng Individual mode khi cần caption riêng.

---

**Last updated:** 2025-11-27
