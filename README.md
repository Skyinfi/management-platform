# 应用管理平台（App Management Platform）

单仓库包含 **Go 后端**（`app-manager`）、**宿主机扫描 Agent**（`discovered-app`）与 **React + Vite 前端**（`manager-frontend-app`），用于在服务器上统一管理 Docker、进程类应用以及宿主机野生进程。详细架构见仓库内 [`app-management-platform.md`](./app-management-platform.md)。

## 仓库结构

| 路径 | 说明 |
|------|------|
| `app-manager/` | Go API 服务（默认 `:8080`） |
| `discovered-app/` | 宿主机扫描 Agent（默认 `:8081`），发现未被 systemd/docker 管理的野生进程 |
| `manager-frontend-app/` | 管理端 Web UI（Vite 开发服务器默认 `:5173`） |
| `scripts/` | 本地开发辅助脚本（PowerShell）与 systemd 服务文件 |
| `package.json` | 根目录脚本入口（一键起停开发环境） |

## 环境要求

- **Go**：1.22 或更高（见 `app-manager/go.mod`）
- **Node.js**：建议使用当前 LTS，并自带 **npm**

## 本地开发

### Windows（推荐）

在仓库根目录执行（需先安装前端依赖）：

```powershell
cd manager-frontend-app; npm install; cd ..
npm run dev
```

- 后端：<http://localhost:8080>
- 前端：<http://localhost:5173>（`/api` 由 Vite 代理到后端）

停止根目录脚本后，若后台 Job 仍在运行，可执行：

```powershell
npm run stop
```

### 手动分别启动（任意系统）

**后端**

```bash
cd app-manager
go run ./cmd/app-manager
```

**前端**

```bash
cd manager-frontend-app
npm install
npm run dev
```

## 后端环境变量（可选）

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_MANAGER_ADDR` | 监听地址 | `:8080` |
| `APP_MANAGER_ENV` | 运行环境标识 | `development` |
| `APP_MANAGER_LOG_LEVEL` | 日志级别 | `info` |
| `APP_MANAGER_ENABLE_CORS` | 是否启用 CORS | `true` |
| `APP_MANAGER_JWT_SECRET` | JWT 密钥 | `dev-secret`（**生产务必修改**） |
| `APP_MANAGER_ALLOW_ORIGIN` | CORS 允许的 Origin | `*` |
| `APP_MANAGER_SCANNER_ADDR` | 扫描 Agent 地址 | `http://localhost:8081` |

## 扫描 Agent（discovered-app）

宿主机扫描 Agent 独立运行，通过读取 `/proc` 文件系统发现未被 systemd/docker 管理的野生进程（如 nohup、screen、手动启动的 Java JAR / Python / Node.js 等），`app-manager` 通过 HTTP 调用其 API 获取扫描结果。

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SCANNER_ADDR` | 监听地址 | `:8081` |
| `PROC_ROOT` | proc 文件系统路径 | `/proc`（容器部署时改为 `/host/proc`） |
| `SCANNER_INTERVAL` | 定时扫描间隔 | `60s` |

### 快速启动（本地开发）

```bash
cd discovered-app
go run ./main.go
```

### 编译

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o discovered-app/discovered-app ./discovered-app/main.go
```

### 部署

详细安装方式（systemd / Docker / docker-compose）见 [`discovered-app/INSTALL.md`](discovered-app/INSTALL.md)。

一键编译部署到目标服务器：

```bash
make deploy-scanner DEPLOY_HOST=user@your-server
```

## 构建

**前端生产构建**

```bash
cd manager-frontend-app
npm run build
```

产物在 `manager-frontend-app/dist/`。部署时由静态服务器或反向代理提供，并将 API 请求转发到 Go 服务。

**后端二进制**（示例）

```bash
cd app-manager
go build -o app-manager ./cmd/app-manager
```

## Docker 部署

### 环境要求

- Docker Engine 20.10+
- Docker Compose v2（`docker compose` 命令可用）
- 服务器已安装 Docker（需挂载 `/var/run/docker.sock`）

### 1. 上传代码

```bash
git clone <repo-url> /opt/app-management-platform
cd /opt/app-management-platform
```

### 2. 配置

```bash
# 环境变量
cp .env.example .env
vim .env
```

`.env` 内容：

```
JWT_SECRET=your-random-secret-key
ADMIN_PASSWORD=your-strong-password
```

```bash
# 后端服务配置
cp app-manager/config.example.yaml app-manager/config.yaml
vim app-manager/config.yaml
```

主要修改项：

| 字段 | 说明 |
|------|------|
| `jwt_secret` | JWT 签名密钥，**必须修改**为随机长字符串 |
| `auth.username` | 登录用户名 |
| `auth.password` | 登录密码，**必须修改** |
| `auth.token_ttl` | Token 有效期，如 `8h`、`24h` |
| `docker.enabled` | 是否启用 Docker 容器管理 |
| `services` | 要管理的 systemd 服务列表 |

`services` 中每个服务：

- `unit`：systemd unit 名称（`systemctl status xxx` 中的名称）
- `log_path`：可选，日志文件路径（不填则用 `journalctl` 读取）
- `endpoint`：仅展示用，标记服务监听地址

### 3. 启动

**使用 Nginx（默认）**

```bash
docker compose up -d --build
```

**使用 Caddy（自动 HTTPS）**

```bash
FRONTEND_DOCKERFILE=Dockerfile.caddy docker compose up -d --build
```

或在 `.env` 中设置：

```
FRONTEND_DOCKERFILE=Dockerfile.caddy
```

### 4. 验证

```bash
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f backend
docker compose logs -f frontend

# 健康检查
curl http://localhost:8080/api/health
```

访问 `http://服务器IP` 打开管理面板，用 `.env` 中配置的用户名密码登录。

### 5. 更新部署

```bash
git pull origin main
docker compose up -d --build
docker image prune -f
```

或使用脚本：

```bash
chmod +x scripts/update.sh
./scripts/update.sh
```

### 6. 停止

```bash
docker compose down
```

### Caddy 启用 HTTPS

编辑 `manager-frontend-app/Caddyfile`，将 `:80` 改为域名：

```
your-domain.com {
    root * /srv
    try_files {path} /index.html
    file_server

    reverse_proxy /api/* backend:8080
    reverse_proxy /ws/* backend:8080 {
        header_up Connection {>Connection}
        header_up Upgrade {>Upgrade}
    }
}
```

Caddy 会自动申请并续期 Let's Encrypt 证书，确保服务器 80、443 端口对外开放。

### 端口说明

| 端口 | 服务 | 说明 |
|------|------|------|
| 80 | frontend | Nginx/Caddy 前端 |
| 443 | frontend | Caddy HTTPS（仅 Caddy） |
| 8080 | backend | Go API 服务 |
| 8081 | scanner | 宿主机扫描 Agent（仅内网/localhost） |

### 注意事项

- 后端容器挂载 `/var/run/docker.sock` 用于管理宿主机 Docker 容器
- 进程管理（systemctl）需要后端以 root 权限运行，或配置 sudoers 白名单
- 生产环境建议修改 `enable_cors: false`，由 Nginx/Caddy 统一处理跨域

---

## 发布到 GitHub

1. 在 GitHub 新建空仓库（不要勾选自动添加 README，避免首次推送冲突）。
2. 在仓库根目录初始化并推送：

```bash
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/<你的用户名>/<仓库名>.git
git push -u origin main
```

首次推送前请确认本机未把 `node_modules`、密钥文件等加入版本库；本仓库根目录 `.gitignore` 已覆盖常见忽略项。
