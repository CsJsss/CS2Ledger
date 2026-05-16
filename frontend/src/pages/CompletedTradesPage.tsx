import React, { useState, useMemo } from "react";
import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  type ColumnDef,
  type SortingState,
} from "@tanstack/react-table";
import Typography from "@mui/material/Typography";
import Skeleton from "@mui/material/Skeleton";
import Box from "@mui/material/Box";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import Divider from "@mui/material/Divider";
import Grid from "@mui/material/Grid";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import IconButton from "@mui/material/IconButton";
import InfoIcon from "@mui/icons-material/Info";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableSortLabel from "@mui/material/TableSortLabel";
import Paper from "@mui/material/Paper";
import Collapse from "@mui/material/Collapse";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowRightIcon from "@mui/icons-material/KeyboardArrowRight";
import ErrorBanner from "../components/ErrorBanner";
import EmptyState from "../components/EmptyState";
import PnlSummaryCards from "../components/PnlSummaryCards";
import PageSearchBar from "../components/PageSearchBar";
import { useCompletedTrades } from "../hooks/useCompletedTrades";
import { useCompletedTradesSummary } from "../hooks/useCompletedTradesSummary";
import { useUnmatchedSells } from "../hooks/useUnmatchedSells";
import { useUIStore } from "../store/uiStore";
import { formatCNY, plHexColor } from "../lib/format";
import type { model, trade } from "../lib/wails";

declare module "@tanstack/react-table" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData, TValue> {
    align?: "left" | "right" | "center";
  }
}

type TabKey = "completed" | "unmatched";

interface GroupedTrade {
  itemName: string;
  count: number;
  trades: trade.CompletedTradeView[];
  totalBuyPrice: number;
  totalSellPrice: number;
  totalGrossPl: number;
  totalFee: number;
  totalNetPl: number;
}

interface GroupedUnmatchedSell {
  itemName: string;
  count: number;
  sells: model.TradeRecord[];
  totalSellPrice: number;
  totalFee: number;
}

// ─── Completed Trade Detail Dialog ───────────────────────────────────────────

function TradeDetailDialog({
  open,
  onClose,
  trade,
}: {
  open: boolean;
  onClose: () => void;
  trade: trade.CompletedTradeView | null;
}) {
  if (!trade) return null;

  const buy = trade.buyTrade;
  const sell = trade.sellTrade;

  const platformLabel = (p: string) =>
    ({ buff: "BUFF", youpin: "悠悠", c5: "C5", igxe: "IGXE", eco: "ECO" }[p] ?? p);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ fontWeight: 600 }}>
        Trade Details — {trade.itemName}
      </DialogTitle>
      <DialogContent dividers>
        <Grid container spacing={2}>
          <Grid item xs={6}>
            <Typography variant="overline" color="text.secondary">Buy Order</Typography>
            <Box sx={{ mt: 0.5, display: "flex", flexDirection: "column", gap: 0.5 }}>
              <Typography variant="body2">Platform: {platformLabel(buy.source)}</Typography>
              <Typography variant="body2">Price: {formatCNY(buy.unitPrice)}</Typography>
              <Typography variant="body2">Qty: {buy.quantity}</Typography>
              <Typography variant="body2">Total: {formatCNY(buy.totalPrice)}</Typography>
              <Typography variant="body2">Fee: {formatCNY(buy.fee)}</Typography>
              <Typography variant="body2">Date: {new Date(buy.tradeAt).toLocaleDateString()}</Typography>
            </Box>
          </Grid>
          <Grid item xs={6}>
            <Typography variant="overline" color="text.secondary">Sell Order</Typography>
            <Box sx={{ mt: 0.5, display: "flex", flexDirection: "column", gap: 0.5 }}>
              <Typography variant="body2">Platform: {platformLabel(sell.source)}</Typography>
              <Typography variant="body2">Price: {formatCNY(sell.unitPrice)}</Typography>
              <Typography variant="body2">Qty: {sell.quantity}</Typography>
              <Typography variant="body2">Total: {formatCNY(sell.totalPrice)}</Typography>
              <Typography variant="body2">Fee: {formatCNY(sell.fee)}</Typography>
              <Typography variant="body2">Date: {new Date(sell.tradeAt).toLocaleDateString()}</Typography>
            </Box>
          </Grid>
        </Grid>
        <Divider sx={{ my: 2 }} />
        <Box sx={{ display: "flex", gap: 3 }}>
          <Box>
            <Typography variant="overline" color="text.secondary">Gross P/L</Typography>
            <Typography variant="body2" color={plHexColor(trade.grossPl)} fontWeight={500}>
              {formatCNY(trade.grossPl)}
            </Typography>
          </Box>
          <Box>
            <Typography variant="overline" color="text.secondary">Total Fees</Typography>
            <Typography variant="body2">{formatCNY(trade.totalFee)}</Typography>
          </Box>
          <Box>
            <Typography variant="overline" color="text.secondary">Net P/L</Typography>
            <Typography variant="body2" color={plHexColor(trade.netPl)} fontWeight={600}>
              {formatCNY(trade.netPl)}
            </Typography>
          </Box>
        </Box>
      </DialogContent>
    </Dialog>
  );
}

