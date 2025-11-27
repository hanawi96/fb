# 🔬 Research: Caption riêng cho từng ảnh

## Vấn đề
Đối thủ có thể đăng carousel với caption riêng cho từng ảnh, nhưng code hiện tại không làm được.

## Các phương án có thể

### ❌ Phương án 1: attached_media[] với description
```
POST /{page-id}/feed
{
    "message": "Main message",
    "attached_media": [
        {"media_fbid": "123", "description": "Caption 1"}
    ]
}

→ Error: Invalid keys "description"
→ KHÔNG HOẠT ĐỘNG
```

### ✅ Phương án 2: Page Albums API
```
Step 1: Tạo album
POST /{page-id}/albums
{
    "name": "Album name",
    "message": "Album description"
}
→ Returns: album_id

Step 2: Upload ảnh vào album với message riêng
POST /{album-id}/photos
{
    "message": "Caption for photo 1",
    "source": <image_data>,
    "published": true
}

POST /{album-id}/photos
{
    "message": "Caption for photo 2", 
    "source": <image_data>,
    "published": true
}

→ Tạo album với nhiều ảnh, mỗi ảnh có caption riêng
→ CÓ THỂ HOẠT ĐỘNG!
```

### ✅ Phương án 3: Upload photos với message, sau đó share
```
Step 1: Upload ảnh với published=true
POST /{page-id}/photos
{
    "message": "Caption 1",
    "published": true,
    "source": <image_data>
}
→ Returns: photo_id_1

Step 2: Share photos trong 1 post
POST /{page-id}/feed
{
    "message": "Main post message",
    "child_attachments": [
        {"link": "https://facebook.com/photo_id_1"},
        {"link": "https://facebook.com/photo_id_2"}
    ]
}

→ CÓ THỂ HOẠT ĐỘNG với child_attachments
```

### ✅ Phương án 4: Batch API
```
POST /
{
    "batch": [
        {
            "method": "POST",
            "relative_url": "{page-id}/photos",
            "body": "message=Caption 1&source=..."
        },
        {
            "method": "POST", 
            "relative_url": "{page-id}/photos",
            "body": "message=Caption 2&source=..."
        }
    ]
}
```

## 🎯 Phương án khả thi nhất: Albums API

Đây có thể là cách đối thủ làm!

### Implementation Plan

```go
// 1. Tạo album trước
func (c *Client) CreateAlbum(pageID, accessToken, name, message string) (string, error) {
    data := url.Values{}
    data.Set("name", name)
    data.Set("message", message)
    data.Set("access_token", accessToken)
    
    resp, err := c.httpClient.PostForm(
        fmt.Sprintf("%s/%s/albums", GraphAPIURL, pageID), 
        data,
    )
    // Parse albumID from response
}

// 2. Upload ảnh vào album với caption riêng
func (c *Client) UploadPhotoToAlbum(albumID, accessToken, caption string, imageData []byte) error {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    // Add image
    part, _ := writer.CreateFormFile("source", "image.jpg")
    part.Write(imageData)
    
    // Add caption
    writer.WriteField("message", caption)
    writer.WriteField("published", "true")
    writer.WriteField("access_token", accessToken)
    
    writer.Close()
    
    resp, err := c.httpClient.Post(
        fmt.Sprintf("%s/%s/photos", GraphAPIURL, albumID),
        writer.FormDataContentType(),
        body,
    )
}
```

### Ưu điểm
- ✅ Mỗi ảnh có caption riêng
- ✅ Tất cả ảnh trong 1 album (giống carousel)
- ✅ User có thể swipe xem từng ảnh với caption

### Nhược điểm
- ❌ Tạo album mới mỗi lần post (có thể spam)
- ❌ Phức tạp hơn (2 API calls thay vì 1)
- ❌ Album khác với post thông thường

## 🧪 Test Plan

1. Test tạo album qua API
2. Test upload ảnh vào album với message
3. Kiểm tra trên Facebook xem có hiển thị caption không
4. So sánh với cách đối thủ làm

## 📚 References

- [Facebook Albums API](https://developers.facebook.com/docs/graph-api/reference/album/)
- [Facebook Photos API](https://developers.facebook.com/docs/graph-api/reference/photo/)
- [Facebook Page Feed](https://developers.facebook.com/docs/graph-api/reference/page/feed/)

---

**Next step:** Implement Albums API và test thực tế
