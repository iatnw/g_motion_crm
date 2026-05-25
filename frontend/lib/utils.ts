import type { Referral } from "@/lib/types";

// Parse a YYYY-MM-DD date string in local time (avoids UTC midnight off-by-one)
export function fmtDate(dateStr?: string | null): string {
  if (!dateStr) return "—";
  const [y, m, d] = dateStr.slice(0, 10).split("-");
  return `${m}/${d}/${y}`;
}

export function groupReferralsByDeal(referrals: Referral[]): Record<string, Referral[]> {
  const map: Record<string, Referral[]> = {};
  for (const ref of referrals) {
    if (!ref.deal_id) continue;
    if (!map[ref.deal_id]) map[ref.deal_id] = [];
    map[ref.deal_id].push(ref);
  }
  return map;
}
