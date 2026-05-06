# Frontend Architecture

## Component Hierarchy

```
<App>
  └── <AppLayout>                    // 侧边栏 + 顶部状态栏
        ├── <Sidebar>                // 导航
        │     ├── <NavItem to="/dashboard" />
        │     ├── <NavItem to="/inventory" />        // 在库物品
        │     ├── <NavItem to="/trades" />           // 交易记录
        │     ├── <NavItem to="/trades/completed" /> // 已完成交易（含利润）
        │     ├── <NavItem to="/accounts" />
        │     └── <NavItem to="/settings" />
        │
        ├── <TopBar>
        │     └── <SyncAllButton />
        │
        └── <main>                   // React Router <Outlet />
              ├── DashboardPage          ( /dashboard )
              ├── InventoryPage          ( /inventory )
              │     └── InventoryDetailPage ( /inventory/:id )
              ├── TradeListPage          ( /trades )
              ├── CompletedTradesPage    ( /trades/completed )
              ├── PnLPage               ( /pnl )
              ├── AccountListPage        ( /accounts )
              ├── AccountDetailPage      ( /accounts/:id )
              └── SettingsPage           ( /settings )
```

## Routing

使用 React Router v7，以 `AppLayout` 为父路由，子路由覆盖 8 个页面路径。根路径 `/` 重定向到 `/dashboard`。

路由表：dashboard、inventory、inventory/:id、trades、trades/completed、pnl、accounts、accounts/:id、settings。

## Page Design

### 1. 已完成交易列表（`/trades/completed`）

展示所有"已买入且已卖出"的饰品，以及每一次完整交易的利润。

**数据来源：** `trade_records` 自连接（sell JOIN buy ON matched_buy_trade_id），每笔卖出即为一条已完成交易。

**表格列：**

| 列 | 说明 |
|----|------|
| 饰品名称 | |
| 磨损 | FN/MW/FT/WW/BS |
| 买入单价 | ¥xx.xx |
| 卖出单价 | ¥xx.xx |
| 数量 | |
| 毛利 | (卖 - 买) × 数量 |
| 手续费 | 买入+卖出手续费 |
| 净利润 | 毛利 - 手续费（绿正红负） |
| 交易时间 | 卖出时间 |
| 平台 | BUFF / YouPin / C5 / IGXE |

**筛选/排序：** 按平台筛选、按时间范围筛选、按利润排序、按饰品名称搜索。

**汇总卡片（页面顶部）：**
```
┌──────────────┬──────────────┬──────────────┬──────────────┐
│  总交易笔数   │   总毛利      │  总手续费     │   总净利润    │
│    156       │  ¥12,350.50  │  ¥1,235.00   │  ¥11,115.50  │
└──────────────┴──────────────┴──────────────┴──────────────┘
```

后端通过 `GetCompletedTrades` 和 `GetCompletedTradesSummary` 提供数据。

### 2. 在库物品详情（`/inventory/:id`）

展示某一件在库饰品的详细状态。数据来源：买入记录 + 出租历史 + 平台当前上架状态。

**页面布局：**

```
┌─────────────────────────────────────────────────────┐
│  ← 返回库存列表                                       │
│                                                       │
│  ┌─────────────────────────────────────────────┐      │
│  │  饰品信息                                     │      │
│  │  名称: AK-47 | 红线 (崭新出厂)                 │      │
│  │  平台: BUFF  |  资产ID: 123456789             │      │
│  │  买入时间: 2025-01-15  |  买入价格: ¥850.00    │      │
│  │  当前状态: 在售 / 出租中 / 仓库中               │      │
│  └─────────────────────────────────────────────┘      │
│                                                       │
│  ┌─────────────────────────────────────────────┐      │
│  │  出租历史                                     │      │
│  │                                               │      │
│  │  总出租天数: 120 天                            │      │
│  │  中断天数: 15 天                               │      │
│  │  总租金收入: ¥240.00                           │      │
│  │                                               │      │
│  │  出租记录表格:                                 │      │
│  │  ┌──────────┬──────────┬──────────┬────────┐ │      │
│  │  │ 开始时间  │ 结束时间  │ 天数     │ 收入    │ │      │
│  │  │ 2025-02  │ 2025-03  │ 30       │ ¥60.00 │ │      │
│  │  │ 2025-03  │ 2025-04  │ 30       │ ¥60.00 │ │      │
│  │  │ 2025-05  │ 2025-06  │ 30       │ ¥60.00 │ │      │
│  │  │ —        │ —        │ (中断15天)│ —      │ │      │
│  │  │ 2025-06  │ 2025-07  │ 30       │ ¥60.00 │ │      │
│  │  └──────────┴──────────┴──────────┴────────┘ │      │
│  └─────────────────────────────────────────────┘      │
│                                                       │
│  ┌─────────────────────────────────────────────┐      │
│  │  当前上架信息（如在售）                        │      │
│  │  上架价格: ¥920.00                            │      │
│  │  上架时间: 2025-04-01                          │      │
│  │  潜在利润: ¥70.00 (卖 ¥920 - 买 ¥850)          │      │
│  └─────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────┘
```

