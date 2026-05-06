# Architecture Overview

## System Positioning

CS2 Ledger 是本地优先的个人 CS2 饰品交易终端。用户提供平台凭证，应用以用户身份拉取交易数据，全部存入本地 SQLite，计算盈亏。

**不做行情看板，不做中心化服务。库存通过 `inventory` 表追踪（含在库/在售/出租中三种状态）。**

```
┌──────────────────────────────────────────────────────────┐
│                    Wails Desktop Shell                    │
│                                                          │
│  ┌────────────────────┐   ┌───────────────────────────┐  │
│  │   React Frontend   │   │       Go Backend           │  │
│  │   (TypeScript)     │   │                            │  │
│  │                    │   │  ┌───────────────────────┐ │  │
│  │  React Router      │   │  │  app.go (Wails Bind)  │ │  │
│  │  TanStack Query ◄──┤   │  ├───────────────────────┤ │  │
│  │  Zustand           │   │  │  pkg/service          │ │  │
│  │  MUI v5            │   │  ├───────────────────────┤ │  │
│  │  Emotion           │   │  │  pkg/platform          │ │  │
│  │                    │   │  ├───────────────────────┤ │  │
│  └────────────────────┘   │  │  pkg/orm (GORM)        │ │  │
│                           │  ├───────────────────────┤ │  │
│                           │  │  SQLite (WAL)          │ │  │
│                           │  └───────────────────────┘ │  │
│                           └───────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

## Technology Stack

| Layer        | Technology                   | Rationale                          |
| ------------ | ---------------------------- | ---------------------------------- |
| Desktop Shell| Wails v2                     | Go + web tech, single binary       |
| Frontend     | React 19 + TypeScript        | Component model, type safety       |
| Build        | Vite                         | Fast HMR, native ESM               |
| Styling      | MUI v5 (Material UI)         | Component library, theme-based       |
| Server State | TanStack Query               | Cache, refetch, stale management   |
| Client State | Zustand                      | Simple, no boilerplate             |
| Routing      | React Router v7              | Standard SPA routing               |
| Backend      | Go **1.25**                  | Wails v2 + Uber Fx                |
| DI           | Uber Fx (per-package Module) | 模块随包定义，main.go 组装         |
| ORM          | GORM                         | 在 pkg/orm 中封装，通过 Fx 提供    |
| Database     | SQLite (WAL mode)            | pkg/utils/dbfx 提供底层连接        |
| HTTP Client  | Go net/http                  | Cookie/Token-based platform auth   |
| CI/CD        | GitHub Actions               | Lint + test + build                |
| Git Hooks    | lefthook                     | pre-commit: golangci-lint, format  |
| Linting      | golangci-lint                | Go code quality                    |

## Project Directory Structure

```
cs2-ledger/
├── main.go                   # 入口：组装 Fx Modules，启动 Wails
├── app.go                    # Wails Bind methods (thin delegates)
├── go.mod / go.sum
│
├── migrations/               # SQL 迁移文件（按时间戳排序）
│
├── pkg/                      # 业务包
│   ├── utils/
│   │   ├── dbfx/              # 纯 SQLite 连接 + 配置（与业务无关）
│   │   │   └── sqlite.go     # Open(), WAL, busy_timeout 等
│   │   └── logfx/             # 结构化日志
│   │
│   ├── model/                # GORM entities（纯 struct，无 Fx）
│   │   ├── account.go
│   │   ├── trade.go
│   │   ├── rental.go
│   │   ├── pnl.go
│   │   └── inventory.go
│   │
│   ├── orm/                  # ORM 抽象 + Repository + Fx Module
│   │   ├── orm.go            # NewORM(cfg) (*gorm.DB, error)
│   │   ├── module.go         # var Module = fx.Module("orm", fx.Provide(NewORM, repos...))
│   │   ├── account.go        # AccountRepo
│   │   ├── trade.go          # TradeRepo
│   │   ├── rental.go         # RentalRepo
│   │   ├── pnl.go            # PnlRepo
│   │   └── inventory.go      # InventoryRepo
│   │
│   ├── platform/             # HTTP clients + Fx Module
│   │   ├── client.go         # Client 接口
│   │   ├── module.go
│   │   ├── factory/           # ClientFactory
│   │   ├── buff/
│   │   ├── youpin/
│   │   ├── c5/                # (规划中)
│   │   └── igxe/              # (规划中)
│   │
│   └── service/              # Business logic（module 合并到 service.go）
│       ├── account/
│       │   └── service.go    # Module + Service + NewService
│       ├── trade/
│       │   └── service.go
│       ├── rental/
│       │   └── service.go
│       ├── pnl/
│       │   └── service.go
│       ├── inventory/
│       │   └── service.go
│       └── sync/
│           └── engine.go     # Module + Engine + NewEngine
│
├── hack/                     # Build & verification scripts
├── frontend/                 # React app
│   ├── Makefile              # 前端构建与质量检查
│   ├── .eslintrc.cjs         # ESLint 配置
│   └── .prettierrc           # Prettier 配置
├── design/                   # Architecture docs
├── .github/workflows/        # CI/CD pipelines
├── lefthook.yml
├── .golangci.yml
├── .gitignore
└── CHANGELOG.md
```

## 层级依赖关系

```
main.go (组装）
  ├── pkg/orm/Module         ← 依赖 pkg/utils/dbfx（非 Fx, 直接 import）
  │     ├── NewORM (*gorm.DB)
  │     └── Repos (AccountRepo, TradeRepo, ...)
  ├── pkg/platform/Module    ← 无内部依赖
  │     └── NewClientFactory
  └── pkg/service/*/Module   ← 依赖 pkg/orm + pkg/platform
        ├── account.Service
        ├── trade.Service
        ├── pnl.Service
        ├── inventory.Service
        └── sync.Engine
```

## Communication Pattern

前端通过 Wails Bind 调用 Go 方法，Bind 层做薄转发，Service 层处理业务逻辑，Repository 层执行参数化查询。数据流为单向：React QueryFn → Wails Bind → Service → Repository → GORM → SQLite。

TanStack Query 管理服务端状态（缓存、refetch、stale 管理），Zustand 管理客户端 UI 状态（侧边栏、选中账户、同步状态）。

## 核心数据视图

```
trade_records (所有买卖记录)
    │
    ├── trade_type = 'buy'
    │   ├── 写入 inventory（在库物品）
    │   └── 供后续卖出做 FIFO 成本匹配
    │
    └── trade_type = 'sell'
        ├── 匹配买入（FIFO by asset_id）→ 写入 matched_buy_trade_id → UPSERT pnl_daily（按日聚合）
        └── DELETE FROM inventory（出库）

inventory（当前库存状态）
    ├── status = 'in_inventory'  → 仓库中
    ├── status = 'listed'        → 已在平台挂售（含 listed_price）
    └── status = 'rented'        → 出租中（配合 rental_records）

pnl_daily（每日盈亏汇总）
    └── 按 (account_id, date) 聚合：trade_count / gross_pl / fee / net_pl
```

## Wails Bind 接口

`app.go` 暴露约 14 个方法给前端，覆盖五个功能域：

- **账户管理**：列表、创建、更新、删除、同步触发
- **库存查询**：列表（按账户和状态筛选）、物品详情（含租赁汇总）
- **交易查询**：已完成交易列表（含单笔盈亏）、已完成交易汇总
- **盈亏分析**：总盈亏摘要、按月分组盈亏
- **仪表盘**：跨账户聚合概览（总净值、库存数、交易数、租金收入）

前端通过 Wails 自动生成的 TypeScript 绑定（`wailsjs/go/main/App`）调用这些方法，无需手动维护接口定义。

## Module Dependency Map

```
Module E: Dashboard (read only)
    ▲
    ├── Module D: P&L
    │       ▲
    ├── Module B: Inventory（库存追踪、出租历史）
    │       ▲
    ├── Module C: Trade Records（买卖记录）
    │       ▲
    └── Module A: Accounts（平台连接，基础）
```

## Key Design Decisions

1. **`pkg/utils/dbfx`** — 纯 SQLite 连接层，不依赖 GORM 或业务，通过 Fx 提供
2. **`pkg/orm`** — ORM 抽象 + Repository 合集，通过 `NewORM` 提供 Fx 注入
3. **Service 的 module 合并到 service.go** — 不创建单独的 `module.go`，保持简洁
4. **Go 1.25** — 项目使用的 Go 版本
5. **SQLite 本地优先** — 无服务端，数据安全
