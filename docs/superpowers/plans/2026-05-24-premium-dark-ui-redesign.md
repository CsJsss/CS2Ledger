# Premium Dark UI Redesign — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform CS2 Ledger frontend from generic MUI light-theme admin panel to a premium dark-themed CS2 trading terminal.

**Architecture:** Pure visual layer overhaul — rewrite MUI theme with dark palette and component overrides, then cascade through layout, pages, and shared components. No data layer, routing, or state management changes. All existing tests must continue to pass.

**Tech Stack:** React 19, TypeScript, MUI v5, Vite, Geist Variable + JetBrains Mono fonts

---

### Task 1: Install JetBrains Mono font

**Files:**
- Modify: `frontend/package.json`

- [ ] **Step 1: Install font dependency**

```bash
cd frontend && npm install @fontsource/jetbrains-mono
```

- [ ] **Step 2: Commit**

```bash
git add frontend/package.json frontend/package-lock.json
git commit -m "chore: add JetBrains Mono font for data tables"
```

---

### Task 2: Rewrite MUI theme with dark palette

**Files:**
- Modify: `frontend/src/theme.ts`

- [ ] **Step 1: Replace theme.ts with full dark theme**

Replace the entire file. The new theme defines a dark palette with semantic tokens, Geist as default font, and component overrides for every MUI surface used in the app.

