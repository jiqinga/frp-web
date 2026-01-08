# FRP Web Panel

<p align="center">
  <img src="docs/screenshots/logo.png" alt="FRP Web Panel Logo" width="120">
</p>

<p align="center">
  <strong>🚀 一个现代化的 FRP 内网穿透管理面板</strong>
</p>

<p align="center">
  <a href="./README_EN.md">English</a> | 简体中文
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-18+-61DAFB?style=flat-square&logo=react" alt="React Version">
  <img src="https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat-square&logo=typescript" alt="TypeScript">
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License">
  <img src="https://img.shields.io/docker/pulls/jiqinga/frp-web-panel?style=flat-square&logo=docker" alt="Docker Pulls">
</p>

---

## 📖 项目简介

FRP Web Panel 是一个功能强大的 FRP (Fast Reverse Proxy) 可视化管理平台，提供直观的 Web 界面来管理 FRP 服务器、客户端和代理配置。支持多服务器管理、实时流量监控、告警通知、证书自动续期等企业级功能。

## ✨ 功能特性

### 🖥️ 服务器管理
- 支持多 FRP 服务器管理
- 本地服务器一键安装、启动、停止
- 远程服务器 SSH 部署和管理
- 服务器运行状态实时监控
- 服务器性能指标查看

### 📱 客户端管理
- 客户端注册和令牌管理
- 在线状态实时监控
- 配置远程同步推送
- 客户端守护进程管理
- 一键生成安装脚本
- 支持批量更新客户端

### 🔗 代理配置
- 支持 TCP/UDP/HTTP/HTTPS/STCP/SUDP/XTCP 等多种代理类型
- 可视化代理规则配置
- 域名和子域名配置
- DNS 自动同步（支持阿里云、腾讯云、Cloudflare）
- 带宽限制配置
- 插件配置支持

### 📊 实时监控
- WebSocket 实时数据推送
- 流量统计和趋势图表
- 代理连接数监控
- 24小时流量趋势分析
- 流量排行榜

### 🔐 证书管理
- SSL/TLS 证书管理
- ACME 自动申请和续期
- 证书到期提醒
- 支持多域名证书

### 🔔 告警系统
- 流量阈值告警
- 客户端离线告警
- 自定义告警规则
- 邮件通知支持
- 告警接收人管理
- 告警历史记录

### ⚙️ 系统设置
- GitHub 镜像加速配置
- DNS 提供商配置
- 邮件服务器配置
- 告警接收人管理
- 检查间隔配置

### 📝 其他功能
- JWT 安全认证
- 完整操作审计日志
- 深色/浅色主题切换
- 响应式设计，支持移动端
- IP 地理位置识别

## 📸 界面截图

<details>
<summary>点击展开截图</summary>

### 登录页面
![登录页面](docs/screenshots/login.png)

### 仪表盘
![仪表盘](docs/screenshots/dashboard.png)

### 服务器管理
![服务器列表](docs/screenshots/servers.png)
![服务器指标](docs/screenshots/server-metrics.png)
![远程安装](docs/screenshots/server-install.png)

### 客户端管理
![客户端列表](docs/screenshots/clients.png)
![客户端表单](docs/screenshots/client-form.png)
![脚本生成器](docs/screenshots/script-generator.png)

### 代理管理
![代理列表](docs/screenshots/proxies.png)
![代理表单](docs/screenshots/proxy-form.png)

### 实时监控
![实时监控](docs/screenshots/realtime-monitor.png)

### 证书管理
![证书列表](docs/screenshots/certificates.png)

### 告警管理
![告警规则](docs/screenshots/alert-rules.png)

### 系统设置
![DNS设置](docs/screenshots/settings-dns.png)
![邮件设置](docs/screenshots/settings-email.png)

### 操作日志
![操作日志](docs/screenshots/logs.png)

</details>

## 🛠️ 技术栈

### 后端
| 技术                | 说明     |
| ------------------- | -------- |
| Go 1.24+            | 编程语言 |
| Gin                 | Web 框架 |
| GORM                | ORM 框架 |
| SQLite / PostgreSQL | 数据库   |
| JWT                 | 身份认证 |
| WebSocket           | 实时通信 |
| Swagger             | API 文档 |

### 前端
| 技术         | 说明      |
| ------------ | --------- |
| React 18     | UI 框架   |
| TypeScript   | 类型安全  |
| Vite         | 构建工具  |
| Ant Design   | UI 组件库 |
| Zustand      | 状态管理  |
| React Router | 路由管理  |
| Recharts     | 图表库    |
| TailwindCSS  | 样式框架  |

### 客户端守护进程
| 技术      | 说明              |
| --------- | ----------------- |
| Go        | 编程语言          |
| WebSocket | 与服务端通信      |
| 进程管理  | frpc 生命周期管理 |

## 🚀 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+
- pnpm / npm / yarn

### Docker 部署（推荐）

#### 方式一：使用预构建镜像

```bash
# 拉取镜像
docker pull jiqinga/frp-web-panel:latest

# 运行容器
docker run -d \
  --name frp-web-panel \
  -p 80:80 \
  -v ./data:/app/data/db \
  --restart unless-stopped \
  jiqinga/frp-web-panel:latest
```

