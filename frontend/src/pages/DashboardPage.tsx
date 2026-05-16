import { useState } from "react";
import { useNavigate } from "react-router";
import Grid from "@mui/material/Grid";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import Skeleton from "@mui/material/Skeleton";
import Box from "@mui/material/Box";
import ErrorBanner from "../components/ErrorBanner";
import EmptyState from "../components/EmptyState";
import { useDashboard } from "../hooks/useDashboard";
import { formatCNY } from "../lib/format";

export default function DashboardPage() {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useState(false);
  const { data, isLoading, error, refetch } = useDashboard();

  return (
    <Box>
      <Typography variant="h4" gutterBottom>Dashboard</Typography>

      {isLoading && (
        <Grid container spacing={2} mt={1}>
          {[1, 2, 3, 4].map((i) => (
            <Grid item xs={3} key={i}>
              <Card>
                <CardContent>
                  <Skeleton width="60%" />
                  <Skeleton height={40} width="40%" />
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {error && !dismissed && (
        <Box mt={3}>
          <ErrorBanner
            message={`Failed to load dashboard: ${String(error)}`}
            onRetry={() => { setDismissed(false); void refetch(); }}
            onDismiss={() => setDismissed(true)}
          />
        </Box>
      )}

      {!isLoading && !error && data && data.inventoryCount === 0 && data.completedTrades === 0 && (
        <Box mt={3}>
          <EmptyState
            title="No data yet"
            description="Add an account and sync to see your dashboard."
            action={{ label: "Go to Accounts", onClick: () => { void navigate("/accounts"); } }}
          />
        </Box>
      )}

      {!isLoading && !error && data && !(data.inventoryCount === 0 && data.completedTrades === 0) && (
        <Grid container spacing={2} mt={1}>
          <Grid item xs={3}>
            <Card>
              <CardContent>
                <Typography variant="body2" color="text.secondary">Net Worth</Typography>
                <Typography variant="h5" mt={1}>{formatCNY(data.totalNetWorth)}</Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={3}>
            <Card>
              <CardContent>
                <Typography variant="body2" color="text.secondary">Inventory Items</Typography>
                <Typography variant="h5" mt={1}>{data.inventoryCount}</Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={3}>
            <Card>
              <CardContent>
                <Typography variant="body2" color="text.secondary">Completed Trades</Typography>
                <Typography variant="h5" mt={1}>{data.completedTrades}</Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={3}>
            <Card>
              <CardContent>
                <Typography variant="body2" color="text.secondary">Rental Income</Typography>
                <Typography variant="h5" mt={1}>{formatCNY(data.totalRentalIncome)}</Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}
    </Box>
  );
}
