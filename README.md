# CS2 Ledger

跨平台 CS2 饰品资产管理与盈亏追踪桌面应用 —— 帮助多平台、多账号的 CS2 饰品持有者回答：**我到底赚了多少钱。**

## 技术栈

- **桌面壳:** Wails v2 (Go + React/TypeScript)
- **前端:** React 19 + TypeScript + Vite + MUI v5 (Material UI) + TanStack Query
- **后端:** Go + GORM + SQLite (WAL) + Uber Fx
- **数据:** 完全本地存储，不上传任何数据

## 快速开始

### 前置条件

- Go 1.25
- Node.js 22+
- Wails CLI (可选，`make install-wails`)

### 开发

```bash
make setup        # 安装依赖 + git hooks
make dev          # 启动 Wails 开发服务器 (热重载)
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
├── cmd/demo/            # 命令行演示工具
├── pkg/
│   ├── model/           # 数据模型
│   ├── orm/             # 数据访问层 (GORM)
│   ├── platform/        # 外部平台客户端 (BUFF/悠悠/C5/IGXE)
│   ├── service/         # 业务逻辑层
│   └── utils/           # 基础设施 (数据库、日志)
├── frontend/            # React 前端
├── migrations/          # 数据库迁移
└── design/              # 架构设计文档
```

## 参考

平台客户端（BUFF/悠悠/C5/IGXE）的 API 交互逻辑参考了 [SteamAuto](https://github.com/jiajiaxd/SteamAuto) 项目的实现。

## 架构文档

详细的架构设计和模块说明见 [design/](design/) 目录。