```typescript
import { createTheme } from '@mui/material/styles';

declare module '@mui/material/styles' {
  interface Palette {
    accent: Palette['primary'];
    profit: Palette['primary'];
    loss: Palette['primary'];
    warning: Palette['primary'];
  }
  interface PaletteOptions {
    accent?: PaletteOptions['primary'];
    profit?: PaletteOptions['primary'];
    loss?: PaletteOptions['primary'];
    warning?: PaletteOptions['primary'];
  }
}

const theme = createTheme({
  palette: {
    mode: 'dark',
    background: {
      default: '#09090b',
      paper: '#18181b',
    },
    text: {
      primary: '#fafafa',
      secondary: '#d4d4d8',
      disabled: '#a1a1aa',
    },
    primary: {
      main: '#f97316',
      dark: '#ea580c',
      contrastText: '#ffffff',
    },
    secondary: {
      main: '#a1a1aa',
    },
    error: { main: '#ef4444' },
    success: { main: '#22c55e' },
    warning: { main: '#f59e0b' },
    divider: 'rgba(255,255,255,0.08)',
    accent: {
      main: '#f97316',
      dark: '#ea580c',
      contrastText: '#ffffff',
    },
    profit: { main: '#22c55e' },
    loss: { main: '#ef4444' },
  },
  typography: {
    fontFamily: '"Geist Variable", sans-serif',
    h4: { fontSize: '1.75rem', fontWeight: 700, letterSpacing: '-0.02em' },
    h5: { fontSize: '1.25rem', fontWeight: 600 },
    h6: { fontSize: '1.1rem', fontWeight: 600 },
    body2: { fontSize: '0.875rem' },
    caption: { fontSize: '0.75rem' },
  },
  shape: { borderRadius: 8 },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: { backgroundColor: '#09090b' },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: { textTransform: 'none', fontWeight: 500 },
        contained: { boxShadow: 'none' },
        outlined: { borderColor: 'rgba(255,255,255,0.12)' },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          backgroundImage: 'linear-gradient(135deg, #18181b, #1f1f23)',
          border: '1px solid rgba(255,255,255,0.08)',
          boxShadow: 'none',
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          border: '1px solid rgba(255,255,255,0.08)',
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: '#0d0d10',
          borderBottom: '1px solid rgba(255,255,255,0.08)',
          boxShadow: 'none',
        },
        colorDefault: {
          backgroundColor: '#0d0d10',
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: { backgroundColor: '#0d0d10', borderRight: '1px solid rgba(255,255,255,0.08)' },
      },
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          borderRadius: 6,
          marginBottom: 2,
          '&.active': {
            backgroundColor: 'rgba(249,115,22,0.12)',
            color: '#f97316',
            borderLeft: '3px solid #f97316',
          },
          '&.active:hover': { backgroundColor: 'rgba(249,115,22,0.18)' },
          '&:hover': { backgroundColor: 'rgba(255,255,255,0.04)' },
        },
      },
    },
    MuiTableRow: {
      styleOverrides: {
        root: {
          '&:hover': { backgroundColor: 'rgba(255,255,255,0.03) !important' },
        },
        head: {
          '&:hover': { backgroundColor: 'transparent !important' },
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        root: { borderBottom: '1px solid rgba(255,255,255,0.06)', color: '#d4d4d8' },
        head: { color: '#a1a1aa', fontWeight: 500, fontSize: '0.7rem', textTransform: 'uppercase', letterSpacing: '0.05em' },
      },
    },
    MuiTableSortLabel: {
      styleOverrides: {
        root: { color: '#a1a1aa', '&.Mui-active': { color: '#fafafa' } },
        icon: { color: '#a1a1aa !important' },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: { fontWeight: 500, fontSize: '0.75rem' },
        outlined: { borderColor: 'rgba(255,255,255,0.12)' },
        filled: {},
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            backgroundColor: '#18181b',
            '& fieldset': { borderColor: 'rgba(255,255,255,0.08)' },
            '&:hover fieldset': { borderColor: 'rgba(255,255,255,0.12)' },
            '&.Mui-focused fieldset': { borderColor: '#f97316' },
          },
          '& .MuiInputLabel-root': { color: '#a1a1aa' },
        },
      },
    },
    MuiSelect: {
      styleOverrides: {
        root: {
          backgroundColor: '#18181b',
          '& .MuiOutlinedInput-notchedOutline': { borderColor: 'rgba(255,255,255,0.08)' },
          '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: 'rgba(255,255,255,0.12)' },
          '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: '#f97316' },
        },
      },
    },
    MuiMenu: {
      styleOverrides: {
        paper: { backgroundColor: '#18181b', border: '1px solid rgba(255,255,255,0.08)' },
      },
    },
    MuiMenuItem: {
      styleOverrides: {
        root: {
          '&:hover': { backgroundColor: 'rgba(255,255,255,0.06)' },
          '&.Mui-selected': { backgroundColor: 'rgba(249,115,22,0.12)' },
        },
      },
    },
    MuiTabs: {
      styleOverrides: {
        indicator: { backgroundColor: '#f97316' },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          color: '#a1a1aa',
          '&.Mui-selected': { color: '#fafafa' },
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: { backgroundColor: '#18181b', border: '1px solid rgba(255,255,255,0.08)' },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: { borderRadius: 8 },
        standardError: { backgroundColor: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.2)' },
        standardSuccess: { backgroundColor: 'rgba(34,197,94,0.1)', border: '1px solid rgba(34,197,94,0.2)' },
        standardWarning: { backgroundColor: 'rgba(245,158,11,0.1)', border: '1px solid rgba(245,158,11,0.2)' },
      },
    },
    MuiTablePagination: {
      styleOverrides: {
        root: { color: '#a1a1aa' },
      },
    },
    MuiSkeleton: {
      styleOverrides: {
        root: { backgroundColor: 'rgba(255,255,255,0.06)' },
      },
    },
  },
});

export default theme;
```

