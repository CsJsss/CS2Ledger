import React, { useState, useMemo } from 'react';
import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  type ColumnDef,
  type SortingState,
} from '@tanstack/react-table';
import Typography from '@mui/material/Typography';
import Skeleton from '@mui/material/Skeleton';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import IconButton from '@mui/material/IconButton';
import InfoIcon from '@mui/icons-material/Info';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TextField from '@mui/material/TextField';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import Paper from '@mui/material/Paper';
import Collapse from '@mui/material/Collapse';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowRightIcon from '@mui/icons-material/KeyboardArrowRight';
import ErrorBanner from '../components/ErrorBanner';
import EmptyState from '../components/EmptyState';
import PnlSummaryCards from '../components/PnlSummaryCards';
import ReceiptIcon from '@mui/icons-material/Receipt';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import PageSearchBar from '../components/PageSearchBar';
import Tooltip from '@mui/material/Tooltip';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import { useCompletedTrades } from '../hooks/useCompletedTrades';
import { useCompletedTradesSummary } from '../hooks/useCompletedTradesSummary';
import { useUnmatchedSells } from '../hooks/useUnmatchedSells';
import { useDailySells } from '../hooks/useDailySells';
import { useExpandableSet } from '../hooks/useExpandableSet';
import { useUIStore } from '../store/uiStore';
import { formatCNY, plHexColor } from '../lib/format';
import { platformLabel } from '../lib/constants';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import type { model, trade } from '../lib/wails';

declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData, TValue> {
    align?: 'left' | 'right' | 'center';
  }
}

type TabKey = 'completed' | 'unmatched' | 'dailySell';

interface GroupedTrade {
  itemName: string;
  exterior: string;
  csqaqGoodsId?: number;
  marketHashName?: string;
  count: number;
  trades: trade.CompletedTradeView[];
  totalBuyPrice: number;
  totalSellPrice: number;
  totalGrossPl: number;
  totalFee: number;
  totalNetPl: number;
  marketPrice?: number;
  marketTotal?: number;
  postTradePl?: number;
}

interface GroupedUnmatchedSell {
  itemName: string;
  exterior: string;
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
    ({ buff: 'BUFF', youpin: '悠悠', c5: 'C5', igxe: 'IGXE', eco: 'ECO' })[p] ?? p;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ fontWeight: 600 }}>交易详情 — {trade.itemName}</DialogTitle>
      <DialogContent dividers>
        <Grid container spacing={2}>
          <Grid item xs={6}>
            <Typography variant="overline" color="text.secondary">
              买入订单
            </Typography>
            <Box sx={{ mt: 0.5, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
              <Typography variant="body2">平台: {platformLabel(buy.source)}</Typography>
              <Typography variant="body2">价格: {formatCNY(buy.unitPrice)}</Typography>
              <Typography variant="body2">数量: {buy.quantity}</Typography>
              <Typography variant="body2">总额: {formatCNY(buy.totalPrice)}</Typography>
              <Typography variant="body2">手续费: {formatCNY(buy.fee)}</Typography>
              <Typography variant="body2">
                日期: {new Date(buy.tradeAt).toLocaleDateString()}
              </Typography>
            </Box>
          </Grid>
          <Grid item xs={6}>
            <Typography variant="overline" color="text.secondary">
              卖出订单
            </Typography>
            <Box sx={{ mt: 0.5, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
              <Typography variant="body2">平台: {platformLabel(sell.source)}</Typography>
              <Typography variant="body2">价格: {formatCNY(sell.unitPrice)}</Typography>
              <Typography variant="body2">数量: {sell.quantity}</Typography>
              <Typography variant="body2">总额: {formatCNY(sell.totalPrice)}</Typography>
              <Typography variant="body2">手续费: {formatCNY(sell.fee)}</Typography>
              <Typography variant="body2">
                日期: {new Date(sell.tradeAt).toLocaleDateString()}
              </Typography>
            </Box>
          </Grid>
        </Grid>
        <Divider sx={{ my: 2 }} />
        <Box sx={{ display: 'flex', gap: 3 }}>
          <Box>
            <Typography variant="overline" color="text.secondary">
              毛利
            </Typography>
            <Typography variant="body2" color={plHexColor(trade.grossPl)} fontWeight={500}>
              {formatCNY(trade.grossPl)}
            </Typography>
          </Box>
          <Box>
            <Typography variant="overline" color="text.secondary">
              手续费
            </Typography>
            <Typography variant="body2">{formatCNY(trade.totalFee)}</Typography>
          </Box>
          <Box>
            <Typography variant="overline" color="text.secondary">
              净利润
            </Typography>
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
    ({ buff: 'BUFF', youpin: '悠悠', c5: 'C5', igxe: 'IGXE', eco: 'ECO' })[p] ?? p;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ fontWeight: 600 }}>卖出订单 — {sell.itemName}</DialogTitle>
      <DialogContent dividers>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
          <Typography variant="body2">平台: {platformLabel(sell.source)}</Typography>
          <Typography variant="body2">磨损: {sell.exterior || '-'}</Typography>
          <Typography variant="body2">价格: {formatCNY(sell.unitPrice)}</Typography>
          <Typography variant="body2">数量: {sell.quantity}</Typography>
          <Typography variant="body2">总额: {formatCNY(sell.totalPrice)}</Typography>
          <Typography variant="body2">手续费: {formatCNY(sell.fee)}</Typography>
          <Typography variant="body2">
            日期: {new Date(sell.tradeAt).toLocaleDateString()}
          </Typography>
          {sell.assetId && <Typography variant="body2">Asset ID: {sell.assetId}</Typography>}
        </Box>
      </DialogContent>
    </Dialog>
  );
}

