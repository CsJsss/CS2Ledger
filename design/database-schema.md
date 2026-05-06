# Database Schema

## 概述

SQLite，WAL 模式。五张核心业务表：`accounts`、`trade_records`、`inventory`、`rental_records`、`pnl_daily`。加一张 `schema_version` 用于迁移跟踪。

数据库文件：`<用户数据目录>/cs2-ledger/data.db`

## 表设计

### `accounts` — 平台账户

存储用户添加的每个平台账户。关键字段：

- **name**（唯一）：账户别名
- **platform**：平台标识（buff / youpin / c5 / igxe）
- **cookie**：加密存储的凭证（BUFF 用 Cookie，悠悠有品用 Bearer Token，C5 用 API Key，IGXE 用 PartnerId+私钥）。该字段永不暴露给前端（`json:"-"`）
- **available_balance / purchase_balance**：平台余额，以分为单位
- **status**：账户状态（active / expired / error），同步失败时标记为 expired
- **last_sync_at**：最后同步时间（unix 毫秒），作为增量拉取的 since 参数

name 上有唯一索引。

### `trade_records` — 交易记录

所有买卖记录的完整账本，是 P&L 计算的核心数据源。关键字段：

- **account_id**：所属账户
- **asset_id**：平台资产 ID（饰品唯一标识）
- **item_name**：物品名称
- **trade_type**：buy 或 sell
- **unit_price / total_price / fee**：单价、总价、手续费，均以分为单位
- **trade_at**：交易时间（unix 毫秒）
- **source**：数据来源（platform 自动拉取 / manual 手动录入）
- **external_id**：平台交易 ID，用于去重（`(account_id, external_id)` 部分唯一索引）
- **matched_buy_trade_id**：自引用外键，卖出记录指向其 FIFO 匹配的买入记录。NULL 表示尚未匹配

辅助索引：account+时间（列表查询）、asset_id（FIFO 匹配）、account+type（买入/卖出筛选）、matched_buy_trade_id（已匹配查询）。

### `inventory` — 当前库存

追踪用户当前持有的物品。每个物品由 `(account_id, asset_id)` 唯一标识。关键字段：

- **exterior / paint_seed**：饰品磨损和图案信息
- **buy_trade_id**：关联到买入交易记录
- **status**：物品状态枚举
  - `in_inventory`：仓库中
  - `listed`：已在平台挂售（含 listed_price 和 listed_at）
  - `rented`：出租中
- **listed_price / listed_at**：挂售价格和时间

库存生命周期：

```
买入 → INSERT (status='in_inventory')
   ├── 上架 → UPDATE status='listed', listed_price, listed_at
   ├── 出租 → UPDATE status='rented'（配合 rental_records）
   ├── 卖出 → DELETE（同时 trade_records 新增 sell → 触发 P&L 计算）
   └── 下架 → UPDATE status='in_inventory'
```

### `rental_records` — 出租记录

记录物品的出租历史。关键字段：

- **asset_id / item_name**：关联的物品
- **income**：租金收入（分）
- **duration_days**：出租天数
- **start_at / end_at**：起止时间（unix 毫秒）
- **external_id**：平台记录 ID，`(account_id, external_id)` 部分唯一索引用于去重

### `pnl_daily` — 每日盈亏汇总

按 `(account_id, date)` 唯一聚合每日盈亏数据。关键字段：

- **date**：日期字符串（YYYY-MM-DD）
- **trade_count**：当日完成的卖出笔数
- **gross_pl**：毛利（分）
- **fee**：手续费合计（分）
- **net_pl**：净利（分）

写入使用 `INSERT ... ON CONFLICT` 累加模式：同一日多次匹配时，trade_count / gross_pl / fee / net_pl 累加而非覆盖。

### `schema_version` — 迁移版本

记录已应用的迁移编号和时间戳。应用启动时读取此表，按序执行未应用的 SQL 迁移文件。

## 数据关系图

```
accounts (1)
    │
    ├──< trade_records (N)       买卖记录（buy + sell）
    │       │
    │       │  sell 记录通过 matched_buy_trade_id 指向对应买入（FIFO）
    │       │
    │       └── pnl_daily (N)    每日盈亏汇总
    │
    ├──< inventory (N)           当前库存
    │       │ buy_trade_id → trade_records
    │       └── rental_records (N)  出租历史
    │
    └──< rental_records (N)      租金记录
```

## P&L 计算逻辑

### 匹配规则：FIFO by asset_id

每一笔**卖出**交易按 asset_id 匹配到最早的未匹配**买入**交易：

1. 查找该 asset_id 的所有买入记录（trade_type='buy'）
2. 筛选尚未被匹配的买入（未被任何 sell 的 matched_buy_trade_id 引用）
3. 按 trade_at 升序（FIFO）
4. 取最早的那笔买入作为成本基准
5. 更新 sell.matched_buy_trade_id = buy.id
6. 计算单笔盈亏：gross_pl = (卖价 - 买价) × 数量，net_pl = gross_pl - 买卖手续费之和
7. UPSERT pnl_daily，按 (account_id, date) 聚合累加（trade_count+1, gross_pl/fee/net_pl 累加）
8. 从 inventory 中删除该 asset_id（库存出库）

### 每日聚合

不单独存储单笔盈亏，而是按天聚合到 `pnl_daily` 表。`ProcessPending()` 遍历所有未匹配卖出 → FIFO 匹配买入 → 计算单笔 → UPSERT 到对应日期的聚合行。

### 增量更新

`SyncAccount()` 流程：拉取新交易（since last_sync_at）→ trade_records UPSERT（external_id 去重）→ 买入创建 inventory → 卖出写入 trade_records → ProcessPending() 做增量 P&L → 更新账户余额和 last_sync_at。

### 查询

- **总盈亏**：对 pnl_daily 按 account_id 聚合 SUM(trade_count / gross_pl / fee / net_pl)
- **按月分组**：对 date 字段取前 7 位（YYYY-MM）分组聚合
- **已完成交易明细**：trade_records 自连接（sell JOIN buy ON matched_buy_trade_id），计算每笔 net_pl

## 迁移策略

`migrations/` 目录中按编号顺序存放 SQL 文件（如 `001_initial.sql`）。应用启动时读取 `schema_version` 表，按序在事务中执行未应用的迁移。

## 价格精度

所有价格以**分（cent）**为单位存储（REAL 类型）。前端展示时除以 100 转为元。