- [ ] **Step 2: Run typecheck to verify theme compiles**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/theme.ts
git commit -m "feat: rewrite MUI theme with premium dark palette and component overrides"
```

---

### Task 3: Add font imports and body styles

**Files:**
- Modify: `frontend/src/main.tsx`
- Modify: `frontend/src/index.css`

- [ ] **Step 1: Add JetBrains Mono import in main.tsx**

Add import line after the existing `@fontsource-variable/geist` import (line 1 of the original, or wherever the Geist import is — check actual file):

```typescript
import '@fontsource/jetbrains-mono';
```

**However**, the current `main.tsx` does not import Geist explicitly — it imports `./index.css` and the Geist font is likely configured in the vite config or HTML. Check if `@fontsource-variable/geist` is imported anywhere. If not found, add both imports to `main.tsx`:

```typescript
import '@fontsource-variable/geist';
import '@fontsource/jetbrains-mono';
```

- [ ] **Step 2: Update index.css to add monospace font family**

Replace `frontend/src/index.css`:

```css
body {
  margin: 0;
}

.mono-num {
  font-family: 'JetBrains Mono', monospace;
  font-variant-numeric: tabular-nums;
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/main.tsx frontend/src/index.css
git commit -m "feat: add JetBrains Mono font, monospace utility class"
```

---

### Task 4: Redesign AppLayout — sidebar icons, dark appbar, account chip

**Files:**
- Modify: `frontend/src/components/AppLayout.tsx`

- [ ] **Step 1: Add icon imports at top of AppLayout.tsx**

Add these imports after the existing MUI icon imports:

```typescript
import DashboardIcon from '@mui/icons-material/Dashboard';
import InventoryIcon from '@mui/icons-material/Inventory';
import ReceiptIcon from '@mui/icons-material/Receipt';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import SettingsIcon from '@mui/icons-material/Settings';
```

- [ ] **Step 2: Add icon key to navItems**

Replace the `navItems` array:

```typescript
const navItems = [
  { to: '/dashboard', label: '仪表盘', icon: <DashboardIcon fontSize="small" /> },
  { to: '/inventory', label: '持仓', icon: <InventoryIcon fontSize="small" /> },
  { to: '/trades/completed', label: '交易记录', icon: <ReceiptIcon fontSize="small" /> },
  { to: '/pnl', label: '盈亏', icon: <TrendingUpIcon fontSize="small" /> },
  { to: '/accounts', label: '账户管理', icon: <AccountBalanceIcon fontSize="small" /> },
  { to: '/settings', label: '设置', icon: <SettingsIcon fontSize="small" /> },
];
```

- [ ] **Step 3: Update the text field background from grey.100 to dark**

Find `bgcolor: 'grey.100'` in the search TextField sx and change to `bgcolor: '#18181b'`.

- [ ] **Step 4: Update ListItemButton in the Drawer to show icons**

Replace the `{navItems.map(...)}` block inside the sidebar with:

```typescript
{navItems.map((item) => (
  <ListItemButton
    key={item.to}
    component={NavLink}
    to={item.to}
    sx={{
      borderRadius: 1,
      mb: 0.5,
      gap: 1.5,
      '&.active': {
        bgcolor: 'rgba(249,115,22,0.12)',
        color: '#f97316',
        borderLeft: '3px solid #f97316',
        borderRadius: '0 6px 6px 0',
        pl: 1.5,
      },
      '&.active:hover': { bgcolor: 'rgba(249,115,22,0.18)' },
      '&:hover': { bgcolor: 'rgba(255,255,255,0.04)' },
    }}
  >
    <Box component="span" sx={{ color: 'inherit', display: 'flex', alignItems: 'center' }}>
      {item.icon}
    </Box>
    <ListItemText
      primary={item.label}
      primaryTypographyProps={{ fontSize: 13, fontWeight: 500 }}
    />
  </ListItemButton>
))}
```

- [ ] **Step 5: Update account Chip styling**

Replace the Chip's sx:

```typescript
sx={{
  cursor: 'pointer',
  maxWidth: 220,
  bgcolor: selectedAccount ? 'rgba(249,115,22,0.12)' : 'transparent',
  borderColor: selectedAccount ? 'rgba(249,115,22,0.3)' : 'rgba(255,255,255,0.12)',
  color: selectedAccount ? '#f97316' : '#d4d4d8',
}}
```

- [ ] **Step 6: Run typecheck**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/AppLayout.tsx
git commit -m "feat: redesign sidebar with icons, dark appbar, accent account chip"
```

---

### Task 5: Redesign PnlSummaryCards shared component

**Files:**
- Modify: `frontend/src/components/PnlSummaryCards.tsx`

- [ ] **Step 1: Replace PnlSummaryCards with dark gradient cards**

Replace the file:

```typescript
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import ReceiptIcon from '@mui/icons-material/Receipt';
import PaymentsIcon from '@mui/icons-material/Payments';
import SavingsIcon from '@mui/icons-material/Savings';
import { formatCNY, plColor } from '../lib/format';

const cardSx = {
  background: 'linear-gradient(135deg, #18181b, #1f1f23)',
  border: '1px solid rgba(255,255,255,0.08)',
  borderRadius: '10px',
  p: 2,
};

interface PnlSummaryCardsProps {
  totalTrades: number;
  totalGrossPl: number;
  totalFee: number;
  totalNetPl: number;
}

export default function PnlSummaryCards({
  totalTrades,
  totalGrossPl,
  totalFee,
  totalNetPl,
}: PnlSummaryCardsProps) {
  return (
    <Grid container spacing={1.5}>
      <Grid item xs={3}>
        <Box sx={cardSx}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography variant="caption" color="text.disabled" sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              总交易数
            </Typography>
            <Box sx={{ width: 32, height: 32, borderRadius: 1, bgcolor: 'rgba(255,255,255,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <ReceiptIcon fontSize="small" sx={{ color: '#a1a1aa' }} />
            </Box>
          </Box>
          <Typography variant="h5" mt={1} fontWeight={700}>{totalTrades}</Typography>
        </Box>
      </Grid>
      <Grid item xs={3}>
        <Box sx={cardSx}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography variant="caption" color="text.disabled" sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              毛利
            </Typography>
            <Box sx={{ width: 32, height: 32, borderRadius: 1, bgcolor: 'rgba(245,158,11,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <TrendingUpIcon fontSize="small" sx={{ color: '#f59e0b' }} />
            </Box>
          </Box>
          <Typography variant="h5" mt={1} fontWeight={700} color={plColor(totalGrossPl)}>
            {formatCNY(totalGrossPl)}
          </Typography>
        </Box>
      </Grid>
      <Grid item xs={3}>
        <Box sx={cardSx}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography variant="caption" color="text.disabled" sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              手续费
            </Typography>
            <Box sx={{ width: 32, height: 32, borderRadius: 1, bgcolor: 'rgba(255,255,255,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <PaymentsIcon fontSize="small" sx={{ color: '#a1a1aa' }} />
            </Box>
          </Box>
          <Typography variant="h5" mt={1} fontWeight={700}>{formatCNY(totalFee)}</Typography>
        </Box>
      </Grid>
      <Grid item xs={3}>
        <Box sx={{
          ...cardSx,
          borderLeft: '3px solid',
          borderLeftColor: totalNetPl >= 0 ? '#22c55e' : '#ef4444',
        }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography variant="caption" color="text.disabled" sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              净利润
            </Typography>
            <Box sx={{ width: 32, height: 32, borderRadius: 1, bgcolor: totalNetPl >= 0 ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <SavingsIcon fontSize="small" sx={{ color: totalNetPl >= 0 ? '#22c55e' : '#ef4444' }} />
            </Box>
          </Box>
          <Typography variant="h5" mt={1} fontWeight={700} color={plColor(totalNetPl)}>
            {formatCNY(totalNetPl)}
          </Typography>
        </Box>
      </Grid>
    </Grid>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/PnlSummaryCards.tsx
git commit -m "feat: redesign P&L summary cards with dark gradients and icons"
```

---

### Task 6: Redesign Dashboard page stat cards

**Files:**
- Modify: `frontend/src/pages/DashboardPage.tsx`

- [ ] **Step 1: Add icon imports to DashboardPage**

```typescript
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import LockIcon from '@mui/icons-material/Lock';
import BoltIcon from '@mui/icons-material/Bolt';
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import InventoryIcon from '@mui/icons-material/Inventory';
import ReceiptIcon from '@mui/icons-material/Receipt';
import RedeemIcon from '@mui/icons-material/Redeem';
```

- [ ] **Step 2: Add stat card helper component above DashboardPage**

Add this component inside the file, before the `export default function DashboardPage()`:

```typescript
function StatCard({
  label,
  value,
  color,
  icon,
  accentLeft,
}: {
  label: string;
  value: React.ReactNode;
  color?: string;
  icon: React.ReactNode;
  accentLeft?: string;
}) {
  return (
    <Box
      sx={{
        background: 'linear-gradient(135deg, #18181b, #1f1f23)',
        border: '1px solid rgba(255,255,255,0.08)',
        borderRadius: '10px',
        p: 2,
        ...(accentLeft ? { borderLeft: `3px solid ${accentLeft}` } : {}),
      }}
    >
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <Typography variant="caption" color="text.disabled" sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          {label}
        </Typography>
        <Box sx={{ width: 32, height: 32, borderRadius: 1, bgcolor: 'rgba(255,255,255,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          {icon}
        </Box>
      </Box>
      <Typography variant="h5" mt={1} fontWeight={700} color={color}>
        {value}
      </Typography>
    </Box>
  );
}
```

- [ ] **Step 3: Replace all Card/CardContent blocks with StatCard**

For each card in the Dashboard, replace `<Grid item xs={...}><Card><CardContent>...</CardContent></Card></Grid>` with `<Grid item xs={...}><StatCard ... /></Grid>`.

First row (balance):
```typescript
<Grid item xs={3}>
  <StatCard
    label="钱包余额"
    value={formatCNY(data.totalAvailableBalance)}
    icon={<AccountBalanceWalletIcon fontSize="small" sx={{ color: '#a1a1aa' }} />}
  />
</Grid>
<Grid item xs={3}>
  <StatCard
    label="冻结余额"
    value={formatCNY(data.totalFrozenBalance)}
    icon={<LockIcon fontSize="small" sx={{ color: '#a1a1aa' }} />}
  />
</Grid>
<Grid item xs={3}>
  <StatCard
    label="秒到账余额"
    value={formatCNY(data.totalInstantBalance)}
    icon={<BoltIcon fontSize="small" sx={{ color: '#a1a1aa' }} />}
  />
</Grid>
<Grid item xs={3}>
  <StatCard
    label="求购余额"
    value={formatCNY(data.totalPurchaseBalance)}
    icon={<ShoppingCartIcon fontSize="small" sx={{ color: '#a1a1aa' }} />}
  />
</Grid>
```

Second row (P&L + counts):
```typescript
<Grid item xs={3}>
  <StatCard
    label="已实现盈亏"
    value={formatCNY(data.realizedPl)}
    color={plColor(data.realizedPl)}
    accentLeft={data.realizedPl >= 0 ? '#22c55e' : '#ef4444'}
    icon={<TrendingUpIcon fontSize="small" sx={{ color: plHexColor(data.realizedPl) }} />}
  />
</Grid>
<Grid item xs={3}>
  <StatCard
    label="持仓物品"
    value={data.inventoryCount}
    icon={<InventoryIcon fontSize="small" sx={{ color: '#a1a1aa' }} />}
  />
</Grid>
<Grid item xs={3}>
  <StatCard
    label="已完成交易"
    value={data.completedTrades}
    icon={<ReceiptIcon fontSize="small" sx={{ color: '#a1a1aa' }} />}
  />
</Grid>
<Grid item xs={3}>
  <StatCard
    label="租赁收入"
    value={formatCNY(data.totalRentalIncome)}
    icon={<RedeemIcon fontSize="small" sx={{ color: '#f97316' }} />}
  />
</Grid>
```

Third row (cost + market value + unrealized):
```typescript
<Grid item xs={4}>
  <StatCard
    label="持仓成本"
    value={formatCNY(data.inventoryCost)}
    icon={<PaymentsIcon fontSize="small" sx={{ color: '#a1a1aa' }} />}
  />
</Grid>
<Grid item xs={4}>
  <StatCard
    label={`持仓市值（${priceSourceLabel[data.priceSource] ?? data.priceSource}）`}
    value={formatCNY(data.inventoryMarketValue)}
    icon={<ShowChartIcon fontSize="small" sx={{ color: '#a1a1aa' }} />}
  />
</Grid>
<Grid item xs={4}>
  <StatCard
    label="未实现盈亏"
    value={formatCNY(data.inventoryMarketValue - data.inventoryCost)}
    color={plColor(data.inventoryMarketValue - data.inventoryCost)}
    accentLeft={(data.inventoryMarketValue - data.inventoryCost) >= 0 ? '#22c55e' : '#ef4444'}
    icon={<TrendingUpIcon fontSize="small" sx={{ color: plHexColor(data.inventoryMarketValue - data.inventoryCost) }} />}
  />
</Grid>
```

- [ ] **Step 4: Remove unused MUI imports (Card, CardContent) from DashboardPage**

Remove `Card` and `CardContent` from the MUI imports. Add `Box` if not already imported, and add the new icon imports.

- [ ] **Step 5: Run typecheck**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/DashboardPage.tsx
git commit -m "feat: redesign Dashboard with dark gradient stat cards and icons"
```

---

### Task 7: Update PnLPage — dark chart theme

**Files:**
- Modify: `frontend/src/pages/PnLPage.tsx`

- [ ] **Step 1: Update ECharts color constants**

Replace the color constants at the top of PnLPage:

```typescript
const POS_COLOR = '#22c55e';
const NEG_COLOR = '#ef4444';
const LINE_COLOR = '#f97316';
```

- [ ] **Step 2: Update chart option for dark theme**

Replace the `chartOption` useMemo block. Key changes: dark background, lighter grid, updated tooltip, orange line color.

The full chartOption replacement:

```typescript
const chartOption = useMemo(() => {
  if (monthly.length === 0) return null;

  const sorted = [...monthly].sort((a, b) => a.month.localeCompare(b.month));
  const months = sorted.map((m) => m.month);
  const values = sorted.map((m) => m.netPl);

  const cumulative: number[] = [];
  let cumSum = 0;
  for (const v of values) {
    cumSum += v;
    cumulative.push(cumSum / 100);
  }

  const barValues = values.map((v) => v / 100);

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#18181b',
      borderColor: 'rgba(255,255,255,0.12)',
      borderWidth: 1,
      textStyle: { color: '#fafafa', fontSize: 13 },
      axisPointer: {
        type: 'cross',
        crossStyle: { color: '#52525b' },
        lineStyle: { color: '#52525b', type: 'dashed' },
      },
    },
    legend: {
      bottom: 8,
      textStyle: { fontSize: 12, color: '#a1a1aa' },
      itemWidth: 14,
      itemHeight: 10,
    },
    grid: {
      top: 16,
      left: 64,
      right: 64,
      bottom: 48,
    },
    xAxis: {
      type: 'category',
      data: months,
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.08)' } },
      axisTick: { show: false },
      axisLabel: { color: '#a1a1aa', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.06)' } },
      axisLabel: {
        fontSize: 11,
        color: '#a1a1aa',
        formatter: (v: number) => {
          if (Math.abs(v) >= 10000) return `¥${(v / 10000).toFixed(1)}w`;
          return `¥${v.toFixed(0)}`;
        },
      },
    },
    series: [
      {
        name: 'Net P/L',
        type: 'bar',
        data: barValues,
        barWidth: '55%',
        itemStyle: {
          borderRadius: [4, 4, 0, 0],
          color: (p: { value: number }) => (p.value >= 0 ? POS_COLOR : NEG_COLOR),
        },
      },
      {
        name: 'Cumulative P/L',
        type: 'line',
        data: cumulative,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        showSymbol: false,
        lineStyle: { color: LINE_COLOR, width: 2.5 },
        itemStyle: { color: LINE_COLOR },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(249,115,22,0.15)' },
            { offset: 1, color: 'rgba(249,115,22,0.02)' },
          ]),
        },
      },
    ],
    dataZoom: [
      {
        type: 'slider',
        height: 20,
        bottom: 4,
        borderColor: 'transparent',
        backgroundColor: 'rgba(255,255,255,0.04)',
        fillerColor: 'rgba(249,115,22,0.1)',
        handleStyle: { color: LINE_COLOR, borderColor: LINE_COLOR },
        textStyle: { fontSize: 10, color: '#a1a1aa' },
      },
      { type: 'inside' },
    ],
  };
}, [monthly]);
```

- [ ] **Step 3: Run typecheck**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/PnLPage.tsx
git commit -m "feat: update P&L chart to dark theme with orange accent"
```

