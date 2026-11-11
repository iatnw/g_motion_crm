"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import {
  dashboard,
  deals,
  clients,
  brokers,
  dealers,
  referrals as referralsApi,
} from "@/lib/api";
import type {
  DealStatus,
  DashboardSummary,
  Deal,
  Client,
  PendingPaymentBucket,
  Referral,
} from "@/lib/types";
import DealStatusSelect from "@/components/DealStatusSelect";
import DealModal from "@/components/DealModal";
import PageHeading from "@/components/PageHeading";
import Modal from "@/components/Modal";

type RangeMode = "all" | "month" | "custom";
type DealScope = "all" | "no_cash";

export default function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [mode, setMode] = useState<RangeMode>("all");
  const [dealScope, setDealScope] = useState<DealScope>("all");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const [bucketModal, setBucketModal] = useState<{
    bucket: PendingPaymentBucket;
    deals: Deal[];
    clientMap: Record<string, Client>;
    brokerMap: Record<string, string>;
    dealerMap: Record<string, string>;
    referralMap: Record<string, Referral[]>;
    allClients: Client[];
    allBrokers: import("@/lib/types").Broker[];
    allDealers: import("@/lib/types").Dealer[];
  } | null>(null);
  const [bucketLoading, setBucketLoading] = useState(false);
  const [collectionError, setCollectionError] = useState("");
  const bucketRequest = useRef(0);
  const [editModal, setEditModal] = useState<{
    deal: Deal;
    clients: Client[];
    brokers: import("@/lib/types").Broker[];
    dealers: import("@/lib/types").Dealer[];
  } | null>(null);

  const params = useMemo<{ from?: string; to?: string; exclude_cash?: boolean }>(() => {
    const base =
      mode === "all"
        ? {}
        : mode === "month"
          ? currentMonthRange()
          : { from: from || undefined, to: to || undefined };
    return dealScope === "no_cash"
      ? { ...base, exclude_cash: true as const }
      : base;
  }, [mode, from, to, dealScope]);

  useEffect(() => {
    let cancelled = false;

    function refresh() {
      dashboard
        .summary({ ...params, limit: 8 })
        .then((data) => {
          if (!cancelled) {
            setSummary(data);
            setError("");
          }
        })
        .catch((e: unknown) => {
          if (!cancelled)
            setError(
              e instanceof Error ? e.message : "Failed to load dashboard",
            );
        })
        .finally(() => {
          if (!cancelled) setLoading(false);
        });
    }

    function refreshWhenVisible() {
      if (document.visibilityState === "visible") refresh();
    }

    refresh();
    window.addEventListener("focus", refresh);
    document.addEventListener("visibilitychange", refreshWhenVisible);
    return () => {
      cancelled = true;
      window.removeEventListener("focus", refresh);
      document.removeEventListener("visibilitychange", refreshWhenVisible);
    };
  }, [params]);

  async function handleDealStatusChange(id: string, status: DealStatus) {
    await deals.updateStatus(id, status);
    const next = await dashboard.summary({ ...params, limit: 8 });
    setSummary(next);
  }

  async function openBucket(bucket: PendingPaymentBucket) {
    const request = ++bucketRequest.current;
    setCollectionError("");
    setBucketLoading(true);
    try {
      const [allDeals, allClients, allBrokers, allDealers, allReferrals] =
        await Promise.all([
          deals.list({ vin: "present" }),
          clients.list(),
          brokers.list(),
          dealers.list(),
          referralsApi.list(),
        ]);
      const filtered = (allDeals ?? []).filter(
        (d) =>
          d.deal_status === "pending" &&
          d.pending_payment_bucket === bucket &&
          isDealInRange(d, params) &&
          !(params.exclude_cash && d.is_cash_deal),
      );
      const clientMap = Object.fromEntries(
        (allClients ?? []).map((c) => [c.id, c]),
      );
      const brokerMap = Object.fromEntries(
        (allBrokers ?? []).map((b) => [b.id, `${b.first_name} ${b.last_name}`]),
      );
      const dealerMap = Object.fromEntries(
        (allDealers ?? []).map((d) => [d.id, d.name]),
      );
      const referralMap: Record<string, Referral[]> = {};
      for (const r of allReferrals ?? []) {
        if (r.deal_id) {
          if (!referralMap[r.deal_id]) referralMap[r.deal_id] = [];
          referralMap[r.deal_id].push(r);
        }
      }
      if (request !== bucketRequest.current) return;
      filtered.sort((a, b) => (b.pending_payment_days ?? 0) - (a.pending_payment_days ?? 0));
      setBucketModal({
        bucket,
        deals: filtered,
        clientMap,
        brokerMap,
        dealerMap,
        referralMap,
        allClients: allClients ?? [],
        allBrokers: allBrokers ?? [],
        allDealers: allDealers ?? [],
      });
    } catch {
      if (request === bucketRequest.current) setCollectionError("Couldn’t load these collections. Please try again.");
    } finally {
      if (request === bucketRequest.current) setBucketLoading(false);
    }
  }

  function closeBucket() {
    bucketRequest.current++;
    setBucketLoading(false);
    setBucketModal(null);
  }

  const totals = summary?.totals;
  const aging = summary?.aging;

  const paidRev = totals?.paid_revenue ?? 0;
  const pendingRev = totals?.pending_revenue ?? 0;
  const grossRev = totals?.gross_revenue ?? 0;
  const collectionPct =
    grossRev > 0 ? Math.round((paidRev / grossRev) * 100) : 0;

  return (
    <div className="p-6">
      <PageHeading eyebrow="Your business at a glance" title="Overview" description="A clear view of your deals, collections, and what comes next." />
      <section className="crm-reporting-panel" aria-label="Report settings">
        <div className="crm-reporting-main">
          <div className="crm-reporting-heading">
            <span className="crm-reporting-icon" aria-hidden="true">
              <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round"><rect x="3" y="5" width="18" height="16" rx="3" /><path d="M8 3v4M16 3v4M3 11h18M8 15h3" /></svg>
            </span>
            <div><h2>Reporting period</h2><p>Choose what your overview includes.</p></div>
          </div>
          <div className="crm-reporting-controls">
            <div className="crm-reporting-control">
              <p>Period</p>
              <div className="crm-segmented-group" role="group" aria-label="Reporting period">
                <SegmentedButton active={mode === "all"} onClick={() => setMode("all")}>All time</SegmentedButton>
                <SegmentedButton active={mode === "month"} onClick={() => setMode("month")}>This month</SegmentedButton>
                <SegmentedButton active={mode === "custom"} onClick={() => setMode("custom")}>Custom range</SegmentedButton>
              </div>
            </div>
            <div className="crm-reporting-control">
              <p>Deal scope</p>
              <div className="crm-segmented-group" role="group" aria-label="Deal scope">
                <SegmentedButton active={dealScope === "all"} onClick={() => setDealScope("all")}>All deals</SegmentedButton>
                <SegmentedButton active={dealScope === "no_cash"} onClick={() => setDealScope("no_cash")}>Exclude cash</SegmentedButton>
              </div>
            </div>
          </div>
        </div>
        {mode === "custom" && (
          <div className="crm-reporting-dates">
            <DateInput label="From" value={from} onChange={setFrom} />
            <span className="crm-date-range-arrow" aria-hidden="true">→</span>
            <DateInput label="To" value={to} onChange={setTo} />
          </div>
        )}
      </section>


      {loading ? (
        <div className="text-center py-12 text-slate-600 text-sm">
          Loading...
        </div>
      ) : error ? (
        <div className="text-red-400 text-sm">{error}</div>
      ) : summary && totals && aging ? (
        <div className="space-y-6">
          {/* Hero row */}
          <div className="crm-dashboard-metrics grid gap-4">
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
              <p className="text-xs font-semibold text-slate-500 uppercase tracking-widest">
                Revenue collection
              </p>
              <span
                className={`text-xs font-semibold px-2.5 py-1 rounded-full ${
                  collectionPct >= 80
                    ? "bg-green-500/15 text-green-400"
                    : collectionPct >= 50
                      ? "bg-yellow-500/15 text-yellow-400"
                      : "bg-red-500/15 text-red-400"
                }`}
              >
                {collectionPct}% collected
              </span>
            </div>
            <div className="w-full h-2.5 bg-[#1e1f21] rounded-full overflow-hidden mb-4">
              <div
                className={`h-full rounded-full transition-all duration-700 ${
                  collectionPct >= 80
                    ? "bg-green-500"
                    : collectionPct >= 50
                      ? "bg-yellow-500"
                      : "bg-red-500"
                }`}
                style={{ width: `${collectionPct}%` }}
              />
            </div>
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <p className="text-xs text-slate-600 uppercase tracking-wider mb-1">
                  Gross
                </p>
                <p className="text-base font-semibold text-white">
                  {money(grossRev)}
                </p>
              </div>
              <div>
                <p className="text-xs text-slate-600 uppercase tracking-wider mb-1">
                  Collected
                </p>
                <p className="text-base font-semibold text-green-400">
                  {money(paidRev)}
                </p>
              </div>
              <div>
                <p className="text-xs text-slate-600 uppercase tracking-wider mb-1">
                  Outstanding
                </p>
                <p className="text-base font-semibold text-yellow-400">
                  {money(pendingRev)}
                </p>
              </div>
            </div>
          </div>

          {/* Aging — most actionable, shown high */}
          <section className="crm-collections" aria-labelledby="collections-title">
            <div className="crm-collections-heading">
              <div><h2 id="collections-title">Collections to follow up</h2><p>See what’s outstanding and where to focus next.</p></div>
              <div className="crm-collections-total"><span>Total outstanding</span><strong>{money(pendingRev)}</strong></div>
            </div>
            {collectionError && <p role="alert" className="crm-collection-error">{collectionError}</p>}
            <div className="crm-collections-grid">
              <AgingCard
                title="0–31 days"
                count={aging.pending_0_31_count}
                amount={aging.pending_0_31_revenue}
                tone="green"
                disabled={bucketLoading}
                badge="On time"
                onClick={() => openBucket("green")}
              />
              <AgingCard
                title="32–40 days"
                count={aging.pending_32_40_count}
                amount={aging.pending_32_40_revenue}
                tone="yellow"
                disabled={bucketLoading}
                badge="Needs attention"
                onClick={() => openBucket("yellow")}
              />
              <AgingCard
                title="41+ days"
                count={aging.pending_41_plus_count}
                amount={aging.pending_41_plus_revenue}
                tone="red"
                disabled={bucketLoading}
                badge="Overdue"
                onClick={() => openBucket("red")}
              />
            </div>
          </section>

          {/* Secondary stats — compact */}
          <div className="space-y-3">
            <SectionLabel>Where your revenue goes</SectionLabel>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <StatChip
                label="Broker payout"
                value={money(totals.broker_payout)}
              />
              <StatChip label="Doc fees" value={money(totals.total_doc_fees)} />
              <StatChip
                label="Referral payout"
                value={money(totals.total_referral_payout)}
              />
            </div>
          </div>

          {/* Leases — ending soonest first */}
          <div className="space-y-3">
            <SectionLabel>Leases</SectionLabel>
            <Panel title="Lease tracking — ending soonest first">
              {summary.leases.length === 0 ? (
                <div className="py-8 text-center text-sm text-slate-600">
                  No leases tracked yet. Set a lease term on a deal to see it
                  here.
                </div>
              ) : (
                <Table
                  headers={[
                    "Client",
                    "Vehicle",
                    "Deal",
                    "Term",
                    "Left",
                    "Status",
                  ]}
                  rows={summary.leases.map((l) => [
                    <Link
                      key={l.id}
                      href={`/clients/${l.client_id}`}
                      className="hover:text-blue-400 transition-colors"
                    >
                      {l.client_name}
                    </Link>,
                    [l.make, l.model].filter(Boolean).join(" ") || "—",
                    l.deal_number ?? "—",
                    `${l.lease_term_months} mo`,
                    l.lease_months_remaining > 0
                      ? `${l.lease_months_remaining} mo`
                      : "Ended",
                    l.deal_status ?? "—",
                  ])}
                />
              )}
            </Panel>
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
                    <Link
                      key={d.id}
                      href={`/dealers/${d.id}`}
                      className="hover:text-blue-400 transition-colors"
                    >
                      {d.name}
                    </Link>,
                    d.deal_count.toLocaleString(),
                    money(d.gross_revenue),
                    formatDays(d.avg_days_to_payment),
                  ])}
                />
              </Panel>
            </div>
          </div>

          {/* Recent deals */}
          <Panel title="Recent deals">
            <div className="overflow-x-auto">
              <div className="divide-y divide-[#1e1f20] min-w-[460px]">
                {summary.recent_deals.length === 0 ? (
                  <div className="py-8 text-center text-sm text-slate-600">
                    No recent deals
                  </div>
                ) : (
                  summary.recent_deals.map((deal) => (
                    <div
                      key={deal.id}
                      className="grid grid-cols-5 gap-3 px-4 py-3 text-sm hover:bg-blue-500/5 transition-colors"
                    >
                      <span className="text-slate-200">
                        {[deal.make, deal.model].filter(Boolean).join(" ") ||
                          "Deal"}
                      </span>
                      <span className="text-slate-500">
                        {deal.deal_date ?? "—"}
                      </span>
                      <span className="text-slate-400">
                        {money(deal.broker_fee)}
                      </span>
                      <DealStatusSelect
                        value={deal.deal_status}
                        days={deal.pending_payment_days}
                        bucket={deal.pending_payment_bucket}
                        onChange={(status) =>
                          handleDealStatusChange(deal.id, status)
                        }
                      />
                      <Link
                        href="/deals"
                        className="text-right text-blue-400 hover:text-blue-300"
                      >
                        View
                      </Link>
                    </div>
                  ))
                )}
              </div>
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
          onClose={closeBucket}
          onEdit={(deal) => {
            setEditModal({
              deal,
              clients: bucketModal?.allClients ?? [],
              brokers: bucketModal?.allBrokers ?? [],
              dealers: bucketModal?.allDealers ?? [],
            });
            setBucketModal(null);
          }}
        />
      )}

      {editModal && (
        <DealModal
          deal={editModal.deal}
          clients={editModal.clients}
          brokers={editModal.brokers}
          dealers={editModal.dealers}
          onSave={async (data) => {
            const updated = await deals.update(editModal.deal.id, data);
            setEditModal(null);
            const next = await dashboard.summary({ ...params, limit: 8 });
            setSummary(next);
            return updated;
          }}
          onClose={() => setEditModal(null)}
        />
      )}
    </div>
  );
}

