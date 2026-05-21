import { useState, useMemo } from "react";
import { type ColumnDef } from "@tanstack/react-table";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Alert from "@mui/material/Alert";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Chip from "@mui/material/Chip";
import FormControl from "@mui/material/FormControl";
import MenuItem from "@mui/material/MenuItem";
import Select, { type SelectChangeEvent } from "@mui/material/Select";
import Skeleton from "@mui/material/Skeleton";
import TextField from "@mui/material/TextField";
import SortableTable from "../components/SortableTable";
import PageSearchBar from "../components/PageSearchBar";
import ErrorBanner from "../components/ErrorBanner";
import EmptyState from "../components/EmptyState";
import { useBillRecords } from "../hooks/useBillRecords";
import { useAccounts } from "../hooks/useAccounts";
import { useUIStore } from "../store/uiStore";
import { platformLabel, PLATFORM_OPTIONS } from "../lib/constants";
import { formatCNY } from "../lib/format";
import type { model } from "../lib/wails";
import ReactEChartsCore from "echarts-for-react/lib/core";
import * as echarts from "echarts/core";
import { BarChart, LineChart } from "echarts/charts";
import {
  GridComponent,
  TooltipComponent,
  DataZoomComponent,
  LegendComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";

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
const TYPE_COLORS: Record<number, "default" | "success" | "error" | "warning" | "info"> = {
  1: "error",
  2: "success",
  3: "success",
  4: "success",
  5: "error",
  6: "info",
  7: "warning",
  8: "info",
  9: "info",
  99: "default",
};

const TYPE_LABELS: Record<number, string> = {
  1: "购买",
  2: "出售",
  3: "收取租金",
  4: "收取续租资金",
  5: "租赁服务费",
  6: "充值",
  7: "提现",
  8: "退款",
  9: "求购账户充值",
  99: "其他",
};

const CHART_COLORS: Record<number, string> = {
  1: "#d32f2f",
  2: "#2e7d32",
  3: "#1565c0",
  4: "#00897b",
  5: "#e65100",
  6: "#6a1b9a",
  7: "#f9a825",
  8: "#00838f",
  9: "#4e342e",
  99: "#78909c",
};

export default function BillPage() {
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const [searchQuery, setSearchQuery] = useState("");
  const [platformFilter, setPlatformFilter] = useState("");
  const [typeIdFilter, setTypeIdFilter] = useState<number | "">("");
  const [startDateStr, setStartDateStr] = useState("");
  const [endDateStr, setEndDateStr] = useState("");

  const { data: bills = [], isLoading, error, refetch } = useBillRecords(selectedAccountId);
  const { data: accounts = [] } = useAccounts();

  const accountMap = useMemo(() => {
    const m = new Map<number, string>();
    for (const acc of accounts) m.set(acc.ID, acc.name);
    return m;
  }, [accounts]);

  const typeFilterOptions = useMemo(() => {
    const ids = new Set(bills.map((b) => b.typeId));
    return Array.from(ids).sort((a, b) => a - b);
  }, [bills]);

  const filteredBills = useMemo(() => {
    let result = bills;

    if (platformFilter) {
      result = result.filter((b) => b.platform === platformFilter);
    }
    if (typeIdFilter !== "") {
      result = result.filter((b) => b.typeId === typeIdFilter);
    }
    if (startDateStr) {
      const startMs = new Date(startDateStr + "T00:00:00+08:00").getTime();
      result = result.filter((b) => b.addTime >= startMs);
    }
    if (endDateStr) {
      const endMs = new Date(endDateStr + "T23:59:59.999+08:00").getTime();
      result = result.filter((b) => b.addTime <= endMs);
    }

    return result;
  }, [bills, platformFilter, typeIdFilter, startDateStr, endDateStr]);

  const typeTotals = useMemo(() => {
    const totals: Record<number, number> = {};
    for (const bill of filteredBills) {
      totals[bill.typeId] = (totals[bill.typeId] ?? 0) + bill.thisMoney;
    }
    return totals;
  }, [filteredBills]);

  const chartOption = useMemo(() => {
    if (filteredBills.length === 0) return null;

    const sorted = [...filteredBills].sort((a, b) => a.addTime - b.addTime);
    const typeIds = [...new Set(sorted.map((b) => b.typeId))].sort((a, b) => a - b);

    const dayBuckets: Record<string, Record<number, number>> = {};
    for (const bill of sorted) {
      const d = new Date(bill.addTime);
      const dateKey = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
      if (!dayBuckets[dateKey]) dayBuckets[dateKey] = {};
      dayBuckets[dateKey][bill.typeId] = (dayBuckets[dateKey][bill.typeId] ?? 0) + bill.thisMoney;
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
      name: "当日合计",
      type: "bar" as const,
      data: dailyTotals,
      barWidth: "55%",
      itemStyle: {
        borderRadius: [4, 4, 0, 0],
        color: (p: { value: number }) => (p.value >= 0 ? "#2e7d32" : "#c62828"),
      },
    };

    const series = [barSeries, ...typeIds.map((tid) => ({
      name: TYPE_LABELS[tid] ?? `类型 ${tid}`,
      type: "line" as const,
      data: cumulative[tid],
      smooth: true,
      symbol: "none" as const,
      lineStyle: { color: CHART_COLORS[tid] ?? "#999", width: 2 },
      itemStyle: { color: CHART_COLORS[tid] ?? "#999" },
    }))];

    return {
      tooltip: {
        trigger: "axis",
        backgroundColor: "rgba(255,255,255,0.96)",
        borderColor: "#e0e0e0",
        borderWidth: 1,
        textStyle: { color: "#333", fontSize: 13 },
      },
      legend: {
        bottom: 8,
        textStyle: { fontSize: 12, color: "#666" },
        itemWidth: 14,
        itemHeight: 10,
      },
      grid: { top: 16, left: 64, right: 64, bottom: 48 },
      xAxis: {
        type: "category",
        data: dateKeys,
        axisLine: { lineStyle: { color: "#e0e0e0" } },
        axisTick: { show: false },
        axisLabel: { color: "#888", fontSize: 11, rotate: 45 },
      },
      yAxis: {
        type: "value",
        splitLine: { lineStyle: { color: "#f0f0f0" } },
        axisLabel: {
          fontSize: 11,
          color: "#888",
          formatter: (v: number) => {
            if (Math.abs(v) >= 10000) return `¥${(v / 10000).toFixed(1)}w`;
            return `¥${v.toFixed(0)}`;
          },
        },
      },
      series,
      dataZoom: [
        {
          type: "slider",
          height: 20,
          bottom: 4,
          borderColor: "transparent",
          backgroundColor: "#f5f5f5",
          fillerColor: "rgba(21,101,192,0.1)",
          handleStyle: { color: "#1565c0", borderColor: "#1565c0" },
          textStyle: { fontSize: 10, color: "#999" },
        },
        { type: "inside" },
      ],
    };
  }, [filteredBills]);

  const columns: ColumnDef<model.BillRecord>[] = useMemo(
    () => [
      {
        accessorKey: "addTime",
        header: "时间",
        cell: (info) => {
          const v = info.getValue() as number;
          return (
            <Typography variant="body2" fontSize={13}>
              {new Date(v).toLocaleString("zh-CN")}
            </Typography>
          );
        },
      },
      {
        accessorKey: "platform",
        header: "平台",
        cell: (info) => (
          <Typography variant="body2" color="text.secondary">
            {platformLabel[info.getValue() as string] ?? (info.getValue() as string)}
          </Typography>
        ),
      },
      {
        id: "account",
        header: "账户",
        cell: (info) => (
          <Typography variant="body2" color="text.secondary">
            {accountMap.get(info.row.original.accountId) ?? String(info.row.original.accountId ?? "—")}
          </Typography>
        ),
      },
      {
        accessorKey: "typeName",
        header: "类型",
        cell: (info) => {
          const typeId = info.row.original.typeId;
          return (
            <Chip
              label={info.getValue() as string}
              size="small"
              color={TYPE_COLORS[typeId] ?? "default"}
              variant="outlined"
            />
          );
        },
      },
      {
        accessorKey: "thisMoney",
        header: "金额",
        meta: { align: "right" },
        cell: (info) => {
          const v = info.getValue() as number;
          return (
            <Typography variant="body2" fontWeight={500} color={v >= 0 ? "success.main" : "error.main"}>
              {formatCNY(v)}
            </Typography>
          );
        },
      },
      {
        accessorKey: "orderNo",
        header: "订单号",
        cell: (info) => {
          const v = info.getValue() as string;
          if (!v) return <Typography variant="body2" color="text.disabled">—</Typography>;
          return (
            <Typography variant="body2" fontSize={12} color="text.secondary">
              {v.length > 24 ? v.slice(0, 24) + "..." : v}
            </Typography>
          );
        },
      },
    ],
    [accountMap],
  );

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3, gap: 1 }}>
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
          <Box sx={{ display: "flex", gap: 2, mb: 3 }}>
            {[1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} variant="rectangular" height={72} sx={{ flex: 1, borderRadius: 1 }} />
            ))}
          </Box>
          <Skeleton variant="rectangular" height={400} sx={{ mb: 3, borderRadius: 1 }} />
          <Skeleton variant="rectangular" height={300} sx={{ borderRadius: 1 }} />
        </Box>
      )}

      {!isLoading && !error && bills.length === 0 && (
        <Alert severity="info">暂无流水记录。同步账户数据后将自动拉取资金流水。</Alert>
      )}

      {!isLoading && !error && bills.length > 0 && (
        <>
          {/* Filter bar */}
          <Box sx={{ display: "flex", gap: 1.5, mb: 3, flexWrap: "wrap", alignItems: "center" }}>
            <FormControl size="small" sx={{ minWidth: 130 }}>
              <Select
                value={platformFilter}
                onChange={(e: SelectChangeEvent<string>) => setPlatformFilter(e.target.value)}
                displayEmpty
              >
                <MenuItem value="">全部平台</MenuItem>
                {PLATFORM_OPTIONS.map((opt) => (
                  <MenuItem key={opt.value} value={opt.value}>{opt.label}</MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl size="small" sx={{ minWidth: 130 }}>
              <Select
                value={typeIdFilter}
                onChange={(e: SelectChangeEvent<number | "">) =>
                  setTypeIdFilter(e.target.value as number | "")
                }
                displayEmpty
              >
                <MenuItem value="">全部类型</MenuItem>
                {typeFilterOptions.map((tid) => (
                  <MenuItem key={tid} value={tid}>{TYPE_LABELS[tid] ?? `类型 ${tid}`}</MenuItem>
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
          {filteredBills.length > 0 && (
            <Box sx={{ display: "flex", gap: 1.5, mb: 3, flexWrap: "wrap" }}>
              {Object.keys(typeTotals)
                .map(Number)
                .sort((a, b) => a - b)
                .map((tid) => {
                  const total = typeTotals[tid];
                  return (
                    <Card key={tid} sx={{ minWidth: 140, flex: 1 }}>
                      <CardContent sx={{ py: 1.5, "&:last-child": { pb: 1.5 } }}>
                        <Typography variant="caption" color="text.secondary">
                          {TYPE_LABELS[tid] ?? `类型 ${tid}`}
                        </Typography>
                        <Typography variant="body1" fontWeight={600} color={total >= 0 ? "success.main" : "error.main"}>
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
              <Typography variant="h6" gutterBottom>累计资金流水趋势</Typography>
              {filteredBills.length === 0 ? (
                <Typography color="text.secondary" textAlign="center" py={4}>
                  所选筛选条件下没有数据
                </Typography>
              ) : chartOption ? (
                <Box sx={{ height: 400 }}>
                  <ReactEChartsCore echarts={echarts} option={chartOption} style={{ height: "100%", width: "100%" }} notMerge />
                </Box>
              ) : null}
            </CardContent>
          </Card>

          {/* Table */}
          {filteredBills.length > 0 && (
            <SortableTable
              columns={columns}
              data={filteredBills}
              globalFilter={searchQuery}
              onGlobalFilterChange={setSearchQuery}
              globalFilterFn={(row, _columnId, filterValue) =>
                (row.original.orderNo ?? "").toLowerCase().includes((filterValue as string).toLowerCase()) ||
                (row.original.typeName ?? "").toLowerCase().includes((filterValue as string).toLowerCase())
              }
              getRowId={(b) => String(b.ID)}
            />
          )}

          {filteredBills.length === 0 && (
            <EmptyState title="没有匹配的流水" description="尝试修改筛选条件或搜索关键词。" />
          )}
        </>
      )}
    </Box>
  );
}
