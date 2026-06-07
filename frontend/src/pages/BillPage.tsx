import { useState, useMemo, useEffect } from 'react';
import { type ColumnDef } from '@tanstack/react-table';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select, { type SelectChangeEvent } from '@mui/material/Select';
import Skeleton from '@mui/material/Skeleton';
import TextField from '@mui/material/TextField';
import TablePagination from '@mui/material/TablePagination';
import SortableTable from '../components/SortableTable';
import PageSearchBar from '../components/PageSearchBar';
import ErrorBanner from '../components/ErrorBanner';
import EmptyState from '../components/EmptyState';
import { useBillRecords, useBillChartData } from '../hooks/useBillRecords';
import { useAccounts } from '../hooks/useAccounts';
import { useUIStore } from '../store/uiStore';
import { platformLabel, PLATFORM_OPTIONS } from '../lib/constants';
import { formatCNY } from '../lib/format';
import type { model } from '../lib/wails';
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

// Color mapping for internal BillType constants.
// TypeName (platform's original label) is always displayed as the Chip label.
// When TypeID == 99 (BillTypeOther), the platform-specific TypeName is shown as-is.
const TYPE_COLORS: Record<number, 'default' | 'success' | 'error' | 'warning' | 'info'> = {
  1: 'error',
  2: 'success',
  3: 'success',
  4: 'success',
  5: 'error',
  6: 'info',
  7: 'warning',
  8: 'info',
  9: 'info',
  10: 'success',
  99: 'default',
};

const TYPE_LABELS: Record<number, string> = {
  1: '购买',
  2: '出售',
  3: '收取租金',
  4: '收取续租资金',
  5: '租赁服务费',
  6: '充值',
  7: '提现',
  8: '退款',
  9: '求购账户充值',
  10: '提现退款',
  99: '其他',
};

const CHART_COLORS: Record<number, string> = {
  1: '#d32f2f',
  2: '#2e7d32',
  3: '#1565c0',
  4: '#00897b',
  5: '#e65100',
  6: '#6a1b9a',
  7: '#f9a825',
  8: '#00838f',
  9: '#4e342e',
  99: '#78909c',
};