// ─── Unmatched Sell Detail Dialog ────────────────────────────────────────────

function UnmatchedSellDetailDialog({
  open,
  onClose,
  sell,
}: {
  open: boolean;
  onClose: () => void;
  sell: model.TradeRecord | null;
}) {
  if (!sell) return null;

  const platformLabel = (p: string) =>
    ({ buff: "BUFF", youpin: "悠悠", c5: "C5", igxe: "IGXE", eco: "ECO" }[p] ?? p);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ fontWeight: 600 }}>
        Sell Order — {sell.itemName}
      </DialogTitle>
      <DialogContent dividers>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <Typography variant="body2">Platform: {platformLabel(sell.source)}</Typography>
          <Typography variant="body2">Exterior: {sell.exterior || "-"}</Typography>
          <Typography variant="body2">Price: {formatCNY(sell.unitPrice)}</Typography>
          <Typography variant="body2">Qty: {sell.quantity}</Typography>
          <Typography variant="body2">Total: {formatCNY(sell.totalPrice)}</Typography>
          <Typography variant="body2">Fee: {formatCNY(sell.fee)}</Typography>
          <Typography variant="body2">Date: {new Date(sell.tradeAt).toLocaleDateString()}</Typography>
          {sell.assetId && <Typography variant="body2">Asset ID: {sell.assetId}</Typography>}
        </Box>
      </DialogContent>
    </Dialog>
  );
}

// ─── Completed Trades Columns ────────────────────────────────────────────────

const groupedColumns: ColumnDef<GroupedTrade>[] = [
  {
    id: "expander",
    header: "",
    cell: ({ row }) => (
      <IconButton size="small">
        {row.getIsExpanded?.() ? (
          <KeyboardArrowDownIcon fontSize="small" />
        ) : (
          <KeyboardArrowRightIcon fontSize="small" />
        )}
      </IconButton>
    ),
  },
  {
    accessorKey: "itemName",
    header: "Item Name",
    cell: (info) => (
      <Typography variant="body2" fontWeight={500}>{info.getValue() as string}</Typography>
    ),
  },
  {
    accessorKey: "count",
    header: "Trades",
    meta: { align: "right" },
    cell: (info) => <Typography variant="body2">{String(info.getValue())}</Typography>,
  },
  {
    accessorKey: "totalBuyPrice",
    header: "Total Buy",
    meta: { align: "right" },
    cell: (info) => formatCNY(info.getValue() as number),
  },
  {
    accessorKey: "totalSellPrice",
    header: "Total Sell",
    meta: { align: "right" },
    cell: (info) => formatCNY(info.getValue() as number),
  },
  {
    accessorKey: "totalGrossPl",
    header: "Gross P/L",
    meta: { align: "right" },
    cell: (info) => {
      const v = info.getValue() as number;
      return <Typography variant="body2" color={plHexColor(v)}>{formatCNY(v)}</Typography>;
    },
  },
  {
    accessorKey: "totalFee",
    header: "Fees",
    meta: { align: "right" },
    cell: (info) => formatCNY(info.getValue() as number),
  },
  {
    accessorKey: "totalNetPl",
    header: "Net P/L",
    meta: { align: "right" },
    cell: (info) => {
      const v = info.getValue() as number;
      return <Typography variant="body2" fontWeight={600} color={plHexColor(v)}>{formatCNY(v)}</Typography>;
    },
  },
];

