# API Documentation

Base URL: `http://localhost:8080/api`

## Authentication

### Get Facebook Auth URL
```
GET /auth/facebook/url
```

Response:
```json
{
  "url": "https://www.facebook.com/v18.0/dialog/oauth?..."
}
```

### Facebook Callback
```
POST /auth/facebook/callback
```

Body:
```json
{
  "code": "facebook_auth_code"
}
```

Response:
```json
{
  "message": "Successfully connected pages",
  "pages": [...]
}
```

## Pages

### Get All Pages
```
GET /pages
```

Response:
```json
[
  {
    "id": "uuid",
    "page_id": "123456789",
    "page_name": "My Page",
    "category": "Business",
    "profile_picture_url": "https://...",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Delete Page
```
DELETE /pages/:id
```

### Toggle Page Active Status
```
PATCH /pages/:id/toggle
```

## Posts

### Create Post
```
POST /posts
```

Body:
```json
{
  "content": "Post content here",
  "media_urls": ["https://...", "https://..."],
  "media_type": "photo",
  "status": "draft"
}
```

### Get Posts
```
GET /posts?limit=20&offset=0
```

### Get Single Post
```
GET /posts/:id
```

### Update Post
```
PUT /posts/:id
```

### Delete Post
```
DELETE /posts/:id
```

## Schedule

### Schedule Post
```
POST /schedule
```

Body:
```json
{
  "post_id": "uuid",
  "page_ids": ["uuid1", "uuid2"],
  "scheduled_time": "2024-12-31T23:59:59Z"
}
```

### Get Scheduled Posts
```
GET /schedule?status=pending&limit=50&offset=0
```

Status values: `pending`, `processing`, `completed`, `failed`

### Delete Scheduled Post
```
DELETE /schedule/:id
```

### Retry Failed Post
```
POST /schedule/:id/retry
```

## Logs

### Get Post Logs
```
GET /logs?limit=50&offset=0
```

Response:
```json
[
  {
    "id": "uuid",
    "post_id": "uuid",
    "page_id": "uuid",
    "facebook_post_id": "123456789_987654321",
    "status": "success",
    "error_message": "",
    "posted_at": "2024-01-01T00:00:00Z",
    "post": {...},
    "page": {...}
  }
]
```

## Upload

### Upload Image
```
POST /upload
```

Content-Type: `multipart/form-data`

Body:
- `image`: File (JPEG/PNG, max 10MB)

Response:
```json
{
  "url": "http://localhost:8080/uploads/filename.jpg"
}
```