export default function BillPage() {
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const [searchQuery, setSearchQuery] = useState('');
  const [platformFilter, setPlatformFilter] = useState('');
  const [typeIdFilter, setTypeIdFilter] = useState<number | ''>('');
  const [startDateStr, setStartDateStr] = useState('');
  const [endDateStr, setEndDateStr] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [jumpInput, setJumpInput] = useState('');

  // Reset page when filters change
  useEffect(() => {
    setPage(1);
  }, [platformFilter, typeIdFilter, startDateStr, endDateStr]);

  // Convert date strings to unix ms for backend filtering
  const startTime = useMemo(
    () => (startDateStr ? new Date(startDateStr + 'T00:00:00+08:00').getTime() : 0),
    [startDateStr],
  );
  const endTime = useMemo(
    () => (endDateStr ? new Date(endDateStr + 'T23:59:59.999+08:00').getTime() : 0),
    [endDateStr],
  );

  const { data, isLoading, error, refetch } = useBillRecords(selectedAccountId, {
    page,
    pageSize,
    typeId: typeIdFilter === '' ? 0 : typeIdFilter,
    platform: platformFilter,
    startTime,
    endTime,
  });
  const bills = useMemo(() => data?.records ?? [], [data?.records]);
  const totalCount = data?.totalCount ?? 0;
  const totalPages = useMemo(
    () => Math.max(1, Math.ceil(totalCount / pageSize)),
    [totalCount, pageSize],
  );

  // Pre-aggregated daily chart data (not pagination-dependent)
  const { data: chartData = [] } = useBillChartData(selectedAccountId, startTime, endTime);

  const { data: accounts = [] } = useAccounts();

  const accountMap = useMemo(() => {
    const m = new Map<number, string>();
    for (const acc of accounts) m.set(acc.ID, acc.name);
    return m;
  }, [accounts]);

  const typeFilterOptions = useMemo(() => {
    const ids = new Set(chartData.map((b) => b.typeId));
    return Array.from(ids).sort((a, b) => a - b);
  }, [chartData]);

  // Table data — already filtered server-side; only search filter applied client-side.
  const filteredBills = useMemo(() => {
    if (!searchQuery) return bills;
    const q = searchQuery.toLowerCase();
    return bills.filter(
      (b) =>
        (b.orderNo ?? '').toLowerCase().includes(q) || (b.typeName ?? '').toLowerCase().includes(q),
    );
  }, [bills, searchQuery]);

  // Summary cards — aggregated from pre-aggregated chart data
  const typeTotals = useMemo(() => {
    const totals: Record<number, number> = {};
    for (const p of chartData) {
      totals[p.typeId] = (totals[p.typeId] ?? 0) + p.thisMoney;
    }
    return totals;
  }, [chartData]);

  const chartOption = useMemo(() => {
    if (chartData.length === 0) return null;

    const typeIds = [...new Set(chartData.map((p) => p.typeId))].sort((a, b) => a - b);

    const dayBuckets: Record<string, Record<number, number>> = {};
    for (const p of chartData) {
      if (!dayBuckets[p.date]) dayBuckets[p.date] = {};
      dayBuckets[p.date][p.typeId] = (dayBuckets[p.date][p.typeId] ?? 0) + p.thisMoney;
    }

    const dateKeys = Object.keys(dayBuckets).sort();
    const running: Record<number, number> = {};
    for (const tid of typeIds) running[tid] = 0;

    const cumulative: Record<number, number[]> = {};
    for (const dateKey of dateKeys) {
      for (const tid of typeIds) {
        running[tid] += dayBuckets[dateKey][tid] ?? 0;
        if (!cumulative[tid]) cumulative[tid] = [];
        cumulative[tid].push(running[tid] / 100);
      }
    }

    // Daily total bar values (yuan)
    const dailyTotals = dateKeys.map((dk) => {
      let sum = 0;
      for (const tid of typeIds) sum += dayBuckets[dk][tid] ?? 0;
      return sum / 100;
    });

    const barSeries = {
      name: '当日合计',
      type: 'bar' as const,
      data: dailyTotals,
      barWidth: '55%',
      itemStyle: {
        borderRadius: [4, 4, 0, 0],
        color: (p: { value: number }) => (p.value >= 0 ? '#2e7d32' : '#c62828'),
      },
    };

    const series = [
      barSeries,
      ...typeIds.map((tid) => ({
        name: TYPE_LABELS[tid] ?? `类型 ${tid}`,
        type: 'line' as const,
        data: cumulative[tid],
        smooth: true,
        symbol: 'none' as const,
        lineStyle: { color: CHART_COLORS[tid] ?? '#999', width: 2 },
        itemStyle: { color: CHART_COLORS[tid] ?? '#999' },
      })),
    ];

    return {
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(255,255,255,0.96)',
        borderColor: '#e0e0e0',
        borderWidth: 1,
        textStyle: { color: '#333', fontSize: 13 },
      },
      legend: {
        bottom: 8,
        textStyle: { fontSize: 12, color: '#666' },
        itemWidth: 14,
        itemHeight: 10,
      },
      grid: { top: 16, left: 64, right: 64, bottom: 48 },
      xAxis: {
        type: 'category',
        data: dateKeys,
        axisLine: { lineStyle: { color: '#e0e0e0' } },
        axisTick: { show: false },
        axisLabel: { color: '#888', fontSize: 11, rotate: 45 },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: '#f0f0f0' } },
        axisLabel: {
          fontSize: 11,
          color: '#888',
          formatter: (v: number) => {
            if (Math.abs(v) >= 10000) return `¥${(v / 10000).toFixed(1)}w`;
            return `¥${v.toFixed(0)}`;
          },
        },
      },
      series,
      dataZoom: [
        {
          type: 'slider',
          height: 20,
          bottom: 4,
          borderColor: 'transparent',
          backgroundColor: '#f5f5f5',
          fillerColor: 'rgba(21,101,192,0.1)',
          handleStyle: { color: '#1565c0', borderColor: '#1565c0' },
          textStyle: { fontSize: 10, color: '#999' },
        },
        { type: 'inside' },
      ],
    };
  }, [chartData]);

  const columns: ColumnDef<model.BillRecord>[] = useMemo(
    () => [
      {
        accessorKey: 'addTime',
        header: '时间',
        cell: (info) => {
          const v = info.getValue() as number;
          return (
            <Typography variant="body2" fontSize={13}>
              {new Date(v).toLocaleString('zh-CN')}
            </Typography>
          );
        },
      },
      {
        accessorKey: 'platform',
        header: '平台',
        cell: (info) => (
          <Typography variant="body2" color="text.secondary">
            {platformLabel[info.getValue() as string] ?? (info.getValue() as string)}
          </Typography>
        ),
      },
      {
        id: 'account',
        header: '账户',
        cell: (info) => (
          <Typography variant="body2" color="text.secondary">
            {accountMap.get(info.row.original.accountId) ??
              String(info.row.original.accountId ?? '—')}
          </Typography>
        ),
      },
      {
        accessorKey: 'typeName',
        header: '类型',
        cell: (info) => {
          const typeId = info.row.original.typeId;
          return (
            <Chip
              label={info.getValue() as string}
              size="small"
              color={TYPE_COLORS[typeId] ?? 'default'}
              variant="outlined"
            />
          );
        },
      },
      {
        accessorKey: 'thisMoney',
        header: '金额',
        meta: { align: 'right' },
        cell: (info) => {
          const v = info.getValue() as number;
          return (
            <Typography
              variant="body2"
              fontWeight={500}
              color={v >= 0 ? 'success.main' : 'error.main'}
            >
              {formatCNY(v)}
            </Typography>
          );
        },
      },
      {
        accessorKey: 'orderNo',
        header: '订单号',
        cell: (info) => {
          const v = info.getValue() as string;
          if (!v)
            return (
              <Typography variant="body2" color="text.disabled">
                —
              </Typography>
            );
          return (
            <Typography variant="body2" fontSize={12} color="text.secondary">
              {v.length > 24 ? v.slice(0, 24) + '...' : v}
            </Typography>
          );
        },
      },
    ],
    [accountMap],
  );

  return (
    <Box>
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          mb: 3,
          gap: 1,
        }}
      >
        <Typography variant="h4">资金流水</Typography>
        <PageSearchBar value={searchQuery} onChange={setSearchQuery} placeholder="搜索订单号..." />
      </Box>

      {error && (
        <Box mb={3}>
          <ErrorBanner message={`加载流水失败: ${String(error)}`} onRetry={() => void refetch()} />
        </Box>
      )}

      {isLoading && (
        <Box mt={3}>
          <Skeleton variant="rectangular" height={40} sx={{ mb: 2, borderRadius: 1 }} />
          <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
            {[1, 2, 3, 4, 5].map((i) => (
              <Skeleton
                key={i}
                variant="rectangular"
                height={72}
                sx={{ flex: 1, borderRadius: 1 }}
              />
            ))}
          </Box>
          <Skeleton variant="rectangular" height={400} sx={{ mb: 3, borderRadius: 1 }} />
          <Skeleton variant="rectangular" height={300} sx={{ borderRadius: 1 }} />
        </Box>
      )}

      {!isLoading && !error && totalCount === 0 && (
        <Alert severity="info">暂无流水记录。同步账户数据后将自动拉取资金流水。</Alert>
      )}

      {!isLoading && !error && totalCount > 0 && (
        <>
          {/* Filter bar */}
          <Box sx={{ display: 'flex', gap: 1.5, mb: 3, flexWrap: 'wrap', alignItems: 'center' }}>
            <FormControl size="small" sx={{ minWidth: 130 }}>
              <Select
                value={platformFilter}
                onChange={(e: SelectChangeEvent<string>) => setPlatformFilter(e.target.value)}
                displayEmpty
              >
                <MenuItem value="">全部平台</MenuItem>
                {PLATFORM_OPTIONS.map((opt) => (
                  <MenuItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl size="small" sx={{ minWidth: 130 }}>
              <Select
                value={typeIdFilter}
                onChange={(e: SelectChangeEvent<number | ''>) =>
                  setTypeIdFilter(e.target.value as number | '')
                }
                displayEmpty
              >
                <MenuItem value="">全部类型</MenuItem>
                {typeFilterOptions.map((tid) => (
                  <MenuItem key={tid} value={tid}>
                    {TYPE_LABELS[tid] ?? `类型 ${tid}`}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              label="开始日期"
              type="date"
              size="small"
              value={startDateStr}
              onChange={(e) => setStartDateStr(e.target.value)}
              InputLabelProps={{ shrink: true }}
              sx={{ width: 160 }}
            />

            <TextField
              label="结束日期"
              type="date"
              size="small"
              value={endDateStr}
              onChange={(e) => setEndDateStr(e.target.value)}
              InputLabelProps={{ shrink: true }}
              sx={{ width: 160 }}
            />
          </Box>

          {/* Summary cards */}
          {chartData.length > 0 && (
            <Box sx={{ display: 'flex', gap: 1.5, mb: 3, flexWrap: 'wrap' }}>
              {Object.keys(typeTotals)
                .map(Number)
                .sort((a, b) => a - b)
                .map((tid) => {
                  const total = typeTotals[tid];
                  return (
                    <Card key={tid} sx={{ minWidth: 140, flex: 1 }}>
                      <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
                        <Typography variant="caption" color="text.secondary">
                          {TYPE_LABELS[tid] ?? `类型 ${tid}`}
                        </Typography>
                        <Typography
                          variant="body1"
                          fontWeight={600}
                          color={total >= 0 ? 'success.main' : 'error.main'}
                        >
                          {formatCNY(total)}
                        </Typography>
                      </CardContent>
                    </Card>
                  );
                })}
            </Box>
          )}

          {/* Chart */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                累计资金流水趋势
              </Typography>
              {chartData.length === 0 ? (
                <Typography color="text.secondary" textAlign="center" py={4}>
                  所选筛选条件下没有数据
                </Typography>
              ) : chartOption ? (
                <Box sx={{ height: 400 }}>
                  <ReactEChartsCore
                    echarts={echarts}
                    option={chartOption}
                    style={{ height: '100%', width: '100%' }}
                    notMerge
                  />
                </Box>
              ) : null}
            </CardContent>
          </Card>

          {/* Table */}
          {filteredBills.length > 0 && (
            <>
              <SortableTable
                columns={columns}
                data={filteredBills}
                getRowId={(b) => String(b.ID)}
              />
              <TablePagination
                component="div"
                count={totalCount}
                page={page - 1}
                onPageChange={(_e, newPage) => setPage(newPage + 1)}
                rowsPerPage={pageSize}
                onRowsPerPageChange={(e) => {
                  setPageSize(Number(e.target.value));
                  setPage(1);
                }}
                rowsPerPageOptions={[20, 50, 100]}
                labelRowsPerPage="每页行数："
                labelDisplayedRows={({ from, to, count }) => (
                  <Box
                    component="span"
                    sx={{ display: 'inline-flex', alignItems: 'center', gap: 0.5 }}
                  >
                    {from}–{to} of {count}
                    <TextField
                      size="small"
                      variant="standard"
                      placeholder={`${page}`}
                      value={jumpInput}
                      onChange={(e) => setJumpInput(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          const n = Number(jumpInput);
                          if (n >= 1 && n <= totalPages) {
                            setPage(n);
                            setJumpInput('');
                          }
                        }
                      }}
                      inputProps={{
                        inputMode: 'numeric',
                        style: { width: 40, textAlign: 'center', padding: 0 },
                      }}
                      sx={{ ml: 1, width: 48, '& .MuiInputBase-root': { fontSize: '0.875rem' } }}
                    />
                  </Box>
                )}
              />
            </>
          )}

          {filteredBills.length === 0 && (
            <EmptyState title="没有匹配的流水" description="尝试修改筛选条件或搜索关键词。" />
          )}
        </>
      )}
    </Box>
  );
}
