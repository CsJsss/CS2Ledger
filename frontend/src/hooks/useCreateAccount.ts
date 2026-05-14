import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CreateAccount } from "../lib/wails";

export function useCreateAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ name, platform, cookie, withdrawalFeeRate }: { name: string; platform: string; cookie: string; withdrawalFeeRate: number }) =>
      CreateAccount(name, platform, cookie, withdrawalFeeRate),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}
