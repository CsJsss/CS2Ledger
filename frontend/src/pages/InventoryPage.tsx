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
import type { model } from '../lib/wails';

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
];

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

      {!selectedAccountId && groups.length === 0 && !isLoading && (
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

      {isLoading && (
        <Box mt={3}>
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} variant="rectangular" height={48} sx={{ mb: 1, borderRadius: 1 }} />
          ))}
        </Box>
      )}

      {error && !dismissed && (
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

      {!isLoading && !error && selectedAccountId && groups.length === 0 && (
        <Box mt={3}>
          <EmptyState
            icon={<InventoryIcon sx={{ fontSize: 48 }} />}
            title="暂无持仓物品"
            description="同步账户数据后将在此显示持仓。"
          />
        </Box>
      )}

      {!isLoading && !error && groups.length > 0 && filteredGroups.length === 0 && (
        <Box mt={3}>
          <EmptyState
            icon={<SearchOffIcon sx={{ fontSize: 48 }} />}
            title="无匹配物品"
            description="请尝试更改类型筛选或搜索条件。"
          />
        </Box>
      )}

      {!isLoading && !error && filteredGroups.length > 0 && (
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
