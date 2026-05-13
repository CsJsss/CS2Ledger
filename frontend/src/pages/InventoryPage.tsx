import React, { useState, useMemo } from "react";
import { useNavigate } from "react-router";
import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  type ColumnDef,
  type SortingState,
} from "@tanstack/react-table";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableSortLabel from "@mui/material/TableSortLabel";
import Paper from "@mui/material/Paper";
import Chip from "@mui/material/Chip";
import IconButton from "@mui/material/IconButton";
import Collapse from "@mui/material/Collapse";
import Skeleton from "@mui/material/Skeleton";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import ErrorBanner from "../components/ErrorBanner";
import EmptyState from "../components/EmptyState";
import PageSearchBar from "../components/PageSearchBar";
import { useInventory } from "../hooks/useInventory";
import { useUIStore } from "../store/uiStore";
import { formatCNY } from "../lib/format";
import { inventoryStatusLabel, inventoryStatusColor, platformLabel } from "../lib/constants";
import FormControl from "@mui/material/FormControl";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowRightIcon from "@mui/icons-material/KeyboardArrowRight";
import type { model } from "../lib/wails";

declare module "@tanstack/react-table" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData, TValue> {
    align?: "left" | "right" | "center";
  }
}

interface GroupedItem {
  itemName: string;
  weaponType: string;
  count: number;
  instances: model.InventoryItem[];
}

const groupedColumns: ColumnDef<GroupedItem>[] = [
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
    accessorKey: "weaponType",
    header: "类型",
    cell: (info) => {
      const wt = info.getValue() as string;
      return wt ? <Chip label={wt} size="small" variant="outlined" /> : null;
    },
  },
  {
    accessorKey: "itemName",
    header: "Item Name",
    cell: (info) => (
      <Typography variant="body2" fontWeight={500}>
        {info.getValue() as string}
      </Typography>
    ),
  },
  {
    accessorKey: "count",
    header: "Count",
    meta: { align: "right" },
    cell: (info) => (
      <Typography variant="body2">{(info.getValue() as number).toLocaleString()}</Typography>
    ),
  },
  {
    id: "statusSummary",
    header: "Status",
    cell: ({ row }) => {
      const instances = row.original.instances;
      const inInv = instances.filter((i) => i.status === "in_inventory").reduce((s, i) => s + (i.quantity ?? 1), 0);
      const listed = instances.filter((i) => i.status === "listed").reduce((s, i) => s + (i.quantity ?? 1), 0);
      return (
        <Box sx={{ display: "flex", gap: 0.5 }}>
          {inInv > 0 && (
            <Chip
              size="small"
              label={`${inInv} in inv`}
              color={inventoryStatusColor["in_inventory"] ?? "default"}
              variant="outlined"
            />
          )}
          {listed > 0 && (
            <Chip
              size="small"
              label={`${listed} listed`}
              color={inventoryStatusColor["listed"] ?? "default"}
              variant="outlined"
            />
          )}
        </Box>
      );
    },
  },
];

