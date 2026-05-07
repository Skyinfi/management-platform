# 应用管理平台（App Management Platform）

单仓库包含 **Go 后端**（`app-manager`）与 **React + Vite 前端**（`manager-frontend-app`），用于在服务器上统一管理 Docker 与进程类应用。详细架构见仓库内 [`app-management-platform.md`](./app-management-platform.md)。

## 仓库结构

| 路径 | 说明 |
|------|------|
| `app-manager/` | Go API 服务（默认 `:8080`） |
| `manager-frontend-app/` | 管理端 Web UI（Vite 开发服务器默认 `:5173`） |
| `scripts/` | 本地开发辅助脚本（PowerShell） |
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
