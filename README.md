# Currency Exchange App

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-v1.12-00ADD8)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-MySQL-blue)](https://gorm.io/)
[![Redis](https://img.shields.io/badge/Redis-cache-red?logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-supported-2496ED?logo=docker)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)

一个基于 **Go + Gin + GORM + MySQL + Redis + JWT** 的汇率与文章社区后端 API 项目，支持用户注册登录、汇率数据管理、文章发布与互动、Redis 缓存和 Docker Compose 本地部署。

项目采用模块化单体结构，将 HTTP 接口、业务逻辑和数据访问拆分为 Controller、Service、Repository 三层，并通过 DTO、统一响应结构和业务错误码规范接口输出。

```text
Controller -> Service -> Repository -> MySQL
                   |
                 Redis
```

---

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [架构说明](#架构说明)
- [前置要求](#前置要求)
- [安装与运行](#安装与运行)
- [配置说明](#配置说明)
- [统一响应格式](#统一响应格式)
- [API 文档](#api-文档)
- [License](#license)

---

## 功能特性

- 用户注册与登录，基于 JWT 的接口鉴权
- bcrypt 密码哈希存储
- 汇率数据创建与查询
- 文章列表、详情、创建、更新、删除
- Redis 文章列表分页缓存
- Redis 文章点赞计数（基于用户 ID 记录点赞状态，使用 Lua 脚本保证点赞/取消点赞原子性）
- Controller / Service / Repository 分层
- DTO 请求体与响应体隔离
- 统一响应结构与业务错误码
- CORS 跨域支持
- Graceful Shutdown 优雅关停
- Docker Compose 一键启动后端、MySQL、Redis

---

## 技术栈

| 层级     | 技术                    |
| -------- | ----------------------- |
| 语言     | Go 1.26+                |
| Web 框架 | Gin v1.12               |
| ORM      | GORM                    |
| 数据库   | MySQL 8                 |
| 缓存     | Redis 7                 |
| 认证     | JWT                     |
| 密码哈希 | bcrypt                  |
| 配置管理 | Viper                   |
| 部署     | Docker / Docker Compose |

---

## 项目结构

```text
Exchangeapp_backend/
├── cmd/
│   └── server/
│       └── main.go                     # 程序入口：依赖装配、HTTP 服务启动、优雅关停
├── configs/
│   ├── config.yml                      # 本地配置文件
│   └── config.yml.example              # 配置模板
├── internal/
│   ├── apperrors/
│   │   └── errors.go                   # 业务错误码与 AppError
│   ├── config/
│   │   └── config.go                   # 配置加载、MySQL 初始化、Redis 初始化、JWT 配置
│   ├── controllers/
│   │   ├── auth_controller.go          # 注册、登录接口
│   │   ├── article_controller.go       # 文章 CRUD 与点赞接口
│   │   └── exchange_rate_controller.go # 汇率查询与创建接口
│   ├── dto/
│   │   └── dto.go                      # 请求 DTO 与响应 DTO
│   ├── middlewares/
│   │   └── auth_middleware.go          # JWT 鉴权中间件
│   ├── models/
│   │   ├── user.go                     # 用户模型
│   │   ├── article.go                  # 文章模型
│   │   └── exchange_rate.go            # 汇率模型
│   ├── repository/
│   │   ├── repositories.go             # Repository 聚合
│   │   ├── userRepository.go           # 用户数据访问
│   │   ├── articleRepository.go        # 文章数据访问
│   │   └── rateRepository.go           # 汇率数据访问
│   ├── response/
│   │   └── response.go                 # 统一响应封装
│   ├── router/
│   │   └── router.go                   # 路由注册与 CORS 中间件
│   ├── service/
│   │   ├── services.go                 # Service 聚合
│   │   ├── authService.go              # 认证业务
│   │   ├── articleService.go           # 文章业务、分页缓存、Lua 点赞
│   │   └── rateService.go              # 汇率业务
│   └── utils/
│       └── utils.go                    # bcrypt 密码哈希工具函数
├── pkg/
│   └── jwtauth/
│       └── jwtauth.go                  # JWT 生成/解析、自定义 Claims
├── Dockerfile                          # Go 服务镜像构建
├── docker-compose.yml                  # backend + MySQL + Redis 编排
├── .env.example                        # Docker 环境变量模板
├── go.mod
└── go.sum
```

---

## 架构说明

### 分层职责

- **Controller**：负责 HTTP 请求绑定、参数校验、调用 Service，并通过统一响应返回结果。
- **Service**：负责业务流程、缓存策略、密码处理、JWT 签发和业务错误转换。
- **Repository**：负责 GORM 数据库访问，不处理 HTTP 和业务响应。
- **DTO**：隔离 API 请求/响应字段和数据库模型，避免直接暴露 GORM Model。
- **Response / AppError**：统一接口返回格式和业务错误码。

### 启动与依赖装配

应用启动时由 `cmd/server/main.go` 统一装配依赖：

```text
config.InitConfig()
  -> 初始化 MySQL / Redis
    -> repository.NewRepositories(appConfig.Db)
      -> service.NewServices(repos, appConfig.RedisDB, appConfig.JWT.Secret, appConfig.JWT.TTL)
        -> router.SetupRouter(appConfig, services)
          -> 启动 HTTP Server
```

这种方式让数据库连接、Redis 客户端和业务服务在启动阶段集中创建，并通过构造函数传递给下层模块。

---

## 前置要求

### Docker 部署

- [Docker](https://docs.docker.com/get-docker/) 24+
- [Docker Compose](https://docs.docker.com/compose/) v2+

### 本地部署

- [Go](https://go.dev/dl/) 1.26+
- [MySQL](https://dev.mysql.com/downloads/) 8.0+
- [Redis](https://redis.io/download/) 7.0+

---

## 安装与运行

### Docker 部署

```bash
git clone https://github.com/bai-shiro/Go-Gin-demo-Currency-Exchange-Web.git
cd Go-Gin-demo-Currency-Exchange-Web/Exchangeapp_backend

cp .env.example .env
cp configs/config.yml.example configs/config.yml

docker compose up --build
```

服务启动后默认监听：

```text
http://localhost:3000
```

停止服务：

```bash
docker compose down
```

如需同时删除 MySQL 和 Redis 数据卷：

```bash
docker compose down -v
```

### 本地部署

#### 1. 克隆项目

```bash
git clone https://github.com/bai-shiro/Go-Gin-demo-Currency-Exchange-Web.git
cd Go-Gin-demo-Currency-Exchange-Web/Exchangeapp_backend
```

#### 2. 创建数据库

```sql
CREATE DATABASE currency_exchange_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

#### 3. 准备配置文件

```bash
cp configs/config.yml.example configs/config.yml
```

根据本机环境修改 `configs/config.yml`：

```yaml
app:
  name: CurrencyExchangeApp
  port: :3000

database:
  dsn: root:123456@tcp(127.0.0.1:3306)/currency_exchange_db?charset=utf8mb4&parseTime=True&loc=Local
  maxIdleConns: 11
  maxOpenConns: 114

cache:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

jwt:
  secret: your-secret
  ttl: 24h
```

#### 4. 启动服务

```bash
go run ./cmd/server
```

---

## 配置说明

配置文件路径：

```text
Exchangeapp_backend/configs/config.yml
```

Docker Compose 使用 `.env` 中的变量覆盖容器内连接地址：

```env
DB_HOST=mysql
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=currency_exchange_db
REDIS_HOST=redis
REDIS_PORT=6379
```

---

## 统一响应格式

成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

错误响应：

```json
{
  "code": 40001,
  "message": "invalid params"
}
```

当前错误码：

| code  | message               | HTTP 状态 |
| ----- | --------------------- | --------- |
| 0     | success               | 200 / 201 |
| 40001 | invalid params        | 400       |
| 40101 | unauthorized          | 401       |
| 40301 | forbidden             | 403       |
| 40401 | not found             | 404       |
| 50001 | internal server error | 500       |

---

## API 文档

### 认证接口

| 方法 | 路径                 | 说明     | 认证 |
| ---- | -------------------- | -------- | ---- |
| POST | `/api/auth/register` | 用户注册 | 否   |
| POST | `/api/auth/login`    | 用户登录 | 否   |

请求体：

```json
{
  "username": "user1",
  "password": "123456"
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "<jwt-token>"
  }
}
```

### 汇率接口

| 方法 | 路径                 | 说明         | 认证 |
| ---- | -------------------- | ------------ | ---- |
| GET  | `/api/exchangeRates` | 查询汇率列表 | 否   |
| POST | `/api/exchangeRates` | 创建汇率     | 是   |

创建汇率请求体：

```json
{
  "fromCurrency": "USD",
  "toCurrency": "CNY",
  "rate": 7.25
}
```

### 文章接口

| 方法   | 路径                                | 说明               | 认证 |
| ------ | ----------------------------------- | ------------------ | ---- |
| GET    | `/api/articles?page=1&page_size=10` | 查询文章列表(分页) | 否   |
| GET    | `/api/articles/:id`                 | 查询文章详情       | 否   |
| GET    | `/api/articles/:id/likes`           | 查询文章点赞数     | 否   |
| POST   | `/api/articles`                     | 创建文章           | 是   |
| PUT    | `/api/articles/:id`                 | 更新文章           | 是   |
| DELETE | `/api/articles/:id`                 | 删除文章           | 是   |
| POST   | `/api/articles/:id/like`            | 点赞文章           | 是   |

创建 / 更新文章请求体：

```json
{
  "title": "hello",
  "content": "article content",
  "preview": "short preview"
}
```

文章响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "title": "hello",
    "content": "article content",
    "preview": "short preview"
  }
}
```

点赞/取消点赞响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "liked": true,
    "likes": 1
  }
}
```

点赞数响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "likes": 1
  }
}
```

### 认证方式

需要认证的接口在请求头中携带 JWT：

```http
Authorization: Bearer <jwt-token>
```

---

## Curl 示例

注册：

```bash
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"user1\",\"password\":\"123456\"}"
```

登录：

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"user1\",\"password\":\"123456\"}"
```

创建文章：

```bash
curl -X POST http://localhost:3000/api/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d "{\"title\":\"hello\",\"content\":\"article content\",\"preview\":\"short preview\"}"
```

查询文章：

```bash
curl http://localhost:3000/api/articles
```

点赞文章：

```bash
curl -X POST http://localhost:3000/api/articles/1/like \
  -H "Authorization: Bearer <jwt-token>"
```

---

## License

[MIT](LICENSE)
