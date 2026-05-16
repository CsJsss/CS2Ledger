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
import { platformLabel } from "../lib/constants";
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
      header: "Name",
      cell: (info) => (
        <Typography variant="body2" fontWeight={500}>{info.getValue() as string}</Typography>
      ),
    },
    {
      accessorKey: "platform",
      header: "Platform",
      cell: (info) => (
        <Typography variant="body2" color="text.secondary">
          {platformLabel[info.getValue() as string] ?? (info.getValue() as string)}
        </Typography>
      ),
    },
    {
      accessorKey: "status",
      header: "Status",
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
      header: "Avail. Balance",
      meta: { align: "right" },
      cell: (info) => <>{formatCNY(info.getValue() as number)}</>,
    },
    {
      accessorKey: "purchaseBalance",
      header: "Purch. Balance",
      meta: { align: "right" },
      cell: (info) => <>{formatCNY(info.getValue() as number)}</>,
    },
    {
      accessorKey: "lastSyncAt",
      header: "Last Sync",
      meta: { align: "right" },
      cell: (info) => {
        const v = info.getValue() as number | undefined;
        return (
          <Typography variant="caption" color="text.secondary">
            {v ? new Date(v * 1000).toLocaleString() : "Never"}
          </Typography>
        );
      },
    },
    {
      id: "actions",
      header: "Actions",
      meta: { align: "right" },
      enableSorting: false,
      cell: (info) => {
        const acc = info.row.original;
        return (
          <Box sx={{ display: "flex", gap: 0.5, justifyContent: "flex-end" }}>
            <Button
              size="small"
              onClick={() => {
                setSyncDialogId(acc.ID);
              }}
              disabled={syncMut.isPending}
            >
              {syncMut.isPending ? "Syncing..." : "Sync"}
            </Button>
            <Button
              size="small"
              onClick={() => setEditingAccount({ ID: acc.ID, name: acc.name, platform: acc.platform })}
            >
              Edit
            </Button>
            <Button
              size="small"
              color="error"
              onClick={() => setDeletingId(acc.ID)}
            >
              Delete
            </Button>
          </Box>
        );
      },
    },
  ];

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3, gap: 1 }}>
        <Typography variant="h4">Accounts</Typography>
        <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
          <PageSearchBar value={searchQuery} onChange={setSearchQuery} placeholder="Filter by name..." />
          <Button variant="contained" onClick={() => setShowAdd(true)}>
            + Add Account
          </Button>
        </Box>
      </Box>

      {isLoading && <Typography color="text.secondary">Loading...</Typography>}

      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          Failed to load accounts. Make sure the app is running.
        </Alert>
      )}

      {!isLoading && !error && accounts.length === 0 && (
        <Box sx={{ textAlign: "center", py: 6 }}>
          <Typography color="text.secondary">No accounts configured.</Typography>
          <Typography variant="body2" color="text.secondary" mt={1}>
            Add your first platform account to get started.
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
          Sync failed: {String(syncMut.error)}
        </Alert>
      )}

      {syncMut.data && syncMut.data.warnings && syncMut.data.warnings.length > 0 && (
        <Alert severity="warning" sx={{ mt: 2 }}>
          Sync completed with warnings:
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
        title="Delete Account"
        actions={
          <>
            <Button onClick={() => setDeletingId(null)}>Cancel</Button>
            <Button
              color="error"
              variant="contained"
              disabled={deleteMut.isPending}
              onClick={() => deletingId !== null && deleteMut.mutate(deletingId, { onSuccess: () => setDeletingId(null) })}
            >
              {deleteMut.isPending ? "Deleting..." : "Delete"}
            </Button>
          </>
        }
      >
        <Typography variant="body2" color="text.secondary">
          Delete this account and all related data? This cannot be undone.
        </Typography>
      </Dialog>

      <Dialog
        open={syncDialogId !== null}
        onClose={() => setSyncDialogId(null)}
        title="Sync Account"
        actions={
          <>
            <Button onClick={() => setSyncDialogId(null)}>Cancel</Button>
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
              {syncMut.isPending ? "Syncing..." : "Sync"}
            </Button>
          </>
        }
      >
        Sync account data from the last synced point.
      </Dialog>
    </Box>
  );
}
