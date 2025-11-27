# Cách Fix Lỗi Facebook App

## Vấn đề
App đang ở chế độ **Development** nên chỉ có người được thêm vào Roles mới dùng được.

## Giải pháp

### Cách 1: Thêm tài khoản vào Roles (Nhanh)

1. Vào Facebook Developer Console: https://developers.facebook.com/apps/4526355974247445
2. Vào menu bên trái: **Roles** > **Roles**
3. Thêm tài khoản Facebook của bạn vào một trong các vai trò:
   - **Admin** (quyền cao nhất)
   - **Developer** (có thể test app)
   - **Tester** (chỉ test)
4. Chấp nhận lời mời trong Facebook
5. Test lại flow kết nối

### Cách 2: Chuyển App sang Live Mode (Cần Review)

**LƯU Ý:** Để chuyển sang Live, bạn cần:
- Có Privacy Policy URL
- Submit app để Facebook review các permissions
- Có thể mất vài ngày để được duyệt

**Các bước:**
1. Vào **App Settings** > **Basic**
2. Thêm **Privacy Policy URL** (bắt buộc)
3. Vào **App Review** > **Permissions and Features**
4. Request các permissions cần thiết:
   - `pages_show_list`
   - `pages_read_engagement`
   - `pages_manage_posts`
5. Submit để review
6. Sau khi được duyệt, chuyển toggle sang **Live**

## Khuyến nghị

**Dùng Cách 1** để test ngay lập tức. Sau khi app hoạt động ổn định, mới submit để chuyển sang Live mode.

## Kiểm tra sau khi fix

1. Restart backend server
2. Clear browser cache
3. Test lại flow kết nối Facebook
4. Kiểm tra backend logs để xem response từ Facebook
