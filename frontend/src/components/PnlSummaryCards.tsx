import Grid from "@mui/material/Grid";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import { formatCNY, plColor } from "../lib/format";

interface PnlSummaryCardsProps {
  totalTrades: number;
  totalGrossPl: number;
  totalFee: number;
  totalNetPl: number;
}

export default function PnlSummaryCards({
  totalTrades,
  totalGrossPl,
  totalFee,
  totalNetPl,
}: PnlSummaryCardsProps) {
  return (
    <Grid container spacing={2}>
      <Grid item xs={3}>
        <Card>
          <CardContent>
            <Typography variant="body2" color="text.secondary">总交易数</Typography>
            <Typography variant="h5" mt={1}>{totalTrades}</Typography>
          </CardContent>
        </Card>
      </Grid>
      <Grid item xs={3}>
        <Card>
          <CardContent>
            <Typography variant="body2" color="text.secondary">毛利</Typography>
            <Typography variant="h5" mt={1} color={plColor(totalGrossPl)}>
              {formatCNY(totalGrossPl)}
            </Typography>
          </CardContent>
        </Card>
      </Grid>
      <Grid item xs={3}>
        <Card>
          <CardContent>
            <Typography variant="body2" color="text.secondary">手续费</Typography>
            <Typography variant="h5" mt={1}>{formatCNY(totalFee)}</Typography>
          </CardContent>
        </Card>
      </Grid>
      <Grid item xs={3}>
        <Card>
          <CardContent>
            <Typography variant="body2" color="text.secondary">净利润</Typography>
            <Typography variant="h5" mt={1} color={plColor(totalNetPl)}>
              {formatCNY(totalNetPl)}
            </Typography>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  );
}
