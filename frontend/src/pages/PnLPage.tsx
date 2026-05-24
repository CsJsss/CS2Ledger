import { useState, useMemo } from 'react';
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
