# Currency Exchange App

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-v1.12-00ADD8)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-MySQL-blue)](https://gorm.io/)
[![Redis](https://img.shields.io/badge/Redis-cache-red?logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-supported-2496ED?logo=docker)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)

一个基于 **Go + Gin + GORM + MySQL + Redis + JWT** 的汇率与文章社区后端 API 项目，支持用户注册登录、汇率数据管理(接入第三方实时汇率API)、文章发布与互动、Redis 缓存、SQL Migration 和 Docker Compose 本地部署。

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
- [测试验证](#测试验证)
- [压测记录](#压测记录)
- [License](#license)

---

## 功能特性

- 用户注册与登录，基于 JWT 的接口鉴权
- bcrypt 密码哈希存储
- 汇率数据创建与查询，使用 `DECIMAL(20,10)` 精确存储汇率，支持按货币对和汇率日期唯一约束
- 支持外部汇率 API 查询与金额换算
- 文章列表、详情、创建、更新、删除
- Redis 文章列表分页缓存
- Redis 汇率热点缓存，缓存最新货币对汇率
- 基于 `golang-migrate` 的版本化 SQL migration，支持数据库结构可追踪升级
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
| 迁移工具 | golang-migrate          |
| 配置管理 | Viper                   |
| 部署     | Docker / Docker Compose |

---

## 项目结构

```text
Exchangeapp_backend/
├── cmd/
│   ├── migrate/
│   │   └── main.go                     # 数据库 migration 命令入口
│   └── server/
│       └── main.go                     # HTTP 服务入口：依赖装配、服务启动、优雅关停
├── configs/
│   ├── config.yml                      # 本地配置文件（不建议提交敏感配置）
│   └── config.yml.example              # 配置模板
├── internal/
│   ├── apperrors/
│   │   └── errors.go                   # 业务错误码与 AppError
│   ├── client/
│   │   └── exchange/
│   │       └── frankfurter.go          # Frankfurter 外部汇率 API 客户端
│   ├── config/
│   │   └── config.go                   # 配置加载、MySQL 初始化、Redis 初始化、JWT 配置
│   ├── controllers/
│   │   ├── auth_controller.go          # 注册、登录接口
│   │   ├── article_controller.go       # 文章 CRUD 与点赞接口
│   │   ├── controllers.go              # Controller 聚合/占位
│   │   └── exchange_rate_controller.go # 汇率查询、创建与换算接口
│   ├── dbmigrate/
│   │   └── dbmigrate.go                # golang-migrate 封装
│   ├── dto/
│   │   └── dto.go                      # 请求 DTO 与响应 DTO
│   ├── middlewares/
│   │   └── auth_middleware.go          # JWT 鉴权中间件
│   ├── models/
│   │   ├── article.go                  # 文章模型
│   │   ├── exchange_rate.go            # 汇率模型
│   │   └── user.go                     # 用户模型
│   ├── repository/
│   │   ├── articleRepository.go        # 文章数据访问
│   │   ├── rateRepository.go           # 汇率数据访问
│   │   ├── repositories.go             # Repository 聚合
│   │   └── userRepository.go           # 用户数据访问
│   ├── response/
│   │   └── response.go                 # 统一响应封装
│   ├── router/
│   │   └── router.go                   # 路由注册与 CORS 中间件
│   └── service/
│       ├── articleLike_test.go         # Redis Lua 点赞/取消点赞集成测试
│       ├── articleService.go           # 文章业务、分页缓存、Lua 点赞
│       ├── authService.go              # 认证业务
│       ├── rateService.go              # 汇率业务、热点缓存、换算
│       ├── rateService_test.go         # 汇率缓存与换算测试
│       └── services.go                 # Service 聚合
├── pkg/
│   ├── jwtauth/
│   │   ├── jwtauth.go                  # JWT 生成/解析、自定义 Claims
│   │   └── jwtauth_test.go             # JWT 单元测试
│   └── passwordbcrypt/
│       ├── passwordBcrypt.go           # bcrypt 密码哈希工具函数
│       └── passwordBcrypt_test.go      # bcrypt 单元测试
├── .dockerignore                       # Docker 构建忽略规则
├── .env.example                        # Docker 环境变量模板
├── Dockerfile                          # Go 服务镜像构建
├── docker-compose.yml                  # backend + MySQL + Redis 编排
├── migrations/                         # SQL migration 文件
│   ├── 000001_init_schema.up.sql       # 当前项目初始表结构 baseline
│   ├── 000001_init_schema.down.sql
│   ├── 000002_update_exchange_rates_dates_and_decimal.up.sql
│   └── 000002_update_exchange_rates_dates_and_decimal.down.sql
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

数据库结构变更通过 `cmd/migrate` 单独执行：

```text
cmd/migrate
  -> config.LoadConfig()
  -> 读取 configs/config.yml 的 database.dsn
  -> 如存在 DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME，则覆盖 DSN
  -> 如存在 DB_URL，则优先使用 DB_URL
  -> 执行 migrations/*.sql
```

生产和团队协作中推荐先执行 migration，再启动或重启后端服务。`database.autoMigrate` 默认保持 `false`，避免服务启动时隐式修改数据库结构。

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
```

先启动 MySQL 和 Redis：

```bash
docker compose up -d mysql redis
```

Docker MySQL 在容器内部使用 `mysql:3306`，映射到宿主机为 `127.0.0.1:3307`。在宿主机执行 migration 时，使用宿主机可访问的 `3307`：

PowerShell：

```powershell
$env:DB_URL = "mysql://root:password@tcp(127.0.0.1:3307)/currency_exchange_db?charset=utf8mb4&parseTime=true&loc=Local"
go run ./cmd/migrate up
Remove-Item Env:DB_URL
```

Bash：

```bash
DB_URL="mysql://root:password@tcp(127.0.0.1:3307)/currency_exchange_db?charset=utf8mb4&parseTime=true&loc=Local" \
  go run ./cmd/migrate up
```

迁移完成后启动后端：

```bash
docker compose up -d --build backend
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

先执行数据库 migration：

```bash
go run ./cmd/migrate up
```

再启动 HTTP 服务：

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

注意：

- `.env` 由 Docker Compose 读取，并通过 `environment` 注入到容器内，供后端程序用 `os.Getenv` 获取。
- 容器内后端连接 Docker MySQL 使用 `mysql:3306`。
- 宿主机上的 migration 命令连接 Docker MySQL 使用 `127.0.0.1:3307`。
- 本地 `go run ./cmd/server` 默认读取 `configs/config.yml`，通常连接本机 MySQL `127.0.0.1:3306`。

数据库迁移命令：

```bash
go run ./cmd/migrate up       # 执行所有未执行的迁移
go run ./cmd/migrate down     # 回滚 1 个版本
go run ./cmd/migrate version  # 查看当前数据库版本和 dirty 状态
```

`cmd/migrate` 默认读取 `configs/config.yml`，也支持 `DB_URL` 显式覆盖目标数据库：

```bash
DB_URL="mysql://root:password@tcp(127.0.0.1:3307)/currency_exchange_db?charset=utf8mb4&parseTime=true&loc=Local" \
  go run ./cmd/migrate version
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

| 方法 | 路径                                         | 说明               | 认证 |
| ---- | -------------------------------------------- | ------------------ | ---- |
| GET  | `/api/exchangeRates`                         | 查询已录入汇率列表 | 否   |
| GET  | `/api/rates/latest?base=USD&quote=CNY`       | 查询最新货币对汇率 | 否   |
| GET  | `/api/convert?base=USD&quote=CNY&amount=100` | 汇率换算           | 否   |
| POST | `/api/exchangeRates`                         | 创建汇率           | 是   |

创建汇率请求体：

```json
{
  "fromCurrency": "USD",
  "toCurrency": "CNY",
  "rate": "7.2500000000",
  "rateDate": "2026-05-30"
}
```

说明：

- `rate` 使用字符串传输，后端使用 `decimal.Decimal` 解析，数据库使用 `DECIMAL(20,10)` 存储，避免浮点精度误差。
- `rateDate` 表示汇率日期，格式为 `YYYY-MM-DD`；不传时默认使用当天日期。
- `fetchedAt` 由服务端写入，表示该汇率记录的抓取/录入时间。

汇率响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "fromCurrency": "USD",
      "toCurrency": "CNY",
      "rate": "7.2500000000",
      "rateDate": "2026-05-30",
      "fetchedAt": "2026-05-30 14:23:45"
    }
  ]
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

## 测试验证

当前项目补充了认证、密码哈希和 Redis 点赞核心链路测试：

- `pkg/jwtauth`：覆盖 JWT 生成解析、错误密钥、缺少 Bearer 前缀、过期 token、缺少 `user_id` claim。
- `pkg/passwordbcrypt`：覆盖密码哈希、正确密码校验和错误密码拒绝。
- `internal/service`：覆盖 Redis Lua 点赞/取消点赞 toggle、多用户点赞计数、汇率缓存命中/回源、坏缓存回源和汇率换算场景。

运行测试：

```bash
cd Exchangeapp_backend
go test ./...
```

本次验证结果：

```text
ok   exchangeapp/internal/service
ok   exchangeapp/pkg/jwtauth
ok   exchangeapp/pkg/passwordbcrypt
```

---

## 压测记录

压测目标：文章分页列表接口。

```text
GET /api/articles?page=1&page_size=10
```

压测环境：Docker 后端容器内访问本机服务，预先创建 100 篇测试文章，并先请求一次接口预热 Redis 分页缓存。

压测参数：

```text
requests: 1000
concurrency: 50
```

压测结果：

```text
seed_articles_created=100
warmup_duration=2.270321ms
requests=1000 concurrency=50 elapsed=166.934669ms qps=5990.37
latency_avg=7.768099ms p50=5.447397ms p95=29.14436ms p99=33.863744ms max=53.363188ms
status_counts=map[200:1000] errors=0
```

说明：该结果主要验证 Redis 命中场景下文章分页接口的吞吐和延迟表现。

---

## License

[MIT](LICENSE)
