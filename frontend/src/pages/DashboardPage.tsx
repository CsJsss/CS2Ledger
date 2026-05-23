import { useState } from 'react';
import { useNavigate } from 'react-router';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Skeleton from '@mui/material/Skeleton';
import Box from '@mui/material/Box';
import ErrorBanner from '../components/ErrorBanner';
import EmptyState from '../components/EmptyState';
import { useDashboard } from '../hooks/useDashboard';
import { formatCNY, plColor } from '../lib/format';
import { priceSourceLabel } from '../lib/constants';

export default function DashboardPage() {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useState(false);
  const { data, isLoading, error, refetch } = useDashboard();

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        仪表盘
      </Typography>

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
            message={`加载仪表盘失败: ${String(error)}`}
            onRetry={() => {
              setDismissed(false);
              void refetch();
            }}
            onDismiss={() => setDismissed(true)}
          />
        </Box>
      )}

      {!isLoading && !error && data && data.inventoryCount === 0 && data.completedTrades === 0 && (
        <Box mt={3}>
          <EmptyState
            title="暂无数据"
            description="添加账户并同步以查看仪表盘。"
            action={{
              label: '前往账户管理',
              onClick: () => {
                void navigate('/accounts');
              },
            }}
          />
        </Box>
      )}

      {!isLoading &&
        !error &&
        data &&
        !(
          data.inventoryCount === 0 &&
          data.completedTrades === 0 &&
          data.totalAvailableBalance === 0 &&
          data.totalFrozenBalance === 0 &&
          data.totalInstantBalance === 0 &&
          data.totalPurchaseBalance === 0
        ) && (
          <Grid container spacing={2} mt={1}>
            <Grid item xs={3}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    钱包余额
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {formatCNY(data.totalAvailableBalance)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={3}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    冻结余额
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {formatCNY(data.totalFrozenBalance)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={3}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    秒到账余额
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {formatCNY(data.totalInstantBalance)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={3}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    求购余额
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {formatCNY(data.totalPurchaseBalance)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        )}
      {!isLoading &&
        !error &&
        data &&
        !(
          data.inventoryCount === 0 &&
          data.completedTrades === 0 &&
          data.totalAvailableBalance === 0 &&
          data.totalFrozenBalance === 0 &&
          data.totalInstantBalance === 0 &&
          data.totalPurchaseBalance === 0
        ) && (
          <Grid container spacing={2} mt={2}>
            <Grid item xs={3}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    已实现盈亏
                  </Typography>
                  <Typography variant="h5" mt={1} color={plColor(data.realizedPl)}>
                    {formatCNY(data.realizedPl)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={3}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    持仓物品
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {data.inventoryCount}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={3}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    已完成交易
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {data.completedTrades}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={3}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    租赁收入
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {formatCNY(data.totalRentalIncome)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        )}
      {!isLoading &&
        !error &&
        data &&
        !(
          data.inventoryCount === 0 &&
          data.completedTrades === 0 &&
          data.totalAvailableBalance === 0 &&
          data.totalFrozenBalance === 0 &&
          data.totalInstantBalance === 0 &&
          data.totalPurchaseBalance === 0
        ) && (
          <Grid container spacing={2} mt={2}>
            <Grid item xs={4}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    持仓成本
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {formatCNY(data.inventoryCost)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={4}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    持仓市值（{priceSourceLabel[data.priceSource] ?? data.priceSource}）
                  </Typography>
                  <Typography variant="h5" mt={1}>
                    {formatCNY(data.inventoryMarketValue)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={4}>
              <Card>
                <CardContent>
                  <Typography variant="body2" color="text.secondary">
                    未实现盈亏
                  </Typography>
                  <Typography
                    variant="h5"
                    mt={1}
                    color={plColor(data.inventoryMarketValue - data.inventoryCost)}
                  >
                    {formatCNY(data.inventoryMarketValue - data.inventoryCost)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        )}
    </Box>
  );
}