export default function InventoryPage() {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useState(false);
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const { data: items = [], isLoading, error, refetch } = useInventory(selectedAccountId);
  const [expandedNames, setExpandedNames] = useState<Set<string>>(new Set());

  const [globalFilter, setGlobalFilter] = useState("");
  const [typeFilter, setTypeFilter] = useState("");

  const grouped = useMemo(() => {
    const map = new Map<string, model.InventoryItem[]>();
    for (const item of items) {
      const name = item.itemName ?? "Unknown";
      const arr = map.get(name);
      if (arr) arr.push(item);
      else map.set(name, [item]);
    }
    return Array.from(map, ([itemName, instances]) => ({
      itemName,
      weaponType: instances[0]?.weaponType ?? "",
      count: instances.reduce((sum, i) => sum + (i.quantity ?? 1), 0),
      instances,
    })).sort((a, b) => a.itemName.localeCompare(b.itemName));
  }, [items]);

  const typeFilterOptions = useMemo(() => {
    const types = new Set<string>();
    for (const g of grouped) {
      if (g.weaponType) types.add(g.weaponType);
    }
    return Array.from(types).sort();
  }, [grouped]);

  const filteredGrouped = useMemo(() => {
    if (!typeFilter) return grouped;
    return grouped.filter((g) => g.weaponType === typeFilter);
  }, [grouped, typeFilter]);

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
    data: filteredGrouped,
    columns: groupedColumns,
    state: { sorting, globalFilter },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) =>
      row.original.itemName.toLowerCase().includes((filterValue as string).toLowerCase()),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getCoreRowModel: getCoreRowModel(),
    getRowId: (row) => row.itemName,
  });

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 2 }}>
        <Typography variant="h4">Inventory</Typography>
        <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <Select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              displayEmpty
              sx={{ bgcolor: "grey.100" }}
            >
              <MenuItem value="">全部类型</MenuItem>
              {typeFilterOptions.map((t) => (
                <MenuItem key={t} value={t}>{t}</MenuItem>
              ))}
            </Select>
          </FormControl>
          <PageSearchBar value={globalFilter} onChange={setGlobalFilter} placeholder="Filter by item name..." />
        </Box>
      </Box>

      {!selectedAccountId && items.length === 0 && !isLoading && (
        <Box mt={3}>
          <EmptyState
            title="No account selected"
            description="Select or add an account to view inventory."
            action={{ label: "Go to Accounts", onClick: () => { void navigate("/accounts"); } }}
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
            message={`Failed to load inventory: ${String(error)}`}
            onRetry={() => { setDismissed(false); void refetch(); }}
            onDismiss={() => setDismissed(true)}
          />
        </Box>
      )}

      {!isLoading && !error && items.length === 0 && selectedAccountId && (
        <Box mt={3}>
          <EmptyState
            title="No inventory items"
            description="Sync your account to populate inventory data."
          />
        </Box>
      )}

      {!isLoading && !error && items.length > 0 && filteredGrouped.length === 0 && (
        <Box mt={3}>
          <EmptyState
            title="No matching items"
            description="Try changing the type filter or search query."
          />
        </Box>
      )}

      {!isLoading && !error && filteredGrouped.length > 0 && (
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
                                if (sorted === "desc") {
                                  header.column.clearSorting();
                                } else {
                                  header.column.toggleSorting(sorted === "asc");
                                }
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
                              row: {
                                ...groupRow,
                                getIsExpanded: () => expanded,
                              },
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
                                    <TableCell sx={{ fontSize: "0.75rem" }}>Item Name</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem" }}>Exterior</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem" }} align="right">Wear</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem" }} align="right">Qty</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem" }}>Status</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem" }} align="right">Buy Price</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem" }}>Buy Date</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem" }}>Platform</TableCell>
                                    <TableCell sx={{ fontSize: "0.75rem" }} align="right">Listed Price</TableCell>
                                  </TableRow>
                                </TableHead>
                                <TableBody>
                                  {groupRow.original.instances.map((inst) => (
                                    <TableRow
                                      key={`${inst.accountId}-${inst.assetId}`}
                                      hover
                                      sx={{ cursor: "pointer" }}
                                      onClick={() => {
                                        void navigate(`/inventory/${inst.accountId}/${inst.assetId}`);
                                      }}
                                    >
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2">{inst.itemName ?? "--"}</Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" color="text.secondary">
                                          {inst.exterior || "--"}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2" color="text.secondary">
                                          {inst.paintWear != null ? inst.paintWear.toFixed(8) : "--"}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2">
                                          {(inst.quantity ?? 1).toLocaleString()}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Chip
                                          label={inventoryStatusLabel[inst.status] ?? inst.status}
                                          size="small"
                                          color={inventoryStatusColor[inst.status] ?? "default"}
                                          variant="outlined"
                                        />
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2">
                                          {inst.buyTrade ? formatCNY(inst.buyTrade.unitPrice) : "--"}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" color="text.secondary">
                                          {inst.buyTrade
                                            ? new Date(inst.buyTrade.tradeAt).toLocaleDateString()
                                            : "--"}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }}>
                                        <Typography variant="body2" color="text.secondary">
                                          {inst.buyTrade
                                            ? (platformLabel[inst.buyTrade.source] ?? inst.buyTrade.source)
                                            : "--"}
                                        </Typography>
                                      </TableCell>
                                      <TableCell sx={{ py: 0.5 }} align="right">
                                        <Typography variant="body2">
                                          {inst.status === "listed" && inst.listedPrice != null
                                            ? formatCNY(inst.listedPrice)
                                            : "--"}
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
        </Box>
      )}
    </Box>
  );
}
