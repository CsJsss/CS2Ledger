import { useState } from "react";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import Skeleton from "@mui/material/Skeleton";
import Box from "@mui/material/Box";
import ErrorBanner from "../components/ErrorBanner";
import EmptyState from "../components/EmptyState";
import PnlSummaryCards from "../components/PnlSummaryCards";
import { useCompletedTrades } from "../hooks/useCompletedTrades";
import { useCompletedTradesSummary } from "../hooks/useCompletedTradesSummary";
import { useUIStore } from "../store/uiStore";
import { formatCNY, plHexColor } from "../lib/format";

export default function CompletedTradesPage() {
  const [dismissed, setDismissed] = useState(false);
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const {
    data: trades = [],
    isLoading: tradesLoading,
    error: tradesError,
    refetch: refetchTrades,
  } = useCompletedTrades(selectedAccountId);

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
  } = useCompletedTradesSummary(selectedAccountId);

  const isLoading = tradesLoading || summaryLoading;
  const error = tradesError || summaryError;

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Completed Trades</Typography>

      {!selectedAccountId && (
        <Box mt={3}>
          <EmptyState title="No account selected" description="Select an account to view completed trades." />
        </Box>
      )}

      {selectedAccountId && isLoading && (
        <Box mt={3}>
          <Box sx={{ display: "flex", gap: 2, mb: 3 }}>
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} variant="rectangular" height={96} sx={{ flex: 1, borderRadius: 1 }} />
            ))}
          </Box>
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} variant="rectangular" height={48} sx={{ mb: 1, borderRadius: 1 }} />
          ))}
        </Box>
      )}

      {selectedAccountId && error && !dismissed && (
        <Box mt={3}>
          <ErrorBanner
            message={`Failed to load trades: ${String(error)}`}
            onRetry={() => { setDismissed(false); void refetchTrades(); }}
            onDismiss={() => setDismissed(true)}
          />
        </Box>
      )}

      {selectedAccountId && !isLoading && !error && summary && (
        <Box mt={3}>
          <PnlSummaryCards
            totalTrades={summary.totalTrades}
            totalGrossPl={summary.totalGrossPl}
            totalFee={summary.totalFee}
            totalNetPl={summary.totalNetPl}
          />

          {trades.length === 0 && (
            <Box mt={3}>
              <EmptyState title="No completed trades" description="Sync to populate trade data and calculate P/L." />
            </Box>
          )}

          {trades.length > 0 && (
            <TableContainer component={Paper} sx={{ mt: 3 }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Item Name</TableCell>
                    <TableCell align="right">Buy Price</TableCell>
                    <TableCell align="right">Sell Price</TableCell>
                    <TableCell align="right">Qty</TableCell>
                    <TableCell align="right">Gross P/L</TableCell>
                    <TableCell align="right">Fees</TableCell>
                    <TableCell align="right">Net P/L</TableCell>
                    <TableCell align="right">Sell Date</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {trades.map((t) => (
                    <TableRow key={t.sellTradeId} hover>
                      <TableCell>
                        <Typography variant="body2" fontWeight={500}>{t.itemName}</Typography>
                      </TableCell>
                      <TableCell align="right">{formatCNY(t.buyUnitPrice)}</TableCell>
                      <TableCell align="right">{formatCNY(t.sellUnitPrice)}</TableCell>
                      <TableCell align="right">{t.quantity}</TableCell>
                      <TableCell align="right">
                        <Typography variant="body2" color={plHexColor(t.grossPl)}>
                          {formatCNY(t.grossPl)}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">{formatCNY(t.totalFee)}</TableCell>
                      <TableCell align="right">
                        <Typography variant="body2" fontWeight={600} color={plHexColor(t.netPl)}>
                          {formatCNY(t.netPl)}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2" color="text.secondary">
                          {new Date(t.sellAt * 1000).toLocaleDateString()}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Box>
      )}
    </Box>
  );
}
