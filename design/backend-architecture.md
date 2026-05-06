# Backend Architecture

## 分层设计

```
┌──────────────────────────────────────────────┐
│  app.go — Wails Bind Layer (thin)            │
│  参数转发，不含业务逻辑                         │
├──────────────────────────────────────────────┤
│  pkg/service/ — Business Logic                │
│  account / trade / pnl / rental / inventory   │
│  sync                                        │
├────────────────────────┬─────────────────────┤
│  pkg/orm/              │  pkg/platform/       │
│  GORM DB 抽象          │  HTTP Client        │
│  (NewORM → *gorm.DB)   │  buff/youpin/c5/igxe│
├────────────────────────┴─────────────────────┤
│  pkg/utils/dbfx/ — 纯 SQLite 连接 + 配置       │
│  pkg/model/   — Domain Entities (GORM)       │
└──────────────────────────────────────────────┘
```

## 目录结构

```
cs2-ledger/
├── main.go                       # 入口：组合 Fx Modules，启动 Wails
├── app.go                        # Wails Bind 方法（薄层）
├── go.mod / go.sum
│
├── migrations/                   # SQL 迁移文件（按顺序执行）
│
├── pkg/                          # 可复用的包
│   ├── utils/
│   │   ├── dbfx/                 # SQLite 连接（WAL 模式，单连接），通过 Fx 提供
│   │   └── logfx/                # 结构化日志（slog + tint），含组件级 Logger
│   │
│   ├── model/                    # GORM 实体（纯 struct，无 Fx）
│   │   ├── account.go            # 平台账户
│   │   ├── trade.go              # 交易记录（含自引用 FIFO 匹配）
│   │   ├── inventory.go          # 库存物品（含状态：在库/在售/出租中）
│   │   ├── rental.go             # 出租记录
│   │   └── pnl.go                # 每日盈亏聚合
│   │
│   ├── orm/                      # ORM 抽象层
│   │   ├── orm.go                # NewORM(*sql.DB, MigrationFS) → 打开 GORM + 执行迁移
│   │   ├── interfaces.go         # ORMInterface 组合所有 Repo 接口
│   │   ├── module.go             # Fx Module，提供 NewORM + 所有 Repository
│   │   ├── account.go            # AccountRepo
│   │   ├── trade.go              # TradeRepo（含 FIFO 匹配查询）
│   │   ├── inventory.go          # InventoryRepo（Upsert / Remove / Find + 状态筛选）
│   │   ├── rental.go             # RentalRepo
│   │   └── pnl.go                # PnlRepo（ON CONFLICT 累加聚合）
│   │
│   ├── platform/                 # 平台 HTTP 客户端
│   │   ├── client.go             # Client 接口：Verify / FetchBuyHistory / FetchSellHistory / FetchBalance
│   │   ├── module.go             # Fx Module 占位
│   │   ├── factory/              # ClientFactory：根据平台名创建客户端
│   │   ├── buff/                 # BUFF 163 客户端（Cookie 认证）
│   │   ├── youpin/               # 悠悠有品客户端（Bearer Token 认证）
│   │   ├── c5/                   # C5Game 客户端（API Key 认证，规划中）
│   │   └── igxe/                 # IGXE/ECOsteam 客户端（RSA 签名认证，规划中）
│   │
│   └── service/                  # 业务逻辑层（Module + Service 合并到 service.go）
│       ├── account/service.go    # 账户 CRUD
│       ├── trade/service.go      # 已完成交易视图 + FIFO 盈亏计算
│       ├── pnl/service.go        # 盈亏聚合（ProcessPending 引擎 + 汇总/按月分组）
│       ├── inventory/service.go  # 库存列表 + 物品详情（含租赁汇总）
│       ├── rental/service.go     # 出租记录查询
│       └── sync/engine.go        # 同步引擎（凭证验证 → 并发拉取 → 入库 → P&L）
│
├── hack/                         # 构建 & 校验脚本（规划中）
├── frontend/                     # React 前端
├── design/                       # 架构设计文档
└── .github/workflows/            # CI/CD
```

## 各层职责

### `pkg/utils/dbfx` — SQLite 连接

提供预配置的 SQLite 连接（`*sql.DB`）：WAL 模式、5000ms busy timeout、单连接限制（`SetMaxOpenConns(1)`）。通过 Fx Module 注入给上层。数据库文件路径等配置通过 `dbfx.Config` 结构体控制。

### `pkg/utils/logfx` — 结构化日志

基于 Go 标准库 `log/slog` + `tint` 彩色终端输出。提供 `Logger` 类型和 `WithComponent` 装饰器，允许每个服务拥有带组件标签的子 Logger。通过 Fx Module 提供。

### `pkg/model` — 领域实体

纯 GORM struct，不依赖 Fx，不包含业务逻辑。五个核心实体：
- **Account**：平台账户（name、platform、cookie、余额、状态、同步时间）
- **TradeRecord**：买卖记录（asset_id、item_name、trade_type、价格/费用以分为单位、external_id 用于去重、matched_buy_trade_id 自引用用于 FIFO 匹配）
- **InventoryItem**：库存物品（status 枚举：in_inventory / listed / rented、listed_price）
- **RentalRecord**：出租记录（asset_id、收入、天数、起止时间）
- **PnlDaily**：每日盈亏聚合（trade_count、gross_pl、fee、net_pl，按 account_id + date 唯一）

### `pkg/orm` — ORM 抽象 + Repository

接收 `*sql.DB`（由 dbfx 通过 Fx 注入），用 GORM 包装，在启动时按顺序执行 `migrations/` 目录中的 SQL 文件（通过 `schema_version` 表跟踪）。