---

### Task 8: Dark component pass — shared UI components

**Files:**
- Modify: `frontend/src/components/EmptyState.tsx`
- Modify: `frontend/src/components/PageSearchBar.tsx`
- Modify: `frontend/src/components/ErrorBanner.tsx`
- Modify: `frontend/src/components/Dialog.tsx`

- [ ] **Step 1: Update EmptyState to use MUI icons instead of emoji strings**

Replace `frontend/src/components/EmptyState.tsx`:

```typescript
import { type ReactNode } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';

interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description?: string;
  action?: { label: string; onClick: () => void };
}

export default function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <Box
      display="flex"
      flexDirection="column"
      alignItems="center"
      justifyContent="center"
      py={8}
      textAlign="center"
    >
      {icon && (
        <Box mb={2} sx={{ color: 'text.disabled' }}>
          {icon}
        </Box>
      )}
      <Typography variant="h6" color="text.primary">
        {title}
      </Typography>
      {description && (
        <Typography variant="body2" color="text.disabled" mt={1} maxWidth={400}>
          {description}
        </Typography>
      )}
      {action && (
        <Button variant="outlined" onClick={action.onClick} sx={{ mt: 3 }}>
          {action.label}
        </Button>
      )}
    </Box>
  );
}
```

- [ ] **Step 2: Update PageSearchBar — dark input background**

