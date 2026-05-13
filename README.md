# 分布式票务抢购系统

## 项目文档

---

## 目录

- [一、项目概述](#一项目概述)
- [二、技术栈](#二技术栈)
- [三、项目结构](#三项目结构)
- [四、系统架构](#四系统架构)
- [五、数据库设计](#五数据库设计)
- [六、后端详细设计](#六后端详细设计)
- [七、前端详细设计](#七前端详细设计)
- [八、API 接口文档](#八api-接口文档)
- [九、核心业务流程](#九核心业务流程)
- [十、安全机制](#十安全机制)
- [十一、性能优化](#十一性能优化)
- [十二、部署方案](#十二部署方案)
- [十三、开发指南](#十三开发指南)
- [十四、功能清单](#十四功能清单)

---

## 一、项目概述

本项目是一个**高并发分布式票务抢购系统**，采用前后端分离架构（Go + React），基于 Redis Stream 消息队列实现异步出票，支持活动管理、秒杀抢票、票务转让、二手市场等完整业务链路。

### 核心特性

- **秒杀抢票**：Redis Lua 脚本原子扣减库存，支持万级并发
- **异步出票**：Redis Stream 消息队列异步处理订单，削峰填谷
- **实时通知**：WebSocket 16 分片 Hub，毫秒级推送抢票结果
- **虚拟排队**：高并发时自动进入排队等待页面
- **二手市场**：官方认证的票务转让交易平台
- **多场次管理**：一个活动支持多个时间场次，独立库存
- **数据仪表盘**：管理员实时销售数据、转化漏斗、趋势分析

---

## 页面展示

<!-- 将截图放入 docs/ 目录后取消注释 -->
<!-- ![系统截图](docs/screenshot.png) -->



## 二、技术栈

### 后端

| 技术                           | 版本       | 用途       |
| ---------------------------- | -------- | -------- |
| Go                           | 1.24+    | 主语言      |
| Gin                          | v1.12.0  | HTTP 框架  |
| GORM                         | v1.25.10 | ORM      |
| PostgreSQL                   | 15       | 关系数据库    |
| Redis                        | 7        | 缓存/队列/限流 |
| WebSocket (gorilla)          | v1.5.3   | 实时通信     |
| JWT (golang-jwt)             | v5.2.1   | 认证       |
| OpenTelemetry                | v1.43.0  | 链路追踪     |
| bcrypt (golang.org/x/crypto) | v0.49.0  | 密码加密     |

### 前端

| 技术                 | 版本     | 用途       |
| ------------------ | ------ | -------- |
| React              | 18.2.0 | UI 框架    |
| TypeScript         | 5.3.3  | 类型系统     |
| Vite               | 5.1.0  | 构建工具     |
| Ant Design         | 5.15.0 | UI 组件库   |
| @ant-design/charts | 2.6.7  | 数据可视化    |
| Zustand            | 4.5.0  | 状态管理     |
| React Router       | 6.22.0 | 路由       |
| Axios              | 1.6.7  | HTTP 客户端 |

### 基础设施

| 技术                      | 用途      |
| ----------------------- | ------- |
| Docker + Docker Compose | 容器化部署   |
| Jaeger                  | 链路追踪 UI |
| GitHub Actions          | CI/CD   |

---

## 三、项目结构

```
├── cmd/api/main.go                  # 后端入口
├── internal/
│   ├── config/config.go             # 配置管理
│   ├── handler/                     # HTTP Handler 层 (12 个)
│   │   ├── auth_handler.go          # 认证
│   │   ├── event_handler.go         # 活动管理
│   │   ├── ticket_handler.go        # 票务购买
│   │   ├── ticket_transfer_handler.go # 票务转让
│   │   ├── show_handler.go          # 场次管理
│   │   ├── marketplace_handler.go   # 二手市场
│   │   ├── queue_handler.go         # 排队系统
│   │   ├── waitlist_handler.go      # 等候名单
│   │   ├── promo_code_handler.go    # 促销码
│   │   ├── stats_handler.go         # 数据统计
│   │   ├── health_handler.go        # 健康检查
│   │   └── helpers.go               # 公共工具函数
│   ├── service/                     # 业务逻辑层 (10 个)
│   ├── repository/                  # 数据访问层 (8 个)
│   ├── middleware/                   # 中间件 (5 个)
│   ├── mq/                          # Redis Stream 消息队列
│   ├── queue/                       # 排队/等候名单管理
│   ├── router/router.go             # 路由注册
│   └── pkg/
│       ├── db/db.go                 # 数据库模型+迁移
│       ├── redis/                   # Redis 客户端
│       ├── ws/ws.go                 # WebSocket Hub
│       ├── otel/otel.go             # OpenTelemetry
│       ├── logger/logger.go         # 结构化日志
│       ├── errors/errors.go         # 统一错误类型
│       └── constants/               # 常量定义
├── proto/                           # Proto 定义 (预留，当前未使用)
├── web/                             # React 前端
│   └── src/
│       ├── App.tsx                  # 路由+主题配置
│       ├── api/                     # API 调用层 (12 个)
│       ├── pages/                   # 页面组件 (13 个)
│       ├── components/              # 通用组件 (15 个)
│       ├── stores/                  # Zustand 状态管理
│       ├── theme/                   # 主题配置
│       ├── styles/                  # CSS 变量+全局样式
│       └── types/index.ts           # TypeScript 类型
├── .github/workflows/ci.yml        # CI/CD
├── docker-compose.yml               # Docker 编排
├── Dockerfile                       # 多阶段构建
├── start.bat                        # Windows 一键启动
└── .env                             # 环境配置
```

---

## 四、系统架构

### 架构图

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   React UI  │────▶│  Gin HTTP   │────▶│ PostgreSQL  │
│  (Vite Dev) │     │   Server    │     │    15       │
└─────────────┘     │   :8080     │     └─────────────┘
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
        ┌─────▼─────┐ ┌───▼───┐ ┌─────▼─────┐
        │   Redis   │ │  WS   │ │  Jaeger   │
        │    7      │ │ Hub   │ │  :16686   │
        │ (缓存/队列)│ │:8080  │ │ (追踪UI)  │
        └───────────┘ └───────┘ └───────────┘
```

### 后端分层架构

```
HTTP Request
    │
    ▼
┌──────────────────────────┐
│     Middleware Layer      │  CORS → JWT → RateLimit → RoleAuth
└──────────┬───────────────┘
           ▼
┌──────────────────────────┐
│      Handler Layer       │  参数绑定、响应格式化
└──────────┬───────────────┘
           ▼
┌──────────────────────────┐
│      Service Layer       │  业务逻辑、事务编排
└──────────┬───────────────┘
           ▼
┌──────────────────────────┐
│    Repository Layer      │  数据访问、SQL 查询
└──────────┬───────────────┘
           ▼
┌──────────────────────────┐
│      GORM + PostgreSQL   │  ORM 映射、数据持久化
└──────────────────────────┘
```

### 秒杀架构

```
用户请求 → SeckillRateLimit (10次/秒/用户)
         → SeckillDeduct (Redis Lua 原子扣减)
         → 成功 → Redis Stream (异步消息)
         → TicketConsumer → 创建 Ticket + 扣减 DB 库存
         → WebSocket 推送结果给用户
```

---

## 五、数据库设计

### 5.1 ER 关系图

```
User ──< Ticket ──> Event ──< TicketType
  │        │                      │
  │        ├──> Show              │
  │        │                      │
  │        └──> TicketTransfer    │
  │                               │
  └──> MarketplaceListing ──> Ticket
                               │
Event ──< PromoCode
```

### 5.2 数据模型

#### User（用户表）

| 字段       | 类型     | 约束               | 说明              |
| -------- | ------ | ---------------- | --------------- |
| ID       | uint   | PK, 自增           | 用户 ID           |
| Username | string | UNIQUE, NOT NULL | 用户名             |
| Password | string | NOT NULL         | bcrypt 加密密码     |
| Email    | string | UNIQUE, NOT NULL | 邮箱              |
| Role     | string | DEFAULT 'user'   | 角色 (user/admin) |

#### Event（活动表）

| 字段          | 类型        | 约束                     | 说明       |
| ----------- | --------- | ---------------------- | -------- |
| ID          | uint      | PK                     | 活动 ID    |
| Title       | string    | NOT NULL               | 活动标题     |
| Description | text      |                        | 活动描述     |
| Location    | string    | NOT NULL               | 活动地点     |
| CoverImage  | string    |                        | 封面图片 URL |
| StartTime   | time.Time | NOT NULL, INDEX        | 开始时间     |
| EndTime     | time.Time | NOT NULL, INDEX        | 结束时间     |
| Status      | string    | DEFAULT 'draft', INDEX | 状态       |
| TotalStock  | int       | DEFAULT 0              | 总库存      |

**状态机**: draft → on_sale → off_sale / ended

#### TicketType（票种表）

| 字段         | 类型      | 约束                  | 说明    |
| ---------- | ------- | ------------------- | ----- |
| ID         | uint    | PK                  | 票种 ID |
| EventID    | uint    | NOT NULL, INDEX     | 关联活动  |
| Name       | string  | NOT NULL            | 票种名称  |
| Price      | float64 | NOT NULL            | 价格    |
| Stock      | int     | NOT NULL, DEFAULT 0 | 库存    |
| MaxPerUser | int     | NOT NULL, DEFAULT 1 | 每人限购  |
| SortOrder  | int     | DEFAULT 0           | 排序    |

#### Ticket（票务/订单表）

| 字段             | 类型      | 约束                                  | 说明     |
| -------------- | ------- | ----------------------------------- | ------ |
| ID             | uint    | PK                                  | 票务 ID  |
| UserID         | uint    | NOT NULL, INDEX                     | 购票用户   |
| EventID        | uint    | NOT NULL, INDEX                     | 关联活动   |
| ShowID         | uint    | INDEX                               | 关联场次   |
| TicketTypeID   | uint    | NOT NULL, INDEX                     | 关联票种   |
| OrderNo        | string  | UNIQUE INDEX                        | 订单号    |
| Quantity       | int     | NOT NULL, DEFAULT 1                 | 数量     |
| TotalPrice     | float64 |                                     | 总价     |
| Status         | string  | NOT NULL, DEFAULT 'reserved', INDEX | 状态     |
| QRCode         | text    |                                     | 电子票二维码 |
| DiscountCode   | string  | INDEX                               | 优惠码    |
| RealName       | string  | INDEX                               | 实名姓名   |
| IDCard         | string  | INDEX                               | 身份证号   |
| Phone          | string  | INDEX                               | 手机号    |
| TransferredTo  | uint    | INDEX                               | 转让目标用户 |
| TransferStatus | string  | DEFAULT 'none'                      | 转让状态   |

**状态机**: reserved → paid → used / expired / cancelled

#### Show（场次表）

| 字段        | 类型        | 约束                     | 说明    |
| --------- | --------- | ---------------------- | ----- |
| ID        | uint      | PK                     | 场次 ID |
| EventID   | uint      | NOT NULL, INDEX        | 关联活动  |
| Name      | string    | NOT NULL               | 场次名称  |
| ShowTime  | time.Time | NOT NULL, INDEX        | 开始时间  |
| EndTime   | time.Time | NOT NULL               | 结束时间  |
| Status    | string    | DEFAULT 'draft', INDEX | 状态    |
| Stock     | int       | NOT NULL, DEFAULT 0    | 库存    |
| SoldCount | int       | DEFAULT 0              | 已售数量  |
| SortOrder | int       | DEFAULT 0              | 排序    |

#### TicketTransfer（票务转让表）

| 字段           | 类型         | 约束                                 | 说明                    |
| ------------ | ---------- | ---------------------------------- | --------------------- |
| ID           | uint       | PK                                 | 转让 ID                 |
| TicketID     | uint       | NOT NULL, INDEX                    | 关联票务                  |
| FromUserID   | uint       | NOT NULL, INDEX                    | 转让方                   |
| ToUserID     | uint       | NOT NULL, INDEX                    | 接收方                   |
| Status       | string     | NOT NULL, DEFAULT 'pending', INDEX | 状态                    |
| TransferType | string     | NOT NULL, DEFAULT 'gift', INDEX    | 类型 (gift/marketplace) |
| Price        | float64    |                                    | 交易价格                  |
| Reason       | string     |                                    | 转让原因                  |
| ReviewedBy   | uint       |                                    | 审核人                   |
| ReviewedAt   | *time.Time |                                    | 审核时间                  |

#### MarketplaceListing（二手市场表）

| 字段          | 类型      | 约束                                | 说明    |
| ----------- | ------- | --------------------------------- | ----- |
| ID          | uint    | PK                                | 上架 ID |
| TicketID    | uint    | NOT NULL, INDEX                   | 关联票务  |
| SellerID    | uint    | NOT NULL, INDEX                   | 卖家    |
| Price       | float64 | NOT NULL                          | 出售价格  |
| Status      | string  | NOT NULL, DEFAULT 'active', INDEX | 状态    |
| BuyerID     | uint    |                                   | 买家    |
| Description | text    |                                   | 商品描述  |

#### PromoCode（促销码表）

| 字段            | 类型        | 约束           | 说明                 |
| ------------- | --------- | ------------ | ------------------ |
| ID            | uint      | PK           | 促销码 ID             |
| Code          | string    | UNIQUE INDEX | 促销码                |
| EventID       | uint      | INDEX        | 关联活动               |
| DiscountType  | string    | NOT NULL     | 类型 (percent/fixed) |
| DiscountValue | float64   | NOT NULL     | 折扣值                |
| MinAmount     | float64   | DEFAULT 0    | 最低消费               |
| MaxUses       | int       | DEFAULT 0    | 最大使用次数 (0=无限)      |
| UsedCount     | int       | DEFAULT 0    | 已使用次数              |
| StartTime     | time.Time |              | 生效时间               |
| EndTime       | time.Time |              | 过期时间               |
| IsActive      | bool      | DEFAULT true | 是否启用               |

### 5.3 复合索引

```sql
CREATE INDEX idx_tickets_user_status ON tickets(user_id, status);
CREATE INDEX idx_tickets_event_status ON tickets(event_id, status);
CREATE INDEX idx_tickets_status_created ON tickets(status, created_at);
CREATE INDEX idx_marketplace_status_created ON marketplace_listings(status, created_at);
```

---

## 六、后端详细设计

### 6.1 Handler 层

| Handler               | 文件                         | 方法                                                                                                                                                          |
| --------------------- | -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| AuthHandler           | auth_handler.go            | Register, Login, GetProfile                                                                                                                                 |
| EventHandler          | event_handler.go           | CreateEvent, UpdateEvent, GetEvent, ListEvents, PublishEvent, UnpublishEvent, EndEvent, CreateTicketType, UpdateTicketType, DeleteTicketType, GetEventStock |
| TicketHandler         | ticket_handler.go          | PurchaseTicket, GetMyTickets, GetTicketDetail, PayTicket, CancelTicket, UseTicket                                                                           |
| ShowHandler           | show_handler.go            | CreateShow, UpdateShow, DeleteShow, PublishShow, UnpublishShow, ListShows, GetShow                                                                          |
| MarketplaceHandler    | marketplace_handler.go     | CreateListing, BuyListing, CancelListing, GetListing, ListActive, ListByEvent, ListMyListings, ListMyPurchases                                              |
| TicketTransferHandler | ticket_transfer_handler.go | RequestTransfer, DirectGift, GetTransferHistory, ApproveTransfer, RejectTransfer, GetPendingTransfers                                                       |
| QueueHandler          | queue_handler.go           | JoinQueue, GetPosition, LeaveQueue                                                                                                                          |
| WaitlistHandler       | waitlist_handler.go        | JoinWaitlist, GetWaitlistPosition, LeaveWaitlist                                                                                                            |
| PromoCodeHandler      | promo_code_handler.go      | CreatePromoCode, ValidatePromoCode, GetPromoCodes, DeletePromoCode                                                                                          |
| StatsHandler          | stats_handler.go           | GetDashboardStats, GetSalesTrend, GetTicketTypeStats, GetConversionFunnel                                                                                   |
| HealthHandler         | health_handler.go          | HealthCheck                                                                                                                                                 |

### 6.2 Service 层

| Service               | 文件                         | 核心职责                             |
| --------------------- | -------------------------- | -------------------------------- |
| AuthService           | auth_service.go            | 注册(bcrypt)、登录(JWT HS256)、获取用户信息  |
| EventService          | event_service.go           | 活动 CRUD、票种管理、发布时初始化 Redis 库存     |
| TicketService         | ticket_service.go          | 秒杀购票(Redis Lua 扣减 + MQ)、支付、取消、使用 |
| ShowService           | show_service.go            | 场次 CRUD、库存管理                     |
| MarketplaceService    | marketplace_service.go     | 上架、购买(所有权转移)、下架、列表查询             |
| TicketTransferService | ticket_transfer_service.go | 转让申请、直接转赠、审核、历史记录                |
| PromoCodeService      | promo_code_service.go      | 促销码创建、验证、折扣计算                    |
| StatsService          | stats_service.go           | 仪表盘统计、销售趋势、票种分布、转化漏斗             |
| TicketExpireChecker   | ticket_expire_service.go   | 定时检查过期票务，自动回收库存                  |

### 6.3 Repository 层

所有 Repository 遵循接口/实现模式，通过构造函数注入依赖：

```go
type TicketRepository interface {
    Create(ticket *db.Ticket) error
    FindByID(id uint) (*db.Ticket, error)
    FindByUserID(userID uint, page, limit int) ([]db.Ticket, int64, error)
    FindExpiredReserved(olderThan time.Time, limit int) ([]db.Ticket, error)
    UpdateStatus(id uint, status string) error
    UpdateOwner(id uint, newUserID uint) error
    // ...
}
```

### 6.4 中间件

| 中间件                        | 文件                   | 功能                          |
| -------------------------- | -------------------- | --------------------------- |
| JWTAuthWithBlacklist       | auth.go              | JWT 验证 + Redis 黑名单检查        |
| RoleAuth                   | auth.go              | 角色鉴权 (admin)                |
| RateLimitMiddleware        | ratelimit.go         | IP 级限流 (100次/分钟, Redis Lua) |
| SeckillRateLimitMiddleware | seckill_ratelimit.go | 用户级秒杀限流 (10次/秒)             |
| ErrorHandler               | error_handler.go     | 统一错误处理                      |
| RecoveryMiddleware         | otel.go              | Panic 恢复 + OTEL 追踪          |

### 6.5 消息队列 (Redis Streams)

```go
// 生产者
type TicketProducer struct {
    redis *pkgredis.RedisClientWrapper
}

func (p *TicketProducer) PublishTicket(ctx context.Context, msg TicketMessage) error {
    // 发布到 Redis Stream "ticket:orders"
}

// 消费者
type TicketConsumer struct {
    redis   *pkgredis.RedisClientWrapper
    db      *gorm.DB
    wsHub   *ws.Hub
}

func (c *TicketConsumer) processTicket(ctx context.Context, msg redis.XMessage) {
    // 1. 查询票种信息
    // 2. 原子扣减数据库库存 (AtomicDeductStock)
    // 3. 创建 Ticket 记录
    // 4. WebSocket 推送结果
}
```

### 6.6 WebSocket Hub

16 分片 Hub，按 user ID 哈希分配：

```go
type Hub struct {
    shards         []*shard          // 16 个分片
    broadcast      chan *BroadcastMsg
    jwtSecret      []byte
    redisClient    *redis.Client     // 用于黑名单检查
    allowedOrigins map[string]bool
}
```

支持功能：

- JWT 认证 (query param)
- Redis 黑名单检查
- 房间广播 (room-based)
- 按用户推送 (SendToUser)
- Ping/Pong 保活 (60s)

### 6.7 Redis Lua 脚本

#### 秒杀扣减 (SeckillDeduct)

```lua
-- 原子操作：检查库存 → 检查重复购买 → 扣减库存 → 记录用户
local stock = tonumber(redis.call('HGET', KEYS[1], ARGV[1]))
if stock <= 0 then return -1 end       -- 售罄
if redis.call('SISMEMBER', KEYS[2], ARGV[2]) == 1 then return -2 end  -- 重复
redis.call('HINCRBY', KEYS[1], ARGV[1], -1)
redis.call('SADD', KEYS[2], ARGV[2])
return 1
```

#### 限流 (RateLimit)

```lua
-- 原子操作：检查计数 → 递增 → 设置过期
local current = tonumber(redis.call('GET', KEYS[1]) or '0')
if current >= limit then return 0 end
current = redis.call('INCR', KEYS[1])
if current == 1 then redis.call('EXPIRE', KEYS[1], window) end
return current <= limit and 1 or 0
```

---

## 七、前端详细设计

### 7.1 路由配置

| 路径                  | 组件              | 权限      | 说明    |
| ------------------- | --------------- | ------- | ----- |
| `/login`            | Login           | 公开      | 登录页   |
| `/register`         | Register        | 公开      | 注册页   |
| `/`                 | Dashboard       | 需登录     | 用户仪表盘 |
| `/events`           | Events          | 需登录     | 活动列表  |
| `/events/:id`       | EventDetail     | 需登录     | 活动详情  |
| `/marketplace`      | Marketplace     | 需登录     | 二手市场  |
| `/tickets`          | Tickets         | 需登录     | 我的票务  |
| `/transfer-records` | TransferRecords | 需登录     | 转让记录  |
| `/profile`          | Profile         | 需登录     | 个人中心  |
| `/admin/dashboard`  | AdminDashboard  | 需 admin | 管理仪表盘 |
| `/admin/events`     | AdminEvents     | 需 admin | 活动管理  |

### 7.2 组件清单

| 组件                 | 文件                     | 说明                     |
| ------------------ | ---------------------- | ---------------------- |
| AppLayout          | AppLayout.tsx          | 全局布局 (侧边栏+Header)      |
| BrandLogo          | BrandLogo.tsx          | SVG 品牌 Logo            |
| SkeletonCard       | SkeletonCard.tsx       | 骨架屏加载 (card/list/stat) |
| PageTransition     | PageTransition.tsx     | 页面进入动画                 |
| Countdown          | Countdown.tsx          | 倒计时组件                  |
| ThemeToggle        | ThemeToggle.tsx        | 亮色/暗色切换                |
| NotificationBell   | NotificationBell.tsx   | 通知铃铛                   |
| ProtectedRoute     | ProtectedRoute.tsx     | 路由守卫                   |
| ErrorBoundary      | ErrorBoundary.tsx      | 全局错误边界                 |
| RouteErrorBoundary | RouteErrorBoundary.tsx | 路由级错误边界                |
| TransferButton     | TransferButton.tsx     | 转让按钮 (转赠/申请)           |
| ShareButton        | ShareButton.tsx        | 分享按钮                   |
| PromoCodeInput     | PromoCodeInput.tsx     | 促销码输入                  |
| QueueWaiting       | QueueWaiting.tsx       | 排队等待展示                 |
| WaitlistButton     | WaitlistButton.tsx     | 等候名单按钮                 |

### 7.3 状态管理 (Zustand)

#### authStore

```typescript
interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  loading: boolean
  login: (username: string, password: string) => Promise<void>
  register: (username: string, password: string, email: string) => Promise<void>
  logout: () => void
  fetchProfile: () => Promise<void>
}
```

#### notificationStore

```typescript
interface NotificationState {
  notifications: Notification[]
  unreadCount: number
  addNotification: (n: Notification) => void
  markAsRead: (id: string) => void
  markAllAsRead: () => void
  clearAll: () => void
}
```

### 7.4 主题系统

双层主题架构：

1. **CSS 变量** (`src/styles/variables.css`)：自定义元素的亮色/暗色适配
2. **Ant Design Token** (`src/theme/index.ts`)：组件级主题配置

品牌色系：

- 主色：`#5B2FE8` (深紫) / 暗色 `#8B6FFF`
- 金色：`#D4A843` / 暗色 `#E6BC5C`
- 成功：`#22C55E`
- 警告：`#F59E0B`
- 错误：`#EF4444`

活动封面渐变系统 (8 种)：

```typescript
const eventGradients = [
  'linear-gradient(135deg, #5B2FE8 0%, #8B6FFF 100%)',   // 紫色
  'linear-gradient(135deg, #D4A843 0%, #F5C862 100%)',   // 金色
  'linear-gradient(135deg, #6366F1 0%, #818CF8 100%)',   // 靛蓝
  'linear-gradient(135deg, #EC4899 0%, #F472B6 100%)',   // 粉色
  'linear-gradient(135deg, #14B8A6 0%, #2DD4BF 100%)',   // 青色
  'linear-gradient(135deg, #F97316 0%, #FB923C 100%)',   // 橙色
  'linear-gradient(135deg, #8B5CF6 0%, #A78BFA 100%)',   // 紫罗兰
  'linear-gradient(135deg, #06B6D4 0%, #22D3EE 100%)',   // 天蓝
]
// 按 eventId % 8 自动分配
```

---

## 八、API 接口文档

### 8.1 公开接口

| 方法   | 路径          | 说明   | 限流        |
| ---- | ----------- | ---- | --------- |
| GET  | `/health`   | 健康检查 | 无         |
| POST | `/register` | 用户注册 | IP 100次/分 |
| POST | `/login`    | 用户登录 | IP 100次/分 |

### 8.2 用户接口 (`/api`)

| 方法                   | 路径                                 | 说明         |
| -------------------- | ---------------------------------- | ---------- |
| GET                  | `/api/profile`                     | 获取个人信息     |
| **活动**               |                                    |            |
| GET                  | `/api/events`                      | 活动列表 (分页)  |
| GET                  | `/api/events/:id`                  | 活动详情       |
| GET                  | `/api/events/:id/stock`            | 实时库存       |
| GET                  | `/api/events/:id/shows`            | 活动场次列表     |
| GET                  | `/api/shows/:id`                   | 场次详情       |
| **排队**               |                                    |            |
| POST                 | `/api/queue/:event_id/join`        | 加入排队       |
| GET                  | `/api/queue/:event_id/position`    | 排队位置       |
| POST                 | `/api/queue/:event_id/leave`       | 离开队列       |
| **等候名单**             |                                    |            |
| POST                 | `/api/waitlist/:event_id/join`     | 加入等候       |
| GET                  | `/api/waitlist/:event_id/position` | 等候位置       |
| POST                 | `/api/waitlist/:event_id/leave`    | 离开等候       |
| **促销码**              |                                    |            |
| POST                 | `/api/promo/validate`              | 验证促销码      |
| GET                  | `/api/promo/:event_id`             | 活动促销码      |
| **票务** (秒杀限流: 10次/秒) |                                    |            |
| POST                 | `/api/tickets/purchase`            | 购票 (秒杀)    |
| GET                  | `/api/tickets`                     | 我的票务       |
| GET                  | `/api/tickets/:id`                 | 票务详情       |
| POST                 | `/api/tickets/:id/pay`             | 支付         |
| POST                 | `/api/tickets/:id/cancel`          | 取消         |
| POST                 | `/api/tickets/:id/use`             | 使用         |
| **转让**               |                                    |            |
| POST                 | `/api/transfer`                    | 申请转让 (需审核) |
| POST                 | `/api/transfer/gift`               | 直接转赠       |
| GET                  | `/api/transfer/history`            | 转让记录       |
| **二手市场**             |                                    |            |
| GET                  | `/api/marketplace`                 | 在售列表       |
| GET                  | `/api/marketplace/my`              | 我的上架       |
| GET                  | `/api/marketplace/purchases`       | 我的购买       |
| GET                  | `/api/marketplace/event/:id`       | 按活动查看      |
| GET                  | `/api/marketplace/:id`             | 商品详情       |
| POST                 | `/api/marketplace`                 | 上架         |
| POST                 | `/api/marketplace/:id/buy`         | 购买         |
| POST                 | `/api/marketplace/:id/cancel`      | 下架         |

### 8.3 管理员接口 (`/admin`)

| 方法       | 路径                                  | 说明    |
| -------- | ----------------------------------- | ----- |
| **活动管理** |                                     |       |
| POST     | `/admin/events`                     | 创建活动  |
| PUT      | `/admin/events/:id`                 | 更新活动  |
| POST     | `/admin/events/:id/publish`         | 发布    |
| POST     | `/admin/events/:id/unpublish`       | 下架    |
| POST     | `/admin/events/:id/end`             | 结束    |
| POST     | `/admin/events/:id/ticket-types`    | 创建票种  |
| PUT      | `/admin/events/ticket-types/:id`    | 更新票种  |
| DELETE   | `/admin/events/ticket-types/:id`    | 删除票种  |
| **场次管理** |                                     |       |
| POST     | `/admin/events/:id/shows`           | 创建场次  |
| PUT      | `/admin/events/shows/:id`           | 更新场次  |
| DELETE   | `/admin/events/shows/:id`           | 删除场次  |
| POST     | `/admin/events/shows/:id/publish`   | 上架场次  |
| POST     | `/admin/events/shows/:id/unpublish` | 下架场次  |
| **促销码**  |                                     |       |
| POST     | `/admin/promo`                      | 创建促销码 |
| DELETE   | `/admin/promo/:id`                  | 删除促销码 |
| **统计**   |                                     |       |
| GET      | `/admin/stats/dashboard`            | 仪表盘数据 |
| GET      | `/admin/stats/sales-trend`          | 销售趋势  |
| GET      | `/admin/stats/ticket-types`         | 票种统计  |
| GET      | `/admin/stats/funnel/:event_id`     | 转化漏斗  |
| **转让审核** |                                     |       |
| GET      | `/admin/transfer/pending`           | 待审核列表 |
| POST     | `/admin/transfer/:id/approve`       | 批准    |
| POST     | `/admin/transfer/:id/reject`        | 拒绝    |

### 8.4 WebSocket

```
ws://localhost:8080/ws?token=<JWT>&room_id=<可选>
```

消息格式：

```json
{
  "type": "ticket_result",
  "payload": {
    "ticket_id": 1,
    "user_id": 1,
    "status": "success",
    "message": "购票成功！",
    "order_no": "TK20260513120000",
    "ticket_type": "VIP",
    "timestamp": 1778668646
  }
}
```

---

## 九、核心业务流程

### 9.1 秒杀抢票流程

```
1. 用户点击"抢购"
2. SeckillRateLimitMiddleware 检查用户请求频率 (10次/秒)
3. TicketService.PurchaseTicket:
   a. 验证活动状态 == on_sale
   b. 验证票种属于该活动
   c. 验证未超出每用户限购
   d. Redis Lua 脚本原子操作:
      - 检查库存 > 0
      - 检查未重复购买
      - HINCRBY 扣减库存
      - SADD 记录已购买用户
   e. 发布消息到 Redis Stream "ticket:orders"
4. TicketConsumer 异步消费:
   a. 查询票种信息
   b. AtomicDeductStock 原子扣减 DB 库存
   c. 创建 Ticket 记录 (status: reserved)
   d. WebSocket 推送结果给用户
5. 用户收到通知，跳转到"我的票务"
6. 30 分钟内未支付，票务自动过期，库存回滚
```

### 9.2 票务转让流程

#### 直接转赠 (gift)

```
1. 用户选择票务，输入目标用户 ID
2. 调用 POST /api/transfer/gift
3. 验证:
   - 票务存在且属于当前用户
   - 票务状态 == paid
   - 目标用户存在
   - 不能转赠给自己
4. 直接更新 Ticket.UserID = 目标用户
5. 创建 TicketTransfer 记录 (status: approved, type: gift)
6. 完成
```

#### 申请转让 (review)

```
1. 用户选择票务，输入目标用户 ID 和原因
2. 调用 POST /api/transfer
3. 创建 TicketTransfer 记录 (status: pending, type: review)
4. 管理员在 /admin/transfer/pending 查看待审核列表
5. 管理员 approve/reject
6. 批准后更新 Ticket.UserID
```

### 9.3 二手市场流程

```
1. 卖家选择已支付的票务，设置价格，点击"上架"
2. 调用 POST /api/marketplace
3. 验证:
   - 票务属于当前用户且已支付
   - 价格 > 0
   - 该票务未在市场中
4. 创建 MarketplaceListing (status: active)
5. 买家在二手市场浏览，点击"购买"
6. 调用 POST /api/marketplace/:id/buy
7. Ticket.UpdateOwner 原子更新所有权
8. MarketplaceListing.Status = "sold"
```

### 9.4 票务过期处理

```
1. StartTicketExpireChecker 每分钟执行
2. 查询: status = "reserved" AND created_at < now - 30min
3. 分页循环处理所有过期票务:
   a. 更新 Ticket.Status = "expired"
   b. 回滚 TicketType.Stock
   c. 回滚 Redis 秒杀库存 (SeckillRollback)
4. WebSocket 通知用户 (如有连接)
```

---

## 十、安全机制

### 10.1 认证授权

- **JWT HS256**：Token 有效期 24 小时
- **Redis 黑名单**：登出时 Token 加入黑名单
- **角色鉴权**：admin 角色才能访问 `/admin/*` 接口
- **WebSocket 认证**：连接时验证 JWT，检查黑名单

### 10.2 密码安全

- **bcrypt** 哈希加密，salt rounds = 默认
- 注册要求：用户名 3-32 位，密码 8-72 位，邮箱格式验证

### 10.3 限流防护

| 限流类型  | 策略        | 实现                           |
| ----- | --------- | ---------------------------- |
| IP 限流 | 100 次/分钟  | Redis INCR + EXPIRE (Lua 原子) |
| 秒杀限流  | 10 次/秒/用户 | Redis Lua 脚本                 |
| 库存感知  | 库存为 0 时拒绝 | Redis Hash 查询                |

### 10.4 CORS 配置

```go
cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
})
```

### 10.5 防黄牛机制

- 每用户每票种限购 (MaxPerUser)
- Redis SET 记录已购买用户，防止重复购买
- 秒杀限流防止脚本刷单
- 实名制字段 (RealName, IDCard, Phone)

---

## 十一、性能优化

### 11.1 Redis 优化

- **Lua 脚本原子操作**：秒杀扣减、限流计数
- **Pipeline 批量操作**：减少网络往返
- **16 分片 WebSocket Hub**：分散锁竞争
- **复合索引**：避免全表扫描

### 11.2 N+1 查询修复

批量加载替代逐条查询：

```go
// 修复前：每条记录 2 次查询，20 条 = 40 次查询
for _, listing := range listings {
    ticket, _ := ticketRepo.FindByID(listing.TicketID)
    tt, _ := ticketTypeRepo.FindByID(ticket.TicketTypeID)
}

// 修复后：2 次批量查询
ticketIDs := collectUniqueIDs(listings)
tickets, _ := ticketRepo.FindByIDs(ticketIDs)
ttIDs := collectUniqueIDs(tickets)
ticketTypes, _ := ticketTypeRepo.FindByIDs(ttIDs)
```

### 11.3 原子库存扣减

```go
// 修复前：检查+更新非原子，可能超卖
if ticketType.Stock >= quantity {
    ticketTypeRepo.UpdateStock(id, -quantity)
}

// 修复后：WHERE stock >= qty 保证原子性
func AtomicDeductStock(id uint, quantity int) error {
    result := db.Where("id = ? AND stock >= ?", id, quantity).
        Update("stock", gorm.Expr("stock - ?", quantity))
    if result.RowsAffected == 0 {
        return fmt.Errorf("库存不足")
    }
}
```

### 11.4 后台任务优雅停机

```go
// 修复前：context.Background() 无法取消
go ticketConsumer.Start(context.Background())

// 修复后：传入可取消 context
ctx, cancel := context.WithCancel(context.Background())
go ticketConsumer.Start(ctx)
// 关机时 cancel() 触发优雅退出
```

---

## 十二、部署方案

### 12.1 Docker Compose

```yaml
services:
  app:        # Go 后端 (端口 8080)
  postgres:   # PostgreSQL 15 (端口 5432)
  redis:      # Redis 7 (端口 6379)
  jaeger:     # Jaeger UI (端口 16686)
```

### 12.2 Dockerfile (多阶段构建)

```dockerfile
# 构建阶段
FROM golang:1.24-alpine AS builder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/order-system ./cmd/api/main.go

# 运行阶段
FROM alpine:3.19
RUN apk add ca-certificates tzdata
COPY --from=builder /app/order-system .
EXPOSE 8080 9090
ENTRYPOINT ["./order-system"]
```

### 12.3 环境变量

| 变量                 | 默认值          | 说明            |
| ------------------ | ------------ | ------------- |
| APP_ENV            | production   | 环境            |
| APP_PORT           | 8080         | HTTP 端口       |
| DB_HOST            | postgres     | 数据库主机         |
| DB_PORT            | 5432         | 数据库端口         |
| DB_USER            | postgres     | 数据库用户         |
| DB_PASSWORD        | (必填)         | 数据库密码         |
| DB_NAME            | order_system | 数据库名          |
| REDIS_HOST         | redis        | Redis 主机      |
| REDIS_PORT         | 6379         | Redis 端口      |
| JWT_SECRET         | (必填)         | JWT 密钥        |
| JWT_EXPIRE         | 86400        | Token 过期时间(秒) |
| OTEL_EXPORTER_TYPE | otlp         | 追踪导出器类型       |

### 12.4 一键启动

```bash
# Windows
start.bat

# Docker Compose
docker-compose up -d
```

---

## 十三、开发指南

### 13.1 本地开发

```bash
# 后端
cd internal/
go mod tidy
go run ./cmd/api/main.go

# 前端
cd web/
npm install
npm run dev
```

### 13.2 CI/CD (GitHub Actions)

```yaml
jobs:
  backend:     # go vet → go build → go test -race -cover
  frontend:    # npm ci → tsc --noEmit → npm run build
  docker:      # docker build
```

### 13.3 项目规范

- **提交格式**: `<type>: <description>` (feat/fix/refactor/docs/test/chore)
- **分支策略**: main/master 为稳定分支
- **代码审查**: 所有变更需经过审查
- **测试覆盖**: 最低 80%

---

## 十四、功能清单

### 已实现功能

| 阶段       | 功能                   | 状态  |
| -------- | -------------------- | --- |
| **基础**   | 用户注册/登录              | ✅   |
|          | JWT 认证 + 黑名单         | ✅   |
|          | 角色鉴权 (admin/user)    | ✅   |
|          | 活动 CRUD              | ✅   |
|          | 票种管理                 | ✅   |
|          | 秒杀抢票 (Redis Lua)     | ✅   |
|          | 异步出票 (Redis Stream)  | ✅   |
|          | WebSocket 实时通知       | ✅   |
|          | 票务支付/取消/使用           | ✅   |
|          | 票务过期自动回收             | ✅   |
| **第一阶段** | 排队系统 (Redis List)    | ✅   |
|          | 倒计时功能                | ✅   |
|          | 等候名单                 | ✅   |
| **第二阶段** | 促销码系统                | ✅   |
|          | 邮件通知模板               | ✅   |
|          | 社交分享按钮               | ✅   |
| **第三阶段** | 管理员数据仪表盘             | ✅   |
|          | 实时销售统计               | ✅   |
|          | 转化漏斗分析               | ✅   |
|          | 防黄牛增强 (限流+限购)        | ✅   |
|          | 票务转让审核               | ✅   |
| **第四阶段** | 多场次管理                | ✅   |
|          | 场次选择界面               | ✅   |
|          | 直接转赠好友               | ✅   |
|          | 二手交易市场               | ✅   |
|          | 转让记录追踪               | ✅   |
| **安全加固** | CORS 配置              | ✅   |
|          | 输入校验增强               | ✅   |
|          | WebSocket 黑名单检查      | ✅   |
|          | 密钥不硬编码               | ✅   |
| **性能优化** | N+1 查询修复             | ✅   |
|          | 原子库存扣减               | ✅   |
|          | 复合索引                 | ✅   |
|          | 限流 Lua 脚本            | ✅   |
|          | 优雅停机                 | ✅   |
| **代码质量** | getUser 公共函数         | ✅   |
|          | 参数校验统一               | ✅   |
|          | 魔法数字常量化              | ✅   |
| **前端美化** | CSS 变量体系             | ✅   |
|          | 双模式主题系统              | ✅   |
|          | 品牌 Logo              | ✅   |
|          | 骨架屏加载                | ✅   |
|          | 页面过渡动画               | ✅   |
|          | 活动封面渐变系统             | ✅   |
| **基础设施** | Docker 多阶段构建         | ✅   |
|          | Docker Compose 编排    | ✅   |
|          | GitHub Actions CI/CD | ✅   |
|          | OpenTelemetry 追踪     | ✅   |

---

*文档生成时间: 2026-05-13*
*项目版本: v1.0*
>>>>>>> master
