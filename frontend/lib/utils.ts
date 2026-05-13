// Parse a YYYY-MM-DD date string in local time (avoids UTC midnight off-by-one)
export function fmtDate(dateStr?: string | null): string {
  if (!dateStr) return "—";
  const [y, m, d] = dateStr.slice(0, 10).split("-");
  return `${m}/${d}/${y}`;
}