Replace the sx prop in `frontend/src/components/PageSearchBar.tsx`:

```typescript
sx={{
  width: 260,
  '& .MuiOutlinedInput-root': { bgcolor: '#18181b' },
}}
```

- [ ] **Step 3: Update ErrorBanner — remove color="inherit" from buttons (theme handles it now)**

In `frontend/src/components/ErrorBanner.tsx`, change both `color="inherit"` on the Retry and Dismiss buttons to `color="inherit"` (keep as-is — MUI Alert already handles dark mode via the theme override). No change needed here.

- [ ] **Step 4: Update Dialog — add dividers and better dark styling**

Replace `frontend/src/components/Dialog.tsx`:

```typescript
import { type ReactNode } from 'react';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';

interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  actions?: ReactNode;
}

export default function AppDialog({ open, onClose, title, children, actions }: DialogProps) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ fontWeight: 600 }}>{title}</DialogTitle>
      <DialogContent dividers sx={{ borderColor: 'rgba(255,255,255,0.08)' }}>
        {children}
      </DialogContent>
      {actions && <DialogActions>{actions}</DialogActions>}
    </Dialog>
  );
}
```

- [ ] **Step 5: Update imports in AccountsPage that references 'Dialog'**

In `frontend/src/pages/AccountsPage.tsx`, change:
```typescript
import Dialog from '../components/Dialog';
```
to:
```typescript
import AppDialog from '../components/Dialog';
```

