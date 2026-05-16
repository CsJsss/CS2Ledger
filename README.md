# CS2 Ledger

跨平台 CS2 饰品资产管理与盈亏追踪桌面应用 —— 帮助多平台、多账号的 CS2 饰品持有者回答：**我到底赚了多少钱。**

不替代 csqaq/steamdt（行情看板），不做中心化服务。所有数据只存本地 SQLite，不上传任何数据。

## 支持的平台

| 平台 | 标识 | 凭据类型 | 说明 |
|------|------|----------|------|
| BUFF | `buff` | Cookie | 网页版登录后的 Cookie |
| 悠悠有品 | `youpin` | Token | 网页版登录后的 Authorization Bearer Token |
| C5 | `c5` | API Key | 商户后台 Open API 的 app-key |
| ECO | `eco` | Partner ID + 私钥 | 同 IGXE 格式 |

## 各平台凭据获取方法

### BUFF — 网页 Cookie

1. 打开浏览器并登录 [buff.163.com](https://buff.163.com)
2. 按 `F12` 打开开发者工具，切换到 **Network（网络）** 标签
3. 刷新页面，点击任意发往 `buff.163.com` 的请求
4. 在 **Request Headers** 中找到 `Cookie` 字段，复制完整的 Cookie 字符串

> **注意：** Cookie 会过期。如果数据同步失败，大概率是 Cookie 失效，重新获取即可。

### 悠悠有品 — 网页 Token

1. 打开浏览器并登录 [youpin898.com](https://www.youpin898.com)
2. 按 `F12` 打开开发者工具，切换到 **Network（网络）** 标签
3. 刷新页面，点击任意发往 `api.youpin898.com` 的请求
4. 在 **Request Headers** 中找到 `Authorization` 字段，复制 `Bearer` 后面的 Token 字符串

### C5 — Open API Key

1. 登录 [C5 开放平台](https://www.c5game.com/open)
2. 进入 **申请开通 → 前往API管理** [页面](https://www.c5game.com/user/user/open-api)
3. 复制 **app-key**（即 API Key）

### ECO — Partner ID + RSA 私钥

ECO 使用开放平台 API，凭据格式为 `PartnerId:私钥PEM`。

1. 手机登录 ECO, 在**我的 -> 开放平台**中, 申请入驻
2. 生成RSA公钥私钥: 请妥善保管RSA私钥
   ```bash
    openssl genrsa -out private_key.pem 2048
    openssl rsa -in private_key.pem -pubout -out public_key.pem
    ```
3. 在ECO 开放平台设置中, 开启签名验证, 并上传RSA公钥
4. 找到 **Partner ID** 和 **RSA 私钥**（通常为 PEM 格式）
5. 在添加账户的 Cookie 输入框中填入：`你的PartnerId:-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----`

> 私钥中的换行可以用 `\n` 表示，直接写在一行或多行粘贴均可。

## 功能概览

- **账户管理：** 连接多平台、多账号，统一管理
- **库存管理：** 库存快照、对比、详情查看、导出
- **交易记录：** 买入/卖出历史自动拉取，支持手动录入，批次成本追踪
- **盈亏计算：** 已实现盈亏、浮动盈亏、综合 P&L
- **仪表盘：** 资产净值、分布图、趋势图

## 技术栈

- **桌面壳：** Wails v2 (Go + React/TypeScript)
- **前端：** React 19 + TypeScript + Vite + MUI v5 + TanStack Query + Zustand
- **后端：** Go + GORM + SQLite (WAL) + Uber Fx
- **平台对接：** 模拟 HTTP 请求，无官方 API（C5/IGXE/ECO 除外）

## 快速开始

### 前置条件

- Go 1.25+
- Node.js 22+
- Wails CLI（可选，`make install-wails` 自动安装）

### 开发

```bash
make setup        # 安装依赖 + git hooks
make dev          # 启动 Wails 开发服务器（热重载）
```

### 构建

```bash
make build        # 生产构建 → build/bin/
make build-dev    # 调试构建
```

### 测试 & 检查

```bash
make test         # 运行所有测试
make lint         # golangci-lint 检查
make fmt          # 格式化代码
make vet          # go vet 检查
```

## 项目结构

```
cs2-ledger/
├── main.go              # Wails 入口
├── app.go               # Wails Bind 方法（前端调用的 Go 函数）
├── cmd/platform-cli/    # 命令行工具（平台调试验证）
├── pkg/
│   ├── model/           # 数据模型
│   ├── orm/             # 数据访问层 (GORM)
│   ├── platform/        # 外部平台客户端 (BUFF/悠悠/C5/IGXE/ECO)
│   ├── service/         # 业务逻辑层
│   └── utils/           # 基础设施 (数据库、日志)
├── frontend/            # React 前端
├── migrations/          # 数据库迁移
└── design/              # 架构设计文档
```

## 贡献指南

### 开发流程

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feat/your-feature`
3. 运行 `make setup` 安装开发依赖
4. 开发并确保 `make lint test` 通过
5. 提交代码并创建 PR

### 代码规范

- Go 代码遵循标准 Go 风格，使用 `gofmt` 格式化
- TypeScript/React 代码使用 Vite 默认配置
- 提交信息使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：
  - `feat:` 新功能
  - `fix:` 修复
  - `refactor:` 重构
  - `chore:` 杂项

### 添加新平台

如需对接新的交易平台，实现 `platform.Client` 接口即可：

参考 `pkg/platform/buff/` 或 `pkg/platform/c5/` 的实现，并在 `pkg/platform/factory/` 中注册新平台。

### 数据库迁移

```bash
make db-new NAME=add_xxx_table   # 创建新迁移文件
```

迁移文件放在 `migrations/` 目录，按时间戳排序，在应用启动时自动执行。

## 参考

BUFF/悠悠有品的 API 交互逻辑参考了 [SteamAuto](https://github.com/jiajiaxd/SteamAuto) 项目的实现。