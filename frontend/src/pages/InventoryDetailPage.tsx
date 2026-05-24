import { useState } from 'react';
import { useNavigate, useParams } from 'react-router';
import { type ColumnDef } from '@tanstack/react-table';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Skeleton from '@mui/material/Skeleton';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import ErrorBanner from '../components/ErrorBanner';
import EmptyState from '../components/EmptyState';
import SortableTable from '../components/SortableTable';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import { useItemDetail } from '../hooks/useItemDetail';
import { formatCNY } from '../lib/format';
import IconButton from '@mui/material/IconButton';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import Tooltip from '@mui/material/Tooltip';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import type { model } from '../lib/wails';

const statusLabel: Record<string, string> = {
  in_inventory: 'In Storage',
  listed: 'Listed',
  rented: 'Rented',
};

const rentalColumns: ColumnDef<model.RentalRecord>[] = [
  {
    accessorKey: 'startAt',
    header: 'Start',
    cell: (info) => new Date((info.getValue() as number) * 1000).toLocaleDateString(),
  },
  {
    accessorKey: 'endAt',
    header: 'End',
    cell: (info) => new Date((info.getValue() as number) * 1000).toLocaleDateString(),
  },
  {
    accessorKey: 'durationDays',
    header: 'Days',
    meta: { align: 'right' },
  },
  {
    accessorKey: 'income',
    header: 'Income',
    meta: { align: 'right' },
    cell: (info) => formatCNY(info.getValue() as number),
  },
];

export default function InventoryDetailPage() {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useState(false);
  const { accountId, assetId } = useParams<{ accountId: string; assetId: string }>();
  const accountIdNum = accountId ? Number(accountId) : null;
  const { data: detail, isLoading, error, refetch } = useItemDetail(accountIdNum, assetId ?? null);

  return (
    <Box>
      <Button
        onClick={() => {
          void navigate('/inventory');
        }}
        sx={{ mb: 2, textTransform: 'none' }}
      >
        &larr; Back to Inventory
      </Button>

      {isLoading && (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          <Skeleton variant="rectangular" height={128} sx={{ borderRadius: 1 }} />
          <Skeleton variant="rectangular" height={192} sx={{ borderRadius: 1 }} />
        </Box>
      )}

      {error && !dismissed && (
        <ErrorBanner
          message={`Failed to load item detail: ${String(error)}`}
          onRetry={() => {
            setDismissed(false);
            void refetch();
          }}
          onDismiss={() => setDismissed(true)}
        />
      )}

      {!isLoading && !error && !detail && (
        <EmptyState
          icon={<ErrorOutlineIcon sx={{ fontSize: 48 }} />}
          title="Item not found"
          description="This item may have been sold or removed."
        />
      )}

      {!isLoading && !error && detail && (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography variant="h6">{detail.item.itemName}</Typography>
                {detail.item.csqaqGoodsId ? (
                  <Tooltip title="csqaq">
                    <IconButton
                      size="small"
                      onClick={() =>
                        BrowserOpenURL(`https://www.csqaq.com/goods/${detail.item.csqaqGoodsId}`)
                      }
                    >
                      <OpenInNewIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                ) : null}
              </Box>
              <Grid container spacing={2} mt={1}>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">
                    Asset ID: {detail.item.assetId}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">
                    Exterior: {detail.item.exterior || '--'}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Chip
                    label={statusLabel[detail.item.status] ?? detail.item.status}
                    size="small"
                    color={detail.item.status === 'listed' ? 'success' : 'default'}
                  />
                </Grid>
                {detail.item.status === 'listed' && detail.item.listedPrice && (
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Listed at: {formatCNY(detail.item.listedPrice)}
                    </Typography>
                  </Grid>
                )}
              </Grid>
            </CardContent>
          </Card>

          {detail.rentalHistory && detail.rentalHistory.length > 0 ? (
            <Box>
              <Typography variant="h6" gutterBottom>
                Rental History
              </Typography>
              <Grid container spacing={2} mb={2}>
                <Grid item xs={4}>
                  <Card>
                    <CardContent sx={{ textAlign: 'center', py: 2 }}>
                      <Typography variant="body2" color="text.secondary">
                        Total Days
                      </Typography>
                      <Typography variant="h6">{detail.rentalSummary?.totalDays ?? 0}</Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={4}>
                  <Card>
                    <CardContent sx={{ textAlign: 'center', py: 2 }}>
                      <Typography variant="body2" color="text.secondary">
                        Total Income
                      </Typography>
                      <Typography variant="h6">
                        {formatCNY(detail.rentalSummary?.totalIncome ?? 0)}
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={4}>
                  <Card>
                    <CardContent sx={{ textAlign: 'center', py: 2 }}>
                      <Typography variant="body2" color="text.secondary">
                        Rent Count
                      </Typography>
                      <Typography variant="h6">{detail.rentalSummary?.rentCount ?? 0}</Typography>
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>

              <SortableTable
                columns={rentalColumns}
                data={detail.rentalHistory}
                getRowId={(r) => String(r.ID)}
              />
            </Box>
          ) : (
            <Card>
              <CardContent>
                <Typography color="text.secondary" textAlign="center">
                  No rental history
                </Typography>
              </CardContent>
            </Card>
          )}
        </Box>
      )}
    </Box>
  );
}