And update all `<Dialog ...` usages to `<AppDialog ...`.

- [ ] **Step 6: Update AddAccountDialog.tsx** 

No changes needed — it uses MUI Dialog directly, which is handled by the theme override.

- [ ] **Step 7: Run typecheck**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 8: Commit**

```bash
git add frontend/src/components/EmptyState.tsx frontend/src/components/PageSearchBar.tsx frontend/src/components/Dialog.tsx frontend/src/pages/AccountsPage.tsx
git commit -m "feat: dark theme pass on EmptyState, PageSearchBar, Dialog components"
```

---

### Task 9: Dark table pass — Inventory, CompletedTrades, Accounts pages

**Files:**
- Modify: `frontend/src/pages/InventoryPage.tsx`
- Modify: `frontend/src/pages/CompletedTradesPage.tsx`
- Modify: `frontend/src/pages/AccountsPage.tsx`
- Modify: `frontend/src/pages/InventoryDetailPage.tsx`
- Modify: `frontend/src/pages/SettingsPage.tsx`

- [ ] **Step 1: Inventory page — add monospace to numeric cells, remove grey.100 from filter select**

In `InventoryPage.tsx`:
- Find `sx={{ bgcolor: 'grey.100' }}` on the type filter Select and change to `sx={{ bgcolor: '#18181b' }}`
- Add `className="mono-num"` to Typography components wrapping numeric values in table cells (prices, quantities, P&L values)