// --- Components ---

function HeroCard({
  label,
  value,
  sub,
  accent,
}: {
  label: string;
  value: string;
  sub: string;
  accent: "blue" | "green" | "default";
}) {
  const border =
    accent === "green"
      ? "border-green-500/25"
      : accent === "blue"
        ? "border-blue-500/20"
        : "border-[#252628]";
  const valueColor =
    accent === "green"
      ? "text-green-400"
      : accent === "blue"
        ? "text-blue-400"
        : "text-white";
  const dot =
    accent === "green"
      ? "bg-green-400"
      : accent === "blue"
        ? "bg-blue-500"
        : "bg-slate-500";
  return (
    <div
      className={`crm-metric-card bg-[#111213] border ${border} rounded-2xl p-6 flex flex-col gap-2`}
    >
      <div className="flex items-center gap-2">
        <span className={`w-1.5 h-1.5 rounded-full ${dot} shrink-0`} />
        <p className="text-xs text-slate-500 uppercase tracking-widest font-medium">
          {label}
        </p>
      </div>
      <p className={`text-4xl font-bold tracking-tight ${valueColor}`}>
        {value}
      </p>
      <p className="text-xs text-slate-600">{sub}</p>
    </div>
  );
}

function StatChip({
  label,
  value,
  warn,
}: {
  label: string;
  value: string;
  warn?: boolean;
}) {
  return (
    <div className="bg-[#111213] border border-[#1e1f21] rounded-xl px-4 py-3 flex items-center justify-between gap-3">
      <p className="text-xs text-slate-500 truncate">{label}</p>
      <p
        className={`text-sm font-semibold shrink-0 ${warn ? "text-red-400" : "text-slate-200"}`}
      >
        {value}
      </p>
    </div>
  );
}

