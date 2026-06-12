import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router';
import { type ColumnDef } from '@tanstack/react-table';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import TablePagination from '@mui/material/TablePagination';
import TextField from '@mui/material/TextField';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Collapse from '@mui/material/Collapse';
import Skeleton from '@mui/material/Skeleton';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import ErrorBanner from '../components/ErrorBanner';
import EmptyState from '../components/EmptyState';
import PageSearchBar from '../components/PageSearchBar';
import SearchOffIcon from '@mui/icons-material/SearchOff';
import InventoryIcon from '@mui/icons-material/Inventory';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import { useInventory } from '../hooks/useInventory';
import { useUIStore } from '../store/uiStore';
import { formatCNY, plHexColor } from '../lib/format';
import { inventoryStatusLabel, inventoryStatusColor, platformLabel } from '../lib/constants';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowRightIcon from '@mui/icons-material/KeyboardArrowRight';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import Tooltip from '@mui/material/Tooltip';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import ReceiptIcon from '@mui/icons-material/Receipt';
import type { model } from '../lib/wails';
import { useDailyBuys } from '../hooks/useDailyBuys';
import { useExpandableSet } from '../hooks/useExpandableSet';
declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData, TValue> {
    align?: 'left' | 'right' | 'center';
  }
}

interface GroupRowData {
  itemName: string;
  exterior: string;
  csqaqGoodsId?: number;
  marketHashName: string;
  weaponType: string;
  count: number;
  totalQuantity: number;
  totalBuyPrice: number;
  avgBuyPrice: number;
  marketPrice?: number;
  marketPriceUpdatedAt?: number;
  unrealizedPl?: number;
  instances: model.InventoryItem[];
}

const groupedColumns: ColumnDef<GroupRowData>[] = [
  {
    id: 'expander',
    header: '',
    enableSorting: false,
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
    accessorKey: 'weaponType',
    header: '类型',
    enableSorting: false,
    cell: (info) => {
      const wt = info.getValue() as string;
      return wt ? <Chip label={wt} size="small" variant="outlined" /> : null;
    },
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
    accessorKey: 'totalQuantity',
    header: '数量',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {(info.getValue() as number).toLocaleString()}
      </Typography>
    ),
  },
  {
    accessorKey: 'totalBuyPrice',
    header: '总价',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {formatCNY(info.getValue() as number)}
      </Typography>
    ),
  },
  {
    accessorKey: 'avgBuyPrice',
    header: '均价',
    meta: { align: 'right' },
    cell: (info) => (
      <Typography variant="body2" className="mono-num">
        {formatCNY(info.getValue() as number)}
      </Typography>
    ),
  },
  {
    id: 'marketPrice',
    header: '市场价',
    meta: { align: 'right' },
    cell: ({ row }) => (
      <Typography variant="body2" color="text.secondary" className="mono-num">
        {row.original.marketPrice != null ? formatCNY(row.original.marketPrice) : '--'}
      </Typography>
    ),
  },
  {
    id: 'marketTotal',
    header: '市场总价',
    meta: { align: 'right' },
    cell: ({ row }) => {
      const { marketPrice, totalQuantity } = row.original;
      if (marketPrice == null)
        return (
          <Typography variant="body2" color="text.secondary" className="mono-num">
            --
          </Typography>
        );
      return (
        <Typography variant="body2" color="text.secondary" className="mono-num">
          {formatCNY(marketPrice * totalQuantity)}
        </Typography>
      );
    },
  },
  {
    id: 'priceUpdatedAt',
    header: '行情时间',
    enableSorting: false,
    meta: { align: 'right' },
    cell: ({ row }) => {
      const ts = row.original.marketPriceUpdatedAt;
      return (
        <Typography variant="caption" color="text.disabled">
          {ts ? new Date(ts * 1000).toLocaleString() : '--'}
        </Typography>
      );
    },
  },
  {
    id: 'unrealizedPl',
    header: '未实现盈亏',
    meta: { align: 'right' },
    cell: ({ row }) => {
      if (row.original.unrealizedPl == null)
        return (
          <Typography variant="body2" color="text.secondary" className="mono-num">
            --
          </Typography>
        );
      const v = row.original.unrealizedPl;
      return (
        <Typography variant="body2" color={plHexColor(v)} className="mono-num">
          {formatCNY(v)}
        </Typography>
      );
    },
  },
  {
    id: 'plPercent',
    header: '盈亏%',
    meta: { align: 'right' },
    cell: ({ row }) => {
      const { avgBuyPrice, marketPrice } = row.original;
      if (marketPrice == null || avgBuyPrice === 0)
        return (
          <Typography variant="body2" color="text.secondary">
            --
          </Typography>
        );
      const pct = ((marketPrice - avgBuyPrice) / avgBuyPrice) * 100;
      return (
        <Typography variant="body2" color={plHexColor(pct)}>
          {pct >= 0 ? '+' : ''}
          {pct.toFixed(1)}%
        </Typography>
      );
    },
  },
];

