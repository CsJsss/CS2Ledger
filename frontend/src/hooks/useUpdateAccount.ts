import { useMutation, useQueryClient } from "@tanstack/react-query";
import { UpdateAccountInfo } from "../lib/wails";

export function useUpdateAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, name, cookie, withdrawalFeeRate }: { id: number; name: string; cookie: string; withdrawalFeeRate: number }) =>
      UpdateAccountInfo(id, name, cookie, withdrawalFeeRate),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}
