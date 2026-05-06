import { useState } from "react";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Alert from "@mui/material/Alert";
import Dialog from "../components/Dialog";
import AddAccountDialog from "../components/AddAccountDialog";
import { useSyncAccount } from "../hooks/useSyncAccount";
import { useUIStore } from "../store/uiStore";
import { useAccounts } from "../hooks/useAccounts";
import { useCreateAccount } from "../hooks/useCreateAccount";
import { useUpdateAccount } from "../hooks/useUpdateAccount";
import { useDeleteAccount } from "../hooks/useDeleteAccount";
import { fmt } from "../lib/format";
import { platformLabel } from "../lib/constants";

export default function AccountsPage() {
  const setSelectedAccount = useUIStore((s) => s.setSelectedAccount);
  const [showAdd, setShowAdd] = useState(false);
  const [editingAccount, setEditingAccount] = useState<{ ID: number; name: string; platform: string } | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);

  const { data: accounts = [], isLoading, error } = useAccounts();
  const createMut = useCreateAccount();
  const updateMut = useUpdateAccount();
  const deleteMut = useDeleteAccount();
  const syncMut = useSyncAccount();

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
        <Typography variant="h4">Accounts</Typography>
        <Button variant="contained" onClick={() => setShowAdd(true)}>
          + Add Account
        </Button>
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
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Platform</TableCell>
                <TableCell>Status</TableCell>
                <TableCell align="right">Avail. Balance</TableCell>
                <TableCell align="right">Purch. Balance</TableCell>
                <TableCell align="right">Last Sync</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {accounts.map((acc) => (
                <TableRow key={acc.ID} hover>
                  <TableCell>
                    <Typography variant="body2" fontWeight={500}>{acc.name}</Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      {platformLabel[acc.platform] ?? acc.platform}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={acc.status}
                      size="small"
                      color={acc.status === "active" ? "success" : "default"}
                      variant="outlined"
                    />
                  </TableCell>
                  <TableCell align="right">&yen;{fmt(acc.availableBalance)}</TableCell>
                  <TableCell align="right">&yen;{fmt(acc.purchaseBalance)}</TableCell>
                  <TableCell align="right">
                    <Typography variant="caption" color="text.secondary">
                      {acc.lastSyncAt ? new Date(acc.lastSyncAt * 1000).toLocaleString() : "Never"}
                    </Typography>
                  </TableCell>
                  <TableCell align="right">
                    <Box sx={{ display: "flex", gap: 0.5, justifyContent: "flex-end" }}>
                      <Button
                        size="small"
                        onClick={() => { setSelectedAccount(acc.ID); syncMut.mutate(acc.ID); }}
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
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {syncMut.isError && (
        <Alert severity="error" sx={{ mt: 2 }}>
          Sync failed: {String(syncMut.error)}
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
    </Box>
  );
}