// ─── Unmatched Sells Columns ─────────────────────────────────────────────────

const unmatchedGroupedColumns: ColumnDef<GroupedUnmatchedSell>[] = [
  {
    id: "expander",
    header: "",
    cell: ({ row }) => (
      <IconButton size="small">
        {row.getIsExpanded?.() ? (
          <KeyboardArrowDownIcon fontSize="small" />
        ) : (
          <KeyboardArrowRightIcon fontSize="small" />
        )}
      </IconButton>
    ),
  },
  {
    accessorKey: "itemName",
    header: "Item Name",
    cell: (info) => (
      <Typography variant="body2" fontWeight={500}>{info.getValue() as string}</Typography>
    ),
  },
  {
    accessorKey: "count",
    header: "Sells",
    meta: { align: "right" },
    cell: (info) => <Typography variant="body2">{String(info.getValue())}</Typography>,
  },
  {
    accessorKey: "totalSellPrice",
    header: "Total Sell",
    meta: { align: "right" },
    cell: (info) => formatCNY(info.getValue() as number),
  },
  {
    accessorKey: "totalFee",
    header: "Fees",
    meta: { align: "right" },
    cell: (info) => formatCNY(info.getValue() as number),
  },
];

// ─── Completed Trades Tab Content ────────────────────────────────────────────

