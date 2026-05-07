# 部署到 Linux 服务器

## 方式一：Docker Compose 部署（推荐）

### 1. 上传代码到服务器

```bash
# 方法 A: git clone
ssh your-server
git clone <your-repo-url> /opt/app-management-platform
cd /opt/app-management-platform

# 方法 B: rsync 上传
rsync -avz --exclude node_modules --exclude .git ./ your-server:/opt/app-management-platform/
```

### 2. 配置环境变量

```bash
cp .env.example .env
vim .env
# 修改:
#   JWT_SECRET=your-random-secret-key
#   ADMIN_PASSWORD=your-strong-password
```

### 3. 按需编辑服务配置

```bash
cp app-manager/config.example.yaml app-manager/config.yaml
vim app-manager/config.yaml
# 添加你需要管理的 systemd 服务
```

### 4. 一键部署

```bash
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

或手动执行：

```bash
docker compose up -d --build
```

### 5. 更新部署

```bash
./scripts/update.sh
```

### 6. 停止服务

```bash
./scripts/stop.sh
```

---

## 方式二：直接二进制部署（无 Docker）

### 1. 本地交叉编译

```bash
# 编译后端 (Linux amd64)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o app-manager ./app-manager/main.go

# 编译前端
cd manager-frontend-app
npm ci && npm run build
```

### 2. 上传到服务器

```bash
# 上传二进制
scp app-manager your-server:/usr/local/bin/

# 上传配置
mkdir -p /etc/app-manager
scp app-manager/config.example.yaml your-server:/etc/app-manager/config.yaml

# 上传前端静态文件
rsync -avz manager-frontend-app/dist/ your-server:/var/www/app-manager/

# 上传 systemd 服务文件
scp scripts/app-manager.service your-server:/etc/systemd/system/
```

### 3. 服务器端配置 Nginx

将 `manager-frontend-app/nginx.conf` 内容放入 `/etc/nginx/sites-available/app-manager`，
然后 `ln -s /etc/nginx/sites-available/app-manager /etc/nginx/sites-enabled/`。

修改 `proxy_pass` 为 `http://127.0.0.1:8080`。

```bash
nginx -t && systemctl reload nginx
```

### 4. 启动后端服务

```bash
ssh your-server
systemctl daemon-reload
systemctl enable app-manager
systemctl start app-manager
systemctl status app-manager
```

### 5. 一行命令部署（使用 Makefile）

```bash
make deploy DEPLOY_HOST=user@your-server
```

---

## 防火墙

```bash
# 仅开放必要端口
ufw allow 80/tcp
ufw allow 443/tcp
# 如果需要直接访问后端（调试用）
ufw allow 8080/tcp
```

## HTTPS（生产环境推荐）

```bash
apt install certbot python3-certbot-nginx
certbot --nginx -d your-domain.com
```

## Docker 权限说明

后端容器挂载了 `/var/run/docker.sock`，以便管理宿主机上的容器。
这是设计书中的要求。如果不需要管理 Docker，可以在 config.yaml 中设置：

```yaml
docker:
  enabled: false
```

## systemctl 权限说明

进程管理通过 `systemctl` 命令实现，后端进程需要 root 权限，或者通过 sudoers 白名单：

```bash
visudo
# 添加:
# app-manager ALL=(root) NOPASSWD: /bin/systemctl start order-service, /bin/systemctl stop order-service, /bin/systemctl restart order-service
# app-manager ALL=(root) NOPASSWD: /usr/bin/journalctl -u order-service *
```
