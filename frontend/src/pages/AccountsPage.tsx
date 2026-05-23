import { useState } from "react";
import { type ColumnDef } from "@tanstack/react-table";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";

import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Alert from "@mui/material/Alert";
import Dialog from "../components/Dialog";
import AddAccountDialog from "../components/AddAccountDialog";
import SortableTable from "../components/SortableTable";
import PageSearchBar from "../components/PageSearchBar";
import { useSyncAccount } from "../hooks/useSyncAccount";
import { useUIStore } from "../store/uiStore";
import { useAccounts } from "../hooks/useAccounts";
import { useCreateAccount } from "../hooks/useCreateAccount";
import { useUpdateAccount } from "../hooks/useUpdateAccount";
import { useDeleteAccount } from "../hooks/useDeleteAccount";
import { formatCNY } from "../lib/format";
import { platformLabel, PLATFORM_CSQAQ } from "../lib/constants";
import type { model } from "../lib/wails";

export default function AccountsPage() {
  const setSelectedAccount = useUIStore((s) => s.setSelectedAccount);
  const [showAdd, setShowAdd] = useState(false);
  const [editingAccount, setEditingAccount] = useState<{ ID: number; name: string; platform: string } | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [syncDialogId, setSyncDialogId] = useState<number | null>(null);

  const { data: accounts = [], isLoading, error } = useAccounts();
  const createMut = useCreateAccount();
  const updateMut = useUpdateAccount();
  const deleteMut = useDeleteAccount();
  const syncMut = useSyncAccount();

  const [searchQuery, setSearchQuery] = useState("");

  const columns: ColumnDef<model.Account>[] = [
    {
      accessorKey: "name",
      header: "名称",
      cell: (info) => (
        <Typography variant="body2" fontWeight={500}>{info.getValue() as string}</Typography>
      ),
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
      accessorKey: "status",
      header: "状态",
      cell: (info) => (
        <Chip
          label={info.getValue() as string}
          size="small"
          color={(info.getValue() as string) === "active" ? "success" : "default"}
          variant="outlined"
        />
      ),
    },
    {
      accessorKey: "availableBalance",
      header: "钱包余额",
      meta: { align: "right" },
      cell: (info) => <>{formatCNY(info.getValue() as number)}</>,
    },
    {
      accessorKey: "frozenBalance",
      header: "冻结余额",
      meta: { align: "right" },
      cell: (info) => <>{formatCNY(info.getValue() as number)}</>,
    },
    {
      accessorKey: "instantBalance",
      header: "秒到账余额",
      meta: { align: "right" },
      cell: (info) => <>{formatCNY(info.getValue() as number)}</>,
    },
    {
      accessorKey: "purchaseBalance",
      header: "求购余额",
      meta: { align: "right" },
      cell: (info) => <>{formatCNY(info.getValue() as number)}</>,
    },
    {
      accessorKey: "lastSyncAt",
      header: "最后同步",
      meta: { align: "right" },
      cell: (info) => {
        const v = info.getValue() as number | undefined;
        return (
          <Typography variant="caption" color="text.secondary">
            {v ? new Date(v * 1000).toLocaleString() : "从未"}
          </Typography>
        );
      },
    },
    {
      id: "actions",
      header: "操作",
      meta: { align: "right" },
      enableSorting: false,
      cell: (info) => {
        const acc = info.row.original;
        return (
          <Box sx={{ display: "flex", gap: 0.5, justifyContent: "flex-end" }}>
            {acc.platform === PLATFORM_CSQAQ ? (
              <Button size="small" disabled title="行情数据自动获取，无需手动同步">
                无需同步
              </Button>
            ) : (
              <Button
                size="small"
                onClick={() => {
                  setSyncDialogId(acc.ID);
                }}
                disabled={syncMut.isPending}
              >
                {syncMut.isPending ? "同步中..." : "同步"}
              </Button>
            )}
            <Button
              size="small"
              onClick={() => setEditingAccount({ ID: acc.ID, name: acc.name, platform: acc.platform })}
            >
              编辑
            </Button>
            <Button
              size="small"
              color="error"
              onClick={() => setDeletingId(acc.ID)}
            >
              删除
            </Button>
          </Box>
        );
      },
    },
  ];

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3, gap: 1 }}>
        <Typography variant="h4">账户管理</Typography>
        <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
          <PageSearchBar value={searchQuery} onChange={setSearchQuery} placeholder="搜索名称..." />
          <Button variant="contained" onClick={() => setShowAdd(true)}>
            + 添加账户
          </Button>
        </Box>
      </Box>

      {isLoading && <Typography color="text.secondary">加载中...</Typography>}

      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          加载账户失败，请确认应用正在运行。
        </Alert>
      )}

      {!isLoading && !error && accounts.length === 0 && (
        <Box sx={{ textAlign: "center", py: 6 }}>
          <Typography color="text.secondary">暂无已配置的账户。</Typography>
          <Typography variant="body2" color="text.secondary" mt={1}>
            添加你的第一个平台账户以开始使用。
          </Typography>
        </Box>
      )}

      {accounts.length > 0 && (
        <SortableTable
          columns={columns}
          data={accounts}
          globalFilter={searchQuery}
          onGlobalFilterChange={setSearchQuery}
          globalFilterFn={(row, _columnId, filterValue) =>
            row.original.name.toLowerCase().includes((filterValue as string).toLowerCase())
          }
          getRowId={(acc) => String(acc.ID)}
        />
      )}

      {syncMut.isError && (
        <Alert severity="error" sx={{ mt: 2 }}>
          同步失败: {String(syncMut.error)}
        </Alert>
      )}

      {syncMut.data && syncMut.data.warnings && syncMut.data.warnings.length > 0 && (
        <Alert severity="warning" sx={{ mt: 2 }}>
          同步完成，但有以下警告:
          <ul style={{ margin: 0, paddingLeft: 20 }}>
            {syncMut.data.warnings.map((w, i) => (
              <li key={i}>{w}</li>
            ))}
          </ul>
        </Alert>
      )}

      {showAdd && (
        <AddAccountDialog
          open={showAdd}
          onClose={() => { setShowAdd(false); createMut.reset(); }}
          onSubmit={(data) => createMut.mutate(data, { onSuccess: () => setShowAdd(false) })}
          isPending={createMut.isPending}
          error={createMut.error ? String(createMut.error) : null}
        />
      )}

      {editingAccount && (
        <AddAccountDialog
          open={true}
          editMode
          initialValues={{ name: editingAccount.name, platform: editingAccount.platform }}
          onClose={() => { setEditingAccount(null); updateMut.reset(); }}
          onSubmit={(data) => updateMut.mutate({ id: editingAccount.ID, name: data.name, cookie: data.cookie }, { onSuccess: () => setEditingAccount(null) })}
          isPending={updateMut.isPending}
          error={updateMut.error ? String(updateMut.error) : null}
        />
      )}

      <Dialog
        open={deletingId !== null}
        onClose={() => setDeletingId(null)}
        title="删除账户"
        actions={
          <>
            <Button onClick={() => setDeletingId(null)}>取消</Button>
            <Button
              color="error"
              variant="contained"
              disabled={deleteMut.isPending}
              onClick={() => deletingId !== null && deleteMut.mutate(deletingId, { onSuccess: () => setDeletingId(null) })}
            >
              {deleteMut.isPending ? "删除中..." : "删除"}
            </Button>
          </>
        }
      >
        <Typography variant="body2" color="text.secondary">
          删除此账户及其所有相关数据？此操作不可撤销。
        </Typography>
      </Dialog>

      <Dialog
        open={syncDialogId !== null}
        onClose={() => setSyncDialogId(null)}
        title="同步账户"
        actions={
          <>
            <Button onClick={() => setSyncDialogId(null)}>取消</Button>
            <Button
              variant="contained"
              disabled={syncMut.isPending}
              onClick={() => {
                if (syncDialogId === null) return;
                setSelectedAccount(syncDialogId);
                syncMut.mutate({ accountId: syncDialogId });
                setSyncDialogId(null);
              }}
            >
              {syncMut.isPending ? "同步中..." : "同步"}
            </Button>
          </>
        }
      >
        从上次同步点同步账户数据。
      </Dialog>
    </Box>
  );
}
