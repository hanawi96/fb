import express from 'express';
import cors from 'cors';

const app = express();
const PORT = 8080;

app.use(cors());
app.use(express.json());

// Mock data
let pages = [
  {
    id: '1',
    page_id: '123456789',
    page_name: 'Demo Page 1',
    category: 'Business',
    profile_picture_url: 'https://via.placeholder.com/100',
    is_active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  },
  {
    id: '2',
    page_id: '987654321',
    page_name: 'Demo Page 2',
    category: 'Community',
    profile_picture_url: 'https://via.placeholder.com/100',
    is_active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  }
];

let posts = [
  {
    id: '1',
    content: 'Đây là bài viết demo đầu tiên',
    media_urls: ['https://via.placeholder.com/400'],
    media_type: 'photo',
    status: 'draft',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  }
];

let scheduledPosts = [
  {
    id: '1',
    post_id: '1',
    page_id: '1',
    scheduled_time: new Date(Date.now() + 3600000).toISOString(),
    status: 'pending',
    retry_count: 0,
    max_retries: 3,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    post: posts[0],
    page: pages[0]
  }
];

let logs = [
  {
    id: '1',
    post_id: '1',
    page_id: '1',
    facebook_post_id: '123456789_111111111',
    status: 'success',
    error_message: '',
    posted_at: new Date().toISOString(),
    post: posts[0],
    page: pages[0]
  }
];

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

// Auth
app.get('/api/auth/facebook/url', (req, res) => {
  res.json({ url: 'https://facebook.com/oauth/demo' });
});

app.post('/api/auth/facebook/callback', (req, res) => {
  res.json({
    message: 'Successfully connected pages',
    pages: pages
  });
});

// Pages
app.get('/api/pages', (req, res) => {
  res.json(pages);
});

app.delete('/api/pages/:id', (req, res) => {
  pages = pages.filter(p => p.id !== req.params.id);
  res.json({ message: 'Page deleted' });
});

app.patch('/api/pages/:id/toggle', (req, res) => {
  const page = pages.find(p => p.id === req.params.id);
  if (page) page.is_active = !page.is_active;
  res.json({ message: 'Page toggled' });
});

// Posts
app.post('/api/posts', (req, res) => {
  const newPost = {
    id: String(posts.length + 1),
    ...req.body,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  };
  posts.push(newPost);
  res.status(201).json(newPost);
});

app.get('/api/posts', (req, res) => {
  res.json(posts);
});

app.get('/api/posts/:id', (req, res) => {
  const post = posts.find(p => p.id === req.params.id);
  res.json(post || null);
});

app.put('/api/posts/:id', (req, res) => {
  const index = posts.findIndex(p => p.id === req.params.id);
  if (index !== -1) {
    posts[index] = { ...posts[index], ...req.body, updated_at: new Date().toISOString() };
    res.json(posts[index]);
  } else {
    res.status(404).json({ error: 'Post not found' });
  }
});

app.delete('/api/posts/:id', (req, res) => {
  posts = posts.filter(p => p.id !== req.params.id);
  res.json({ message: 'Post deleted' });
});

// Schedule
app.post('/api/schedule', (req, res) => {
  const { post_id, page_ids, scheduled_time } = req.body;
  const post = posts.find(p => p.id === post_id);
  
  const scheduled = page_ids.map((page_id, index) => {
    const page = pages.find(p => p.id === page_id);
    return {
      id: String(scheduledPosts.length + index + 1),
      post_id,
      page_id,
      scheduled_time,
      status: 'pending',
      retry_count: 0,
      max_retries: 3,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      post,
      page
    };
  });
  
  scheduledPosts.push(...scheduled);
  res.status(201).json({ message: 'Post scheduled successfully', scheduled });
});

app.get('/api/schedule', (req, res) => {
  const { status } = req.query;
  let filtered = scheduledPosts;
  if (status) {
    filtered = scheduledPosts.filter(s => s.status === status);
  }
  res.json(filtered);
});

app.delete('/api/schedule/:id', (req, res) => {
  scheduledPosts = scheduledPosts.filter(s => s.id !== req.params.id);
  res.json({ message: 'Scheduled post deleted' });
});

app.post('/api/schedule/:id/retry', (req, res) => {
  const scheduled = scheduledPosts.find(s => s.id === req.params.id);
  if (scheduled) {
    scheduled.status = 'pending';
    scheduled.retry_count = 0;
  }
  res.json({ message: 'Post queued for retry' });
});

// Logs
app.get('/api/logs', (req, res) => {
  res.json(logs);
});

// Upload
app.post('/api/upload', (req, res) => {
  res.json({ url: 'https://via.placeholder.com/600x400' });
});

app.listen(PORT, () => {
  console.log(`🚀 Mock Backend running on http://localhost:${PORT}`);
  console.log(`✅ Health check: http://localhost:${PORT}/health`);
});
