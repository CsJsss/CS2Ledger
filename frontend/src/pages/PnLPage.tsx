import { useState, useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Skeleton from '@mui/material/Skeleton';
import Box from '@mui/material/Box';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import ErrorBanner from '../components/ErrorBanner';
import PnlSummaryCards from '../components/PnlSummaryCards';
import { usePnlSummary } from '../hooks/usePnlSummary';
import { useMonthlyBreakdown } from '../hooks/useMonthlyBreakdown';
import { useUIStore } from '../store/uiStore';
import ReactEChartsCore from 'echarts-for-react/lib/core';
import * as echarts from 'echarts/core';
import { BarChart, LineChart } from 'echarts/charts';
import {
  GridComponent,
  TooltipComponent,
  DataZoomComponent,
  LegendComponent,
} from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';

echarts.use([
  BarChart,
  LineChart,
  GridComponent,
  TooltipComponent,
  DataZoomComponent,
  LegendComponent,
  CanvasRenderer,
]);

const POS_COLOR = '#22c55e';
const NEG_COLOR = '#ef4444';
const LINE_COLOR = '#f97316';

export default function PnLPage() {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const currentYear = new Date().getFullYear();
  const [dismissed, setDismissed] = useState(false);
  const [year, setYear] = useState<number | 'all'>('all');

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = usePnlSummary(selectedAccountId);

  const {
    data: monthly = [],
    isLoading: monthlyLoading,
    error: monthlyError,
  } = useMonthlyBreakdown(selectedAccountId, year === 'all' ? 0 : year);

  const isLoading = summaryLoading || monthlyLoading;
  const error = summaryError || monthlyError;

  const years: (number | 'all')[] = ['all'];
  for (let y = currentYear; y >= currentYear - 3; y--) years.push(y);

  const chartOption = useMemo(() => {
    if (monthly.length === 0) return null;

    const sorted = [...monthly].sort((a, b) => a.month.localeCompare(b.month));
    const months = sorted.map((m) => m.month);
    const values = sorted.map((m) => m.netPl);

    // Cumulative P&L in yuan
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
        backgroundColor: isDark ? '#18181b' : '#ffffff',
        borderColor: isDark ? 'rgba(255,255,255,0.12)' : '#e2e8f0',
        borderWidth: 1,
        textStyle: { color: isDark ? '#fafafa' : '#0f172a', fontSize: 13 },
        axisPointer: {
          type: 'cross',
          crossStyle: { color: isDark ? '#52525b' : '#cbd5e1' },
          lineStyle: { color: isDark ? '#52525b' : '#cbd5e1', type: 'dashed' },
        },
      },
      legend: {
        bottom: 8,
        textStyle: { fontSize: 12, color: isDark ? '#a1a1aa' : '#64748b' },
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
        axisLine: { lineStyle: { color: isDark ? 'rgba(255,255,255,0.08)' : '#e2e8f0' } },
        axisTick: { show: false },
        axisLabel: { color: isDark ? '#a1a1aa' : '#64748b', fontSize: 11 },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: isDark ? 'rgba(255,255,255,0.06)' : '#f1f5f9' } },
        axisLabel: {
          fontSize: 11,
          color: isDark ? '#a1a1aa' : '#64748b',
          formatter: (v: number) => {
            if (Math.abs(v) >= 10000) return `¥${(v / 10000).toFixed(1)}w`;
            return `¥${v.toFixed(0)}`;
          },
        },
      },
      series: [
        {
          name: '月度净盈亏',
          type: 'bar',
          data: barValues,
          barWidth: '55%',
          itemStyle: {
            borderRadius: [4, 4, 0, 0],
            color: (p: { value: number }) => (p.value >= 0 ? POS_COLOR : NEG_COLOR),
          },
        },
        {
          name: '累计盈亏',
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
          backgroundColor: isDark ? 'rgba(255,255,255,0.04)' : '#f1f5f9',
          fillerColor: 'rgba(249,115,22,0.1)',
          handleStyle: { color: LINE_COLOR, borderColor: LINE_COLOR },
          textStyle: { fontSize: 10, color: isDark ? '#a1a1aa' : '#64748b' },
        },
        { type: 'inside' },
      ],
    };
  }, [monthly, isDark]);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        盈亏
      </Typography>

      {isLoading && (
        <Box mt={3}>
          <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
            {[1, 2, 3, 4].map((i) => (
              <Skeleton
                key={i}
                variant="rectangular"
                height={96}
                sx={{ flex: 1, borderRadius: 1 }}
              />
            ))}
          </Box>
          <Skeleton variant="rectangular" height={400} sx={{ borderRadius: 1 }} />
        </Box>
      )}

      {error && !dismissed && (
        <Box mt={3}>
          <ErrorBanner
            message={`加载盈亏数据失败: ${String(error)}`}
            onRetry={() => {
              setDismissed(false);
              void refetchSummary();
            }}
            onDismiss={() => setDismissed(true)}
          />
        </Box>
      )}

      {!isLoading && !error && summary && (
        <Box mt={3}>
          <PnlSummaryCards
            totalTrades={summary.totalTrades}
            totalGrossPl={summary.totalGrossPl}
            totalFee={summary.totalFee}
            totalNetPl={summary.totalNetPl}
          />

          <Card sx={{ mt: 3 }}>
            <CardContent>
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  mb: 2,
                  flexWrap: 'wrap',
                  gap: 1,
                }}
              >
                <Typography variant="h6">盈亏图表</Typography>
                <Box sx={{ display: 'flex', gap: 1 }}>
                  <FormControl size="small" sx={{ minWidth: 100 }}>
                    <Select
                      value={year}
                      onChange={(e) => setYear(e.target.value as number | 'all')}
                    >
                      {years.map((y) => (
                        <MenuItem key={String(y)} value={y}>
                          {y === 'all' ? '全部' : String(y)}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Box>
              </Box>

              {monthly.length === 0 && (
                <Typography color="text.secondary" textAlign="center" py={4}>
                  {year === 'all' ? '全部年份' : year} 暂无数据
                </Typography>
              )}

              {monthly.length > 0 && chartOption && (
                <Box sx={{ height: 500 }}>
                  <ReactEChartsCore
                    echarts={echarts}
                    option={chartOption}
                    style={{ height: '100%', width: '100%' }}
                    notMerge
                  />
                </Box>
              )}
            </CardContent>
          </Card>
        </Box>
      )}
    </Box>
  );
}
