import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router';
import Card from '@mui/material/Card';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Skeleton from '@mui/material/Skeleton';
import Box from '@mui/material/Box';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import LockIcon from '@mui/icons-material/Lock';
import BoltIcon from '@mui/icons-material/Bolt';
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import InventoryIcon from '@mui/icons-material/Inventory';
import ReceiptIcon from '@mui/icons-material/Receipt';
import RedeemIcon from '@mui/icons-material/Redeem';
import PaymentsIcon from '@mui/icons-material/Payments';
import DashboardIcon from '@mui/icons-material/Dashboard';
import ErrorBanner from '../components/ErrorBanner';
import EmptyState from '../components/EmptyState';
import { useDashboard } from '../hooks/useDashboard';
import { useMonthlyBreakdown } from '../hooks/useMonthlyBreakdown';
import { useInventory } from '../hooks/useInventory';
import { useUIStore } from '../store/uiStore';
import { formatCNY, plColor, plHexColor } from '../lib/format';
import { priceSourceLabel } from '../lib/constants';
import ShowChartIcon from '@mui/icons-material/ShowChart';
import ReactEChartsCore from 'echarts-for-react/lib/core';
import * as echarts from 'echarts/core';
import { BarChart, LineChart, PieChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';

echarts.use([
  BarChart,
  LineChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer,
]);

const POS_COLOR = '#22c55e';
const NEG_COLOR = '#ef4444';

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
    <Card
      sx={{
        borderRadius: '10px',
        p: 2,
        ...(accentLeft ? { borderLeft: `3px solid ${accentLeft}` } : {}),
      }}
    >
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <Typography
          variant="caption"
          color="text.disabled"
          sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}
        >
          {label}
        </Typography>
        <Box
          sx={{
            width: 32,
            height: 32,
            borderRadius: 1,
            bgcolor: 'action.hover',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {icon}
        </Box>
      </Box>
      <Typography variant="h5" mt={1} fontWeight={700} color={color}>
        {value}
      </Typography>
    </Card>
  );
}

