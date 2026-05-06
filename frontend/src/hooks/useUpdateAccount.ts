import { useMutation, useQueryClient } from "@tanstack/react-query";
import { UpdateAccountInfo } from "../lib/wails";

export function useUpdateAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, name, cookie }: { id: number; name: string; cookie: string }) =>
      UpdateAccountInfo(id, name, cookie),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}
