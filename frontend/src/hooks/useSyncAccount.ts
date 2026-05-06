import { useMutation, useQueryClient } from "@tanstack/react-query";
import { SyncAccount } from "../lib/wails";

export function useSyncAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (accountId: number) => SyncAccount(accountId),
    onSuccess: (result) => {
      if (result.NewTrades > 0) {
        void queryClient.invalidateQueries({ queryKey: ["completedTrades"] });
        void queryClient.invalidateQueries({ queryKey: ["completedTradesSummary"] });
        void queryClient.invalidateQueries({ queryKey: ["pnlSummary"] });
      }
      if (result.NewPnl > 0) {
        void queryClient.invalidateQueries({ queryKey: ["inventory"] });
        void queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      }
      void queryClient.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}