#### 方式二：使用 Docker Compose（SQLite）

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  frp-web-panel:
    image: jiqinga/frp-web-panel:latest
    container_name: frp-web-panel
    ports:
      - "80:80"      # Web 前端 (Nginx 反向代理后端 API)
      # 如需暴露 FRP 服务端口，请添加相应端口映射
      # - "7000:7000"  # frps bind_port
      # - "7500:7500"  # frps dashboard
    volumes:
      - ./data:/app/data/db       # 数据持久化
      - ./configs:/app/configs # 配置文件
    environment:
      - LOG_LEVEL=info
      - GIN_MODE=release
      - TZ=Asia/Shanghai
    restart: unless-stopped
```

启动服务:
```bash
docker-compose up -d
```

#### 方式三：使用 Docker Compose（PostgreSQL）

创建 `docker-compose-postgres.yml`:

```yaml
version: '3.8'

services:
  frp-web-panel:
    image: jiqinga/frp-web-panel:latest
    container_name: frp-web-panel
    ports:
      - "80:80"
    volumes:
      - ./data:/app/data/db
      - ./configs:/app/configs
    environment:
      - DATABASE_TYPE=postgres
      - DATABASE_POSTGRES_HOST=postgres
      - DATABASE_POSTGRES_PORT=5432
      - DATABASE_POSTGRES_USER=frp
      - DATABASE_POSTGRES_PASSWORD=frp123
      - DATABASE_POSTGRES_DBNAME=frp_panel
      - JWT_SECRET=your-secret-key-change-in-production
      - SECURITY_ENCRYPTION_KEY=12345678901234567890123456789012
      - LOG_LEVEL=info
      - TZ=Asia/Shanghai
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    container_name: frp-panel-postgres
    environment:
      - POSTGRES_USER=frp
      - POSTGRES_PASSWORD=frp123
      - POSTGRES_DB=frp_panel
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U frp -d frp_panel"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  postgres_data:
```

启动服务:
```bash
docker-compose -f docker-compose-postgres.yml up -d
```

#### 方式三：本地构建镜像

```bash
# 克隆项目
git clone https://github.com/your-username/frp-web-panel.git
cd frp-web-panel

# 构建镜像
docker build -t frp-web-panel:local .

# 运行容器
docker run -d \
  --name frp-web-panel \
  -p 80:80 \
  -v ./data:/app/data/db \
  frp-web-panel:local
```

### 手动部署

#### 1. 克隆项目

```bash
git clone https://github.com/your-username/frp-web-panel.git
cd frp-web-panel
```

#### 2. 启动后端

```bash
cd backend

# 下载依赖
go mod download

# 复制配置文件
cp configs/config.yaml.example configs/config.yaml

# 编辑配置文件
vim configs/config.yaml

# 运行服务
go run cmd/server/main.go
```

#### 3. 启动前端

```bash
cd web

# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 或构建生产版本
pnpm build
```

#### 4. 访问面板

```
地址: http://localhost:5173 (开发模式) 或 http://localhost (生产模式)
默认账号: admin
默认密码: admin123
```

> ⚠️ **安全提示**: 首次登录后请立即修改默认密码！

## ⚙️ 配置说明

### 后端配置

配置文件位于 `backend/configs/config.yaml`:

```yaml
server:
  port: 8080              # API 服务端口
  mode: release           # debug / release
  public_url: 'http://localhost:8080'  # 公网访问地址

log:
  level: info
  format: console

database:
  type: sqlite            # sqlite / postgres
  sqlite:
    path: ./data/db/frp_panel.db
  postgres:
    host: localhost
    port: 5432
    user: frp
    password: your-password
    dbname: frp_panel

jwt:
  secret: your-secret-key-change-in-production  # JWT 密钥，生产环境必须修改
  expire_hours: 24        # Token 过期时间

security:
  encryption_key: '12345678901234567890123456789012'  # 32字符加密密钥

frps:
  binary_dir: ./data/frps           # frps 二进制文件目录
  config_dir: ./data/frps/configs   # frps 配置文件目录
  log_dir: ./data/frps/logs         # frps 日志目录
  default_version: latest
  github_api: https://api.github.com/repos/fatedier/frp
```

### 环境变量

支持通过环境变量覆盖配置：

```bash
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=release
SERVER_PUBLIC_URL=https://your-domain.com

# 数据库配置
DB_TYPE=sqlite
DB_SQLITE_PATH=./data/db/frp_panel.db

# JWT 配置
JWT_SECRET=your-super-secret-key
JWT_EXPIRE_HOURS=24

