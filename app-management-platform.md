# 应用管理平台架构设计文档

## 概述

本文档描述一个部署在 Ubuntu 服务器上的统一应用管理平台，用于同时管理 Docker 容器化应用与静态部署（进程型）应用。

- **后端语言**：Go
- **前端框架**：React + Ant Design
- **目标环境**：Ubuntu Server（x86_64）
- **管理对象**：Docker 容器 + systemd/supervisord 管理的进程

---

## 系统架构

```
┌─────────────────────────────────────────────────────┐
│                浏览器 / Web UI                       │
│         React + Ant Design 管理面板                  │
│    实时日志 · 状态监控 · 容器/进程操作面板            │
└───────────────────┬─────────────────────────────────┘
                    │ HTTP REST + WebSocket
┌───────────────────▼─────────────────────────────────┐
│              后端 API 服务（Go）                      │
│   Gin / Echo 框架 · JWT 认证 · 统一 REST 接口         │
│   goroutine 并发处理 · WebSocket 日志流               │
└───────────┬─────────────────────┬───────────────────┘
            │                     │
┌───────────▼──────────┐ ┌───────▼───────────────────┐
│   Docker 管理模块     │ │   静态应用管理模块          │
│  docker SDK (Go)     │ │  os/exec · systemctl       │
│  启停容器 · 查看日志  │ │  supervisorctl · 日志读取  │
│  镜像管理 · 资源监控  │ │  端口监控 · 进程状态       │
└───────────┬──────────┘ └───────┬───────────────────┘
            │                     │
┌───────────▼─────────────────────▼───────────────────┐
│                  Ubuntu Server                       │
│     文件系统 · 网络 · systemd · Docker Engine         │
│                                                      │
│  [容器 A: Nginx]  [容器 B: Node App]                 │
│  [进程 C: Java JAR]  [进程 D: Python App]            │
└─────────────────────────────────────────────────────┘
```

---

## 目录结构

```
app-manager/
├── backend/                  # Go 后端
│   ├── main.go               # 程序入口，路由注册
│   ├── config/
│   │   └── config.go         # 配置加载（yaml/env）
│   ├── api/
│   │   ├── auth.go           # 登录、JWT 签发
│   │   ├── docker.go         # Docker 相关接口
│   │   └── process.go        # 静态进程相关接口
│   ├── service/
│   │   ├── docker_service.go # 封装 docker SDK 操作
│   │   └── process_service.go# 封装 systemctl/exec 操作
│   ├── middleware/
│   │   └── auth.go           # JWT 验证中间件
│   ├── ws/
│   │   └── log_stream.go     # WebSocket 日志流
│   └── Dockerfile            # 后端镜像构建
│
├── frontend/                 # React 前端
│   ├── src/
│   │   ├── App.tsx
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx # 总览页：所有应用状态
│   │   │   ├── DockerApps.tsx# Docker 容器管理页
│   │   │   └── ProcessApps.tsx# 静态进程管理页
│   │   ├── components/
│   │   │   ├── LogViewer.tsx # 实时日志组件（WebSocket）
│   │   │   ├── StatusBadge.tsx
│   │   │   └── AppCard.tsx
│   │   └── api/
│   │       └── client.ts     # Axios 封装，统一请求
│   ├── package.json
│   └── Dockerfile            # 前端 Nginx 镜像
│
├── docker-compose.yml        # 整体编排
└── nginx.conf                # 反向代理配置
```

---

## 后端技术栈（Go）

### 核心依赖

| 依赖 | 用途 |
|------|------|
| `github.com/gin-gonic/gin` | HTTP 路由框架 |
| `github.com/docker/docker/client` | Docker 官方 Go SDK |
| `github.com/golang-jwt/jwt/v5` | JWT 认证 |
| `github.com/gorilla/websocket` | WebSocket 日志流 |
| `gopkg.in/yaml.v3` | 配置文件解析 |
| `os/exec`（标准库）| 调用 systemctl / supervisorctl |

### 关键设计

**1. Docker 管理**

通过挂载 `/var/run/docker.sock` 与 Docker 引擎通信，无需远程 TCP，安全且低开销。

```go
// 初始化 Docker 客户端
cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

// 列出所有容器（含停止状态）
containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})

// 启动容器
err = cli.ContainerStart(ctx, containerID, container.StartOptions{})

// 实时日志流（通过 goroutine + WebSocket 推送）
reader, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
    ShowStdout: true,
    ShowStderr: true,
    Follow:     true,
    Tail:       "100",
})
```

**2. 静态进程管理**

通过 `os/exec` 调用 `systemctl`，以 goroutine 并发执行，避免阻塞主线程。

