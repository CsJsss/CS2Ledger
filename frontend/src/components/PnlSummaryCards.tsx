import Card from '@mui/material/Card';
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import ReceiptIcon from '@mui/icons-material/Receipt';
import PaymentsIcon from '@mui/icons-material/Payments';
import SavingsIcon from '@mui/icons-material/Savings';
import { formatCNY, plColor } from '../lib/format';

const iconBoxSx = {
  width: 32,
  height: 32,
  borderRadius: 1,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
};

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
    <Grid container spacing={1.5}>
      <Grid item xs={3}>
        <Card sx={{ borderRadius: '10px', p: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography
              variant="caption"
              color="text.disabled"
              sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}
            >
              总交易数
            </Typography>
            <Box sx={{ ...iconBoxSx, bgcolor: 'action.hover' }}>
              <ReceiptIcon fontSize="small" color="action" />
            </Box>
          </Box>
          <Typography variant="h5" mt={1} fontWeight={700}>
            {totalTrades}
          </Typography>
        </Card>
      </Grid>
      <Grid item xs={3}>
        <Card sx={{ borderRadius: '10px', p: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography
              variant="caption"
              color="text.disabled"
              sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}
            >
              毛利
            </Typography>
            <Box sx={{ ...iconBoxSx, bgcolor: 'rgba(245,158,11,0.1)' }}>
              <TrendingUpIcon fontSize="small" sx={{ color: '#f59e0b' }} />
            </Box>
          </Box>
          <Typography variant="h5" mt={1} fontWeight={700} color={plColor(totalGrossPl)}>
            {formatCNY(totalGrossPl)}
          </Typography>
        </Card>
      </Grid>
      <Grid item xs={3}>
        <Card sx={{ borderRadius: '10px', p: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography
              variant="caption"
              color="text.disabled"
              sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}
            >
              手续费
            </Typography>
            <Box sx={{ ...iconBoxSx, bgcolor: 'action.hover' }}>
              <PaymentsIcon fontSize="small" color="action" />
            </Box>
          </Box>
          <Typography variant="h5" mt={1} fontWeight={700}>
            {formatCNY(totalFee)}
          </Typography>
        </Card>
      </Grid>
      <Grid item xs={3}>
        <Card
          sx={{
            borderRadius: '10px',
            p: 2,
            borderLeft: '3px solid',
            borderLeftColor: totalNetPl >= 0 ? '#22c55e' : '#ef4444',
          }}
        >
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography
              variant="caption"
              color="text.disabled"
              sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}
            >
              净利润
            </Typography>
            <Box
              sx={{
                ...iconBoxSx,
                bgcolor: totalNetPl >= 0 ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)',
              }}
            >
              <SavingsIcon
                fontSize="small"
                sx={{ color: totalNetPl >= 0 ? '#22c55e' : '#ef4444' }}
              />
            </Box>
          </Box>
          <Typography variant="h5" mt={1} fontWeight={700} color={plColor(totalNetPl)}>
            {formatCNY(totalNetPl)}
          </Typography>
        </Card>
      </Grid>
    </Grid>
  );
}
