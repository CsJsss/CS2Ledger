# CI/CD & Development Infrastructure

## GitHub Actions

### CI Pipeline

每次 push 到 main 或 PR 时触发，包含三个 Job：**lint**（golangci-lint, 5 分钟超时）、**test**（`go test ./pkg/... -v -race -count=1`）、**build**（依赖 lint + test 通过，执行 `npm ci` 后 `go build`）。使用 Go 1.25 和 Node 22。

### Lint 专项

仅在 PR 时运行，输出格式更详细（colored-line-number），便于在 PR diff 中定位问题。

## golangci-lint

启用的 linter：errcheck、gosimple、govet、ineffassign、staticcheck、unused、gofmt、goimports、misspell、unconvert、unparam、prealloc。govet 启用全部检查，goimports 配置 local-prefixes 为 `cs2-ledger`。超时 5 分钟。

## lefthook（Git Hooks）

开发者 clone 后运行 `lefthook install` 安装 hooks。

- **pre-commit**（并行执行）：gofmt、goimports（自动格式化并 stage）、golangci-lint（仅检查相对于 HEAD~1 的变更）
- **commit-msg**：强制 conventional commit 前缀（feat / fix / refactor / chore / docs / test / ci / style）

## .gitignore

忽略依赖目录（node_modules/）、构建产物（dist/、build/、二进制文件）、Wails 生成代码（wailsjs/）、IDE 和 OS 文件、数据库文件（*.db*）、日志、环境变量文件、压缩包。

## CHANGELOG.md

遵循 Keep a Changelog 格式，按语义化版本组织。当前 [Unreleased] 记录项目初始化。