- [ ] **Step 2: CompletedTrades page — same treatment**

In `CompletedTradesPage.tsx`:
- Add `className="mono-num"` to all Typography components wrapping monetary values
- The `TradeDetailDialog` and `UnmatchedSellDetailDialog` are already using MUI Dialog which is themed — no changes needed

- [ ] **Step 3: Accounts page action buttons — switch to outlined variant for dark**

In `AccountsPage.tsx`, for the action buttons ("同步", "编辑", "删除"), change their `variant` to `"outlined"`:

```typescript
<Button size="small" variant="outlined" onClick={...}>同步</Button>
<Button size="small" variant="outlined" onClick={...}>编辑</Button>
<Button size="small" variant="outlined" color="error" onClick={...}>删除</Button>
```

- [ ] **Step 4: InventoryDetailPage — remove outlined variant="outlined" from cards (theme handles it)**

In `InventoryDetailPage.tsx`, change `<Card variant="outlined">` to `<Card>` for rental summary cards. The theme override handles all cards uniformly.

- [ ] **Step 5: SettingsPage — remove explicit variant styling (theme handles it)**

No code changes needed — the Select and TextField use default MUI styling which the theme now overrides. The `sx` props for explicit styling can stay as-is since they're local overrides.

- [ ] **Step 6: Run typecheck**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/InventoryPage.tsx frontend/src/pages/CompletedTradesPage.tsx frontend/src/pages/AccountsPage.tsx frontend/src/pages/InventoryDetailPage.tsx
git commit -m "feat: dark table and page polish across all pages"
```

---

### Task 10: Final verification — typecheck, tests, visual check

**Files:** None (verification only)

- [ ] **Step 1: Run full typecheck**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

- [ ] **Step 2: Run tests**

```bash
cd frontend && npx vitest run
```
Expected: All tests pass.

- [ ] **Step 3: Run dev server for visual check**

```bash
cd frontend && npx vite --host 0.0.0.0 &
```
Check http://localhost:5173 in browser. Verify:
- Dark background on all pages
- Sidebar has icons with orange active accent
- Dashboard has gradient cards with icons
- P&L page has dark chart
- Tables are dark themed
- Dialogs are dark themed

- [ ] **Step 4: Commit any final tweaks**

If visual check reveals issues, fix and commit them.
```
