import { useState } from "react";
import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  useReactTable,
  type ColumnDef,
  type FilterFn,
  type SortingState,
} from "@tanstack/react-table";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableSortLabel from "@mui/material/TableSortLabel";
import TablePagination from "@mui/material/TablePagination";
import Paper from "@mui/material/Paper";

declare module "@tanstack/react-table" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData, TValue> {
    align?: "left" | "right" | "center";
  }
}

interface PaginationProps {
  pageIndex: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

interface SortableTableProps<T> {
  columns: ColumnDef<T>[];
  data: T[];
  getRowId: (row: T) => string;
  globalFilter?: string;
  onGlobalFilterChange?: (value: string) => void;
  globalFilterFn?: FilterFn<T>;
  onRowClick?: (row: T) => void;
  meta?: Record<string, unknown>;
  pagination?: PaginationProps;
  manualSorting?: boolean;
  onSortingChange?: (sortBy: string, sortDir: string) => void;
}

export default function SortableTable<T>({
  columns,
  data,
  getRowId,
  globalFilter,
  onGlobalFilterChange,
  globalFilterFn,
  onRowClick,
  meta,
  pagination,
  manualSorting,
  onSortingChange,
}: SortableTableProps<T>) {
  const [sorting, setSorting] = useState<SortingState>([]);

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      globalFilter: globalFilter ?? "",
      ...(pagination ? { pagination: { pageIndex: pagination.pageIndex, pageSize: pagination.pageSize } } : {}),
    },
    onSortingChange: (updater) => {
      const next = typeof updater === "function" ? updater(sorting) : updater;
      setSorting(next);
      if (manualSorting && onSortingChange && next.length > 0) {
        onSortingChange(next[0].id, next[0].desc ? "desc" : "asc");
      } else if (manualSorting && onSortingChange && next.length === 0) {
        onSortingChange("itemName", "asc");
      }
    },
    onGlobalFilterChange: onGlobalFilterChange ?? (() => {}),
    globalFilterFn: globalFilterFn ?? "auto",
    getCoreRowModel: getCoreRowModel(),
    getRowId,
    meta,
    ...(pagination
      ? {
          manualPagination: true,
          pageCount: Math.ceil(pagination.total / pagination.pageSize),
          getPaginationRowModel: getPaginationRowModel(),
        }
      : {
          getSortedRowModel: getSortedRowModel(),
          getFilteredRowModel: getFilteredRowModel(),
        }),
  });

  return (
    <Paper>
      <TableContainer>
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
            {table.getRowModel().rows.map((row) => (
              <TableRow
                key={row.id}
                hover
                sx={onRowClick ? { cursor: "pointer" } : undefined}
                onClick={onRowClick ? () => onRowClick(row.original) : undefined}
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    align={cell.column.columnDef.meta?.align}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      {pagination && (
        <TablePagination
          component="div"
          count={pagination.total}
          page={pagination.pageIndex}
          rowsPerPage={pagination.pageSize}
          onPageChange={(_, page) => pagination.onPageChange(page)}
          onRowsPerPageChange={(e) => pagination.onPageSizeChange(Number(e.target.value))}
          rowsPerPageOptions={[20, 50, 100]}
        />
      )}
    </Paper>
  );
}