```go
// 查询服务状态
func GetServiceStatus(name string) (string, error) {
    out, err := exec.Command("systemctl", "is-active", name).Output()
    return strings.TrimSpace(string(out)), err
}

// 启动服务
func StartService(name string) error {
    return exec.Command("systemctl", "start", name).Run()
}

// 读取日志（journalctl）
func TailServiceLog(name string, lines int) (string, error) {
    out, err := exec.Command("journalctl", "-u", name,
        fmt.Sprintf("-n%d", lines), "--no-pager").Output()
    return string(out), err
}
```

**3. WebSocket 日志流**

每个日志请求开启独立 goroutine，通过 channel 控制生命周期，客户端断开时自动清理。

```go
func StreamLogs(c *gin.Context) {
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
    defer conn.Close()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 从 Docker 或文件读取日志，逐行推送
    go func() {
        for line := range logChannel {
            conn.WriteMessage(websocket.TextMessage, []byte(line))
        }
    }()
}
```

### REST API 设计

```
POST   /api/auth/login              # 登录，返回 JWT

GET    /api/docker/containers        # 列出所有容器
POST   /api/docker/containers/:id/start   # 启动容器
POST   /api/docker/containers/:id/stop    # 停止容器
POST   /api/docker/containers/:id/restart # 重启容器
DELETE /api/docker/containers/:id         # 删除容器
GET    /api/docker/images            # 列出镜像
GET    /ws/docker/logs/:id           # WebSocket: 容器日志流

GET    /api/process/services         # 列出所有托管服务
POST   /api/process/services/:name/start   # 启动服务
POST   /api/process/services/:name/stop    # 停止服务
POST   /api/process/services/:name/restart # 重启服务
GET    /ws/process/logs/:name        # WebSocket: 服务日志流
```

---

## 前端技术栈（React + Ant Design）

### 核心依赖

| 依赖 | 用途 |
|------|------|
| `react` + `typescript` | 基础框架 |
| `antd` | UI 组件库（表格、卡片、标签、按钮） |
| `axios` | HTTP 请求 |
| `xterm.js` | 终端风格日志展示组件 |
| `react-router-dom` | 页面路由 |
| `zustand` 或 `redux-toolkit` | 状态管理 |

### 主要页面

**Dashboard（总览）**
- 以卡片形式展示所有应用（Docker + 静态），颜色区分运行状态
- 指标汇总：运行中 N 个 / 停止 M 个 / 异常 X 个

**Docker 管理页**
- Ant Design Table 展示容器列表（名称、镜像、状态、端口、CPU/内存）
- 操作列：启动 / 停止 / 重启 / 删除
- 点击容器名展开实时日志面板（xterm.js + WebSocket）

**静态应用管理页**
- 同上结构，数据来自 systemd/supervisord
- 日志读取 journalctl 或指定日志文件路径

---

## 部署方案

### docker-compose.yml

```yaml
version: "3.9"

services:
  backend:
    build: ./backend
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock  # 挂载 Docker socket
    environment:
      - JWT_SECRET=your-secret-key
      - GIN_MODE=release

  frontend:
    build: ./frontend
    restart: unless-stopped
    ports:
      - "3000:80"
    depends_on:
      - backend
```

### nginx.conf（反向代理）

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        proxy_pass http://frontend:80;
    }

    # 后端 API
    location /api/ {
        proxy_pass http://backend:8080;
    }

    # WebSocket 日志流
    location /ws/ {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

---

## 安全设计

| 措施 | 说明 |
|------|------|
| JWT 认证 | 所有 API 请求须携带有效 Token，有效期可配置 |
| Docker socket 权限 | 后端容器以最小权限运行，仅挂载 socket |
| systemctl 权限 | 通过 `sudoers` 配置白名单，仅允许特定服务名操作 |
| HTTPS | 生产环境通过 Nginx + Let's Encrypt 启用 TLS |
| 操作日志 | 所有启停操作记录操作人、时间、目标服务 |

---

## 性能特性（选择 Go 的理由）

- **内存占用**：常驻约 10–20 MB，对宿主机几乎无感知
- **并发模型**：goroutine 天然支持同时 stream 多个容器日志，无线程切换开销
- **单二进制部署**：`go build` 产出单个可执行文件，无运行时依赖
- **启动时间**：毫秒级启动，适合管理工具场景
- **Docker SDK**：官方维护，API 完整，类型安全

---

## 开发启动步骤

```bash
# 1. 启动后端（开发模式）
cd backend
go mod tidy
go run main.go

# 2. 启动前端（开发模式）
cd frontend
npm install
npm run dev

# 3. 生产部署
docker-compose up -d --build
```

---

## 扩展方向

- **多服务器支持**：通过 SSH 远程管理多台 Ubuntu 机器
- **告警通知**：服务异常时通过钉钉 / Slack / 邮件推送
- **定时任务管理**：集成 crontab 或 Go 内置调度器
- **资源监控图表**：接入 Prometheus + Grafana，或自研简易图表
- **部署流水线**：支持 git pull + 重启的简易 CI 能力
