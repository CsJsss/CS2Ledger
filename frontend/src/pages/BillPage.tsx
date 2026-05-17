import { useState } from "react";
import { type ColumnDef } from "@tanstack/react-table";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Alert from "@mui/material/Alert";
import Chip from "@mui/material/Chip";
import SortableTable from "../components/SortableTable";
import PageSearchBar from "../components/PageSearchBar";
import ErrorBanner from "../components/ErrorBanner";
import { useBillRecords } from "../hooks/useBillRecords";
import { useUIStore } from "../store/uiStore";
import { platformLabel } from "../lib/constants";
import { formatCNY } from "../lib/format";
import type { model } from "../lib/wails";

const TYPE_COLORS: Record<number, "default" | "success" | "error" | "warning" | "info"> = {
  1: "error",    // 购买
  2: "success",  // 出售
  3: "success",  // 收取租金
  4: "error",    // 租赁服务费
  5: "info",     // 充值
  6: "warning",  // 提现
  7: "info",     // 退款
  99: "default", // 其他
};

export default function BillPage() {
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const [searchQuery, setSearchQuery] = useState("");

  const { data: bills = [], isLoading, error, refetch } = useBillRecords(selectedAccountId);

  const columns: ColumnDef<model.BillRecord>[] = [
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
          {info.row.original.accountId ?? "—"}
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
          <Typography
            variant="body2"
            fontWeight={500}
            color={v >= 0 ? "success.main" : "error.main"}
          >
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
  ];

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3, gap: 1 }}>
        <Typography variant="h4">资金流水</Typography>
        <PageSearchBar value={searchQuery} onChange={setSearchQuery} placeholder="搜索订单号..." />
      </Box>

      {error && (
        <Box mb={3}>
          <ErrorBanner
            message={`加载流水失败: ${String(error)}`}
            onRetry={() => void refetch()}
          />
        </Box>
      )}

      {isLoading && <Typography color="text.secondary">Loading...</Typography>}

      {!isLoading && !error && bills.length === 0 && (
        <Alert severity="info">
          暂无流水记录。同步账户数据后将自动拉取资金流水。
        </Alert>
      )}

      {bills.length > 0 && (
        <SortableTable
          columns={columns}
          data={bills}
          globalFilter={searchQuery}
          onGlobalFilterChange={setSearchQuery}
          globalFilterFn={(row, _columnId, filterValue) =>
            (row.original.orderNo ?? "").toLowerCase().includes((filterValue as string).toLowerCase()) ||
            (row.original.typeName ?? "").toLowerCase().includes((filterValue as string).toLowerCase())
          }
          getRowId={(b) => String(b.ID)}
        />
      )}
    </Box>
  );
}
