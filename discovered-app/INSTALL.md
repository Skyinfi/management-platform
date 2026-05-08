# discovered-app 安装指南

宿主机扫描 Agent，用于发现未被 systemd/docker 管理的野生进程。

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SCANNER_ADDR` | `:8081` | 监听地址 |
| `PROC_ROOT` | `/proc` | proc 文件系统路径（容器部署时改为 `/host/proc`） |
| `SCANNER_INTERVAL` | `60s` | 定时扫描间隔（支持 `60s` 或纯数字秒数 `60`） |

## 方式一：systemd 直接部署（推荐）

适用于宿主机直接部署，无需 Docker，天然可读 `/proc`。

### 1. 交叉编译

在开发机或 CI 上执行：

```bash
# 在项目根目录
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o discovered-app/discovered-app ./discovered-app/main.go
```

### 2. 上传到目标服务器

```bash
# 上传二进制
scp discovered-app/discovered-app your-server:/usr/local/bin/

# 上传 systemd 服务文件
scp scripts/discovered-app.service your-server:/etc/systemd/system/
```

### 3. 启动服务

```bash
ssh your-server

systemctl daemon-reload
systemctl enable discovered-app
systemctl start discovered-app
systemctl status discovered-app
```

### 4. 验证

```bash
# 检查服务状态
systemctl status discovered-app

# 查看日志
journalctl -u discovered-app -f

# 测试 API
curl http://localhost:8081/api/health
curl -X POST http://localhost:8081/api/scanner/run
curl http://localhost:8081/api/scanner/apps
```

### 5. 自定义配置

编辑 systemd 服务文件中的环境变量：

```bash
systemctl edit discovered-app
```

添加覆盖：

```ini
[Service]
Environment=SCANNER_ADDR=:9090
Environment=SCANNER_INTERVAL=30s
```

然后重启：

```bash
systemctl daemon-reload
systemctl restart discovered-app
```

---

## 方式二：Makefile 一键编译部署

```bash
# 编译 + 上传 + 启动
make deploy-scanner DEPLOY_HOST=user@your-server
```

---

## 方式三：Docker 部署

如果需要容器化部署，必须挂载宿主机 `/proc` 只读。

### Dockerfile

在 `discovered-app/` 目录下创建 `Dockerfile`：

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /discovered-app ./main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /discovered-app /usr/local/bin/discovered-app
EXPOSE 8081
ENTRYPOINT ["discovered-app"]
```

### docker run

```bash
docker build -t discovered-app ./discovered-app

docker run -d \
  --name discovered-app \
  --restart unless-stopped \
  --network host \
  --pid host \
  -v /proc:/host/proc:ro \
  -e PROC_ROOT=/host/proc \
  -e SCANNER_ADDR=:8081 \
  -e SCANNER_INTERVAL=60s \
  discovered-app
```

### docker-compose

在项目根目录的 `docker-compose.yml` 中追加：

```yaml
  scanner:
    build:
      context: ./discovered-app
      dockerfile: Dockerfile
    restart: unless-stopped
    network_mode: host
    pid: host
    volumes:
      - /proc:/host/proc:ro
    environment:
      - PROC_ROOT=/host/proc
      - SCANNER_ADDR=:8081
      - SCANNER_INTERVAL=60s
```

然后：

```bash
docker compose up -d --build scanner
```

---

## 与 app-manager 集成

`discovered-app` 启动后，`app-manager` 通过 HTTP 调用其 API 获取扫描结果。

在 `app-manager` 的 `config.yaml` 中配置 scanner 地址：

```yaml
scanner:
  addr: http://localhost:8081
```

或通过环境变量：

```bash
APP_MANAGER_SCANNER_ADDR=http://localhost:8081
```

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查 |
| POST | `/api/scanner/run` | 手动触发一次全量扫描 |
| GET | `/api/scanner/apps` | 获取上次扫描发现的应用列表 |
| GET | `/api/scanner/apps/:pid` | 获取单个发现应用的详情 |
| POST | `/api/scanner/apps/:pid/watch` | 将发现的进程纳入监控 |
| GET | `/ws/scanner/progress` | WebSocket 实时推送扫描进度 |

---

## 权限说明

- `discovered-app` 仅需**读取** `/proc`，无需 root 权限
- 如果以非 root 用户运行，确保该用户有 `/proc` 下各进程目录的读权限（通常默认满足）
- Docker 挂载 `/proc` 时务必加 `:ro` 只读标记

## 防火墙

`discovered-app` 默认监听 `8081` 端口，仅需对 `app-manager` 开放：

```bash
# 如果 app-manager 在同一台机器，无需额外配置（localhost 通信）
# 如果跨机器，需开放端口
ufw allow from <app-manager-ip> to any port 8081
```

生产环境建议不对外暴露 8081 端口，仅通过 localhost 或内网访问。
