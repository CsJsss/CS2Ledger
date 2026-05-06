# CS2 Ledger

跨平台 CS2 饰品资产管理与盈亏追踪桌面应用。

## 产品定位

个人本地交易终端 —— 帮助多平台、多账号的 CS2 饰品持有者回答一个问题：**我到底赚了多少钱。**

不替代 csqaq/steamdt（行情看板），不做中心化服务。

## 技术栈

- **桌面壳:** Wails (Go + React/TypeScript)
- **前端:** React 19 + TypeScript + Vite + MUI v5
- **后端:** Go (HTTP 客户端、SQLite、数据同步)
- **数据库:** SQLite (WAL 模式)，数据完全本地
- **前端状态:** TanStack Query + Zustand

## 架构原则

- 所有数据只存本地 SQLite，不做任何上传
- 用户提供平台 Cookie/Token，Go 后端以用户身份请求外部平台
- Wails Bind 通信 —— Go 暴露函数，前端直接调用，无需 REST 层
- 平台对接通过模拟 HTTP 请求（逆向），无官方 API

## Feature Phases

### Phase 1（当前）: 资产管理 & P&L
- Module A: 账户与平台连接（BUFF/悠悠/C5/IGXE）
- Module B: 库存管理（快照、对比、详情、导出）
- Module C: 交易记录 & 成本基准（自动拉取 + 手动录入 + 批次追踪）
- Module D: 盈亏计算（已实现盈亏、浮动盈亏、租金收入、综合 P&L）
- Module E: 仪表盘（净值、分布图、排行榜、趋势图）
- Module F: 数据管理（备份/恢复）

### Phase 2（规划中）: 租赁分析 + 自动化
### Phase 3（规划中）: 策略引擎 + 风控