export default function DashboardPage() {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useState(false);
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const { data, isLoading, error, refetch } = useDashboard();

  const currentYear = new Date().getFullYear();
  const { data: monthly = [] } = useMonthlyBreakdown(selectedAccountId, currentYear);

  const { data: chartData } = useInventory(null, {
    page: 1,
    pageSize: 200,
    weaponType: '',
    sortBy: 'itemName',
    sortDir: 'asc',
  });

  const marketChartOption = useMemo(() => {
    const allGroups = chartData?.groups ?? [];
    if (allGroups.length === 0) return null;

    const typeMap = new Map<string, number>();
    let grandTotal = 0;
    for (const g of allGroups) {
      if (g.marketPrice == null) continue;
      const wt = g.weaponType || '其他';
      const mv = g.marketPrice * g.totalQuantity;
      typeMap.set(wt, (typeMap.get(wt) ?? 0) + mv);
      grandTotal += mv;
    }
    if (grandTotal === 0) return null;

    const types = Array.from(typeMap.keys());
    const colors = [
      '#f97316',
      '#14b8a6',
      '#3b82f6',
      '#a855f7',
      '#eab308',
      '#ec4899',
      '#06b6d4',
      '#84cc16',
      '#f43f5e',
      '#8b5cf6',
      '#22d3ee',
      '#f59e0b',
    ];

    return {
      totalCents: grandTotal,
      option: {
        color: colors,
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'item' as const,
          backgroundColor: '#18181b',
          borderColor: '#27272a',
          textStyle: { color: '#d4d4d8', fontSize: 13, fontFamily: "'Geist Variable', sans-serif" },
          formatter: (p: { name: string; value: number; percent: number }) =>
            `${p.name}<br/><b>¥${p.value.toLocaleString()}</b>  ${p.percent}%`,
        },
        series: [
          {
            type: 'pie',
            radius: ['55%', '78%'],
            center: ['50%', '50%'],
            selectedMode: 'single',
            selectedOffset: 4,
            itemStyle: { borderColor: '#09090b', borderWidth: 2, borderRadius: 2 },
            data: types.map((t) => ({ name: t, value: typeMap.get(t)! / 100 })),
            label: {
              show: true,
              position: 'outside' as const,
              formatter: '{b} {d}%',
              fontFamily: "'Geist Variable', sans-serif",
              fontSize: 12,
              color: '#a1a1aa',
            },
            labelLine: { lineStyle: { color: '#3f3f46' } },
            emphasis: {
              label: { show: true, fontSize: 14, fontWeight: 600 },
              scaleSize: 6,
            },
          },
        ],
      },
    };
  }, [chartData?.groups]);

  const netWorth = data
    ? data.totalAvailableBalance +
      data.totalFrozenBalance +
      data.totalPurchaseBalance +
      data.totalInstantBalance +
      data.inventoryMarketValue
    : 0;

  const chartOption = useMemo(() => {
    if (monthly.length === 0) return null;
    const sorted = [...monthly].sort((a, b) => a.month.localeCompare(b.month));
    const months = sorted.map((m) => m.month.slice(5)); // "MM" only
    const values = sorted.map((m) => m.netPl / 100);

    return {
      backgroundColor: 'transparent',
      grid: { top: 8, left: 0, right: 0, bottom: 0 },
      xAxis: {
        type: 'category',
        data: months,
        show: false,
      },
      yAxis: { type: 'value', show: false },
      series: [
        {
          type: 'bar',
          data: values,
          barWidth: '60%',
          itemStyle: {
            borderRadius: [2, 2, 0, 0],
            color: (p: { value: number }) => (p.value >= 0 ? POS_COLOR : NEG_COLOR),
          },
        },
      ],
    };
  }, [monthly]);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        仪表盘
      </Typography>

      {isLoading && (
        <Grid container spacing={2} mt={1}>
          {[1, 2, 3, 4].map((i) => (
            <Grid item xs={3} key={i}>
              <Card sx={{ borderRadius: '10px', p: 2 }}>
                <Skeleton width="60%" />
                <Skeleton height={40} width="40%" sx={{ mt: 1 }} />
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {error && !dismissed && (
        <Box mt={3}>
          <ErrorBanner
            message={`加载仪表盘失败: ${String(error)}`}
            onRetry={() => {
              setDismissed(false);
              void refetch();
            }}
            onDismiss={() => setDismissed(true)}
          />
        </Box>
      )}

      {!isLoading && !error && data && data.inventoryCount === 0 && data.completedTrades === 0 && (
        <Box mt={3}>
          <EmptyState
            icon={<DashboardIcon sx={{ fontSize: 48 }} />}
            title="暂无数据"
            description="添加账户并同步以查看仪表盘。"
            action={{
              label: '前往账户管理',
              onClick: () => {
                void navigate('/accounts');
              },
            }}
          />
        </Box>
      )}

      {!isLoading &&
        !error &&
        data &&
        !(
          data.inventoryCount === 0 &&
          data.completedTrades === 0 &&
          data.totalAvailableBalance === 0 &&
          data.totalFrozenBalance === 0 &&
          data.totalInstantBalance === 0 &&
          data.totalPurchaseBalance === 0
        ) && (
          <>
            {/* Net Worth Hero */}
            <Card
              sx={{
                borderRadius: '10px',
                p: 2.5,
                mb: 2,
                display: 'flex',
                alignItems: 'center',
                gap: 3,
                borderLeft: '3px solid #f97316',
              }}
            >
              <Box
                sx={{
                  width: 48,
                  height: 48,
                  borderRadius: '12px',
                  bgcolor: 'rgba(249,115,22,0.1)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  flexShrink: 0,
                }}
              >
                <AccountBalanceWalletIcon sx={{ color: '#f97316' }} />
              </Box>
              <Box sx={{ flex: 1 }}>
                <Typography
                  variant="caption"
                  color="text.disabled"
                  sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}
                >
                  Net Worth · 净资产
                </Typography>
                <Typography variant="h4" fontWeight={700}>
                  {formatCNY(netWorth)}
                </Typography>
              </Box>
              {chartOption && (
                <Box sx={{ width: 200, height: 56, flexShrink: 0 }}>
                  <ReactEChartsCore
                    echarts={echarts}
                    option={chartOption}
                    style={{ height: '100%', width: '100%' }}
                    notMerge
                  />
                </Box>
              )}
            </Card>

            {/* Balance row */}
            <Grid container spacing={2}>
              <Grid item xs={3}>
                <StatCard
                  label="钱包余额"
                  value={formatCNY(data.totalAvailableBalance)}
                  icon={<AccountBalanceWalletIcon fontSize="small" color="action" />}
                />
              </Grid>
              <Grid item xs={3}>
                <StatCard
                  label="冻结余额"
                  value={formatCNY(data.totalFrozenBalance)}
                  icon={<LockIcon fontSize="small" color="action" />}
                />
              </Grid>
              <Grid item xs={3}>
                <StatCard
                  label="秒到账余额"
                  value={formatCNY(data.totalInstantBalance)}
                  icon={<BoltIcon fontSize="small" color="action" />}
                />
              </Grid>
              <Grid item xs={3}>
                <StatCard
                  label="求购余额"
                  value={formatCNY(data.totalPurchaseBalance)}
                  icon={<ShoppingCartIcon fontSize="small" color="action" />}
                />
              </Grid>
            </Grid>

            {/* P&L + counts row */}
            <Grid container spacing={2} mt={2}>
              <Grid item xs={3}>
                <StatCard
                  label="已实现盈亏"
                  value={formatCNY(data.realizedPl)}
                  color={plColor(data.realizedPl)}
                  accentLeft={data.realizedPl >= 0 ? '#22c55e' : '#ef4444'}
                  icon={
                    <TrendingUpIcon fontSize="small" sx={{ color: plHexColor(data.realizedPl) }} />
                  }
                />
              </Grid>
              <Grid item xs={3}>
                <StatCard
                  label="持仓物品"
                  value={data.inventoryCount}
                  icon={<InventoryIcon fontSize="small" color="action" />}
                />
              </Grid>
              <Grid item xs={3}>
                <StatCard
                  label="已完成交易"
                  value={data.completedTrades}
                  icon={<ReceiptIcon fontSize="small" color="action" />}
                />
              </Grid>
              <Grid item xs={3}>
                <StatCard
                  label="租赁收入"
                  value={formatCNY(data.totalRentalIncome)}
                  icon={<RedeemIcon fontSize="small" sx={{ color: '#f97316' }} />}
                />
              </Grid>
            </Grid>

            {/* Cost / Market Value / P&L + Chart row */}
            <Grid container spacing={2} mt={2}>
              <Grid item xs={4}>
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                  <StatCard
                    label="持仓成本"
                    value={formatCNY(data.inventoryCost)}
                    icon={<PaymentsIcon fontSize="small" color="action" />}
                  />
                  <StatCard
                    label={`持仓市值（${priceSourceLabel[data.priceSource] ?? data.priceSource}）`}
                    value={formatCNY(data.inventoryMarketValue)}
                    icon={<ShowChartIcon fontSize="small" color="action" />}
                  />
                  <StatCard
                    label="未实现盈亏"
                    value={formatCNY(data.inventoryMarketValue - data.inventoryCost)}
                    color={plColor(data.inventoryMarketValue - data.inventoryCost)}
                    accentLeft={
                      data.inventoryMarketValue - data.inventoryCost >= 0 ? '#22c55e' : '#ef4444'
                    }
                    icon={
                      <TrendingUpIcon
                        fontSize="small"
                        sx={{
                          color: plHexColor(data.inventoryMarketValue - data.inventoryCost),
                        }}
                      />
                    }
                  />
                </Box>
              </Grid>
              <Grid item xs={8}>
                {marketChartOption && (
                  <Card sx={{ borderRadius: '10px', p: 2, height: '100%' }}>
                    <Typography
                      variant="overline"
                      color="text.disabled"
                      sx={{ letterSpacing: '0.08em' }}
                    >
                      持仓市值分布
                    </Typography>
                    <Typography variant="h6" fontWeight={600} mt={0.5}>
                      {formatCNY(marketChartOption.totalCents)}
                    </Typography>
                    <Box sx={{ height: 220, mt: 1 }}>
                      <ReactEChartsCore
                        echarts={echarts}
                        option={marketChartOption.option}
                        style={{ height: '100%', width: '100%' }}
                        notMerge
                      />
                    </Box>
                  </Card>
                )}
              </Grid>
            </Grid>
          </>
        )}
    </Box>
  );
}
