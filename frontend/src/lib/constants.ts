export const PLATFORM_OPTIONS = [
  { value: "buff", label: "BUFF" },
  { value: "youpin", label: "悠悠有品" },
  { value: "c5", label: "C5" },
  { value: "igxe", label: "IGXE" },
] as const;

export const platformLabel: Record<string, string> = {
  buff: "BUFF",
  youpin: "悠悠有品",
  c5: "C5",
  igxe: "IGXE",
};

export const inventoryStatusLabel: Record<string, string> = {
  in_inventory: "In Storage",
  listed: "Listed",
  rented: "Rented",
};

export const inventoryStatusColor: Record<string, "default" | "success" | "warning"> = {
  in_inventory: "default",
  listed: "success",
  rented: "warning",
};