function formatDays(value?: number) {
  if (value == null) return "—";
  const days = Math.round(Number(value));
  return `${days} ${days === 1 ? "day" : "days"}`;
}

function AgingCard({ title, count, amount, tone, badge, disabled, onClick }: {
  title: string;
  count: number;
  amount?: number;
  tone: "green" | "yellow" | "red";
  badge: string;
  disabled: boolean;
  onClick: () => void;
}) {
  const description = count === 0 ? "No outstanding payments in this group." : tone === "red" ? "Prioritize these outstanding payments." : tone === "yellow" ? "Follow up before these become overdue." : "Within the expected collection window.";
  return (
    <button type="button" onClick={onClick} disabled={disabled || count === 0} aria-haspopup="dialog"
      aria-label={`Review ${badge.toLowerCase()} collections: ${count} ${count === 1 ? "deal" : "deals"}, ${title}`}
      className={`crm-collection-card crm-collection-${tone}`}>
      <div className="crm-collection-card-top">
        <span className="crm-collection-title">{badge}</span>
        <span className="crm-collection-indicator" aria-hidden="true" />
      </div>
      <strong className="crm-collection-amount">{money(amount)}</strong>
      <span className="crm-collection-count">{count} {count === 1 ? "deal awaiting collection" : "deals awaiting collection"}</span>
      <p className="crm-collection-description">{description}</p>
      <span className="crm-collection-action"><span>{count ? "Review collections" : "All clear"}</span><span aria-hidden="true">{count ? "↗" : "✓"}</span></span>
    </button>
  );
}

