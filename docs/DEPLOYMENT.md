# Hướng dẫn Deploy Production

## Option 1: Deploy lên Railway (Recommended - Dễ nhất)

### Backend + Database

1. **Tạo tài khoản Railway**: https://railway.app/

2. **Deploy PostgreSQL**:
   - New Project > Add PostgreSQL
   - Copy DATABASE_URL từ Variables

3. **Deploy Backend**:
   - New Service > GitHub Repo
   - Chọn repo của bạn
   - Root Directory: `/backend`
   - Build Command: `go build -o main cmd/server/main.go`
   - Start Command: `./main`
   
4. **Set Environment Variables**:
   ```
   DATABASE_URL=<from railway postgres>
   FACEBOOK_APP_ID=<your app id>
   FACEBOOK_APP_SECRET=<your app secret>
   FACEBOOK_REDIRECT_URI=https://your-frontend.vercel.app/auth/callback
   PORT=8080
   FRONTEND_URL=https://your-frontend.vercel.app
   ```

5. **Generate Domain**: Railway sẽ tự tạo domain cho backend

### Frontend

1. **Deploy lên Vercel**: https://vercel.com/

2. **Import GitHub Repo**:
   - Root Directory: `/frontend`
   - Framework: SvelteKit
   
3. **Set Environment Variable**:
   ```
   PUBLIC_API_URL=https://your-backend.railway.app
   ```

4. **Deploy**: Vercel tự động build và deploy

### Update Facebook App Settings

1. Vào Facebook App Settings
2. Update OAuth Redirect URIs:
   - `https://your-frontend.vercel.app/auth/callback`
3. Update App Domains:
   - `your-frontend.vercel.app`

---

## Option 2: Deploy lên VPS (Ubuntu)

### 1. Setup Server

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# Install Golang
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt install -y nodejs

# Install Nginx
sudo apt install nginx -y
```

### 2. Setup Database

```bash
sudo -u postgres psql
CREATE DATABASE fbscheduler;
CREATE USER fbuser WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE fbscheduler TO fbuser;
\q
```

### 3. Deploy Backend

```bash
# Clone repo
git clone <your-repo>
cd backend

# Create .env
nano .env
# Paste environment variables

# Run migrations
psql -U fbuser -d fbscheduler -f migrations/001_init.sql

# Build
go build -o server cmd/server/main.go

# Create systemd service
sudo nano /etc/systemd/system/fbscheduler.service
```

Content của service file:
```ini
[Unit]
Description=FB Scheduler Backend
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/backend
ExecStart=/home/ubuntu/backend/server
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
# Start service
sudo systemctl enable fbscheduler
sudo systemctl start fbscheduler
```

### 4. Deploy Frontend

```bash
cd frontend
npm install
npm run build

# Copy build to nginx
sudo cp -r build/* /var/www/html/
```

### 5. Configure Nginx

```bash
sudo nano /etc/nginx/sites-available/default
```

Content:
```nginx
server {
    listen 80;
    server_name your-domain.com;

    # Frontend
    location / {
        root /var/www/html;
        try_files $uri $uri/ /index.html;
    }

    # Backend API
    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

```bash
sudo nginx -t
sudo systemctl restart nginx
```

### 6. Setup SSL (Let's Encrypt)

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d your-domain.com
```

---

## Option 3: Deploy lên Fly.io

### Backend

```bash
cd backend

# Install flyctl
curl -L https://fly.io/install.sh | sh

# Login
flyctl auth login

# Launch app
flyctl launch

# Set secrets
flyctl secrets set DATABASE_URL=<your-db-url>
flyctl secrets set FACEBOOK_APP_ID=<your-app-id>
flyctl secrets set FACEBOOK_APP_SECRET=<your-secret>

# Deploy
flyctl deploy
```

### Frontend

Deploy lên Vercel như Option 1.

---

## Checklist sau khi Deploy

- [ ] Backend health check: `curl https://your-backend/health`
- [ ] Frontend accessible
- [ ] Database migrations chạy thành công
- [ ] Facebook OAuth redirect URIs đã update
- [ ] Environment variables đã set đúng
- [ ] SSL certificate đã cài (nếu dùng VPS)
- [ ] Test flow: Login > Connect Pages > Create Post > Schedule
- [ ] Scheduler đang chạy (check logs)
- [ ] Upload ảnh hoạt động

## Monitoring

### Backend Logs
- Railway: Xem trong dashboard
- VPS: `sudo journalctl -u fbscheduler -f`
- Fly.io: `flyctl logs`

### Database
- Railway: Xem metrics trong dashboard
- VPS: `sudo -u postgres psql -d fbscheduler`

## Backup Database

```bash
# Backup
pg_dump -U fbuser fbscheduler > backup.sql

# Restore
psql -U fbuser fbscheduler < backup.sql
```