const SKELETON_COUNT = 5;
const PAGE_SIZE = 12;

const dailyBuyStatusLabel: Record<string, string> = {
  in_inventory: '持有中',
  listed: '已上架',
};

const dailyBuyStatusColor = (status: string): 'success' | 'warning' | 'default' =>
  status === 'listed' ? 'warning' : 'success';

function DailyBuysContent({ accountId }: { accountId: number | null }) {
  const [dismissed, setDismissed] = useState(false);
  const [jumpPage, setJumpPage] = useState('');

  const [page, setPage] = useState(0);

  const {
    data: paginated,
    isLoading,
    error,
    refetch,
  } = useDailyBuys(accountId, page + 1, PAGE_SIZE);
  const { isExpanded, toggle } = useExpandableSet();

  const months = useMemo(() => paginated?.months ?? [], [paginated?.months]);
  const totalMonths = paginated?.total ?? 0;

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
          message={`加载每日买入数据失败: ${String(error)}`}
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
      {months.length === 0 && (
        <EmptyState
          icon={<ReceiptIcon sx={{ fontSize: 48 }} />}
          title="暂无买入记录"
          description="同步账户数据后将在此显示每日买入详情。"
        />
      )}

      {months.length > 0 && (
        <React.Fragment>
          <Paper>
            <TableContainer>
              <Table size="small">
                <TableBody>
                  {months.map((m) => {
                    const monthExpanded = isExpanded(m.month);
                    const pl = m.totalMarketValue != null ? m.totalMarketValue - m.totalCost : null;
                    const plRate = m.totalCost > 0 && pl != null ? (pl / m.totalCost) * 100 : null;
                    const mLabel = m.month.replace('-', '年') + '月';

                    return (
                      <React.Fragment key={m.month}>
                        <TableRow
                          hover
                          sx={{
                            bgcolor: 'background.paper',
                            cursor: 'pointer',
                            borderBottom: '2px solid',
                            borderColor: 'divider',
                          }}
                          onClick={() => toggle(m.month)}
                        >
                          <TableCell sx={{ py: 1, width: 40 }}>
                            <IconButton size="small">
                              {monthExpanded ? (
                                <KeyboardArrowDownIcon fontSize="small" />
                              ) : (
                                <KeyboardArrowRightIcon fontSize="small" />
                              )}
                            </IconButton>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} colSpan={2}>
                            <Box sx={{ display: 'flex', gap: 3, alignItems: 'center' }}>
                              <Typography variant="body2" fontWeight={600}>
                                {mLabel}
                              </Typography>
                              <Typography variant="body2" color="text.secondary">
                                {m.totalCount} 件
                              </Typography>
                              <Typography variant="body2">成本 {formatCNY(m.totalCost)}</Typography>
                              {m.totalMarketValue != null && pl != null && (
                                <>
                                  <Typography variant="body2">
                                    市值 {formatCNY(m.totalMarketValue)}
                                  </Typography>
                                  <Typography variant="body2">
                                    盈亏{' '}
                                    <span style={{ color: plHexColor(pl), fontWeight: 600 }}>
                                      {formatCNY(pl)}
                                    </span>
                                  </Typography>
                                  {plRate != null && (
                                    <Typography variant="body2">
                                      盈亏率{' '}
                                      <span style={{ color: plHexColor(plRate) }}>
                                        {plRate >= 0 ? '+' : ''}
                                        {plRate.toFixed(1)}%
                                      </span>
                                    </Typography>
                                  )}
                                </>
                              )}
                            </Box>
                          </TableCell>
                        </TableRow>
                        <TableRow sx={{ '& td': { border: 0 } }}>
                          <TableCell colSpan={3} sx={{ p: 0 }}>
                            <Collapse in={monthExpanded}>
                              <Box>
                                {m.dayGroups.map((group) => {
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
                                            买入 {group.totalCount} 件
                                          </Typography>
                                          <Box
                                            sx={{
                                              display: 'flex',
                                              gap: 2,
                                              flexWrap: 'wrap',
                                              mt: 0.5,
                                            }}
                                          >
                                            <Typography variant="body2">
                                              成本 {formatCNY(group.totalCost)}
                                            </Typography>
                                            {group.totalMarketValue != null ? (
                                              <>
                                                <Typography variant="body2">
                                                  市值{' '}
                                                  <span
                                                    style={{
                                                      color: plHexColor(
                                                        group.totalMarketValue - group.totalCost,
                                                      ),
                                                    }}
                                                  >
                                                    {formatCNY(group.totalMarketValue)}
                                                  </span>
                                                </Typography>
                                                <Typography variant="body2">
                                                  浮动盈亏{' '}
                                                  <span
                                                    style={{
                                                      color: plHexColor(
                                                        group.totalMarketValue - group.totalCost,
                                                      ),
                                                      fontWeight: 600,
                                                    }}
                                                  >
                                                    {formatCNY(
                                                      group.totalMarketValue - group.totalCost,
                                                    )}
                                                  </span>
                                                </Typography>
                                                {group.totalCost > 0 && (
                                                  <Typography variant="body2">
                                                    盈亏率{' '}
                                                    <span
                                                      style={{
                                                        color: plHexColor(
                                                          ((group.totalMarketValue -
                                                            group.totalCost) /
                                                            group.totalCost) *
                                                            100,
                                                        ),
                                                      }}
                                                    >
                                                      {((group.totalMarketValue - group.totalCost) /
                                                        group.totalCost) *
                                                        100 >=
                                                      0
                                                        ? '+'
                                                        : ''}
                                                      {(
                                                        ((group.totalMarketValue -
                                                          group.totalCost) /
                                                          group.totalCost) *
                                                        100
                                                      ).toFixed(1)}
                                                      %
                                                    </span>
                                                  </Typography>
                                                )}
                                              </>
                                            ) : null}
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
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                    >
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
                                                      总额
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      当前市价
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      浮动盈亏
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                      align="right"
                                                    >
                                                      浮动率
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                    >
                                                      平台
                                                    </TableCell>
                                                    <TableCell
                                                      sx={{ fontSize: '0.75rem', py: 0.5 }}
                                                    >
                                                      状态
                                                    </TableCell>
                                                  </TableRow>
                                                </TableHead>
                                                <TableBody>
                                                  {group.items.map((item, idx) => {
                                                    const upl = item.unrealizedPl;
                                                    const uplRate =
                                                      item.totalCost > 0 && upl != null
                                                        ? (upl / item.totalCost) * 100
                                                        : null;
                                                    return (
                                                      <TableRow key={`${group.date}-${idx}`} hover>
                                                        <TableCell sx={{ py: 0.5 }}>
                                                          <Typography
                                                            variant="body2"
                                                            fontWeight={500}
                                                          >
                                                            {item.itemName}
                                                            {item.exterior
                                                              ? ` (${item.exterior})`
                                                              : ''}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography
                                                            variant="body2"
                                                            className="mono-num"
                                                          >
                                                            {item.quantity}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography
                                                            variant="body2"
                                                            className="mono-num"
                                                          >
                                                            {formatCNY(item.buyPrice)}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography
                                                            variant="body2"
                                                            className="mono-num"
                                                          >
                                                            {formatCNY(item.totalCost)}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          <Typography
                                                            variant="body2"
                                                            className="mono-num"
                                                          >
                                                            {item.marketPrice != null
                                                              ? formatCNY(item.marketPrice)
                                                              : '--'}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          {upl != null ? (
                                                            <Typography
                                                              variant="body2"
                                                              color={plHexColor(upl)}
                                                              fontWeight={600}
                                                              className="mono-num"
                                                            >
                                                              {formatCNY(upl)}
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
                                                        <TableCell sx={{ py: 0.5 }} align="right">
                                                          {uplRate != null ? (
                                                            <Typography
                                                              variant="body2"
                                                              color={plHexColor(uplRate)}
                                                              className="mono-num"
                                                            >
                                                              {uplRate >= 0 ? '+' : ''}
                                                              {uplRate.toFixed(1)}%
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
                                                        <TableCell sx={{ py: 0.5 }}>
                                                          <Typography
                                                            variant="body2"
                                                            color="text.secondary"
                                                          >
                                                            {platformLabel[item.platform] ??
                                                              item.platform}
                                                          </Typography>
                                                        </TableCell>
                                                        <TableCell sx={{ py: 0.5 }}>
                                                          <Chip
                                                            label={
                                                              dailyBuyStatusLabel[item.status] ??
                                                              item.status
                                                            }
                                                            size="small"
                                                            color={dailyBuyStatusColor(item.status)}
                                                            variant="outlined"
                                                          />
                                                        </TableCell>
                                                      </TableRow>
                                                    );
                                                  })}
                                                  <TableRow>
                                                    <TableCell colSpan={9} sx={{ py: 0.5 }}>
                                                      <Typography
                                                        variant="caption"
                                                        color="text.secondary"
                                                      >
                                                        当日合计：成本 {formatCNY(group.totalCost)}
                                                        {group.totalMarketValue != null && (
                                                          <span>
                                                            {' '}
                                                            · 当前市值{' '}
                                                            {formatCNY(group.totalMarketValue)} ·
                                                            浮动盈亏{' '}
                                                            <span
                                                              style={{
                                                                color: plHexColor(
                                                                  group.totalMarketValue -
                                                                    group.totalCost,
                                                                ),
                                                              }}
                                                            >
                                                              {formatCNY(
                                                                group.totalMarketValue -
                                                                  group.totalCost,
                                                              )}
                                                            </span>
                                                          </span>
                                                        )}
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
              count={totalMonths}
              page={page}
              rowsPerPage={PAGE_SIZE}
              onPageChange={(_, p) => setPage(p)}
              rowsPerPageOptions={[12]}
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
                    Math.min(Math.ceil(totalMonths / PAGE_SIZE), Number(jumpPage)),
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

export default function InventoryPage() {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useState(false);
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);

  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [sortBy, setSortBy] = useState('itemName');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');

  const [globalFilter, setGlobalFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');

  const { data, isLoading, error, refetch } = useInventory(selectedAccountId, {
    page: page + 1,
    pageSize,
    sortBy,
    sortDir,
    weaponType: typeFilter,
  });

  const groups: GroupRowData[] = useMemo(() => data?.groups ?? [], [data?.groups]);
  const total = data?.total ?? 0;

  const typeFilterOptions = useMemo(() => {
    const types = new Set<string>();
    for (const g of groups) {
      if (g.weaponType) types.add(g.weaponType);
    }
    return Array.from(types).sort();
  }, [groups]);

  const filteredGroups = useMemo(() => {
    let result = groups;
    if (globalFilter)
      result = result.filter((g) => g.itemName.toLowerCase().includes(globalFilter.toLowerCase()));
    return result;
  }, [groups, globalFilter]);

  const [tab, setTab] = useState<'list' | 'dailyBuy'>('list');

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

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h4">持仓</Typography>
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <Select
              value={typeFilter}
              onChange={(e) => {
                setTypeFilter(e.target.value);
                setPage(0);
              }}
              displayEmpty
              sx={{ bgcolor: 'background.paper' }}
            >
              <MenuItem value="">全部类型</MenuItem>
              {typeFilterOptions.map((t) => (
                <MenuItem key={t} value={t}>
                  {t}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <PageSearchBar
            value={globalFilter}
            onChange={setGlobalFilter}
            placeholder="搜索物品名称..."
          />
        </Box>
      </Box>

      <Tabs value={tab} onChange={(_, v) => setTab(v as 'list' | 'dailyBuy')} sx={{ mb: 1 }}>
        <Tab label="持仓列表" value="list" />
        <Tab label="每日买入" value="dailyBuy" />
      </Tabs>

      {tab === 'dailyBuy' && <DailyBuysContent accountId={selectedAccountId} />}

      {tab === 'list' && !selectedAccountId && groups.length === 0 && !isLoading && (
        <Box mt={3}>
          <EmptyState
            icon={<AccountBalanceIcon sx={{ fontSize: 48 }} />}
            title="未选择账户"
            description="选择或添加一个账户以查看持仓。"
            action={{
              label: '前往账户管理',
              onClick: () => {
                void navigate('/accounts');
              },
            }}
          />
        </Box>
      )}

      {tab === 'list' && isLoading && (
        <Box mt={3}>
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} variant="rectangular" height={48} sx={{ mb: 1, borderRadius: 1 }} />
          ))}
        </Box>
      )}

      {tab === 'list' && error && !dismissed && (
        <Box mt={3}>
          <ErrorBanner
            message={`加载持仓数据失败: ${String(error)}`}
            onRetry={() => {
              setDismissed(false);
              void refetch();
            }}
            onDismiss={() => setDismissed(true)}
          />
        </Box>
      )}

      {tab === 'list' && !isLoading && !error && selectedAccountId && groups.length === 0 && (
        <Box mt={3}>
          <EmptyState
            icon={<InventoryIcon sx={{ fontSize: 48 }} />}
            title="暂无持仓物品"
            description="同步账户数据后将在此显示持仓。"
          />
        </Box>
      )}

      {tab === 'list' &&
        !isLoading &&
        !error &&
        groups.length > 0 &&
        filteredGroups.length === 0 && (
          <Box mt={3}>
            <EmptyState
              icon={<SearchOffIcon sx={{ fontSize: 48 }} />}
              title="无匹配物品"
              description="请尝试更改类型筛选或搜索条件。"
            />
          </Box>
        )}

      {tab === 'list' && !isLoading && !error && filteredGroups.length > 0 && (
        <Box mt={3}>
          <Paper>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    {groupedColumns.map((col) => {
                      const headerId =
                        (col as { id?: string }).id ??
                        (col as { accessorKey?: string }).accessorKey ??
                        '';
                      const canSort = col.enableSorting !== false;
                      const isSorted = canSort && sortBy === headerId ? sortDir : false;
                      return (
                        <TableCell
                          key={headerId}
                          align={col.meta?.align}
                          sortDirection={isSorted || false}
                          sx={{ py: 1 }}
                        >
                          {canSort ? (
                            <TableSortLabel
                              active={!!isSorted}
                              direction={isSorted === 'desc' ? 'desc' : 'asc'}
                              onClick={() => {
                                if (isSorted === 'asc') handleSort(headerId, 'desc');
                                else if (isSorted === 'desc') handleSort('itemName', 'asc');
                                else handleSort(headerId, 'asc');
                              }}
                            >
                              {typeof col.header === 'string' ? col.header : headerId}
                            </TableSortLabel>
                          ) : typeof col.header === 'string' ? (
                            col.header
                          ) : (
                            headerId
                          )}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredGroups.map((group) => {
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
                            {group.weaponType ? (
                              <Chip label={group.weaponType} size="small" variant="outlined" />
                            ) : null}
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
                                    <OpenInNewIcon fontSize="small" />
                                  </IconButton>
                                </Tooltip>
                              ) : null}
                            </Box>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" className="mono-num">
                              {group.totalQuantity.toLocaleString()}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" className="mono-num">
                              {formatCNY(group.totalBuyPrice)}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" className="mono-num">
                              {formatCNY(group.avgBuyPrice)}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" color="text.secondary" className="mono-num">
                              {group.marketPrice != null ? formatCNY(group.marketPrice) : '--'}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="body2" color="text.secondary" className="mono-num">
                              {group.marketPrice != null
                                ? formatCNY(group.marketPrice * group.totalQuantity)
                                : '--'}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            <Typography variant="caption" color="text.disabled">
                              {group.marketPriceUpdatedAt
                                ? new Date(group.marketPriceUpdatedAt * 1000).toLocaleString()
                                : '--'}
                            </Typography>
                          </TableCell>
                          <TableCell sx={{ py: 1 }} align="right">
                            {group.unrealizedPl != null ? (
                              <Typography
                                variant="body2"
                                color={plHexColor(group.unrealizedPl)}
                                className="mono-num"
                              >
                                {formatCNY(group.unrealizedPl)}
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
                            {group.marketPrice != null && group.avgBuyPrice > 0 ? (
                              <Typography
                                variant="body2"
                                color={plHexColor(
                                  ((group.marketPrice - group.avgBuyPrice) / group.avgBuyPrice) *
                                    100,
                                )}
                              >
                                {((group.marketPrice - group.avgBuyPrice) / group.avgBuyPrice) *
                                  100 >=
                                0
                                  ? '+'
                                  : ''}
                                {(
                                  ((group.marketPrice - group.avgBuyPrice) / group.avgBuyPrice) *
                                  100
                                ).toFixed(1)}
                                %
                              </Typography>
                            ) : (
                              <Typography variant="body2" color="text.secondary">
                                --
                              </Typography>
                            )}
                          </TableCell>
                        </TableRow>
                        <TableRow sx={{ '& td': { border: 0 } }}>
                          <TableCell colSpan={groupedColumns.length} sx={{ p: 0 }}>
                            <Collapse in={expanded}>
                              <Box sx={{ mx: 2, my: 1 }}>
                                <Table size="small">
                                  <TableHead>
                                    <TableRow>
                                      <TableCell sx={{ fontSize: '0.75rem' }}>物品名称</TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem' }}>磨损</TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem' }} align="right">
                                        磨损值
                                      </TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem' }} align="right">
                                        数量
                                      </TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem' }}>状态</TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem' }} align="right">
                                        买入价
                                      </TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem' }}>买入日期</TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem' }}>平台</TableCell>
                                      <TableCell sx={{ fontSize: '0.75rem' }} align="right">
                                        上架价
                                      </TableCell>
                                    </TableRow>
                                  </TableHead>
                                  <TableBody>
                                    {group.instances.map((inst) => (
                                      <TableRow
                                        key={`${inst.accountId}-${inst.assetId}`}
                                        hover
                                        sx={{ cursor: 'pointer' }}
                                        onClick={() => {
                                          void navigate(
                                            `/inventory/${inst.accountId}/${inst.assetId}`,
                                          );
                                        }}
                                      >
                                        <TableCell sx={{ py: 0.5 }}>
                                          <Typography variant="body2">
                                            {inst.itemName ?? '--'}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }}>
                                          <Typography variant="body2" color="text.secondary">
                                            {inst.exterior || '--'}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" color="text.secondary">
                                            {inst.paintWear != null
                                              ? inst.paintWear.toFixed(8)
                                              : '--'}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {(inst.quantity ?? 1).toLocaleString()}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }}>
                                          <Chip
                                            label={inventoryStatusLabel[inst.status] ?? inst.status}
                                            size="small"
                                            color={inventoryStatusColor[inst.status] ?? 'default'}
                                            variant="outlined"
                                          />
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {inst.buyTrade
                                              ? formatCNY(inst.buyTrade.unitPrice)
                                              : '--'}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }}>
                                          <Typography variant="body2" color="text.secondary">
                                            {inst.buyTrade
                                              ? new Date(inst.buyTrade.tradeAt).toLocaleDateString()
                                              : '--'}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }}>
                                          <Typography variant="body2" color="text.secondary">
                                            {inst.buyTrade
                                              ? (platformLabel[inst.buyTrade.source] ??
                                                inst.buyTrade.source)
                                              : '--'}
                                          </Typography>
                                        </TableCell>
                                        <TableCell sx={{ py: 0.5 }} align="right">
                                          <Typography variant="body2" className="mono-num">
                                            {inst.status === 'listed' && inst.listedPrice != null
                                              ? formatCNY(inst.listedPrice)
                                              : '--'}
                                          </Typography>
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
        </Box>
      )}
    </Box>
  );
}