**出租中断天数计算：** 中断 = 两次连续出租之间的间隔天数。总持有天数 = 当前时间 - 买入时间；出租天数 = Σ(每次出租的 end_at - start_at)；中断天数 = 总持有天数 - 出租天数（粗略，精确需逐段判断）。

后端通过 `GetItemDetail` 和 `GetRentalHistory` 提供数据，返回 `ItemDetail`（含买入信息、出租历史、`RentalSummary` 汇总指标、当前上架信息）。

### 3. 库存列表（`/inventory`）

显示所有在库物品。

**表格列：**

| 列 | 说明 |
|----|------|
| 饰品名称 | |
| 磨损 | |
| 买入价格 | 单价（元） |
| 买入时间 | |
| 平台 | |
| 出租天数 | 累计出租天数 |
| 租金收入 | 累计租金（元） |
| 当前状态 | 在售 / 出租中 / 仓库中 |
| 浮动盈亏 | 如在售：当前售价 - 买入价 |

点击行 → 进入 `/inventory/:id` 详情页。

## State Management Strategy

### TanStack Query (Server State)

服务端状态（交易数据、库存、盈亏、仪表盘）通过 TanStack Query 管理。数据缓存策略按数据类型分层：

| 数据类型 | staleTime | 原因 |
| ------------ | ------------ | ---------------------------------- |
| 已完成交易 | 2 min | 相对静态，同步后刷新 |
| 在库物品 | 5 min | 不频繁变动 |
| 物品详情 | 5 min | 含出租历史，不频繁 |
| 仪表盘 | 5 min | 聚合数据 |
| 账户列表 | Infinity | 仅手动修改时变化 |

Mutation 操作（创建/删除账户、触发同步）在成功后 invalidate 相关查询缓存，确保数据一致性。

### Zustand (Client State)

客户端 UI 状态（侧边栏折叠、选中账户、同步进度、筛选条件）通过 Zustand 管理。每个状态字段有对应的 setter action。

## Wails Bindings

前端通过 Wails 自动生成的 TypeScript 绑定（`wailsjs/go/main/App`）调用 Go 方法。所有绑定从统一的 `lib/wails.ts` 模块重新导出，为其他模块提供单一导入路径。

## Feature Module Structure

```
features/
├── accounts/
│   ├── hooks/useAccounts.ts
│   ├── components/AccountForm.tsx, AccountCard.tsx
│   └── pages/AccountListPage.tsx, AccountDetailPage.tsx
├── inventory/
│   ├── hooks/useInventory.ts, useItemDetail.ts
│   ├── components/InventoryTable.tsx, ItemDetailCard.tsx,
│   │             RentalHistoryTable.tsx, ListingInfo.tsx
│   └── pages/InventoryPage.tsx, InventoryDetailPage.tsx
├── trades/
│   ├── hooks/useCompletedTrades.ts, useTradeList.ts
│   ├── components/CompletedTradesTable.tsx, PnlSummaryCards.tsx
│   └── pages/TradeListPage.tsx, CompletedTradesPage.tsx
├── pnl/
│   ├── hooks/usePnl.ts
│   └── pages/PnLPage.tsx
├── dashboard/
│   ├── hooks/useDashboard.ts
│   └── pages/DashboardPage.tsx
└── settings/
    └── pages/SettingsPage.tsx
```

注意：当前实现阶段使用扁平 `pages/` 和 `components/` 目录结构。上述 `features/` 结构是目标架构，在页面数量达到 15+ 时迁移。

## UI Component Strategy

- **MUI v5 (Material UI)** 作为基础组件库：Button, Table, Dialog, Select, Card, TextField, Chip, Skeleton, Alert, Box, Grid, Typography
- 利润数字着色：绿色（正利润 `success.main`）、红色（亏损 `error.main`）、灰色（零 `text.secondary`）
- `PnlSummaryCards` 作为共享组件，在 CompletedTradesPage 和 PnLPage 中复用
- 出租中断天数用 Tooltip 解释计算方式
- 金额统一使用 `formatCNY(cents)` 工具函数（分转元，千分位）

## Shared Components

每个页面需要处理四种状态：

- **Loading**：使用 MUI Skeleton 组件，形状匹配页面布局
- **Error**：`ErrorBanner` 组件，红色提示框 + 错误信息 + 重试按钮
- **Empty**：`EmptyState` 组件，灰色提示框 + 说明文字 + 可选操作按钮
- **Success**：正常数据展示

## Development Tooling

### Makefile

前端有独立的 `frontend/Makefile`，包含：dev（Vite HMR）、install（npm ci）、lint（ESLint）、format（Prettier）、typecheck（`tsc --noEmit`）、build（生产构建）、clean。

### Lint & Format

ESLint 配置 typescript-eslint、react-hooks、react-refresh 插件。Prettier 配置单引号、分号、尾逗号、100 字符宽。

### Git Hooks

项目使用 lefthook 统一管理 Go 和前端 hooks（不使用 husky）。
