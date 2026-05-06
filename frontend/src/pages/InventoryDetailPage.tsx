import { useState } from "react";
import { useNavigate, useParams } from "react-router";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import Chip from "@mui/material/Chip";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Skeleton from "@mui/material/Skeleton";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Grid from "@mui/material/Grid";
import ErrorBanner from "../components/ErrorBanner";
import EmptyState from "../components/EmptyState";
import { useItemDetail } from "../hooks/useItemDetail";
import { formatCNY } from "../lib/format";

const statusLabel: Record<string, string> = {
  in_inventory: "In Storage",
  listed: "Listed",
  rented: "Rented",
};

export default function InventoryDetailPage() {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useState(false);
  const { accountId, assetId } = useParams<{ accountId: string; assetId: string }>();
  const accountIdNum = accountId ? Number(accountId) : null;
  const { data: detail, isLoading, error, refetch } = useItemDetail(accountIdNum, assetId ?? null);

  return (
    <Box>
      <Button
        onClick={() => { void navigate("/inventory"); }}
        sx={{ mb: 2, textTransform: "none" }}
      >
        &larr; Back to Inventory
      </Button>

      {isLoading && (
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          <Skeleton variant="rectangular" height={128} sx={{ borderRadius: 1 }} />
          <Skeleton variant="rectangular" height={192} sx={{ borderRadius: 1 }} />
        </Box>
      )}

      {error && !dismissed && (
        <ErrorBanner
          message={`Failed to load item detail: ${String(error)}`}
          onRetry={() => { setDismissed(false); void refetch(); }}
          onDismiss={() => setDismissed(true)}
        />
      )}

      {!isLoading && !error && !detail && (
        <EmptyState title="Item not found" description="This item may have been sold or removed." />
      )}

      {!isLoading && !error && detail && (
        <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
          <Card>
            <CardContent>
              <Typography variant="h6">{detail.item.itemName}</Typography>
              <Grid container spacing={2} mt={1}>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">Asset ID: {detail.item.assetId}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">Exterior: {detail.item.exterior || "--"}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Chip
                    label={statusLabel[detail.item.status] ?? detail.item.status}
                    size="small"
                    color={detail.item.status === "listed" ? "success" : "default"}
                  />
                </Grid>
                {detail.item.status === "listed" && detail.item.listedPrice && (
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">Listed at: {formatCNY(detail.item.listedPrice)}</Typography>
                  </Grid>
                )}
              </Grid>
            </CardContent>
          </Card>

          {detail.rentalHistory && detail.rentalHistory.length > 0 ? (
            <Box>
              <Typography variant="h6" gutterBottom>Rental History</Typography>
              <Grid container spacing={2} mb={2}>
                <Grid item xs={4}>
                  <Card variant="outlined">
                    <CardContent sx={{ textAlign: "center", py: 2 }}>
                      <Typography variant="body2" color="text.secondary">Total Days</Typography>
                      <Typography variant="h6">{detail.rentalSummary?.totalDays ?? 0}</Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={4}>
                  <Card variant="outlined">
                    <CardContent sx={{ textAlign: "center", py: 2 }}>
                      <Typography variant="body2" color="text.secondary">Total Income</Typography>
                      <Typography variant="h6">{formatCNY(detail.rentalSummary?.totalIncome ?? 0)}</Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={4}>
                  <Card variant="outlined">
                    <CardContent sx={{ textAlign: "center", py: 2 }}>
                      <Typography variant="body2" color="text.secondary">Rent Count</Typography>
                      <Typography variant="h6">{detail.rentalSummary?.rentCount ?? 0}</Typography>
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>

              <TableContainer component={Paper}>
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>Start</TableCell>
                      <TableCell>End</TableCell>
                      <TableCell align="right">Days</TableCell>
                      <TableCell align="right">Income</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {detail.rentalHistory.map((r, i) => (
                      <TableRow key={r.ID || i}>
                        <TableCell>{new Date(r.startAt * 1000).toLocaleDateString()}</TableCell>
                        <TableCell>{new Date(r.endAt * 1000).toLocaleDateString()}</TableCell>
                        <TableCell align="right">{r.durationDays}</TableCell>
                        <TableCell align="right">{formatCNY(r.income)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Box>
          ) : (
            <Card variant="outlined">
              <CardContent>
                <Typography color="text.secondary" textAlign="center">No rental history</Typography>
              </CardContent>
            </Card>
          )}
        </Box>
      )}
    </Box>
  );
}