// ─── Completed Trades Columns ────────────────────────────────────────────────

const groupedColumns: ColumnDef<GroupedTrade>[] = [
  {
    id: 'expander',
    header: '',
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
    accessorKey: 'itemName',
    header: '物品名称',
    cell: ({ row }) => (
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
        <Typography variant="body2" fontWeight={500}>
          {row.original.itemName}
          {row.original.exterior ? ` (${row.original.exterior})` : ''}
        </Typography>
        {row.original.csqaqGoodsId ? (
          <Tooltip title="csqaq">
            <IconButton
              size="small"
              onClick={(e) => {
                e.stopPropagation();
                BrowserOpenURL(`https://www.csqaq.com/goods/${row.original.csqaqGoodsId}`);
              }}
            >
              <OpenInNewIcon sx={{ fontSize: 14 }} />
            </IconButton>
          </Tooltip>
        ) : null}
      </Box>
    ),
  },
  {
    accessorKey: 'count',
    header: '交易数',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {String(info.getValue())}
      </Typography>
    ),
  },
  {
    accessorKey: 'totalBuyPrice',
    header: '买入总额',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {formatCNY(info.getValue() as number)}
      </Typography>
    ),
  },
  {
    accessorKey: 'totalSellPrice',
    header: '卖出总额',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {formatCNY(info.getValue() as number)}
      </Typography>
    ),
  },
  {
    id: 'marketTotal',
    header: '市场总价',
    meta: { align: 'right' },
    cell: ({ row }) => (
      <Typography variant="body2" color="text.secondary" className="mono-num">
        {row.original.marketTotal != null ? formatCNY(row.original.marketTotal) : '--'}
      </Typography>
    ),
  },
  {
    id: 'postTradePl',
    header: '交易后盈亏',
    meta: { align: 'right' },
    cell: ({ row }) => {
      if (row.original.postTradePl == null)
        return (
          <Typography variant="body2" color="text.secondary" className="mono-num">
            --
          </Typography>
        );
      const v = row.original.postTradePl;
      return (
        <Typography variant="body2" color={plHexColor(v)} className="mono-num">
          {formatCNY(v)}
        </Typography>
      );
    },
  },
  {
    accessorKey: 'totalGrossPl',
    header: '毛利',
    meta: { align: 'right' },
    cell: (info) => {
      const v = info.getValue() as number;
      return (
        <Typography variant="body2" color={plHexColor(v)} className="mono-num">
          {formatCNY(v)}
        </Typography>
      );
    },
  },
  {
    accessorKey: 'totalFee',
    header: '手续费',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {formatCNY(info.getValue() as number)}
      </Typography>
    ),
  },
  {
    accessorKey: 'totalNetPl',
    header: '净利润',
    meta: { align: 'right' },
    cell: (info) => {
      const v = info.getValue() as number;
      return (
        <Typography variant="body2" fontWeight={600} color={plHexColor(v)} className="mono-num">
          {formatCNY(v)}
        </Typography>
      );
    },
  },
];

