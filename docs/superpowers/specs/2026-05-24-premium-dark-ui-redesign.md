# Premium Dark UI Redesign — CS2 Ledger

## Overview

Full visual redesign of CS2 Ledger frontend. Transform from generic MUI light-theme admin panel to a premium dark-themed CS2 trading terminal. Data layer and component structure are preserved; changes are purely visual (theme, components, layout).

**Aesthetic**: Premium Dark — deep black surfaces, orange primary accent, green/red P&L tinting. Linear.app meets CS2 trading.

## Design System

### Color Tokens

| Token | Hex | Tailwind | Usage |
|-------|-----|----------|-------|
| Page BG | `#09090b` | zinc-950 | Root background |
| Card BG | `#18181b` | zinc-900 | Cards, Paper, Drawer |
| Elevated | `#27272a` | zinc-800 | Hover states, menus, dialogs |
| Primary | `#f97316` | orange-500 | CTAs, active nav, accent bars |
| Primary Hover | `#ea580c` | orange-600 | Button hover |
| Profit | `#22c55e` | green-500 | Positive P&L, gains |
| Loss | `#ef4444` | red-500 | Negative P&L, destructive actions |
| Warning | `#f59e0b` | amber-500 | Unrealized P&L |
| Text Primary | `#fafafa` | zinc-50 | Headings, body |
| Text Secondary | `#d4d4d8` | zinc-300 | Descriptions |
| Text Muted | `#a1a1aa` | zinc-400 | Labels, hints |
| Border Default | `rgba(255,255,255,0.08)` | — | Card/table borders |
| Border Strong | `rgba(255,255,255,0.12)` | — | Hover borders |
| Border Accent | `rgba(249,115,22,0.2)` | — | Active/selected borders |

### Typography

- **Family**: Geist Variable (already installed via `@fontsource-variable/geist`)
- **Mono**: JetBrains Mono (add to dependencies, for data tables and numbers)
- **Scale**:
  - Display: 32px / 700 / -0.5px — page titles
  - H5: 24px / 600 — card values
  - H6: 18px / 600 — section headings
  - Body: 14px / 400 / 1.5 — content
  - Label: 11px / 500 / uppercase / +0.5px — card labels
  - Mono: 13px / 500 — table numbers, financial data

### Spacing
   - 4px base unit (MUI default)
   - Cards: 16-18px internal padding
   - Grid gaps: 12px (stat cards), 16px (sections)
   - Page padding: 24px

### Elevation & Depth
   - No heavy box-shadows — use border contrast + subtle gradients
   - Card depth: `linear-gradient(135deg, #18181b, #1f1f23)` + `border: 1px solid #3f3f46`
   - Elevated: `#27272a` background
   - P&L cards: colored left border (3px) + tinted icon background

### Corner Radius
   - Cards/Buttons: 8px
   - Stat cards: 10px
   - Sidebar items: 6px
   - Tables: 8px container, no rounding on rows

### Icons
   - MUI Icons (already installed), outlined style, 20px
   - Sidebar nav items get icons
   - Dashboard stat cards get semantic icons (wallet, chart, inventory, etc.)

## Component Changes

### 1. Theme (`theme.ts`)
- Complete rewrite: dark palette, custom component overrides for all MUI surfaces
- Add `CssBaseline` dark mode via palette mode
- Override MuiCard, MuiPaper, MuiTable, MuiDrawer, MuiAppBar, MuiButton, MuiChip, MuiTextField
- Export semantic color constants for ECharts and inline styles

### 2. AppLayout (`AppLayout.tsx`)
- Dark background for AppBar + Drawer + main content
- Sidebar: add MUI icons to each nav item (Dashboard, Inventory, Trades, P&L, Accounts, Settings)
- Sidebar active state: orange left accent bar + tinted background (instead of full blue block)
- AppBar: darker background, subtle bottom border instead of elevation shadow
- Search bar: dark themed with icon
- Account selector chip: orange-tinted when an account is selected

### 3. Dashboard (`DashboardPage.tsx`)
- Stat cards: dark gradient background, uppercase labels, icons, trend indicators
- P&L cards: colored left border (green for profit, red for loss)
- Use MUI icons per card type (AccountBalanceWallet, Inventory, TrendingUp, etc.)
- Loading skeletons: match dark card dimensions

### 4. P&L Page (`PnLPage.tsx`)
- Summary cards: same dark stat card treatment
- ECharts: update color palette to match theme (dark background, orange/green bars, lighter grid)
- Chart tooltip: dark themed
- DataZoom slider: dark themed

### 5. Inventory (`InventoryPage.tsx`)
- Data table: dark Paper container, dark rows with subtle borders
- Monospace font for numeric columns (prices, quantities, P&L)
- Group header rows: slightly lighter background (#1a1a1e)
- Expandable sub-rows: indent with subtle left border
- Type filter chip: dark themed
- Search bar: dark themed

### 6. Completed Trades (`CompletedTradesPage.tsx`)
- Same table treatment as Inventory
- Tabs: orange underline indicator
- Detail dialog: dark themed

### 7. Accounts (`AccountsPage.tsx`)
- Same table treatment
- Action buttons: outlined style (default for dark)
- Status chips: dark variant

### 8. Settings (`SettingsPage.tsx`)
- Form controls: dark themed inputs
- Select: dark dropdown
- Success/error alerts: dark variant

### 9. Empty States (`EmptyState.tsx`)
- Replace emoji icon support with MUI icon component
- Dark themed text
- Outlined button for actions

### 10. Error Banner (`ErrorBanner.tsx`)
- Dark Alert variant

## Implementation Order

1. **Theme** — rewrite `theme.ts` with full dark palette + component overrides
2. **AppLayout** — sidebar icons, dark appbar, account chip
3. **Dashboard** — stat cards redesign
4. **PnLSummaryCards** — shared component update
5. **ECharts** — dark theme colors
6. **Tables** — dark table styling across Inventory, Trades, Accounts
7. **Dialogs & Forms** — dark theme pass
8. **EmptyState & ErrorBanner** — icon + dark pass

## What Does NOT Change
- All hooks (`useDashboard`, `useInventory`, etc.)
- Data fetching logic (TanStack Query)
- State management (Zustand store)
- Routing structure
- ECharts component choice (keep `echarts-for-react`)
- Table libraries (keep `@tanstack/react-table`)
- Wails bindings

## Dependencies to Add
- `@fontsource/jetbrains-mono` — monospace font for data tables

## Verification
- Run `npm run typecheck` — zero errors
- Run `npm run test` — all existing tests pass
- Run `npm run dev` — visual inspection of all 6 pages
- Check dark mode contrast ratios meet WCAG AA (4.5:1 for body text)