function CompletedTradesContent({ accountId, searchQuery }: { accountId: number | null; searchQuery: string }) {
  const [dismissed, setDismissed] = useState(false);
  const [detailTrade, setDetailTrade] = useState<trade.CompletedTradeView | null>(null);

  const {
    data: trades = [],
    isLoading: tradesLoading,
    error: tradesError,
    refetch: refetchTrades,
  } = useCompletedTrades(accountId);

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
  } = useCompletedTradesSummary(accountId);

  const isLoading = tradesLoading || summaryLoading;
  const error = tradesError || summaryError;

  const [expandedNames, setExpandedNames] = useState<Set<string>>(new Set());

  const grouped = useMemo(() => {
    const map = new Map<string, trade.CompletedTradeView[]>();
    for (const t of trades) {
      const name = t.itemName ?? "Unknown";
      const arr = map.get(name);
      if (arr) arr.push(t);
      else map.set(name, [t]);
    }
    return Array.from(map, ([itemName, trades]) => {
      let totalGrossPl = 0;
      let totalFee = 0;
      let totalNetPl = 0;
      let totalBuyPrice = 0;
      let totalSellPrice = 0;
      for (const t of trades) {
        totalGrossPl += t.grossPl;
        totalFee += t.totalFee;
        totalNetPl += t.netPl;
        totalBuyPrice += t.buyTrade.totalPrice;
        totalSellPrice += t.sellTrade.totalPrice;
      }
      return { itemName, count: trades.length, trades, totalBuyPrice, totalSellPrice, totalGrossPl, totalFee, totalNetPl };
    }).sort((a, b) => a.itemName.localeCompare(b.itemName));
  }, [trades]);

  const toggle = (name: string) => {
    setExpandedNames((prev) => {
      const next = new Set(prev);
      if (next.has(name)) next.delete(name);
      else next.add(name);
      return next;
    });
  };

  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable({
    data: grouped,
    columns: groupedColumns,
    state: { sorting, globalFilter: searchQuery },
    onSortingChange: setSorting,
    onGlobalFilterChange: () => {},
    globalFilterFn: (row, _columnId, filterValue) =>
      row.original.itemName.toLowerCase().includes((filterValue as string).toLowerCase()),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getCoreRowModel: getCoreRowModel(),
    getRowId: (row) => row.itemName,
  });

  if (isLoading) {
    return (
      <Box mt={3}>
        <Box sx={{ display: "flex", gap: 2, mb: 3 }}>
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} variant="rectangular" height={96} sx={{ flex: 1, borderRadius: 1 }} />
          ))}
        </Box>
        {[1, 2, 3, 4, 5].map((i) => (
          <Skeleton key={i} variant="rectangular" height={48} sx={{ mb: 1, borderRadius: 1 }} />
        ))}
      </Box>
    );
  }

  if (error && !dismissed) {
    return (
      <Box mt={3}>
        <ErrorBanner
          message={`Failed to load trades: ${String(error)}`}
          onRetry={() => { setDismissed(false); void refetchTrades(); }}
          onDismiss={() => setDismissed(true)}
        />
      </Box>
    );
  }

  if (!summary) return null;

  return (
    <Box mt={3}>
      <PnlSummaryCards
        totalTrades={summary.totalTrades}
        totalGrossPl={summary.totalGrossPl}
        totalFee={summary.totalFee}
        totalNetPl={summary.totalNetPl}
        withdrawalFee={summary.withdrawalFee}
        withdrawalFeeRate={summary.withdrawalFeeRate}
      />

      {trades.length === 0 && (
        <Box mt={3}>
          <EmptyState title="No completed trades" description="Sync to populate trade data and calculate P/L." />
        </Box>
      )}

      {trades.length > 0 && (
        <Box mt={3}>
          <TableContainer component={Paper}>
            <Table size="small">
              <TableHead>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => {
                      const canSort = header.column.getCanSort();
                      const sorted = header.column.getIsSorted();
                      return (
                        <TableCell
                          key={header.id}
                          align={header.column.columnDef.meta?.align}
                          sortDirection={sorted || false}
                          sx={{ py: 1 }}
                        >
                          {canSort ? (
                            <TableSortLabel
                              active={!!sorted}
                              direction={sorted || undefined}
                              onClick={() => {
                                if (sorted === "desc") header.column.clearSorting();
                                else header.column.toggleSorting(sorted === "asc");
                              }}
                            >
                              {flexRender(header.column.columnDef.header, header.getContext())}
                            </TableSortLabel>
                          ) : (
                            flexRender(header.column.columnDef.header, header.getContext())
                          )}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                ))}
              </TableHead>
              <TableBody>
                {table.getRowModel().rows.map((groupRow) => {
                  const expanded = expandedNames.has(groupRow.original.itemName);
                  return (
                    <React.Fragment key={groupRow.id}>
                      <TableRow
                        hover
                        sx={{ bgcolor: "grey.50", cursor: "pointer" }}
                        onClick={() => toggle(groupRow.original.itemName)}
                      >
                        {groupRow.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id} align={cell.column.columnDef.meta?.align} sx={{ py: 1 }}>
                            {flexRender(cell.column.columnDef.cell, {
                              ...cell.getContext(),
                              row: { ...groupRow, getIsExpanded: () => expanded },
                            })}
                          </TableCell>
                        ))}
                      </TableRow>
                      <TableRow sx={{ "& td": { border: 0 } }}>
                        <TableCell colSpan={groupedColumns.length} sx={{ p: 0 }}>
                          <Collapse in={expanded}>
                            <Box sx={{ mx: 2, my: 1 }}>
                              <Table size="small">
                                <TableHead>
                                  <TableRow>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5, width: 40 }} />
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }}>Item Name</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }}>Exterior</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Buy Price</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Sell Price</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Qty</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Buy Total</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Sell Total</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Gross P/L</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Fees</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Net P/L</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Sell Date</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }}>Details</TableCell>
                                  </TableRow>
                                </TableHead>
                                <TableBody>
                                  {groupRow.original.trades.map((t) => (
                                    <TableRow key={String(t.sellTrade.ID)} hover>
                                      <TableCell sx={{ py: 0.5, width: 40 }} />
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" fontWeight={500}>{t.itemName}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" color="text.secondary">{t.exterior || "-"}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">{formatCNY(t.buyTrade.unitPrice)}</TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">{formatCNY(t.sellTrade.unitPrice)}</TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2">{t.quantity}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">{formatCNY(t.buyTrade.totalPrice)}</TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">{formatCNY(t.sellTrade.totalPrice)}</TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" color={plHexColor(t.grossPl)}>{formatCNY(t.grossPl)}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">{formatCNY(t.totalFee)}</TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" fontWeight={600} color={plHexColor(t.netPl)}>{formatCNY(t.netPl)}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" color="text.secondary">
                                          {new Date(t.sellTrade.tradeAt).toLocaleDateString()}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <IconButton size="small" onClick={() => setDetailTrade(t)}>
                                          <InfoIcon fontSize="small" />
                                        </IconButton>
                                      </TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </Box>
                          </Collapse>
                        </TableCell>
                      </TableRow>
                    </React.Fragment>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>

          <TradeDetailDialog
            open={!!detailTrade}
            onClose={() => setDetailTrade(null)}
            trade={detailTrade}
          />
        </Box>
      )}
    </Box>
  );
}

