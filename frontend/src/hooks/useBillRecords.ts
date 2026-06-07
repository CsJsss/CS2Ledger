import { useQuery } from '@tanstack/react-query';
import { GetBillRecords, GetBillChartData } from '../lib/wails';

interface UseBillRecordsOpts {
  page?: number;
  pageSize?: number;
  typeId?: number;
  platform?: string;
  startTime?: number;
  endTime?: number;
}

export function useBillRecords(accountId: number | null, opts?: UseBillRecordsOpts) {
  const page = opts?.page ?? 1;
  const pageSize = opts?.pageSize ?? 20;
  const typeId = opts?.typeId ?? 0;
  const platform = opts?.platform ?? '';
  const startTime = opts?.startTime ?? 0;
  const endTime = opts?.endTime ?? 0;

  return useQuery({
    queryKey: ['billRecords', accountId ?? 0, page, pageSize, typeId, platform, startTime, endTime],
    queryFn: () =>
      GetBillRecords(accountId ?? 0, page, pageSize, typeId, platform, startTime, endTime),
    staleTime: 2 * 60 * 1000,
    placeholderData: (prev) => prev,
  });
}

/** Pre-aggregated daily bill data for chart rendering. */
export function useBillChartData(accountId: number | null, startTime?: number, endTime?: number) {
  return useQuery({
    queryKey: ['billChartData', accountId ?? 0, startTime ?? 0, endTime ?? 0],
    queryFn: () => GetBillChartData(accountId ?? 0, startTime ?? 0, endTime ?? 0),
    staleTime: 2 * 60 * 1000,
  });
}
