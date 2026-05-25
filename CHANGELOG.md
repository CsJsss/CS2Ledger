# Changelog

All notable changes to CS2 Ledger will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.4] - 2026-05-25

### Fixed

- Race condition in market price service by adding mutex lock
- Item name matching discrepancies between platform naming conventions
- CSQAQ client issues

## [0.1.3] - 2026-05-25

### Changed

- Merge feat/upgrade-ui (PR #7, #8)

## [0.1.2] - 2026-05-24

### Added

- Premium dark theme with orange accent, light/dark mode toggle
- Sidebar with collapsible toggle and active indicator
- Dashboard net worth hero card with monthly P&L sparkline
- Stat cards with gradient backgrounds, semantic P&L colors
- Dark-themed ECharts P&L chart (theme-aware)
- Market price update timestamp column in inventory
- P&L percentage column with server-side sorting
- Inventory group expand/collapse by item name + exterior
- CSS transitions for theme toggle, sidebar, and cards

### Changed

- Monospace numeric columns (JetBrains Mono) in all data tables
- Updated README and screenshots

### Fixed

- Missing `updated_at` in GetAllPrices Select

## [0.1.1] - 2026-05-23

### Added

- CSQAQ market price data integration for unrealized P&L calculation
- Server-side pagination for better performance
- Chinese (i18n) localization
- CSQAQ goods name resolution and unified HTTP client
- Per-exterior inventory grouping
- Account balance display features
- Sync-triggered data refresh

### Fixed

- `ensureProvider` on CSQAQ account addition
- Goods name normalization for cross-platform matching

### Changed

- Removed unused CI lint workflows

## [0.1.0] - 2026-05-16

### Added

- Multi-platform account management (BUFF / 悠悠有品 / C5 / ECO)
- Inventory management
- Trade record auto-fetch from platform buy/sell histories
- Realized P&L and comprehensive P&L summary
- Dashboard with net worth, distribution charts, and trend view
- Monthly P&L breakdown across all accounts
- CLI tool (`platform-cli`) for platform API debugging
- Wails-based desktop application shell (Go + React/TypeScript + MUI)
- Local SQLite (WAL mode) storage — all data stays on-device
