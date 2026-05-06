import { useState } from "react";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import Skeleton from "@mui/material/Skeleton";
import Box from "@mui/material/Box";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import FormControl from "@mui/material/FormControl";
import ErrorBanner from "../components/ErrorBanner";
import EmptyState from "../components/EmptyState";
import PnlSummaryCards from "../components/PnlSummaryCards";
import { usePnlSummary } from "../hooks/usePnlSummary";
import { useMonthlyBreakdown } from "../hooks/useMonthlyBreakdown";
import { useUIStore } from "../store/uiStore";
import { formatCNY, plHexColor } from "../lib/format";

export default function PnLPage() {
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const currentYear = new Date().getFullYear();
  const [dismissed, setDismissed] = useState(false);
  const [year, setYear] = useState(currentYear);

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = usePnlSummary(selectedAccountId);

  const {
    data: monthly = [],
    isLoading: monthlyLoading,
    error: monthlyError,
  } = useMonthlyBreakdown(selectedAccountId, year);

  const isLoading = summaryLoading || monthlyLoading;
  const error = summaryError || monthlyError;
  const maxAbsPl = monthly.reduce((max, m) => Math.max(max, Math.abs(m.netPl)), 1);

  const years: number[] = [];
  for (let y = currentYear; y >= currentYear - 3; y--) years.push(y);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Profit &amp; Loss</Typography>

      {!selectedAccountId && (
        <Box mt={3}>
          <EmptyState title="No account selected" description="Select an account to view P&amp;L data." />
        </Box>
      )}

      {selectedAccountId && isLoading && (
        <Box mt={3}>
          <Box sx={{ display: "flex", gap: 2, mb: 3 }}>
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} variant="rectangular" height={96} sx={{ flex: 1, borderRadius: 1 }} />
            ))}
          </Box>
          <Skeleton variant="rectangular" height={256} sx={{ borderRadius: 1 }} />
        </Box>
      )}

      {selectedAccountId && error && !dismissed && (
        <Box mt={3}>
          <ErrorBanner
            message={`Failed to load P&L data: ${String(error)}`}
            onRetry={() => { setDismissed(false); void refetchSummary(); }}
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

          <Card sx={{ mt: 3 }}>
            <CardContent>
              <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 2 }}>
                <Typography variant="h6">Monthly Breakdown</Typography>
                <FormControl size="small" sx={{ minWidth: 100 }}>
                  <Select value={year} onChange={(e) => setYear(Number(e.target.value))}>
                    {years.map((y) => (
                      <MenuItem key={y} value={y}>{y}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Box>

              {monthly.length === 0 && (
                <Typography color="text.secondary" textAlign="center" py={4}>
                  No data for {year}
                </Typography>
              )}

              {monthly.length > 0 && (
                <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
                  {monthly.map((m) => (
                    <Box key={m.month} sx={{ display: "flex", alignItems: "center", gap: 1.5 }}>
                      <Typography variant="body2" color="text.secondary" sx={{ width: 64, textAlign: "right" }}>
                        {m.month}
                      </Typography>
                      <Box sx={{ flex: 1, height: 24, bgcolor: "grey.100", borderRadius: 0.5, overflow: "hidden" }}>
                        <Box
                          sx={{
                            height: "100%",
                            borderRadius: 0.5,
                            transition: "width 0.3s",
                            bgcolor: m.netPl >= 0 ? "success.main" : "error.main",
                            width: `${Math.min((Math.abs(m.netPl) / maxAbsPl) * 100, 100)}%`,
                          }}
                        />
                      </Box>
                      <Typography
                        variant="body2"
                        fontWeight={500}
                        color={plHexColor(m.netPl)}
                        sx={{ width: 96 }}
                      >
                        {formatCNY(m.netPl)}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              )}
            </CardContent>
          </Card>
        </Box>
      )}
    </Box>
  );
}
