export function formatCNY(cents: number | null | undefined): string {
  if (cents == null) return "¥ --";
  const yuan = cents / 100;
  return `¥${yuan.toLocaleString("zh-CN", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export function plColor(value: number): string {
  if (value > 0) return "success";
  if (value < 0) return "error";
  return "text.secondary";
}

export function fmt(n: number | null | undefined): string {
  if (n == null) return "0.00";
  return Number(n).toFixed(2);
}

export function plHexColor(value: number): string {
  if (value > 0) return "#16a34a";
  if (value < 0) return "#dc2626";
  return "#6b7280";
}
