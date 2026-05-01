"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { dashboard, deals } from "@/lib/api";
import type { DealStatus, DashboardSummary } from "@/lib/types";
import DealStatusSelect from "@/components/DealStatusSelect";

type RangeMode = "all" | "month" | "custom";

export default function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [mode, setMode] = useState<RangeMode>("all");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const params = useMemo(() => {
    if (mode === "all") return {};
    if (mode === "month") return currentMonthRange();
    return { from: from || undefined, to: to || undefined };
  }, [mode, from, to]);

  useEffect(() => {
    let cancelled = false;
    dashboard
      .summary({ ...params, limit: 8 })
      .then((data) => {
        if (!cancelled) setSummary(data);
      })
      .catch((e: unknown) => {
        if (!cancelled) setError(e instanceof Error ? e.message : "Failed to load dashboard");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [params]);

  async function handleDealStatusChange(id: string, status: DealStatus) {
    await deals.updateStatus(id, status);
    const next = await dashboard.summary({ ...params, limit: 8 });
    setSummary(next);
  }

  const totals = summary?.totals;
  const aging = summary?.aging;

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-start justify-between gap-4 mb-6">
        <div>
          <h1 className="text-xl font-semibold text-white">Dashboard</h1>
          <p className="text-slate-500 text-xs mt-0.5">Company overview, pending money, broker and dealer performance</p>
        </div>
        <div className="flex items-center gap-2">
          <SegmentedButton active={mode === "all"} onClick={() => setMode("all")}>All time</SegmentedButton>
          <SegmentedButton active={mode === "month"} onClick={() => setMode("month")}>Current month</SegmentedButton>
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
          <section className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
            <Metric label="Deals" value={totals.total_deals.toLocaleString()} />
            <Metric label="Gross revenue" value={money(totals.gross_revenue)} />
            <Metric label="Company profit" value={money(totals.company_profit_before_tax)} tone="green" />
            <Metric label="Pending revenue" value={money(totals.pending_revenue)} tone="yellow" />
            <Metric label="Paid revenue" value={money(totals.paid_revenue)} tone="green" />
            <Metric label="Broker payout" value={money(totals.broker_payout)} />
            <Metric label="Doc fees" value={money(totals.total_doc_fees)} />
            <Metric label="Avg revenue / deal" value={money(totals.avg_revenue_per_deal)} />
            <Metric label="Avg profit / deal" value={money(totals.avg_profit_per_deal)} />
          </section>

          <section className="grid grid-cols-1 lg:grid-cols-3 gap-4">
            <AgingCard title="0-31 days" count={aging.pending_0_31_count} amount={aging.pending_0_31_revenue} tone="green" />
            <AgingCard title="32-40 days" count={aging.pending_32_40_count} amount={aging.pending_32_40_revenue} tone="yellow" />
            <AgingCard title="41+ days" count={aging.pending_41_plus_count} amount={aging.pending_41_plus_revenue} tone="red" />
          </section>

          <section className="grid grid-cols-1 xl:grid-cols-2 gap-5">
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
                  d.avg_days_to_payment == null ? "-" : Number(d.avg_days_to_payment).toFixed(1),
                ])}
              />
            </Panel>
          </section>

          <Panel title="Recent deals">
            <div className="divide-y divide-[#1e1f20]">
              {summary.recent_deals.length === 0 ? (
                <div className="py-8 text-center text-sm text-slate-600">No recent deals</div>
              ) : (
                summary.recent_deals.map((deal) => (
                  <div
                    key={deal.id}
                    className="grid grid-cols-5 gap-3 px-4 py-3 text-sm hover:bg-blue-500/5 transition-colors"
                  >
                    <span className="text-slate-200">{[deal.make, deal.model].filter(Boolean).join(" ") || "Deal"}</span>
                    <span className="text-slate-500">{deal.deal_date ?? "-"}</span>
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
    </div>
  );
}

function currentMonthRange() {
  const now = new Date();
  const first = new Date(now.getFullYear(), now.getMonth(), 1);
  const last = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  return { from: formatDateInput(first), to: formatDateInput(last) };
}

function formatDateInput(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function money(value?: number) {
  return `$${Number(value ?? 0).toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
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

function DateInput({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
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

function Metric({ label, value, tone = "default" }: { label: string; value: string; tone?: "default" | "green" | "yellow" }) {
  const border = tone === "green" ? "border-green-500/25" : tone === "yellow" ? "border-yellow-500/25" : "border-[#252628]";
  const glow = tone === "green" ? "shadow-green-500/5" : tone === "yellow" ? "shadow-yellow-500/5" : "shadow-blue-500/5";
  const dot = tone === "green" ? "bg-green-400" : tone === "yellow" ? "bg-yellow-400" : "bg-blue-500";
  const valueColor = tone === "green" ? "text-green-400" : tone === "yellow" ? "text-yellow-400" : "text-white";
  return (
    <div className={`bg-[#111213] border ${border} rounded-2xl p-5 shadow-lg ${glow} flex flex-col gap-3`}>
      <div className="flex items-center gap-2">
        <span className={`w-1.5 h-1.5 rounded-full ${dot} shrink-0`} />
        <p className="text-xs text-slate-500 uppercase tracking-widest font-medium truncate">{label}</p>
      </div>
      <p className={`text-3xl font-bold tracking-tight ${valueColor}`}>{value}</p>
    </div>
  );
}

function AgingCard({ title, count, amount, tone }: { title: string; count: number; amount?: number; tone: "green" | "yellow" | "red" }) {
  const bg = tone === "green" ? "bg-green-500/5" : tone === "yellow" ? "bg-yellow-500/5" : "bg-red-500/5";
  const border = tone === "green" ? "border-green-500/20" : tone === "yellow" ? "border-yellow-500/20" : "border-red-500/20";
  const countColor = tone === "green" ? "text-green-400" : tone === "yellow" ? "text-yellow-400" : "text-red-400";
  const badge = tone === "green" ? "bg-green-500/15 text-green-400" : tone === "yellow" ? "bg-yellow-500/15 text-yellow-400" : "bg-red-500/15 text-red-400";
  const badgeLabel = tone === "green" ? "On time" : tone === "yellow" ? "Watch" : "Overdue";
  return (
    <div className={`${bg} border ${border} rounded-2xl p-5 flex flex-col gap-4`}>
      <div className="flex items-center justify-between">
        <p className="text-xs text-slate-500 uppercase tracking-widest font-medium">{title}</p>
        <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${badge}`}>{badgeLabel}</span>
      </div>
      <div>
        <p className={`text-4xl font-bold tracking-tight ${countColor}`}>{count}</p>
        <p className="text-sm text-slate-400 mt-1">{money(amount)}</p>
      </div>
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
        {rows.map((row, index) => (
          <tr key={index}>
            {row.map((cell, i) => (
              <td key={i} className={i === 0 ? "px-4 py-3 text-slate-200" : "px-4 py-3 text-slate-400"}>{cell}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
