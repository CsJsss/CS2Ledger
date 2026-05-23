import { useQuery } from '@tanstack/react-query';
import { GetDashboardSummary } from '../lib/wails';

export function useDashboard() {
  return useQuery({
    queryKey: ['dashboard'],
    queryFn: GetDashboardSummary,
    staleTime: 5 * 60 * 1000,
  });
}
