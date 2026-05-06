import { useState } from "react";
import { useNavigate } from "react-router";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Chip from "@mui/material/Chip";
import Skeleton from "@mui/material/Skeleton";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import ErrorBanner from "../components/ErrorBanner";
import EmptyState from "../components/EmptyState";
import { useInventory } from "../hooks/useInventory";
import { useUIStore } from "../store/uiStore";
import { formatCNY } from "../lib/format";
import { inventoryStatusLabel, inventoryStatusColor } from "../lib/constants";

export default function InventoryPage() {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useState(false);
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const { data: items = [], isLoading, error, refetch } = useInventory(selectedAccountId);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Inventory</Typography>

      {!selectedAccountId && (
        <Box mt={3}>
          <EmptyState
            title="No account selected"
            description="Select or add an account to view inventory."
            action={{ label: "Go to Accounts", onClick: () => { void navigate("/accounts"); } }}
          />
        </Box>
      )}

      {selectedAccountId && isLoading && (
        <Box mt={3}>
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} variant="rectangular" height={48} sx={{ mb: 1, borderRadius: 1 }} />
          ))}
        </Box>
      )}

      {selectedAccountId && error && !dismissed && (
        <Box mt={3}>
          <ErrorBanner
            message={`Failed to load inventory: ${String(error)}`}
            onRetry={() => { setDismissed(false); void refetch(); }}
            onDismiss={() => setDismissed(true)}
          />
        </Box>
      )}

      {selectedAccountId && !isLoading && !error && items.length === 0 && (
        <Box mt={3}>
          <EmptyState title="No inventory items" description="Sync your account to populate inventory data." />
        </Box>
      )}

      {selectedAccountId && !isLoading && !error && items.length > 0 && (
        <TableContainer component={Paper} sx={{ mt: 3 }}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Item Name</TableCell>
                <TableCell>Exterior</TableCell>
                <TableCell>Status</TableCell>
                <TableCell align="right">Listed Price</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {items.map((item) => (
                <TableRow
                  key={`${item.accountId}-${item.assetId}`}
                  hover
                  sx={{ cursor: "pointer" }}
                  onClick={() => { void navigate(`/inventory/${item.accountId}/${item.assetId}`); }}
                >
                  <TableCell>
                    <Typography variant="body2" fontWeight={500}>{item.itemName}</Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">{item.exterior || "--"}</Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={inventoryStatusLabel[item.status] ?? item.status}
                      size="small"
                      color={inventoryStatusColor[item.status] ?? "default"}
                      variant="outlined"
                    />
                  </TableCell>
                  <TableCell align="right">
                    {item.status === "listed" && item.listedPrice
                      ? formatCNY(item.listedPrice)
                      : "--"}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
}