// ─── Unmatched Sells Columns ─────────────────────────────────────────────────

const unmatchedGroupedColumns: ColumnDef<GroupedUnmatchedSell>[] = [
  {
    id: 'expander',
    header: '',
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
    accessorKey: 'itemName',
    header: '物品名称',
    cell: ({ row }) => (
      <Typography variant="body2" fontWeight={500}>
        {row.original.itemName}
        {row.original.exterior ? ` (${row.original.exterior})` : ''}
      </Typography>
    ),
  },
  {
    accessorKey: 'count',
    header: '卖出数',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {String(info.getValue())}
      </Typography>
    ),
  },
  {
    accessorKey: 'totalSellPrice',
    header: '卖出总额',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {formatCNY(info.getValue() as number)}
      </Typography>
    ),
  },
  {
    accessorKey: 'totalFee',
    header: '手续费',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {formatCNY(info.getValue() as number)}
      </Typography>
    ),
  },
];

// ─── Completed Trades Tab Content ────────────────────────────────────────────

function CompletedTradesContent({
  accountId,
  searchQuery,
}: {
  accountId: number | null;
  searchQuery: string;
}) {
  const [dismissed, setDismissed] = useState(false);
  const [detailTrade, setDetailTrade] = useState<trade.CompletedTradeView | null>(null);

  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(50);
  const [sortBy, setSortBy] = useState('itemName');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');

  const {
    data,
    isLoading: tradesLoading,
    error: tradesError,
    refetch: refetchTrades,
  } = useCompletedTrades(accountId, { page: page + 1, pageSize, sortBy, sortDir });

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
  } = useCompletedTradesSummary(accountId);

  const isLoading = tradesLoading || summaryLoading;
  const error = tradesError || summaryError;

  const allGroups: GroupedTrade[] = data?.groups ?? [];
  const total = data?.total ?? 0;

  const groups = searchQuery
    ? allGroups.filter((g) => g.itemName.toLowerCase().includes(searchQuery.toLowerCase()))
    : allGroups;

  const [expandedNames, setExpandedNames] = useState<Set<string>>(new Set());

  const toggle = (name: string) => {
    setExpandedNames((prev) => {
      const next = new Set(prev);
      if (next.has(name)) next.delete(name);
      else next.add(name);
      return next;
    });
  };

  const handleSort = (sb: string, sd: string) => {
    setSortBy(sb);
    setSortDir(sd as 'asc' | 'desc');
    setPage(0);
  };

  if (isLoading) {
    return (
      <Box mt={3}>
        <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
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
          message={`加载交易数据失败: ${String(error)}`}
          onRetry={() => {
            setDismissed(false);
            void refetchTrades();
          }}
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
      />

      {groups.length === 0 && (
        <Box mt={3}>
          <EmptyState
            icon={<ReceiptIcon sx={{ fontSize: 48 }} />}
            title="暂无已完成交易"
            description="同步账户数据后将在此显示交易和盈亏。"
          />
        </Box>
      )}

      {groups.length > 0 && (
        <Box mt={3}>
          <Paper>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell sx={{ py: 1 }} />
                    {(
                      [
                        ['itemName', '物品名称', false],
                        ['count', '交易数', true],
                        ['totalBuy', '买入总额', true],
                        ['totalSell', '卖出总额', true],
                        ['marketTotal', '市场总价', true],
                        ['postTradePl', '交易后盈亏', true],
                        ['grossPl', '毛利', true],
                        ['fees', '手续费', true],
                        ['netPl', '净利润', true],
                      ] as [string, string, boolean][]
                    ).map(([key, label, right]) => (
                      <TableCell key={key} sx={{ py: 1 }} align={right ? 'right' : 'left'}>
                        <TableSortLabel
                          active={sortBy === key}
                          direction={sortBy === key ? sortDir : 'asc'}
                          onClick={() => {
                            if (sortBy === key && sortDir === 'asc') handleSort(key, 'desc');
                            else handleSort(key, 'asc');
                          }}
                        >
                          {label}
                        </TableSortLabel>
                      </TableCell>
                    ))}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {groups.map((group) => {
                    const groupKey = `${group.itemName}|${group.exterior}`;
                    const expanded = expandedNames.has(groupKey);
                    return (
                      <React.Fragment key={groupKey}>
                        <TableRow
                          hover
                          sx={{ bgcolor: 'background.default', cursor: 'pointer' }}
                          onClick={() => toggle(groupKey)}
                        >
                          <TableCell sx={{ py: 1 }}>
                            <IconButton size="small">
                              {expanded ? (
                                <KeyboardArrowDownIcon fontSize="small" />
                              ) : (
                                <KeyboardArrowRightIcon fontSize="small" />
                              )}
                            </IconButton>
                          </TableCell>
                          <TableCell sx={{ py: 1 }}>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                              <Typography variant="body2" fontWeight={500}>
                                {group.itemName}
                                {group.exterior ? ` (${group.exterior})` : ''}
                              </Typography>
                              {group.csqaqGoodsId ? (
                                <Tooltip title="csqaq">
                                  <IconButton
                                    size="small"
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      BrowserOpenURL(
                                        `https://www.csqaq.com/goods/${group.csqaqGoodsId}`,
                                      );
                                    }}
                                  >
                                    <OpenInNewIcon sx={{ fontSize: 14 }} />
                                  </IconButton>
                                </Tooltip>
                              ) : null}
                            </Box>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" className="mono-num">
                              {String(group.count)}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" className="mono-num">
                              {formatCNY(group.totalBuyPrice)}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" className="mono-num">
                              {formatCNY(group.totalSellPrice)}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" color="text.secondary" className="mono-num">
                              {group.marketTotal != null ? formatCNY(group.marketTotal) : '--'}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            {group.postTradePl != null ? (
                              <Typography
                                variant="body2"
                                color={plHexColor(group.postTradePl)}
                                className="mono-num"
                              >
                                {formatCNY(group.postTradePl)}
                              </Typography>
                            ) : (
                              <Typography
                                variant="body2"
                                color="text.secondary"
                                className="mono-num"
                              >
                                --
                              </Typography>
                            )}
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography
                              variant="body2"
                              color={plHexColor(group.totalGrossPl)}
                              className="mono-num"
                            >
                              {formatCNY(group.totalGrossPl)}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" className="mono-num">
                              {formatCNY(group.totalFee)}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography
                              variant="body2"
                              fontWeight={600}
                              color={plHexColor(group.totalNetPl)}
                              className="mono-num"
                            >
                              {formatCNY(group.totalNetPl)}
                            </Typography>
                          </TableCell>
                        </TableRow>
                        <TableRow sx={{ '& td': { border: 0 } }}>
                          <TableCell colSpan={groupedColumns.length} sx={{ p: 0 }}>
                            <Collapse in={expanded}>
                              <Box sx={{ mx: 2, my: 1 }}>
                                <Table size="small">
                                  <TableHead>
                                    <TableRow>
                                      <TableCell sx={{ fontSize: '0.75rem', py: 0.5, width: 40 }} />
                                      <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                        物品名称
                                      </TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                        磨损
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        买入价
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        卖出价
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        数量
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        买入总额
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        卖出总额
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        毛利
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        手续费
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        净利润
                                      </TableCell>
                                      <TableCell
                                        sx={{ fontSize: '0.75rem', py: 0.5 }}
                                        align="right"
                                      >
                                        卖出日期
                                      </TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                        详情
                                      </TableCell>
                                    </TableRow>
                                  </TableHead>
                                  <TableBody>
                                    {group.trades.map((t) => (
                                      <TableRow key={String(t.sellTrade.ID)} hover>
                                        <TableCell sx={{ py: 0.5, width: 40 }} />
                                        <TableCell sx={{ py: 0.5 }}>
                                          <Typography variant="body2" fontWeight={500}>
                                            {t.itemName}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }}>
                                          <Typography variant="body2" color="text.secondary">
                                            {t.exterior || '-'}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {formatCNY(t.buyTrade.unitPrice)}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {formatCNY(t.sellTrade.unitPrice)}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {t.quantity}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {formatCNY(t.buyTrade.totalPrice)}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {formatCNY(t.sellTrade.totalPrice)}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography
                                            variant="body2"
                                            color={plHexColor(t.grossPl)}
                                            className="mono-num"
                                          >
                                            {formatCNY(t.grossPl)}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {formatCNY(t.totalFee)}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography
                                            variant="body2"
                                            fontWeight={600}
                                            color={plHexColor(t.netPl)}
                                            className="mono-num"
                                          >
                                            {formatCNY(t.netPl)}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" color="text.secondary">
                                            {new Date(t.sellTrade.tradeAt).toLocaleDateString()}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }}>
                                          <IconButton
                                            size="small"
                                            onClick={() => setDetailTrade(t)}
                                          >
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
            <TablePagination
              component="div"
              count={total}
              page={page}
              rowsPerPage={pageSize}
              onPageChange={(_, p) => setPage(p)}
              onRowsPerPageChange={(e) => {
                setPageSize(Number(e.target.value));
                setPage(0);
              }}
              rowsPerPageOptions={[20, 50, 100]}
            />
          </Paper>

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

function UnmatchedSellsContent({
  accountId,
  searchQuery,
}: {
  accountId: number | null;
  searchQuery: string;
}) {
  const [dismissed, setDismissed] = useState(false);
  const [detailSell, setDetailSell] = useState<model.TradeRecord | null>(null);

  const { data: sells = [], isLoading, error, refetch } = useUnmatchedSells(accountId);

  const [expandedNames, setExpandedNames] = useState<Set<string>>(new Set());

  const grouped = useMemo(() => {
    const map = new Map<string, model.TradeRecord[]>();
    for (const s of sells) {
      const name = s.itemName ?? 'Unknown';
      const exterior = s.exterior ?? '';
      const key = `${name}|${exterior}`;
      const arr = map.get(key);
      if (arr) arr.push(s);
      else map.set(key, [s]);
    }
    return Array.from(map, ([key, sells]) => {
      let totalSellPrice = 0;
      let totalFee = 0;
      for (const s of sells) {
        totalSellPrice += s.totalPrice;
        totalFee += s.fee;
      }
      const [itemName, exterior] = key.split('|', 2);
      return { itemName, exterior, count: sells.length, sells, totalSellPrice, totalFee };
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
    getRowId: (row) => `${row.itemName}|${row.exterior}`,
  });

  const totalSells = sells.length;
  const totalSellValue = useMemo(() => sells.reduce((sum, s) => sum + s.totalPrice, 0), [sells]);
  const totalFees = useMemo(() => sells.reduce((sum, s) => sum + s.fee, 0), [sells]);

  const platformLabel = (p: string) =>
    ({ buff: 'BUFF', youpin: '悠悠', c5: 'C5', igxe: 'IGXE', eco: 'ECO' })[p] ?? p;

  if (isLoading) {
    return (
      <Box mt={3}>
        <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
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
          message={`加载未匹配卖出数据失败: ${String(error)}`}
          onRetry={() => {
            setDismissed(false);
            void refetch();
          }}
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
              <Typography variant="body2" color="text.secondary">
                未匹配卖出
              </Typography>
              <Typography variant="h5" mt={1}>
                {totalSells}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={4}>
          <Card>
            <CardContent>
              <Typography variant="body2" color="text.secondary">
                卖出总额
              </Typography>
              <Typography variant="h5" mt={1}>
                {formatCNY(totalSellValue)}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={4}>
          <Card>
            <CardContent>
              <Typography variant="body2" color="text.secondary">
                手续费总额
              </Typography>
              <Typography variant="h5" mt={1}>
                {formatCNY(totalFees)}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {sells.length === 0 && (
        <Box mt={3}>
          <EmptyState
            icon={<CheckCircleIcon sx={{ fontSize: 48 }} />}
            title="无未匹配卖出"
            description="所有卖出订单均已匹配到买入订单。"
          />
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
                                if (sorted === 'desc') header.column.clearSorting();
                                else header.column.toggleSorting(sorted === 'asc');
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
                  const groupKey = `${groupRow.original.itemName}|${groupRow.original.exterior}`;
                  const expanded = expandedNames.has(groupKey);
                  return (
                    <React.Fragment key={groupRow.id}>
                      <TableRow
                        hover
                        sx={{
                          bgcolor: (t) => (t.palette.mode === 'dark' ? '#111114' : '#f8fafc'),
                          cursor: 'pointer',
                        }}
                        onClick={() => toggle(groupKey)}
                      >
                        {groupRow.getVisibleCells().map((cell) => (
                          <TableCell
                            key={cell.id}
                            align={cell.column.columnDef.meta?.align}
                            sx={{ py: 1 }}
                          >
                            {flexRender(cell.column.columnDef.cell, {
                              ...cell.getContext(),
                              row: { ...groupRow, getIsExpanded: () => expanded },
                            })}
                          </TableCell>
                        ))}
                      </TableRow>
                      <TableRow sx={{ '& td': { border: 0 } }}>
                        <TableCell colSpan={unmatchedGroupedColumns.length} sx={{ p: 0 }}>
                          <Collapse in={expanded}>
                            <Box sx={{ mx: 2, my: 1 }}>
                              <Table size="small">
                                <TableHead>
                                  <TableRow>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5, width: 40 }} />
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                      物品名称
                                    </TableCell>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                      磨损
                                    </TableCell>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }} align="right">
                                      价格
                                    </TableCell>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }} align="right">
                                      数量
                                    </TableCell>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }} align="right">
                                      总额
                                    </TableCell>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }} align="right">
                                      手续费
                                    </TableCell>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }} align="right">
                                      卖出日期
                                    </TableCell>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                      平台
                                    </TableCell>
                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                      详情
                                    </TableCell>
                                  </TableRow>
                                </TableHead>
                                <TableBody>
                                  {groupRow.original.sells.map((s) => (
                                    <TableRow key={String(s.ID)} hover>
                                      <TableCell sx={{ py: 0.5, width: 40 }} />
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" fontWeight={500}>
                                          {s.itemName}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" color="text.secondary">
                                          {s.exterior || '-'}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" className="mono-num">
                                          {formatCNY(s.unitPrice)}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" className="mono-num">
                                          {s.quantity}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" className="mono-num">
                                          {formatCNY(s.totalPrice)}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" className="mono-num">
                                          {formatCNY(s.fee)}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" color="text.secondary">
                                          {new Date(s.tradeAt).toLocaleDateString()}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" color="text.secondary">
                                          {platformLabel(s.source)}
                                        </Typography>
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

// ─── Daily Sells Tab Content ──────────────────────────────────────────────────

const SKELETON_COUNT = 5;

function DailySellsContent({ accountId }: { accountId: number | null }) {
  const [dismissed, setDismissed] = useState(false);
  const [jumpPage, setJumpPage] = useState('');

  const [page, setPage] = useState(0);
  const PAGE_SIZE = 30;

  const {
    data: paginated,
    isLoading,
    error,
    refetch,
  } = useDailySells(accountId, 0, 0, page + 1, PAGE_SIZE);
  const { isExpanded, toggle } = useExpandableSet();

  const groups = paginated?.groups ?? [];
  const totalGroups = paginated?.total ?? 0;

  const monthlyGroups = useMemo(() => {
    const map = new Map<string, { count: number; profit: number; fee: number }>();
    for (const g of groups) {
      const m = g.date.substring(0, 7);
      const entry = map.get(m) || { count: 0, profit: 0, fee: 0 };
      entry.count += g.totalCount;
      entry.profit += g.totalProfit;
      entry.fee += g.totalFee;
      map.set(m, entry);
    }
    return map;
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [paginated?.groups]);

  if (isLoading) {
    return (
      <Box mt={3}>
        {Array.from({ length: SKELETON_COUNT }, (_, i) => (
          <Skeleton key={i} variant="rectangular" height={48} sx={{ mb: 1, borderRadius: 1 }} />
        ))}
      </Box>
    );
  }

  if (error && !dismissed) {
    return (
      <Box mt={3}>
        <ErrorBanner
          message={`加载每日卖出数据失败: ${String(error)}`}
          onRetry={() => {
            setDismissed(false);
            void refetch();
          }}
          onDismiss={() => setDismissed(true)}
        />
      </Box>
    );
  }

  return (
    <Box mt={3}>
      {groups.length === 0 && (
        <EmptyState
          icon={<ReceiptIcon sx={{ fontSize: 48 }} />}
          title="暂无卖出记录"
          description="同步账户数据后将在此显示每日卖出详情。"
        />
      )}

      {groups.length > 0 && (
        <React.Fragment>
          <Paper>
            <TableContainer>
              <Table size="small">
                <TableBody>
                  {Array.from(monthlyGroups.entries()).map(([monthKey, mData]) => {
                    const monthGroups = groups.filter(g => g.date.substring(0, 7) === monthKey);
                    const monthExpanded = isExpanded(monthKey);
                    const mLabel = monthKey.replace('-', '年') + '月';
                    const netPl = mData.profit - mData.fee;

                    return (
                      <React.Fragment key={monthKey}>
                        <TableRow
                          hover
                          sx={{ bgcolor: 'background.paper', cursor: 'pointer', borderBottom: '2px solid', borderColor: 'divider' }}
                          onClick={() => toggle(monthKey)}
                        >
                          <TableCell sx={{ py: 1, width: 40 }}>
                            <IconButton size="small">
                              {monthExpanded ? <KeyboardArrowDownIcon fontSize="small" /> : <KeyboardArrowRightIcon fontSize="small" />}
                            </IconButton>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} colSpan={2}>
                            <Box sx={{ display: 'flex', gap: 3, alignItems: 'center' }}>
                              <Typography variant="body2" fontWeight={600}>{mLabel}</Typography>
                              <Typography variant="body2" color="text.secondary">卖出 {mData.count} 件</Typography>
                              <Typography variant="body2">
                                利润 <span style={{ color: plHexColor(mData.profit), fontWeight: 600 }}>{formatCNY(mData.profit)}</span>
                              </Typography>
                              <Typography variant="body2">手续费 {formatCNY(mData.fee)}</Typography>
                              <Typography variant="body2">
                                净利 <span style={{ color: plHexColor(netPl), fontWeight: 600 }}>{formatCNY(netPl)}</span>
                              </Typography>
                            </Box>
                          </TableCell>
                        </TableRow>
                        <TableRow sx={{ '& td': { border: 0 } }}>
                          <TableCell colSpan={3} sx={{ p: 0 }}>
                            <Collapse in={monthExpanded}>
                              <Box>
                                {monthGroups.map((group) => {
                                  const expanded = isExpanded(group.date);
                                  return (
                                    <React.Fragment key={group.date}>
                                      <TableRow
                                        hover
                                        sx={{ bgcolor: 'background.default', cursor: 'pointer' }}
                                        onClick={() => toggle(group.date)}
                                      >
                                        <TableCell sx={{ py: 1, width: 40 }}>
                                          <IconButton size="small">
                                            {expanded ? (
                                              <KeyboardArrowDownIcon fontSize="small" />
                                            ) : (
                                              <KeyboardArrowRightIcon fontSize="small" />
                                            )}
                                          </IconButton>
                                        </TableCell>
                                        <TableCell sx={{ py: 1, width: 240, textAlign: 'center' }}>
                                          <Typography variant="body2" fontWeight={700}>
                                            {(() => {
                                              const d = new Date(group.date);
                                              return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`;
                                            })()}
                                          </Typography>
                                          <Typography variant="caption" color="text.secondary">
                                            {group.dayOfWeek}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 1 }}>
                                          <Typography variant="body2" fontWeight={600}>
                                            卖出 {group.totalCount} 件
                                          </Typography>
                                          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', mt: 0.5 }}>
                                            <Typography variant="body2">
                                              利润{' '}
                                              <span
                                                style={{ color: plHexColor(group.totalProfit), fontWeight: 600 }}
                                              >
                                                {formatCNY(group.totalProfit)}
                                              </span>
                                            </Typography>
                                            <Typography variant="body2">
                                              手续费 {formatCNY(group.totalFee)}
                                            </Typography>
                                            <Typography variant="body2">
                                              净利{' '}
                                              <span
                                                style={{
                                                  color: plHexColor(group.totalProfit - group.totalFee),
                                                  fontWeight: 600,
                                                }}
                                              >
                                                {formatCNY(group.totalProfit - group.totalFee)}
                                              </span>
                                            </Typography>
                                          </Box>
                                        </TableCell>
                                      </TableRow>
                                      <TableRow sx={{ '& td': { border: 0 } }}>
                                        <TableCell colSpan={3} sx={{ p: 0 }}>
                                          <Collapse in={expanded}>
                                            <Box sx={{ mx: 2, my: 1 }}>
                                              <Table size="small">
                                                <TableHead>
                                                  <TableRow>
                                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                                      物品
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      数量
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      买入价
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      卖出价
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      手续费
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      利润
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      利润率
                                                    </TableCell>
                                                    <TableCell sx={{ fontSize: '0.75rem', py: 0.5 }}>
                                                      平台
                                                    </TableCell>
                                                  </TableRow>
                                                </TableHead>
                                                <TableBody>
                                                  {group.items.map((item, idx) => {
                                                    const costBasis = item.buyPrice * item.quantity;
                                                    const profitRate =
                                                      costBasis > 0 ? (item.profit / costBasis) * 100 : 0;
                                                    return (
                                                      <TableRow key={`${group.date}-${idx}`} hover>
                                                        <TableCell sx={{ py: 0.5 }}>
                                                          <Typography variant="body2" fontWeight={500}>
                                                            {item.itemName}
                                                            {item.exterior ? ` (${item.exterior})` : ''}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography variant="body2" className="mono-num">
                                                            {item.quantity}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography variant="body2" className="mono-num">
                                                            {formatCNY(item.buyPrice)}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography variant="body2" className="mono-num">
                                                            {formatCNY(item.sellPrice)}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography variant="body2" className="mono-num">
                                                            {formatCNY(item.totalFee)}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography
                                                            variant="body2"
                                                            color={plHexColor(item.profit)}
                                                            fontWeight={600}
                                                            className="mono-num"
                                                          >
                                                            {formatCNY(item.profit)}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography
                                                            variant="body2"
                                                            color={plHexColor(profitRate)}
                                                            className="mono-num"
                                                          >
                                                            {profitRate >= 0 ? '+' : ''}
                                                            {profitRate.toFixed(1)}%
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }}>
                                                          <Typography variant="body2" color="text.secondary">
                                                            {platformLabel[item.platform] ?? item.platform}
                                                          </Typography>
                                                        </TableCell>
                                                      </TableRow>
                                                    );
                                                  })}
                                                  <TableRow>
                                                    <TableCell colSpan={8} sx={{ py: 0.5 }}>
                                                      <Typography variant="caption" color="text.secondary">
                                                        当日合计：利润 {formatCNY(group.totalProfit)} · 手续费{' '}
                                                        {formatCNY(group.totalFee)} · 净利{' '}
                                                        {formatCNY(group.totalProfit - group.totalFee)}
                                                      </Typography>
                                                    </TableCell>
                                                  </TableRow>
                                                </TableBody>
                                              </Table>
                                            </Box>
                                          </Collapse>
                                        </TableCell>
                                      </TableRow>
                                    </React.Fragment>
                                  );
                                })}
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
          </Paper>
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'flex-end',
              alignItems: 'center',
              mt: 2,
              gap: 1,
            }}
          >
            <TablePagination
              component="div"
              count={totalGroups}
              page={page}
              rowsPerPage={PAGE_SIZE}
              onPageChange={(_, p) => setPage(p)}
              rowsPerPageOptions={[30]}
              labelRowsPerPage="每页"
            />
            <TextField
              size="small"
              type="number"
              placeholder="跳转"
              value={jumpPage}
              onChange={(e) => setJumpPage(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && jumpPage) {
                  const p = Math.max(
                    1,
                    Math.min(Math.ceil(totalGroups / PAGE_SIZE), Number(jumpPage)),
                  );
                  setPage(p - 1);
                  setJumpPage('');
                }
              }}
              sx={{ width: 70 }}
              inputProps={{ min: 1, style: { textAlign: 'center', padding: '4px 8px' } }}
            />
          </Box>
        </React.Fragment>
      )}
    </Box>
  );
}

// ─── Page ────────────────────────────────────────────────────────────────────

export default function CompletedTradesPage() {
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const [tab, setTab] = useState<TabKey>('completed');
  const [searchQuery, setSearchQuery] = useState('');

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h4" gutterBottom>
          交易记录
        </Typography>
        <PageSearchBar
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="搜索物品名称..."
        />
      </Box>

      <Tabs value={tab} onChange={(_, v) => setTab(v as TabKey)} sx={{ mb: 1 }}>
        <Tab label="已完成" value="completed" />
        <Tab label="未匹配卖出" value="unmatched" />
        <Tab label="每日卖出" value="dailySell" />
      </Tabs>

      {tab === 'completed' && (
        <CompletedTradesContent accountId={selectedAccountId} searchQuery={searchQuery} />
      )}
      {tab === 'unmatched' && (
        <UnmatchedSellsContent accountId={selectedAccountId} searchQuery={searchQuery} />
      )}
      {tab === 'dailySell' && <DailySellsContent accountId={selectedAccountId} />}
    </Box>
  );
}
