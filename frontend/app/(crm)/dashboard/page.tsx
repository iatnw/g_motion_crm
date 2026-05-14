"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { dashboard, deals, clients, brokers, dealers, referrals as referralsApi } from "@/lib/api";
import type { DealStatus, DashboardSummary, Deal, Client, PendingPaymentBucket, Referral } from "@/lib/types";
import DealStatusSelect from "@/components/DealStatusSelect";
import DealModal from "@/components/DealModal";

type RangeMode = "all" | "month" | "custom";

export default function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [mode, setMode] = useState<RangeMode>("all");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const [bucketModal, setBucketModal] = useState<{ bucket: PendingPaymentBucket; deals: Deal[]; clientMap: Record<string, Client>; brokerMap: Record<string, string>; dealerMap: Record<string, string>; referralMap: Record<string, Referral[]>; allClients: Client[]; allBrokers: import("@/lib/types").Broker[]; allDealers: import("@/lib/types").Dealer[] } | null>(null);
  const [bucketLoading, setBucketLoading] = useState(false);
  const [editModal, setEditModal] = useState<{ deal: Deal } | null>(null);

  const params = useMemo(() => {
    if (mode === "all") return {};
    if (mode === "month") return currentMonthRange();
    return { from: from || undefined, to: to || undefined };
  }, [mode, from, to]);

  useEffect(() => {
    let cancelled = false;
    dashboard
      .summary({ ...params, limit: 8 })
      .then((data) => { if (!cancelled) setSummary(data); })
      .catch((e: unknown) => { if (!cancelled) setError(e instanceof Error ? e.message : "Failed to load dashboard"); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [params]);

  async function handleDealStatusChange(id: string, status: DealStatus) {
    await deals.updateStatus(id, status);
    const next = await dashboard.summary({ ...params, limit: 8 });
    setSummary(next);
  }

  async function openBucket(bucket: PendingPaymentBucket) {
    setBucketLoading(true);
    try {
      const [allDeals, allClients, allBrokers, allDealers, allReferrals] = await Promise.all([
        deals.list(), clients.list(), brokers.list(), dealers.list(), referralsApi.list(),
      ]);
      const filtered = (allDeals ?? []).filter((d) => d.deal_status === "pending" && d.pending_payment_bucket === bucket);
      const clientMap = Object.fromEntries((allClients ?? []).map((c) => [c.id, c]));
      const brokerMap = Object.fromEntries((allBrokers ?? []).map((b) => [b.id, `${b.first_name} ${b.last_name}`]));
      const dealerMap = Object.fromEntries((allDealers ?? []).map((d) => [d.id, d.name]));
      const referralMap: Record<string, Referral[]> = {};
      for (const r of allReferrals ?? []) {
        if (r.deal_id) { if (!referralMap[r.deal_id]) referralMap[r.deal_id] = []; referralMap[r.deal_id].push(r); }
      }
      setBucketModal({ bucket, deals: filtered, clientMap, brokerMap, dealerMap, referralMap, allClients: allClients ?? [], allBrokers: allBrokers ?? [], allDealers: allDealers ?? [] });
    } finally {
      setBucketLoading(false);
    }
  }

  const totals = summary?.totals;
  const aging = summary?.aging;

  const paidRev = totals?.paid_revenue ?? 0;
  const pendingRev = totals?.pending_revenue ?? 0;
  const grossRev = totals?.gross_revenue ?? 0;
  const collectionPct = grossRev > 0 ? Math.round((paidRev / grossRev) * 100) : 0;

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4 mb-6">
        <div>
          <h1 className="text-xl font-semibold text-white">Dashboard</h1>
          <p className="text-slate-500 text-xs mt-0.5">Company overview, pending money, broker and dealer performance</p>
        </div>
        <div className="flex items-center gap-2">
          <SegmentedButton active={mode === "all"} onClick={() => setMode("all")}>All time</SegmentedButton>
          <SegmentedButton active={mode === "month"} onClick={() => setMode("month")}>This month</SegmentedButton>
          <SegmentedButton active={mode === "custom"} onClick={() => setMode("custom")}>Custom</SegmentedButton>
        </div>
      </div>

      {mode === "custom" && (
        <div className="flex items-center gap-3 mb-5">
          <DateInput label="From" value={from} onChange={setFrom} />
          <DateInput label="To" value={to} onChange={setTo} />
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading...</div>
      ) : error ? (
        <div className="text-red-400 text-sm">{error}</div>
      ) : summary && totals && aging ? (
        <div className="space-y-6">

          {/* Hero row */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <HeroCard
              label="Total deals"
              value={totals.total_deals.toLocaleString()}
              sub={`Avg ${money(totals.avg_revenue_per_deal)} / deal`}
              accent="blue"
            />
            <HeroCard
              label="Gross revenue"
              value={money(totals.gross_revenue)}
              sub={`Avg profit ${money(totals.avg_profit_per_deal)} / deal`}
              accent="default"
            />
            <HeroCard
              label="Company net"
              value={money(totals.company_profit_before_tax)}
              sub={`Pending ${money(totals.pending_company_profit)}`}
              accent="green"
            />
          </div>

          {/* Collection status */}
          <div className="bg-[#111213] border border-[#1e1f21] rounded-2xl p-5">
            <div className="flex items-center justify-between mb-4">
              <p className="text-xs font-semibold text-slate-500 uppercase tracking-widest">Revenue collection</p>
              <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${
                collectionPct >= 80
                  ? "bg-green-500/15 text-green-400"
                  : collectionPct >= 50
                    ? "bg-yellow-500/15 text-yellow-400"
                    : "bg-red-500/15 text-red-400"
              }`}>{collectionPct}% collected</span>
            </div>
            <div className="w-full h-2.5 bg-[#1e1f21] rounded-full overflow-hidden mb-4">
              <div
                className={`h-full rounded-full transition-all duration-700 ${
                  collectionPct >= 80 ? "bg-green-500" : collectionPct >= 50 ? "bg-yellow-500" : "bg-red-500"
                }`}
                style={{ width: `${collectionPct}%` }}
              />
            </div>
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <p className="text-xs text-slate-600 uppercase tracking-wider mb-1">Gross</p>
                <p className="text-base font-semibold text-white">{money(grossRev)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-600 uppercase tracking-wider mb-1">Collected</p>
                <p className="text-base font-semibold text-green-400">{money(paidRev)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-600 uppercase tracking-wider mb-1">Outstanding</p>
                <p className="text-base font-semibold text-yellow-400">{money(pendingRev)}</p>
              </div>
            </div>
          </div>

          {/* Aging — most actionable, shown high */}
          <div className="space-y-3">
            <SectionLabel>Pending by date</SectionLabel>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
              <AgingCard title="0–31 days" count={aging.pending_0_31_count} amount={aging.pending_0_31_revenue} tone="green" badge="On time" onClick={() => openBucket("green")} />
              <AgingCard title="32–40 days" count={aging.pending_32_40_count} amount={aging.pending_32_40_revenue} tone="yellow" badge="Watch" onClick={() => openBucket("yellow")} />
              <AgingCard title="41+ days" count={aging.pending_41_plus_count} amount={aging.pending_41_plus_revenue} tone="red" badge="Overdue" onClick={() => openBucket("red")} />
            </div>
          </div>

          {/* Secondary stats — compact */}
          <div className="space-y-3">
            <SectionLabel>Costs</SectionLabel>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <StatChip label="Broker payout" value={money(totals.broker_payout)} />
              <StatChip label="Doc fees" value={money(totals.total_doc_fees)} />
              <StatChip label="Referral payout" value={money(totals.total_referral_payout)} />
            </div>
          </div>

          {/* Brokers + Dealers */}
          <div className="space-y-3">
            <SectionLabel>Performance</SectionLabel>
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-5">
              <Panel title="Brokers">
                <Table
                  headers={["Broker", "Deals", "Revenue", "Payout", "Profit"]}
                  rows={summary.brokers.map((b) => [
                    `${b.first_name} ${b.last_name}`,
                    b.deal_count.toLocaleString(),
                    money(b.gross_revenue),
                    money(b.broker_payout),
                    money(b.company_profit_before_tax),
                  ])}
                />
              </Panel>
              <Panel title="Dealers">
                <Table
                  headers={["Dealer", "Deals", "Revenue", "Avg days"]}
                  rows={summary.dealers.map((d) => [
                    d.name,
                    d.deal_count.toLocaleString(),
                    money(d.gross_revenue),
                    d.avg_days_to_payment == null ? "—" : Number(d.avg_days_to_payment).toFixed(1),
                  ])}
                />
              </Panel>
            </div>
          </div>

          {/* Recent deals */}
          <Panel title="Recent deals">
            <div className="divide-y divide-[#1e1f20]">
              {summary.recent_deals.length === 0 ? (
                <div className="py-8 text-center text-sm text-slate-600">No recent deals</div>
              ) : (
                summary.recent_deals.map((deal) => (
                  <div key={deal.id} className="grid grid-cols-5 gap-3 px-4 py-3 text-sm hover:bg-blue-500/5 transition-colors">
                    <span className="text-slate-200">{[deal.make, deal.model].filter(Boolean).join(" ") || "Deal"}</span>
                    <span className="text-slate-500">{deal.deal_date ?? "—"}</span>
                    <span className="text-slate-400">{money(deal.broker_fee)}</span>
                    <DealStatusSelect
                      value={deal.deal_status}
                      days={deal.pending_payment_days}
                      bucket={deal.pending_payment_bucket}
                      onChange={(status) => handleDealStatusChange(deal.id, status)}
                    />
                    <Link href="/deals" className="text-right text-blue-400 hover:text-blue-300">View</Link>
                  </div>
                ))
              )}
            </div>
          </Panel>

        </div>
      ) : null}

      {(bucketLoading || bucketModal) && (
        <BucketModal
          loading={bucketLoading}
          bucket={bucketModal?.bucket}
          deals={bucketModal?.deals ?? []}
          clientMap={bucketModal?.clientMap ?? {}}
          brokerMap={bucketModal?.brokerMap ?? {}}
          dealerMap={bucketModal?.dealerMap ?? {}}
          referralMap={bucketModal?.referralMap ?? {}}
          onClose={() => setBucketModal(null)}
          onEdit={(deal) => { setBucketModal(null); setEditModal({ deal }); }}
        />
      )}

      {editModal && (
        <DealModal
          deal={editModal.deal}
          clients={bucketModal?.allClients ?? []}
          brokers={bucketModal?.allBrokers ?? []}
          dealers={bucketModal?.allDealers ?? []}
          onSave={async (data) => {
            const updated = await deals.update(editModal.deal.id, data);
            setEditModal(null);
            const next = await dashboard.summary({ ...params, limit: 8 });
            setSummary(next);
            return updated;
          }}
          onClose={() => setEditModal(null)}
          onStatusChange={() => {}}
        />
      )}
    </div>
  );
}

// --- Components ---

function HeroCard({ label, value, sub, accent }: { label: string; value: string; sub: string; accent: "blue" | "green" | "default" }) {
  const border = accent === "green" ? "border-green-500/25" : accent === "blue" ? "border-blue-500/20" : "border-[#252628]";
  const valueColor = accent === "green" ? "text-green-400" : accent === "blue" ? "text-blue-400" : "text-white";
  const dot = accent === "green" ? "bg-green-400" : accent === "blue" ? "bg-blue-500" : "bg-slate-500";
  return (
    <div className={`bg-[#111213] border ${border} rounded-2xl p-6 flex flex-col gap-2`}>
      <div className="flex items-center gap-2">
        <span className={`w-1.5 h-1.5 rounded-full ${dot} shrink-0`} />
        <p className="text-xs text-slate-500 uppercase tracking-widest font-medium">{label}</p>
      </div>
      <p className={`text-4xl font-bold tracking-tight ${valueColor}`}>{value}</p>
      <p className="text-xs text-slate-600">{sub}</p>
    </div>
  );
}

function StatChip({ label, value, warn }: { label: string; value: string; warn?: boolean }) {
  return (
    <div className="bg-[#111213] border border-[#1e1f21] rounded-xl px-4 py-3 flex items-center justify-between gap-3">
      <p className="text-xs text-slate-500 truncate">{label}</p>
      <p className={`text-sm font-semibold shrink-0 ${warn ? "text-red-400" : "text-slate-200"}`}>{value}</p>
    </div>
  );
}

function AgingCard({ title, count, amount, tone, badge, onClick }: { title: string; count: number; amount?: number; tone: "green" | "yellow" | "red"; badge: string; onClick: () => void }) {
  const bg = tone === "green" ? "bg-green-500/5" : tone === "yellow" ? "bg-yellow-500/5" : "bg-red-500/5";
  const border = tone === "green" ? "border-green-500/20" : tone === "yellow" ? "border-yellow-500/20" : "border-red-500/20";
  const countColor = tone === "green" ? "text-green-400" : tone === "yellow" ? "text-yellow-400" : "text-red-400";
  const badgeCls = tone === "green" ? "bg-green-500/15 text-green-400" : tone === "yellow" ? "bg-yellow-500/15 text-yellow-400" : "bg-red-500/15 text-red-400";
  return (
    <button onClick={onClick} className={`${bg} border ${border} rounded-2xl p-5 flex flex-col gap-4 text-left w-full hover:brightness-110 transition-all cursor-pointer`}>
      <div className="flex items-center justify-between">
        <p className="text-xs text-slate-500 uppercase tracking-widest font-medium">{title}</p>
        <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${badgeCls}`}>{badge}</span>
      </div>
      <div>
        <p className={`text-4xl font-bold tracking-tight ${countColor}`}>{count}</p>
        <p className="text-sm text-slate-400 mt-1">{money(amount)}</p>
      </div>
    </button>
  );
}

const bucketLabel: Record<string, string> = { green: "0–31 days", yellow: "32–40 days", red: "41+ days" };

function BucketModal({ loading, bucket, deals, clientMap, brokerMap, dealerMap, referralMap, onClose, onEdit }: {
  loading: boolean;
  bucket?: string;
  deals: Deal[];
  clientMap: Record<string, Client>;
  brokerMap: Record<string, string>;
  dealerMap: Record<string, string>;
  referralMap: Record<string, Referral[]>;
  onClose: () => void;
  onEdit: (deal: Deal) => void;
}) {
  const tone = bucket === "green" ? "text-green-400" : bucket === "yellow" ? "text-yellow-400" : "text-red-400";
  const borderColor = bucket === "green" ? "border-l-green-500" : bucket === "yellow" ? "border-l-amber-500" : "border-l-red-500";

  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [onClose]);

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-[#0b0c0d] rounded-2xl w-full max-w-3xl max-h-[80vh] flex flex-col border border-[#252628] shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-6 py-4 border-b border-[#252628] shrink-0">
          <div>
            <h2 className="text-base font-semibold text-white">Pending deals</h2>
            {bucket && <p className={`text-xs mt-0.5 font-medium ${tone}`}>{bucketLabel[bucket]}</p>}
          </div>
          <button onClick={onClose} className="text-slate-600 hover:text-slate-300 text-xl transition-colors">✕</button>
        </div>
        <div className="overflow-y-auto flex-1 p-4 space-y-2">
          {loading ? (
            <div className="py-16 text-center text-sm text-slate-600">Loading...</div>
          ) : deals.length === 0 ? (
            <div className="py-16 text-center text-sm text-slate-600">No deals in this range</div>
          ) : deals.map((d) => {
            const client = clientMap[d.client_id];
            const vehicle = [d.make, d.model].filter(Boolean).join(" ");
            const dealReferrals = referralMap[d.id] ?? [];
            const referralTotal = dealReferrals.reduce((s, r) => s + (r.amount ?? 0), 0);
            const base = (d.broker_fee ?? 0) + (d.collect_amount ?? 0) - (d.doc_fee ?? 0);
            const commission = d.broker_commission || (d.broker_percent != null && base > 0 ? base * d.broker_percent / 100 : 0);
            const computedNet = base > 0 ? base - commission - referralTotal : null;
            return (
              <div key={d.id} className={`bg-[#111213] border border-[#252628] border-l-4 ${borderColor} rounded-xl`}>
                {/* Top bar */}
                <div className="flex items-center justify-between px-4 py-3 border-b border-[#1e1f20]">
                  <div className="flex items-center gap-3 flex-wrap min-w-0">
                    <span className="text-white font-semibold">
                      {client ? `${client.first_name} ${client.last_name}` : "Unknown"}
                    </span>
                    {d.deal_number && (
                      <span className="text-xs font-mono bg-[#0b0c0d] border border-[#252628] text-slate-400 px-2 py-0.5 rounded-md">
                        #{d.deal_number}
                      </span>
                    )}
                    {vehicle && <span className="text-slate-300 text-sm">{vehicle}</span>}
                  </div>
                  <div className="flex items-center gap-3 shrink-0 ml-3">
                    <span className={`text-sm font-semibold ${tone}`}>
                      {d.pending_payment_days ?? "—"}d pending
                    </span>
                    <button
                      onClick={() => onEdit(d)}
                      className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                    >
                      Edit
                    </button>
                  </div>
                </div>
                {/* Stats */}
                <div className="flex divide-x divide-[#1e1f20]">
                  {d.broker_id && <BucketStat label="Broker" value={brokerMap[d.broker_id]} />}
                  {d.dealer_id && <BucketStat label="Dealer" value={dealerMap[d.dealer_id]} />}
                  {d.broker_fee != null && <BucketStat label="Revenue" value={money(d.broker_fee)} highlight="green" />}
                  {computedNet != null && <BucketStat label="Net" value={money(computedNet)} highlight="blue" />}
                  {d.broker_commission != null && <BucketStat label="Commission" value={money(d.broker_commission)} />}
                  {d.doc_fee != null && <BucketStat label="Doc Fee" value={money(d.doc_fee)} />}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

function BucketStat({ label, value, highlight }: { label: string; value?: string; highlight?: "green" | "blue" }) {
  if (!value) return null;
  const cls = highlight === "green" ? "text-green-400 font-bold" : highlight === "blue" ? "text-blue-400 font-bold" : "text-slate-300";
  return (
    <div className="flex-1 px-3 py-2 flex flex-col gap-0.5 min-w-0">
      <span className="text-xs text-slate-600 uppercase tracking-wider">{label}</span>
      <span className={`text-sm truncate ${cls}`}>{value}</span>
    </div>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-3">
      <p className="text-xs font-semibold text-slate-500 uppercase tracking-widest">{children}</p>
      <div className="flex-1 h-px bg-[#1e1f21]" />
    </div>
  );
}

function Panel({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="bg-[#111213] border border-[#1e1f21] rounded-2xl overflow-hidden">
      <div className="px-5 py-4 border-b border-[#1e1f21] flex items-center gap-2">
        <span className="w-1.5 h-1.5 rounded-full bg-blue-500 shrink-0" />
        <h2 className="text-sm font-semibold text-white">{title}</h2>
      </div>
      {children}
    </section>
  );
}

function Table({ headers, rows }: { headers: string[]; rows: string[][] }) {
  if (!rows.length) return <div className="py-8 text-center text-sm text-slate-600">No data</div>;
  return (
    <table className="w-full text-sm">
      <thead className="border-b border-[#252628]">
        <tr>
          {headers.map((h) => (
            <th key={h} className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">{h}</th>
          ))}
        </tr>
      </thead>
      <tbody className="divide-y divide-[#1e1f20]">
        {rows.map((row, i) => (
          <tr key={i}>
            {row.map((cell, j) => (
              <td key={j} className={j === 0 ? "px-4 py-3 text-slate-200" : "px-4 py-3 text-slate-400"}>{cell}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function SegmentedButton({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-colors ${
        active ? "bg-blue-600/15 text-blue-400 border-blue-500/30" : "text-slate-500 border-[#252628] hover:text-slate-300"
      }`}
    >
      {children}
    </button>
  );
}

function DateInput({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <label className="text-xs text-slate-500 uppercase tracking-wider">
      <span className="mr-2">{label}</span>
      <input
        type="date"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white focus:outline-none"
      />
    </label>
  );
}

function money(value?: number) {
  return `$${Number(value ?? 0).toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
}

function currentMonthRange() {
  const now = new Date();
  const first = new Date(now.getFullYear(), now.getMonth(), 1);
  const last = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  return { from: fmt(first), to: fmt(last) };
}

function fmt(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}
