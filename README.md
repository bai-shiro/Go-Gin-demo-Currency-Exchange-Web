# Go Exchange Hub

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-v1.12-00ADD8)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-MySQL-blue)](https://gorm.io/)
[![Redis](https://img.shields.io/badge/Redis-cache-red?logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-supported-2496ED?logo=docker)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)

基于 **Go + Gin + GORM + MySQL + Redis + JWT** 的汇率查询与内容管理后端项目。

项目围绕“汇率数据查询与持久化”展开，同时包含用户认证、文章管理、分页缓存、点赞互动、外部 API 接入、历史汇率查询、测试和压测验证。

```text
HTTP API
  -> Controller
  -> Service
       -> Redis cache
       -> External exchange-rate API
       -> Repository
            -> MySQL
```

## 功能特性

- 用户注册、登录、JWT 鉴权
- bcrypt 密码哈希存储
- 文章创建、查询、更新、删除
- 文章分页查询与 Redis 缓存
- 文章点赞 / 取消点赞，使用 Redis Lua 脚本保证原子性
- 外部汇率 API 接入，支持最新汇率查询
- 汇率换算，使用 `decimal.Decimal` 处理金额和汇率精度
- 最新汇率 Redis 热点缓存
- 汇率数据持久化到 MySQL
- 历史汇率区间查询
- 基于 `golang-migrate` 的 SQL migration
- Controller / Service / Repository 分层
- 统一响应结构和业务错误码
- Docker Compose 本地部署
- 单元测试、Redis 集成测试、基础压测记录

## 技术栈

| 分类       | 技术                    |
| ---------- | ----------------------- |
| 语言       | Go 1.26+                |
| Web 框架   | Gin                     |
| ORM        | GORM                    |
| 数据库     | MySQL 8                 |
| 缓存       | Redis                   |
| 认证       | JWT                     |
| 密码处理   | bcrypt                  |
| 精度处理   | shopspring/decimal      |
| 配置       | Viper                   |
| 数据库迁移 | golang-migrate          |
| 部署       | Docker / Docker Compose |

## 项目结构

```text
Exchangeapp_backend/
├── cmd/
│   ├── migrate/
│   │   └── main.go                     # 数据库 migration 命令入口
│   └── server/
│       └── main.go                     # HTTP 服务入口
├── configs/
│   └── config.yml.example              # 配置文件模板
├── internal/
│   ├── apperrors/
│   │   └── errors.go                   # 业务错误码
│   ├── client/
│   │   └── exchange/
│   │       └── frankfurter.go          # 外部汇率 API 客户端
│   ├── config/
│   │   └── config.go                   # 配置加载、MySQL/Redis 初始化
│   ├── controllers/
│   │   ├── article_controller.go       # 文章接口
│   │   ├── auth_controller.go          # 注册登录接口
│   │   ├── controllers.go              # Controller 聚合
│   │   └── exchange_rate_controller.go # 汇率接口
│   ├── dbmigrate/
│   │   └── dbmigrate.go                # migration 封装
│   ├── dto/
│   │   └── dto.go                      # 请求和响应 DTO
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
│   │   └── router.go                   # 路由注册
│   └── service/
│       ├── articleLike_test.go         # Redis Lua 点赞测试
│       ├── articleService.go           # 文章业务、分页缓存、点赞
│       ├── authService.go              # 认证业务
│       ├── rateService.go              # 汇率缓存、换算、历史查询
│       ├── rateService_test.go         # 汇率 service 测试
│       └── services.go                 # Service 聚合
├── migrations/
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_update_exchange_rates_dates_and_decimal.up.sql
│   └── 000002_update_exchange_rates_dates_and_decimal.down.sql
├── pkg/
│   ├── jwtauth/
│   │   ├── jwtauth.go                  # JWT 生成与解析
│   │   └── jwtauth_test.go
│   └── passwordbcrypt/
│       ├── passwordBcrypt.go           # bcrypt 工具函数
│       └── passwordBcrypt_test.go
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## 核心设计说明

### 分层设计

- `Controller`：处理 HTTP 参数、调用 service、返回统一响应。
- `Service`：承载业务规则，例如缓存回源、日期校验、金额精度计算、JWT 相关逻辑。
- `Repository`：封装 GORM 数据库访问，不处理 HTTP 语义。
- `DTO`：隔离 API 字段和数据库模型。

### 汇率模块

汇率模块当前包含三条主要链路：

```text
最新汇率查询:
GET /api/rates/latest
  -> Redis 查询 rates:latest:{BASE}:{QUOTE}
  -> 命中则直接返回
  -> 未命中 / 坏缓存 / Redis 异常则请求外部 API
  -> 写入 Redis
  -> Upsert 到 MySQL

汇率换算:
GET /api/rates/convert
  -> 查询最新汇率
  -> 使用 decimal.Decimal 计算 amount * rate

历史汇率:
GET /api/rates/history
  -> 校验币种和日期范围
  -> 查询 MySQL 中已经持久化的历史记录
```

汇率表使用 `(from_currency, to_currency, rate_date)` 作为唯一约束，保证同一天同一币种对只有一条记录；同时这个组合字段也服务于历史区间查询。

### Redis 缓存策略

- 文章分页缓存：缓存分页列表，降低重复查询数据库的压力。
- 汇率热点缓存：缓存最新币种对汇率，减少外部 API 调用。
- 坏缓存处理：缓存反序列化失败时删除缓存并回源。
- Redis 异常降级：Redis 查询异常时记录日志，继续走外部 API 或数据库路径。
- 点赞操作：通过 Redis Set + Lua 脚本实现点赞 / 取消点赞的原子切换。

## 快速启动

### Docker Compose

```bash
git clone https://github.com/bai-shiro/Go-Gin-demo-Currency-Exchange-Web.git
cd Go-Gin-demo-Currency-Exchange-Web/Exchangeapp_backend

cp .env.example .env
cp configs/config.yml.example configs/config.yml
docker compose up -d mysql redis
```

在宿主机执行 migration：

```powershell
$env:DB_URL = "mysql://root:password@tcp(127.0.0.1:3307)/currency_exchange_db?charset=utf8mb4&parseTime=true&loc=Local"
go run ./cmd/migrate up
Remove-Item Env:DB_URL
```

启动后端：

```bash
docker compose up -d --build backend
```

默认服务地址：

```text
http://localhost:3000
```

停止服务：

```bash
docker compose down
```

如需清理 MySQL 和 Redis 数据卷：

```bash
docker compose down -v
```

### 本地启动

```bash
cd Exchangeapp_backend
cp configs/config.yml.example configs/config.yml
go run ./cmd/migrate up
go run ./cmd/server
```

## 配置说明

配置文件路径：

```text
Exchangeapp_backend/configs/config.yml
```

示例：

```yaml
app:
  name: CurrencyExchangeApp
  port: :3000

database:
  dsn: root:123456@tcp(127.0.0.1:3306)/currency_exchange_db?charset=utf8mb4&parseTime=True&loc=Local
  maxIdleConns: 11
  maxOpenConns: 114
  autoMigrate: false

cache:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

jwt:
  secret: your-secret
  ttl: 24h
```

推荐通过 migration 管理表结构，`database.autoMigrate` 默认保持关闭，避免服务启动时隐式修改数据库结构。

## 统一响应

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

| code  | message               | HTTP 状态 |
| ----- | --------------------- | --------- |
| 0     | success               | 200 / 201 |
| 40001 | invalid params        | 400       |
| 40101 | unauthorized          | 401       |
| 40301 | forbidden             | 403       |
| 40401 | not found             | 404       |
| 50001 | internal server error | 500       |

## API 文档

### 认证

| 方法 | 路径                 | 说明     | 鉴权 |
| ---- | -------------------- | -------- | ---- |
| POST | `/api/auth/register` | 用户注册 | 否   |
| POST | `/api/auth/login`    | 用户登录 | 否   |

请求示例：

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

### 汇率

| 方法 | 路径                                                                            | 说明                         | 鉴权 |
| ---- | ------------------------------------------------------------------------------- | ---------------------------- | ---- |
| GET  | `/api/rates/latest/all`                                                         | 查询数据库中已记录的汇率列表 | 否   |
| GET  | `/api/rates/latest?base=USD&quote=CNY`                                          | 查询最新币种对汇率           | 否   |
| GET  | `/api/rates/convert?base=USD&quote=CNY&amount=100`                              | 汇率换算                     | 否   |
| GET  | `/api/rates/history?base=USD&quote=CNY&startDate=2026-05-01&endDate=2026-05-31` | 历史汇率查询                 | 是   |
| POST | `/api/rates/create`                                                             | 手动创建汇率记录             | 是   |

创建汇率请求：

```json
{
  "fromCurrency": "USD",
  "toCurrency": "CNY",
  "rate": "7.2500000000",
  "rateDate": "2026-05-30"
}
```

历史汇率响应示例：

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

说明：

- `base` 和 `quote` 必须是 3 位货币代码，后端会统一转成大写。
- `base` 和 `quote` 不能相同。
- `startDate` 和 `endDate` 格式为 `YYYY-MM-DD`。
- 历史查询要求 `startDate <= endDate`。
- 单次历史查询范围限制为不超过 1 年。
- `rate` 使用字符串返回，避免前端 JavaScript number 精度问题。

### 文章

| 方法   | 路径                                | 说明             | 鉴权 |
| ------ | ----------------------------------- | ---------------- | ---- |
| GET    | `/api/articles?page=1&page_size=10` | 分页查询文章列表 | 否   |
| GET    | `/api/articles/:id`                 | 查询文章详情     | 否   |
| GET    | `/api/articles/:id/likes`           | 查询文章点赞数   | 否   |
| POST   | `/api/articles`                     | 创建文章         | 是   |
| PUT    | `/api/articles/:id`                 | 更新文章         | 是   |
| DELETE | `/api/articles/:id`                 | 删除文章         | 是   |
| POST   | `/api/articles/:id/like`            | 点赞或取消点赞   | 是   |

创建 / 更新文章请求：

```json
{
  "title": "hello",
  "content": "article content",
  "preview": "short preview"
}
```

点赞响应示例：

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

### 鉴权方式

需要认证的接口在请求头携带 JWT：

```http
Authorization: Bearer <jwt-token>
```

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

查询最新汇率：

```bash
curl "http://localhost:3000/api/rates/latest?base=USD&quote=CNY"
```

汇率换算：

```bash
curl "http://localhost:3000/api/rates/convert?base=USD&quote=CNY&amount=100"
```

历史汇率：

```bash
curl "http://localhost:3000/api/rates/history?base=USD&quote=CNY&startDate=2026-05-01&endDate=2026-05-31" \
  -H "Authorization: Bearer <jwt-token>"
```

创建文章：

```bash
curl -X POST http://localhost:3000/api/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d "{\"title\":\"hello\",\"content\":\"article content\",\"preview\":\"short preview\"}"
```

点赞 / 取消点赞：

```bash
curl -X POST http://localhost:3000/api/articles/1/like \
  -H "Authorization: Bearer <jwt-token>"
```

## 测试验证

运行测试：

```bash
cd Exchangeapp_backend
go test ./...
```

当前测试覆盖：

- `pkg/jwtauth`：JWT 生成、解析、错误密钥、缺少 Bearer 前缀、过期 token、缺少 `user_id` claim。
- `pkg/passwordbcrypt`：密码哈希、正确密码校验、错误密码拒绝。
- `internal/service`：
  - Redis Lua 点赞 / 取消点赞 toggle。
  - 多用户点赞计数。
  - 最新汇率缓存命中与回源。
  - 坏缓存回源。
  - 汇率换算。
  - 历史汇率查询参数校验。
  - 历史汇率查询对 repository 的调用参数验证。

Redis 相关测试默认连接 `127.0.0.1:6379` 的 DB 15；如果本地没有 Redis，会自动跳过 Redis 集成测试。也可以通过环境变量指定测试 Redis：

```bash
TEST_REDIS_ADDR=127.0.0.1:6379 go test ./internal/service
```

最近一次验证结果：

```text
?    exchangeapp/cmd/migrate        [no test files]
?    exchangeapp/cmd/server         [no test files]
ok   exchangeapp/internal/service
ok   exchangeapp/pkg/jwtauth
ok   exchangeapp/pkg/passwordbcrypt
```

## 压测记录

压测目标：

```text
GET /api/articles?page=1&page_size=10
```

压测场景：

- Docker 环境运行后端、MySQL、Redis。
- 预先创建 100 篇测试文章。
- 先请求一次接口预热 Redis 分页缓存。
- 对缓存命中场景进行并发请求。

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

说明：该结果主要验证 Redis 缓存命中场景下文章分页接口的吞吐和延迟表现。数据库回源场景、缓存失效场景和写操作压测可作为后续补充。

## 后续规划

- 增加定时任务，周期性拉取指定币种对汇率并写入数据库。
- 增加汇率阈值监控和通知能力。
- 接入 Prometheus 指标，观察缓存命中率、外部 API 延迟和错误率。
- 补充 OpenAPI / Swagger 文档。