// ─── Unmatched Sells Tab Content ─────────────────────────────────────────────

function UnmatchedSellsContent({ accountId, searchQuery }: { accountId: number | null; searchQuery: string }) {
  const [dismissed, setDismissed] = useState(false);
  const [detailSell, setDetailSell] = useState<model.TradeRecord | null>(null);

  const {
    data: sells = [],
    isLoading,
    error,
    refetch,
  } = useUnmatchedSells(accountId);

  const [expandedNames, setExpandedNames] = useState<Set<string>>(new Set());

  const grouped = useMemo(() => {
    const map = new Map<string, model.TradeRecord[]>();
    for (const s of sells) {
      const name = s.itemName ?? "Unknown";
      const arr = map.get(name);
      if (arr) arr.push(s);
      else map.set(name, [s]);
    }
    return Array.from(map, ([itemName, sells]) => {
      let totalSellPrice = 0;
      let totalFee = 0;
      for (const s of sells) {
        totalSellPrice += s.totalPrice;
        totalFee += s.fee;
      }
      return { itemName, count: sells.length, sells, totalSellPrice, totalFee };
    }).sort((a, b) => a.itemName.localeCompare(b.itemName));
  }, [sells]);

  const toggle = (name: string) => {
    setExpandedNames((prev) => {
      const next = new Set(prev);
      if (next.has(name)) next.delete(name);
      else next.add(name);
      return next;
    });
  };

  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable({
    data: grouped,
    columns: unmatchedGroupedColumns,
    state: { sorting, globalFilter: searchQuery },
    onSortingChange: setSorting,
    onGlobalFilterChange: () => {},
    globalFilterFn: (row, _columnId, filterValue) =>
      row.original.itemName.toLowerCase().includes((filterValue as string).toLowerCase()),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getCoreRowModel: getCoreRowModel(),
    getRowId: (row) => row.itemName,
  });

  const totalSells = sells.length;
  const totalSellValue = useMemo(() => sells.reduce((sum, s) => sum + s.totalPrice, 0), [sells]);
  const totalFees = useMemo(() => sells.reduce((sum, s) => sum + s.fee, 0), [sells]);

  const platformLabel = (p: string) =>
    ({ buff: "BUFF", youpin: "悠悠", c5: "C5", igxe: "IGXE", eco: "ECO" }[p] ?? p);

  if (isLoading) {
    return (
      <Box mt={3}>
        <Box sx={{ display: "flex", gap: 2, mb: 3 }}>
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} variant="rectangular" height={96} sx={{ flex: 1, borderRadius: 1 }} />
          ))}
        </Box>
        {[1, 2, 3, 4, 5].map((i) => (
          <Skeleton key={i} variant="rectangular" height={48} sx={{ mb: 1, borderRadius: 1 }} />
        ))}
      </Box>
    );
  }

  if (error && !dismissed) {
    return (
      <Box mt={3}>
        <ErrorBanner
          message={`Failed to load unmatched sells: ${String(error)}`}
          onRetry={() => { setDismissed(false); void refetch(); }}
          onDismiss={() => setDismissed(true)}
        />
      </Box>
    );
  }

  return (
    <Box mt={3}>
      <Grid container spacing={2}>
        <Grid item xs={4}>
          <Card>
            <CardContent>
              <Typography variant="body2" color="text.secondary">Unmatched Sells</Typography>
              <Typography variant="h5" mt={1}>{totalSells}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={4}>
          <Card>
            <CardContent>
              <Typography variant="body2" color="text.secondary">Total Sell Value</Typography>
              <Typography variant="h5" mt={1}>{formatCNY(totalSellValue)}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={4}>
          <Card>
            <CardContent>
              <Typography variant="body2" color="text.secondary">Total Fees</Typography>
              <Typography variant="h5" mt={1}>{formatCNY(totalFees)}</Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {sells.length === 0 && (
        <Box mt={3}>
          <EmptyState title="No unmatched sells" description="All sells have been matched to buy orders." />
        </Box>
      )}

      {sells.length > 0 && (
        <Box mt={3}>
          <TableContainer component={Paper}>
            <Table size="small">
              <TableHead>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => {
                      const canSort = header.column.getCanSort();
                      const sorted = header.column.getIsSorted();
                      return (
                        <TableCell
                          key={header.id}
                          align={header.column.columnDef.meta?.align}
                          sortDirection={sorted || false}
                          sx={{ py: 1 }}
                        >
                          {canSort ? (
                            <TableSortLabel
                              active={!!sorted}
                              direction={sorted || undefined}
                              onClick={() => {
                                if (sorted === "desc") header.column.clearSorting();
                                else header.column.toggleSorting(sorted === "asc");
                              }}
                            >
                              {flexRender(header.column.columnDef.header, header.getContext())}
                            </TableSortLabel>
                          ) : (
                            flexRender(header.column.columnDef.header, header.getContext())
                          )}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                ))}
              </TableHead>
              <TableBody>
                {table.getRowModel().rows.map((groupRow) => {
                  const expanded = expandedNames.has(groupRow.original.itemName);
                  return (
                    <React.Fragment key={groupRow.id}>
                      <TableRow
                        hover
                        sx={{ bgcolor: "grey.50", cursor: "pointer" }}
                        onClick={() => toggle(groupRow.original.itemName)}
                      >
                        {groupRow.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id} align={cell.column.columnDef.meta?.align} sx={{ py: 1 }}>
                            {flexRender(cell.column.columnDef.cell, {
                              ...cell.getContext(),
                              row: { ...groupRow, getIsExpanded: () => expanded },
                            })}
                          </TableCell>
                        ))}
                      </TableRow>
                      <TableRow sx={{ "& td": { border: 0 } }}>
                        <TableCell colSpan={unmatchedGroupedColumns.length} sx={{ p: 0 }}>
                          <Collapse in={expanded}>
                            <Box sx={{ mx: 2, my: 1 }}>
                              <Table size="small">
                                <TableHead>
                                  <TableRow>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5, width: 40 }} />
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }}>Item Name</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }}>Exterior</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Price</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Qty</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Total</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Fee</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }} align="right">Sell Date</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }}>Platform</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem", py: 0.5 }}>Details</TableCell>
                                  </TableRow>
                                </TableHead>
                                <TableBody>
                                  {groupRow.original.sells.map((s) => (
                                    <TableRow key={String(s.ID)} hover>
                                      <TableCell sx={{ py: 0.5, width: 40 }} />
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" fontWeight={500}>{s.itemName}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" color="text.secondary">{s.exterior || "-"}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">{formatCNY(s.unitPrice)}</TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2">{s.quantity}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">{formatCNY(s.totalPrice)}</TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">{formatCNY(s.fee)}</TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" color="text.secondary">
                                          {new Date(s.tradeAt).toLocaleDateString()}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" color="text.secondary">{platformLabel(s.source)}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <IconButton size="small" onClick={() => setDetailSell(s)}>
                                          <InfoIcon fontSize="small" />
                                        </IconButton>
                                      </TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </Box>
                          </Collapse>
                        </TableCell>
                      </TableRow>
                    </React.Fragment>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>

          <UnmatchedSellDetailDialog
            open={!!detailSell}
            onClose={() => setDetailSell(null)}
            sell={detailSell}
          />
        </Box>
      )}
    </Box>
  );
}

// ─── Page ────────────────────────────────────────────────────────────────────

export default function CompletedTradesPage() {
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const [tab, setTab] = useState<TabKey>("completed");
  const [searchQuery, setSearchQuery] = useState("");

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 2 }}>
        <Typography variant="h4" gutterBottom>Trades</Typography>
        <PageSearchBar value={searchQuery} onChange={setSearchQuery} placeholder="Filter by item name..." />
      </Box>

      <Tabs value={tab} onChange={(_, v) => setTab(v as TabKey)} sx={{ mb: 1 }}>
        <Tab label="Completed" value="completed" />
        <Tab label="Unmatched Sells" value="unmatched" />
      </Tabs>

      {tab === "completed" && <CompletedTradesContent accountId={selectedAccountId} searchQuery={searchQuery} />}
      {tab === "unmatched" && <UnmatchedSellsContent accountId={selectedAccountId} searchQuery={searchQuery} />}
    </Box>
  );
}