每个实体对应一个 Repository，封装参数化 GORM 查询。所有 Repository 通过独立的 Fx provider 注册，服务层通过组合接口 `ORMInterface` 注入。

关键操作：
- **TradeRepo**：UpsertByExternalID（去重写入）、FindUnmatchedSells/Buys（FIFO 待匹配）、SetMatchedBuy（建立买卖关联）
- **InventoryRepo**：Upsert（冲突时更新）、RemoveByAssetID、FindByAccount（可选 status 筛选）
- **PnlRepo**：UpsertDailyPnl（`INSERT ... ON CONFLICT` 累加聚合字段）

### `pkg/platform` — 平台 HTTP 客户端

定义统一的 `Client` 接口：`Verify()`、`FetchBuyHistory(since)`、`FetchSellHistory(since)`、`FetchBalance()`。所有金额以分为单位，时间戳为 unix 毫秒。

`ClientFactory` 根据平台名（"buff" / "youpin" / "c5" / "igxe"）和凭证创建对应的客户端实例。凭证格式因平台而异（Cookie、Token、API Key 等）。

### `pkg/service` — 业务逻辑层

每个 service 是一个包，包含一个 `Service` struct、一个 `NewService` 构造函数、业务方法、以及一个 Fx `Module`。所有代码合并在单个 `service.go` 文件中（不拆 `module.go`）。

**核心服务：**
- **account**：薄 CRUD 包装
- **trade**：`ListCompletedTrades` 通过代码层匹配卖出与其买入，计算每笔交易的盈亏，生成 `CompletedTradeView`；`GetCompletedTradesSummary` 提供汇总
- **pnl**：`ProcessPending` 是核心引擎——获取未匹配卖出 → FIFO 匹配买入（按 asset_id、按时间升序）→ 计算 grossPl/fee/netPl → UPSERT pnl_daily（按日聚合）→ 从 inventory 移除已卖出物品。同时提供 `GetSummary`（全量汇总）和 `GetMonthlyBreakdown`（按月分组）
- **inventory**：`List` 按账户和状态筛选库存；`GetItemDetail` 合并物品信息 + 租赁历史 + 租赁汇总指标
- **rental**：薄查询包装
- **sync**：协调整个拉取管道——加载账户 → 创建平台客户端 → 验证凭证 → 并发拉取 buy/sell/balance → 买入写入 trade_records + 创建 inventory → 卖出写入 trade_records → 调用 PnL ProcessPending → 更新账户余额和同步时间

## Fx 依赖注入

### 模式

每个包通过 `var Module = fx.Module("name", fx.Provide(...))` 声明自己的 provider。Service 层同时依赖 `ORMInterface`（查询）和 `logfx.Logger`（日志）。

顶层 `main.go` 将各模块注册到 `fx.New(...)` 容器：基础设施模块（dbfx、logfx、orm、platform、factory）→ 业务服务模块（account、trade、pnl、rental、inventory、sync）→ `fx.Provide(NewApp)` 提供 Wails 绑定对象 → `fx.Populate` 提取全局 App 实例。

### 模块依赖

```
main.go
  ├── dbfx.Module          → *sql.DB
  ├── logfx.Module         → *Logger
  ├── orm.Module           → ORMInterface (依赖 *sql.DB)
  ├── platform.Module      → 占位
  ├── factory.Module       → ClientFactoryInterface
  ├── account.Module       → *account.Service (依赖 ORMInterface)
  ├── trade.Module         → *trade.Service
  ├── pnl.Module           → *pnl.Service
  ├── rental.Module        → *rental.Service
  ├── inventory.Module     → *inventory.Service
  ├── sync.Module          → *sync.Engine (依赖 ClientFactory + ORMInterface + pnl.Service)
  └── NewApp              → *App (依赖所有 Service + Engine)
```

## Sync Engine 流程

同步单个账户的完整生命周期：

1. 从数据库加载账户信息
2. 通过 ClientFactory 创建对应平台的 HTTP 客户端
3. 调用 `Verify()` 验证凭证有效性，失败则将账户状态标记为 "expired"
4. 并发发起三个请求：FetchBuyHistory(since)、FetchSellHistory(since)、FetchBalance()（使用 `sync.WaitGroup`）
5. 买入记录：写入 trade_records（按 external_id 去重）+ 创建 inventory 条目（状态="in_inventory"）
6. 卖出记录：写入 trade_records（按 external_id 去重）
7. 调用 `pnlSvc.ProcessPending(accountID)` 执行增量 FIFO 匹配和盈亏计算
8. 更新账户的 available_balance、purchase_balance 和 last_sync_at
9. 返回 `SyncResult{NewTrades, NewPnl}`

## Wails Bind Layer

`app.go` 中的 `App` struct 持有所有 Service 和 SyncEngine 的引用（通过构造函数注入）。每个公开方法是一个薄委托：验证输入参数 → 调用对应 Service 方法 → 返回结果。不包含业务逻辑。

暴露的方法覆盖：账户 CRUD（4 个）、同步触发（1 个）、库存查询（2 个）、交易查询（2 个）、盈亏查询（2 个）、仪表盘汇总（1 个）、出租历史（1 个）。

## 错误处理（待实现）

`pkg/errors/` 包计划定义统一的 `AppError` 结构体，包含：机器可读的 `Code`（如 "PLATFORM_AUTH_FAILED"、"SYNC_FAILED"、"NOT_FOUND"）、用户可读的中文 `Message`、以及内部 `Cause`（不暴露给前端）。当前尚未实现，各层直接返回 Go 标准 error。