# 安全配置
SECURITY_ENCRYPTION_KEY=your-32-character-encryption-key
```

### Docker 数据卷说明

| 路径           | 说明     |
| -------------- | -------- |
| `/app/data/db` | 数据库   |
| `/app/configs` | 配置文件 |

### 端口说明

| 端口 | 说明                                  |
| ---- | ------------------------------------- |
| 80   | Web 前端界面 (Nginx 反向代理后端 API) |
| 7000 | frps 默认绑定端口 (需自行映射)        |
| 7500 | frps Dashboard 端口 (需自行映射)      |

## 📁 项目结构

```
frp-web-panel/
├── backend/                    # Go 后端服务
│   ├── cmd/server/            # 程序入口
│   │   ├── main.go            # 主函数
│   │   ├── bootstrap.go       # 初始化
│   │   └── scheduler.go       # 定时任务
│   ├── configs/               # 配置文件
│   ├── data/                  # 运行时数据（IP库等）
│   ├── docs/                  # Swagger API 文档
│   ├── internal/              # 内部模块
│   │   ├── config/            # 配置加载
│   │   ├── container/         # 依赖注入容器
│   │   ├── errors/            # 错误定义
│   │   ├── events/            # 事件总线
│   │   ├── frp/               # FRP 客户端封装
│   │   ├── handler/           # HTTP 处理器
│   │   ├── logger/            # 日志模块
│   │   ├── middleware/        # 中间件
│   │   ├── model/             # 数据模型
│   │   ├── repository/        # 数据访问层
│   │   ├── router/            # 路由定义
│   │   ├── service/           # 业务逻辑
│   │   ├── util/              # 工具函数
│   │   └── websocket/         # WebSocket 处理
│   ├── migrations/            # 数据库迁移脚本
│   └── pkg/                   # 可复用包
├── web/                       # React 前端应用
│   ├── src/
│   │   ├── api/              # API 接口封装
│   │   ├── assets/           # 静态资源
│   │   ├── components/       # 公共组件
│   │   ├── constants/        # 常量定义
│   │   ├── hooks/            # 自定义 Hooks
│   │   ├── pages/            # 页面组件
│   │   ├── router/           # 路由配置
│   │   ├── store/            # 状态管理
│   │   ├── styles/           # 样式文件
│   │   ├── types/            # TypeScript 类型
│   │   └── utils/            # 工具函数
│   └── public/               # 静态资源
├── docker/                    # Docker 相关配置
│   └── s6-rc.d/              # s6 进程管理配置
├── docs/                      # 项目文档
│   └── screenshots/          # 截图文件
├── .github/                   # GitHub 配置
├── Dockerfile                 # Docker 构建文件
└── README.md                  # 项目说明
```

## 📖 API 文档

启动后端后访问 Swagger 文档:

```
http://localhost:8080/swagger/index.html
```

## 🔧 客户端守护进程

客户端守护进程 (frpc-daemon-ws) 用于管理远程 frpc 客户端，支持配置同步、健康检查和自动更新。

### 功能特性

- WebSocket 长连接通信
- 配置自动同步
- frpc 进程生命周期管理
- 心跳检测
- 自动更新支持

### 安装方式

#### 方式一：通过面板生成安装脚本

1. 在面板中添加客户端
2. 点击"生成脚本"按钮
3. 复制生成的安装脚本到目标机器执行

#### 方式二：手动安装

1. 从面板下载对应平台的守护进程二进制文件
2. 创建配置文件 `daemon.yaml`:

```yaml
client_id: 1                                              # 客户端 ID
server_url: ws://your-panel-server:8080/api/ws/client-daemon  # 面板 WebSocket 地址
token: your-client-token                                  # 客户端令牌
frpc_path: ./frpc                                         # frpc 二进制路径
frpc_config: ./frpc.toml                                  # frpc 配置文件路径
frpc_admin_addr: 127.0.0.1                               # frpc admin 地址
frpc_admin_port: 7400                                     # frpc admin 端口
heartbeat_sec: 30                                         # 心跳间隔（秒）
```

3. 运行守护进程:

```bash
./frpc-daemon-ws -c daemon.yaml
```

### 支持的平台

- Linux (amd64, arm64, arm)
- Windows (amd64, 386)
- macOS (amd64, arm64)

## 🔒 安全建议

1. **修改默认密码**: 首次登录后立即修改 admin 默认密码
2. **配置 JWT 密钥**: 生产环境必须修改 `jwt.secret` 配置
3. **配置加密密钥**: 修改 `security.encryption_key` 为随机 32 字符字符串
4. **使用 HTTPS**: 生产环境建议配置 SSL/TLS 证书
5. **限制访问**: 使用防火墙限制面板访问来源
6. **定期备份**: 定期备份 `/app/data/db` 目录

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

### 开发环境设置

```bash
# 后端开发
cd backend
go mod download
go run cmd/server/main.go

# 前端开发
cd web
pnpm install
pnpm dev
```

## 📄 开源协议

本项目采用 [MIT](LICENSE) 开源协议。

## 🙏 致谢

- [frp](https://github.com/fatedier/frp) - 快速反向代理
- [Gin](https://github.com/gin-gonic/gin) - Go Web 框架
- [GORM](https://gorm.io/) - Go ORM 框架
- [Ant Design](https://ant.design/) - React UI 组件库
- [Recharts](https://recharts.org/) - React 图表库

## 📞 联系方式

- 提交 Issue: [GitHub Issues](https://github.com/your-username/frp-web-panel/issues)

---
