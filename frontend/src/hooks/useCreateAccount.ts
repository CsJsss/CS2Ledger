import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CreateAccount } from "../lib/wails";

export function useCreateAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ name, platform, cookie }: { name: string; platform: string; cookie: string }) =>
      CreateAccount(name, platform, cookie),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}