const bucketLabel: Record<string, string> = {
  green: "0–31 days",
  yellow: "32–40 days",
  red: "41+ days",
};

function BucketModal({
  loading,
  bucket,
  deals,
  clientMap,
  brokerMap,
  dealerMap,
  referralMap,
  onClose,
  onEdit,
}: {
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
  const tone =
    bucket === "green"
      ? "text-green-400"
      : bucket === "yellow"
        ? "text-yellow-400"
        : "text-red-400";
  const borderColor =
    bucket === "green"
      ? "border-l-green-500"
      : bucket === "yellow"
        ? "border-l-amber-500"
        : "border-l-red-500";

  return (
    <Modal title="Pending deals" description={bucket ? bucketLabel[bucket] : "Review outstanding payments and follow up."} onClose={onClose} size="wide">
        <div className="crm-payment-list p-5 space-y-3">

          {loading ? (
            <div className="py-16 text-center text-sm text-slate-600">
              Loading...
            </div>
          ) : deals.length === 0 ? (
            <div className="py-16 text-center text-sm text-slate-600">
              No deals in this range
            </div>
          ) : (
            deals.map((d) => {
              const client = clientMap[d.client_id];
              const vehicle = [d.make, d.model].filter(Boolean).join(" ");
              const dealReferrals = referralMap[d.id] ?? [];
              const referralTotal = dealReferrals.reduce(
                (s, r) => s + (r.amount ?? 0),
                0,
              );
              const base =
                (d.broker_fee ?? 0) +
                (d.collect_amount ?? 0) -
                (d.doc_fee ?? 0);
              const commission =
                d.broker_commission ||
                (d.broker_percent != null && base > 0
                  ? (base * d.broker_percent) / 100
                  : 0);
              const computedNet =
                base > 0 ? base - commission - referralTotal : null;
              return (
                <div
                  key={d.id}
                  className={`bg-[#111213] border border-[#252628] border-l-4 ${borderColor} rounded-xl`}
                >
                  {/* Top bar */}
                  <div className="flex items-center justify-between px-4 py-3 border-b border-[#1e1f20]">
                    <div className="flex items-center gap-3 flex-wrap min-w-0">
                      <span className="text-white font-semibold">
                        {client
                          ? `${client.first_name} ${client.last_name}`
                          : "Unknown"}
                      </span>
                      {d.deal_number && (
                        <span className="text-xs font-mono bg-[#0b0c0d] border border-[#252628] text-slate-400 px-2 py-0.5 rounded-md">
                          #{d.deal_number}
                        </span>
                      )}
                      {vehicle && (
                        <span className="text-slate-300 text-sm">
                          {vehicle}
                        </span>
                      )}
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
                  <div className="crm-payment-stats flex flex-wrap divide-x divide-[#1e1f20]">
                    {d.broker_id && (
                      <BucketStat
                        label="Broker"
                        value={brokerMap[d.broker_id]}
                      />
                    )}
                    {d.dealer_id && (
                      <BucketStat
                        label="Dealer"
                        value={dealerMap[d.dealer_id]}
                      />
                    )}
                    {d.broker_fee != null && (
                      <BucketStat
                        label="Broker Fee"
                        value={money(d.broker_fee)}
                        highlight="green"
                      />
                    )}
                    {computedNet != null && (
                      <BucketStat
                        label="Net"
                        value={money(computedNet)}
                        highlight="blue"
                      />
                    )}
                    {d.broker_commission != null && (
                      <BucketStat
                        label="Commission"
                        value={money(d.broker_commission)}
                      />
                    )}
                    {d.doc_fee != null && (
                      <BucketStat label="Doc Fee" value={money(d.doc_fee)} />
                    )}
                  </div>
                </div>
              );
            })
          )}
        </div>
    </Modal>
  );
}

function BucketStat({
  label,
  value,
  highlight,
}: {
  label: string;
  value?: string;
  highlight?: "green" | "blue";
}) {
  if (!value) return null;
  const cls =
    highlight === "green"
      ? "text-green-400 font-bold"
      : highlight === "blue"
        ? "text-blue-400 font-bold"
        : "text-slate-300";
  return (
    <div className="flex-1 px-3 py-2 flex flex-col gap-0.5 min-w-0">
      <span className="text-xs text-slate-600 uppercase tracking-wider">
        {label}
      </span>
      <span className={`text-sm truncate ${cls}`}>{value}</span>
    </div>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-3">
      <p className="text-xs font-semibold text-slate-500 uppercase tracking-widest">
        {children}
      </p>
      <div className="flex-1 h-px bg-[#1e1f21]" />
    </div>
  );
}

function Panel({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
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

function Table({
  headers,
  rows,
}: {
  headers: string[];
  rows: React.ReactNode[][];
}) {
  if (!rows.length)
    return (
      <div className="py-8 text-center text-sm text-slate-600">No data</div>
    );
  return (
    <table className="w-full text-sm">
      <thead className="border-b border-[#252628]">
        <tr>
          {headers.map((h) => (
            <th
              key={h}
              className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider"
            >
              {h}
            </th>
          ))}
        </tr>
      </thead>
      <tbody className="divide-y divide-[#1e1f20]">
        {rows.map((row, i) => (
          <tr key={i}>
            {row.map((cell, j) => (
              <td
                key={j}
                className={
                  j === 0
                    ? "px-4 py-3 text-slate-200"
                    : "px-4 py-3 text-slate-400"
                }
              >
                {cell}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function SegmentedButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      aria-pressed={active}
      className="crm-filter-text crm-segment-button"
    >
      {children}
    </button>
  );
}

function DateInput({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
}) {
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

function isDealInRange(deal: Deal, range: { from?: string; to?: string }) {
  const date = deal.deal_date ?? deal.created_at.slice(0, 10);
  return (!range.from || date >= range.from) && (!range.to || date <= range.to);
}
