# Currency Exchange App

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-v1.12-00ADD8)](https://gin-gonic.com/)
[![Docker](https://img.shields.io/badge/Docker-✓-2496ED?logo=docker)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)

一个基于 **Go + Gin + GORM + Redis + MySQL** 的汇率兑换与新闻文章管理后端 API 项目，支持 Docker 一键部署，配套 Vue 3 前端。

(重构代码目录为经典Golang项目目录,预更新)

---

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [前置要求](#前置要求)
- [安装与运行](#安装与运行)
- [API 文档](#api-文档)
- [相关项目](#相关项目)

---

## 功能特性

- ✅ 用户注册 & 登录（JWT 认证）
- ✅ 汇率查询 & 创建（货币兑换）
- ✅ 文章 CRUD（含 Redis 缓存）
- ✅ 文章点赞功能（Redis 存储）
- ✅ CORS 跨域支持
- ✅ Graceful Shutdown（优雅关停）
- ✅ Redis 旁路缓存（Cache-Aside 模式）

---

## 技术栈

| 层级     | 技术                |
| -------- | ------------------- |
| 语言     | Go 1.26+            |
| Web 框架 | Gin v1.12           |
| ORM      | GORM v1.31          |
| 数据库   | MySQL 8             |
| 缓存     | Redis 7（go-redis） |
| 认证     | JWT（HS256）        |
| 密码哈希 | bcrypt              |
| 配置管理 | Viper               |

---

## 项目结构

```
Exchangeapp_backend/
├── cmd/server/
│   └── main.go                 # 入口：优雅关停 + HTTP 服务启动
├── configs/
│   ├── config.go               # 配置文件加载（Viper）
│   ├── config.yml.example      # 配置模板（clone 后复制为 config.yml）
│   ├── db.go                   # MySQL 连接初始化
│   └── redis.go                # Redis 连接初始化
├── internal/
│   ├── controllers/
│   │   ├── auth_controller.go      # 注册 & 登录
│   │   ├── exchange_rate_controller.go  # 汇率查询 & 创建
│   │   ├── article_controller.go   # 文章 CRUD（含 Redis 缓存）
│   │   └── like_controller.go      # 文章点赞 & 查看点赞数
│   ├── global/
│   │   └── global.go               # 全局变量（DB、RedisDB）
│   ├── middlewares/
│   │   └── auth_middleware.go      # JWT 鉴权中间件
│   ├── models/
│   │   ├── user.go                 # 用户模型
│   │   ├── exchange_rate.go        # 汇率模型
│   │   └── article.go              # 文章模型
│   ├── router/
│   │   └── router.go               # 路由注册 + CORS 中间件
│   └── utils/
│   │   └── utils.go                # JWT 生成/解析 + bcrypt 哈希
├── Dockerfile                  # 多阶段构建（Go 编译 → Alpine 运行）
├── docker-compose.yml          # 一键启动（backend + MySQL + Redis）
├── .dockerignore               # 排除敏感文件进入 Docker 镜像
├── .env.example                # 环境变量模板（clone 后复制为 .env）
├── go.mod
└── go.sum

其他文档（暂未上线）：
├── Exchangeapp_frontend/       # Vue 3 前端（独立项目）
├── 学习文档记录/               # 学习笔记
```

---

## 前置要求

### Docker 部署（推荐）

- [Docker](https://docs.docker.com/get-docker/) 24+
- [Docker Compose](https://docs.docker.com/compose/install/) v2+

### 传统本地部署

- [Go](https://go.dev/dl/) 1.26+
- [MySQL](https://dev.mysql.com/downloads/) 8.0+
- [Redis](https://redis.io/download/) 7.0+
- （可选）[Node.js](https://nodejs.org/) 18+（运行前端需要）

---

## 安装与运行

### 🐳 Docker 部署（推荐）

无需安装 Go、MySQL、Redis，只需 Docker 即可一键启动。

```bash
# 1. 克隆仓库
git clone https://github.com/bai-shiro/Go-Gin-demo-Currency-Exchange-Web.git
cd Go-Gin-demo-Currency-Exchange-Web/Exchangeapp_backend

# 2. 从模板创建配置文件
cp config/config.yml.example config/config.yml
cp .env.example .env
# （可选）编辑 .env 修改数据库密码等

# 3. 一键启动
docker-compose up --build -d

# 4. 查看日志
docker-compose logs -f backend
```

`docker-compose up` 会自动完成以下工作：

- 拉取 MySQL 8.0、Redis 7.4 镜像
- 编译 Go 项目并启动后端服务（端口 3000）
- MySQL 健康检查通过后 backend 才启动，避免连接失败

停止服务：

```bash
docker-compose down
```

> 注意：`docker-compose down -v` 会删除数据库数据，谨慎使用。

### 传统本地部署

#### 1. 克隆仓库

```bash
git clone https://github.com/bai-shiro/Go-Gin-demo-Currency-Exchange-Web.git
cd Go-Gin-demo-Currency-Exchange-Web/Exchangeapp_backend
```

#### 2. 配置数据库与 Redis

确保 MySQL 和 Redis 已启动。

```bash
# 从模板创建配置文件，填入你自己的数据库密码
cp config/config.yml.example config/config.yml
```

编辑 `config/config.yml`：

```yaml
database:
  dsn: root:your_password@tcp(127.0.0.1:3306)/currency_exchange_db?charset=utf8mb4&parseTime=True&loc=Local
```

> 注意：需先在 MySQL 中创建数据库 `currency_exchange_db`。

#### 3. 启动后端

```bash
cd Exchangeapp_backend/cmd/server
go run main.go
```

服务启动后输出：无报错即成功。按 `Ctrl+C` 优雅关停。

#### 4. （可选）启动前端

```bash
cd Exchangeapp_frontend
npm install
npm run dev
```

浏览器访问 `http://localhost:5173`。

---

## API 文档

### 认证接口

| 方法 | 路径                 | 说明     | 认证 |
| ---- | -------------------- | -------- | ---- |
| POST | `/api/auth/register` | 用户注册 | ❌   |
| POST | `/api/auth/login`    | 用户登录 | ❌   |

**注册请求示例：**

```bash
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"123456"}'
```

**响应：**

```json
{ "token": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
```

**登录同理**，请求体与注册一致。

### 汇率接口

| 方法 | 路径                 | 说明         | 认证 |
| ---- | -------------------- | ------------ | ---- |
| GET  | `/api/exchangeRates` | 获取所有汇率 | ❌   |
| POST | `/api/exchangeRates` | 创建汇率     | ✅   |

### 文章接口

| 方法 | 路径                     | 说明                       | 认证 |
| ---- | ------------------------ | -------------------------- | ---- |
| GET  | `/api/articles`          | 获取文章列表（Redis 缓存） | ❌   |
| GET  | `/api/articles/:id`      | 获取单篇文章               | ❌   |
| POST | `/api/articles`          | 创建文章（会清理缓存）     | ✅   |
| GET  | `/api/articles/:id/like` | 查看点赞数（Redis）        | ❌   |
| POST | `/api/articles/:id/like` | 点赞文章（Redis）          | ✅   |

### 认证方式

需登录的接口在请求头中携带 JWT：

```bash
curl -X POST http://localhost:3000/api/exchangeRates \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{"from":"USD","to":"CNY","rate":7.25}'
```

---

## 相关项目

- **前端仓库（暂未上线）**：[Exchangeapp Frontend](https://github.com/<用户名>/<前端仓库>) — Vue 3 + Vite 前端页面

---

## License

[MIT](LICENSE)
