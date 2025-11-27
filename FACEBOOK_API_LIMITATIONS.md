# 🚫 Facebook Graph API Limitations

## Những gì Facebook KHÔNG hỗ trợ

### 1. Caption riêng cho từng ảnh trong Album/Carousel

**Vấn đề:**
```
❌ Không thể có caption khác nhau cho mỗi ảnh trong 1 post
```

**Lý do:**
- Facebook Graph API endpoint `/feed` với `attached_media[]` chỉ chấp nhận field `media_fbid`
- Field `description` bị reject với error: `Invalid keys "description" were found in param "attached_media[0]"`

**Workaround:**
```
✅ Dùng Individual mode: Mỗi ảnh = 1 post riêng với caption riêng
❌ Album mode: Chỉ có 1 message chung cho tất cả ảnh
```

**Code example:**
```go
// ❌ KHÔNG HOẠT ĐỘNG
attached_media[] = {
    "media_fbid": "123",
    "description": "Caption 1"  // ← Facebook reject!
}

// ✅ HOẠT ĐỘNG
attached_media[] = {
    "media_fbid": "123"  // Chỉ có media_fbid
}
```

---

### 2. Caption cho unpublished photos

**Vấn đề:**
```
❌ Caption bị bỏ qua khi upload ảnh với published=false
```

**Lý do:**
- Khi upload ảnh để attach vào post sau, phải set `published=false`
- Facebook không lưu caption cho unpublished photos
- Caption chỉ work khi `published=true` (nhưng không thể attach vào post sau)

**Code example:**
```go
// Upload ảnh để attach vào post
POST /photos
{
    "source": <image_data>,
    "published": "false",
    "caption": "My caption"  // ← Bị bỏ qua!
}
```

---

### 3. Edit caption sau khi đăng

**Vấn đề:**
```
❌ Không thể edit caption của ảnh trong carousel
```

**Lý do:**
- Facebook chỉ cho phép edit message của post
- Không có API để edit caption riêng của từng ảnh

---

## ✅ Những gì Facebook HỖ TRỢ

### 1. Individual posts với caption riêng
```
POST /photos (published=true)
{
    "message": "Caption for this image",
    "source": <image_data>
}

→ Tạo 1 post với 1 ảnh và caption
```

### 2. Album với message chung
```
POST /feed
{
    "message": "Common message for all images",
    "attached_media": [
        {"media_fbid": "123"},
        {"media_fbid": "456"}
    ]
}

→ Tạo 1 post với nhiều ảnh, 1 message chung
```

### 3. Video với description
```
POST /videos
{
    "description": "Video description",
    "source": <video_data>
}

→ Video hỗ trợ description field
```

---

## 🎯 Recommendations

### Khi nào dùng Album mode?
- ✅ Ảnh cùng chủ đề, không cần caption riêng
- ✅ Muốn user swipe xem nhiều ảnh trong 1 post
- ✅ Tiết kiệm space trên timeline

### Khi nào dùng Individual mode?
- ✅ Mỗi ảnh cần caption/mô tả riêng
- ✅ Ảnh khác chủ đề hoàn toàn
- ✅ Muốn mỗi ảnh có engagement riêng (likes, comments)
- ✅ Product showcase với giá/mô tả khác nhau

---

## 📚 References

- [Facebook Graph API - Page Posts](https://developers.facebook.com/docs/graph-api/reference/page/feed/)
- [Facebook Graph API - Photos](https://developers.facebook.com/docs/graph-api/reference/photo/)
- [Facebook Graph API - Videos](https://developers.facebook.com/docs/graph-api/reference/video/)

---

## 💡 Lessons Learned

1. **Always check API docs** - Không phải tất cả features đều được hỗ trợ
2. **Test with real API** - Mock data có thể không phản ánh đúng limitations
3. **Provide alternatives** - Nếu feature không được hỗ trợ, đưa ra workaround
4. **Clear UI warnings** - Cảnh báo user về limitations trước khi họ sử dụng

---

**Last updated:** 2025-11-27
