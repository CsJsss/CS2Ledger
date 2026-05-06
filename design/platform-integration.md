# Platform Integration

## 设计原则

无官方 API，所有对接通过模拟 HTTP 请求（逆向工程）实现。对外部平台而言，请求等同于用户在浏览器/App 中的操作。

- **遵守平台协议**：不刷接口，仅拉取个人数据，合理频率
- **凭证本地加密存储**：不传输，不上传
- **失败兜底**：一个平台挂了不影响其他平台
- **幂等同步**：通过 external_id 去重，重复拉取不产生重复数据

## Common Interface

所有平台实现统一的 `Client` 接口，包含四个方法：

- **Verify()** — 测试凭证是否有效（失败时将账户标记为 expired）
- **FetchBuyHistory(since)** — 增量拉取买入记录，since 为 unix 毫秒时间戳
- **FetchSellHistory(since)** — 增量拉取卖出记录
- **FetchBalance()** — 获取账户余额（可用余额 + 求购余额）

所有金额以**分（cent）**为单位，时间戳为 unix 毫秒。每个平台返回统一的 `TradeRecord` 和 `Balance` 结构体，屏蔽平台差异。

`ClientFactory` 根据平台名（"buff" / "youpin" / "c5" / "igxe"）和凭证创建对应的客户端。凭证格式因平台而异。

## BUFF (buff.163.com)

### 认证

Cookie-based 认证。用户提供 `session` cookie，客户端自动管理 `csrf_token`（从 API 响应的 Set-Cookie 中提取，用于写操作的 `X-CSRFToken` header）。

### 已实现端点

- **凭证验证**：GET 用户信息接口，返回 `code: "OK"` 表示有效
- **余额**：GET 资产摘要接口，`balance` → available_balance，`frozen_balance` → purchase_balance
- **买入历史**：分页 GET，参数 `game=csgo`，按 `state=SUCCESS` 筛选，`income` 字段为元需 ×100 转分。响应含 `goods_infos` 用于解析物品名称
- **卖出历史**：分页 GET `/api/market/sell_order/history`，参数 `appid=730&mode=1`，不需 CSRF。响应结构与买入历史类似，含 `goods_infos`
- **在售物品**：GET 当前挂售列表（含售价、上架时间）

### 反爬措施

- 分页请求间隔 ≥ 2 秒
- 出现验证码时暂停同步并通知用户
- 不需要写操作 CSRF token（CS2 Ledger 只读）

## 悠悠有品 YouPin (api.youpin898.com)

### 认证

Token-based（JWT Bearer）。用户在悠悠有品 App 中抓包获取 `Authorization: Bearer {token}`。客户端模拟 Android App 请求头（okhttp User-Agent、app-version、devicetoken 等）。

### 已实现端点

- **凭证验证**：POST 用户信息接口，存在 `Data` 或 `code == 0` 表示有效
- **买入历史**：分页 POST，`orderStatus=340`（已完成），`paymentAmount` 单位为分
- **卖出历史**：分页 POST，结构与买入相同
- **余额**：暂未找到专用余额端点（Steamauto 参考实现中也未包含）。当前返回空 Balance，待后续补充

### 待补充（Phase 2）

- **租金/出租 API**：悠悠有品有租赁业务。Steamauto 参考实现中包含租赁上架/下架/改价端点，以及出租记录列表（`/api/youpin/bff/trade/v1/order/lease/out/list`）。需要用户确认端点后实现
- **充值/提现记录**：用于完整资金流水追踪

## C5 (openapi.c5game.com)

### 认证

API Key 认证。用户在 C5Game 网站 `https://www.c5game.com/user/user/open-api` 获取 AppKey，作为 HTTP header `app-key` 传递。最简洁的认证方式——无 Cookie、无 CSRF、无加密。

### 端点设计

- **Verify / FetchBalance**：`GET /merchant/account/v1/balance`，返回 `{"success": true, "data": {"amount": ...}}`
- **FetchBuyHistory / FetchSellHistory**：`GET /merchant/order/v1/list`，参数 `status=10`（已完成）、`page`。按订单方向字段区分买卖

注意：C5 使用 **HTTP**（非 HTTPS）协议。

## IGXE / ECOsteam (openapi.ecosteam.cn)

### 认证

RSA-SHA256 签名认证。用户提供 PartnerId 和 RSA 私钥（PEM 格式）。每个请求在 JSON body 中附带 `PartnerId`、`Timestamp`（unix 秒）、`Sign`（对所有参数按键名排序后拼接 `key=value&...`、SHA256 哈希、RSA PKCS1v15 签名、Base64 编码）。

### 端点设计

- **Verify / FetchBalance**：`POST /Api/Merchant/GetTotalMoney`，返回 `{"ResultCode": "0", "ResultData": {"Money": ...}}`
- **FetchSellHistory**：`POST /Api/open/order/SellerOrderList`，分页参数 `PageIndex`/`PageSize`，可选 `StartTime`/`EndTime`
- **FetchBuyHistory**：同上端点（ECOsteam 可能不区分买卖列表，需按订单字段区分）
- **出租相关**（Phase 2）：`POST /Api/Rent/QuerySelfRentGoods`、`POST /Api/Rent/PublishRentAndSaleGoods`

## 实现状态

| 平台 | 买入历史 | 卖出历史 | 余额 | 租金 | 备注 |
|------|---------|---------|------|------|------|
| BUFF | 已完成 | 已完成 | 已完成 | — | Cookie 认证 |
| 悠悠有品 | 已完成 | 已完成 | 待补充 | Phase 2 | Bearer Token 认证 |
| C5 | 待实现 | 待实现 | 待实现 | — | API Key 认证 |
| IGXE | 待实现 | 待实现 | 待实现 | Phase 2 | RSA 签名认证 |

## 同步策略

### 增量同步流程

1. 加载账户信息（platform、credential、last_sync_at）
2. 通过工厂创建对应平台客户端
3. 验证凭证有效性，失败则标记账户为 expired
4. 并发拉取：buy history、sell history、balance（使用 WaitGroup）
5. 买入写入 trade_records（external_id 去重）+ 创建 inventory
6. 卖出写入 trade_records（external_id 去重）
7. 调用 PnL ProcessPending 做增量 FIFO 匹配和盈亏计算
8. 更新账户余额和 last_sync_at

### 频率限制

| 操作 | 最小间隔 | 说明 |
|------|---------|------|
| BUFF 分页请求 | 2 s | 反爬要求 |
| YouPin 分页请求 | 1 s | App 模拟 |
| 全量同步 | 30 min | 避免频繁请求 |
| 手动同步 | 10 min | 用户主动触发 |

## 错误场景处理

| 场景 | BUFF 检测 | YouPin 检测 | 处理 |
|------|----------|------------|------|
| 凭证过期 | `code != "OK"` 或重定向登录 | JWT 过期（code 84101） | 标记 status='expired'，提示用户 |
| 平台不可用 | 5xx / timeout | 同 | 返回 error，不标记账户 |
| 风控/验证码 | 响应体异常或空 data | 同 | 暂停同步，提示用户手动解除 |
| 频率限制 | 空 items 或异常 | 同 | 等待后重试 |
| 格式变化 | JSON 解析失败 | 同 | 记录原始响应，告警 |
