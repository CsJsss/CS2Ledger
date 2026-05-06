import { useMutation, useQueryClient } from "@tanstack/react-query";
import { DeleteAccount } from "../lib/wails";

export function useDeleteAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => DeleteAccount(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}
