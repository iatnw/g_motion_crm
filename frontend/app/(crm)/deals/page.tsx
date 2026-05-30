"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { deals, clients, brokers, dealers, referrals as referralsApi } from "@/lib/api";
import type { Deal, Client, Broker, Dealer, DealStatus, Referral, ReferralStatus } from "@/lib/types";
import DealModal from "@/components/DealModal";
import DealStatusSelect from "@/components/DealStatusSelect";
import { fmtDate, groupReferralsByDeal } from "@/lib/utils";

type FilterTab = "all" | "pending" | "paid" | "awaiting_vin";
const PAGE_SIZE = 20;
const filterTabs: Array<{ value: FilterTab; label: string }> = [
  { value: "all", label: "All" },
  { value: "pending", label: "Pending" },
  { value: "paid", label: "Paid" },
  { value: "awaiting_vin", label: "Awaiting VIN" },
];

const statusBorder: Record<string, string> = {
  paid:    "border-l-green-500",
  pending: "border-l-amber-500",
};

function belongsToTab(deal: Deal, tab: FilterTab) {
  if (deal.deal_status === "cancelled") return false;
  const hasVIN = !!deal.vin?.trim();
  if (tab === "awaiting_vin") return !hasVIN;
  if (!hasVIN) return false;
  return tab === "all" || deal.deal_status === tab;
}

export default function DealsPage() {
  const [dealList, setDealList] = useState<Deal[]>([]);
  const [allClients, setAllClients] = useState<Client[]>([]);
  const [clientMap, setClientMap] = useState<Record<string, Client>>({});
  const [allBrokers, setAllBrokers] = useState<Broker[]>([]);
  const [allDealers, setAllDealers] = useState<Dealer[]>([]);
  const [brokerMap, setBrokerMap] = useState<Record<string, string>>({});
  const [dealerMap, setDealerMap] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [querySearch, setQuerySearch] = useState("");
  const [tab, setTab] = useState<FilterTab>("all");
  const [modal, setModal] = useState<{ open: boolean; deal?: Deal; nextNumber?: string }>({ open: false });
  const [referralMap, setReferralMap] = useState<Record<string, Referral[]>>({});
  const [counts, setCounts] = useState({ all: 0, pending: 0, paid: 0, awaiting_vin: 0 });
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [refreshVersion, setRefreshVersion] = useState(0);

  useEffect(() => {
    Promise.all([clients.list(), brokers.list(), dealers.list()])
      .then(([c, b, dl]) => {
        setAllClients(c ?? []);
        setClientMap(Object.fromEntries((c ?? []).map((x) => [x.id, x])));
        setAllBrokers(b ?? []);
        setAllDealers(dl ?? []);
        setBrokerMap(Object.fromEntries((b ?? []).map((x) => [x.id, `${x.first_name} ${x.last_name}`])));
        setDealerMap(Object.fromEntries((dl ?? []).map((x) => [x.id, x.name])));
      });
  }, []);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      setOffset(0);
      setQuerySearch(search.trim());
    }, 250);
    return () => window.clearTimeout(timeout);
  }, [search]);

  useEffect(() => {
    let cancelled = false;
    deals.listPaged({
      limit: PAGE_SIZE,
      offset,
      status: tab === "pending" || tab === "paid" ? tab : undefined,
      search: querySearch,
      vin: tab === "awaiting_vin" ? "missing" : "present",
    }).then(async (res) => {
      const refs = await referralsApi.listByDeals(res.data.map((deal) => deal.id));
      if (cancelled) return;
      setDealList((prev) => offset === 0 ? res.data : [...prev, ...res.data]);
      setReferralMap((prev) => {
        const pageMap = groupReferralsByDeal(refs ?? []);
        return offset === 0 ? pageMap : { ...prev, ...pageMap };
      });
      setHasMore(res.has_more);
      if (res.counts) setCounts(res.counts);
    }).finally(() => {
      if (!cancelled) setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [tab, querySearch, offset, refreshVersion]);

  async function handleSave(data: object): Promise<Deal> {
    if (modal.deal) {
      const updated = await deals.update(modal.deal.id, data);
      setLoading(true);
      setOffset(0);
      setRefreshVersion((version) => version + 1);
      return updated;
    } else {
      const created = await deals.create(data);
      setLoading(true);
      setOffset(0);
      setRefreshVersion((version) => version + 1);
      return created;
    }
  }

  async function handleStatusChange(deal: Deal, status: DealStatus) {
    const updated = await deals.updateStatus(deal.id, status);
    setDealList((prev) => {
      if (!belongsToTab(updated, tab)) {
        return prev.filter((d) => d.id !== updated.id);
      }
      return prev.map((d) => (d.id === updated.id ? updated : d));
    });
    setRefreshVersion((version) => version + 1);
  }

  async function handleCommissionMethod(deal: Deal, status: "paid" | "pending") {
    const updated = await deals.updateBrokerCommissionStatus(deal.id, status);
    setDealList((prev) => prev.map((d) => (d.id === updated.id ? updated : d)));
  }

  async function handleReferralStatus(ref: Referral, status: ReferralStatus) {
    const updated = await referralsApi.updateStatus(ref.id, status);
    setReferralMap((prev) => {
      if (!ref.deal_id) return prev;
      return {
        ...prev,
        [ref.deal_id]: (prev[ref.deal_id] ?? []).map((r) => (r.id === ref.id ? { ...r, ...updated } : r)),
      };
    });
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-white">Deals</h1>
          <p className="text-slate-500 text-xs mt-0.5">{counts.all} total</p>
        </div>
        <button
          onClick={async () => {
            const num = await deals.nextNumber().catch(() => "");
            setModal({ open: true, nextNumber: num });
          }}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition-all glow-blue-sm hover:glow-blue"
        >
          + New Deal
        </button>
      </div>

      {/* Filter tabs + search */}
      <div className="flex items-center gap-3 mb-5">
        <div className="flex bg-[#0b0c0d] border border-[#252628] rounded-lg p-0.5 gap-0.5">
          {filterTabs.map(({ value, label }) => (
            <button
              key={value}
              onClick={() => {
                if (value === tab) return;
                setLoading(true);
                setTab(value);
                setOffset(0);
              }}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                tab === value
                  ? "bg-[#1a1b1d] text-white"
                  : "text-slate-500 hover:text-slate-300"
              }`}
            >
              {label} <span className={`ml-1 ${tab === value ? "text-slate-400" : "text-slate-600"}`}>{counts[value]}</span>
            </button>
          ))}
        </div>
        <div className="relative flex-1">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
          </svg>
          <input
            type="text"
            placeholder="Search by client, broker, dealer, car, VIN, deal #…"
            value={search}
            onChange={(e) => {
              setLoading(true);
              setSearch(e.target.value);
            }}
            className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg pl-9 pr-3 py-2 text-sm text-white placeholder:text-slate-700 focus:outline-none"
          />
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      ) : dealList.length === 0 ? (
        <div className="text-center py-12 text-slate-600 text-sm">No deals found</div>
      ) : (
        <div className="space-y-3">
          {dealList.map((d) => {
            const client = clientMap[d.client_id];
            const vehicle = [d.make, d.model].filter(Boolean).join(" ");
            const borderColor = statusBorder[d.deal_status ?? ""] ?? "border-l-[#252628]";
            const dealReferrals = referralMap[d.id] ?? [];
            const referralTotal = dealReferrals.reduce((sum, r) => sum + (r.amount ?? 0), 0);
            const base = (d.broker_fee ?? 0) + (d.collect_amount ?? 0) - (d.doc_fee ?? 0);
            const commission = d.broker_commission || (d.broker_percent != null && base > 0 ? base * d.broker_percent / 100 : 0);
            const displayPercent = d.broker_percent || (commission > 0 && base > 0 ? commission / base * 100 : null);
            const computedNet = base - commission - referralTotal;

            return (
              <div key={d.id} className={`bg-[#0b0c0d] border border-[#252628] border-l-4 ${borderColor} rounded-xl hover:border-[#333436] hover:border-l-4 transition-colors`}>

                {/* Top bar */}
                <div className="flex items-center justify-between px-5 py-3.5 border-b border-[#1e1f20]">
                  <div className="flex items-center gap-4 min-w-0 flex-wrap">
                    {client ? (
                      <Link href={`/clients/${d.client_id}`} className="text-white font-semibold hover:text-blue-400 transition-colors whitespace-nowrap">
                        {client.first_name} {client.last_name}
                      </Link>
                    ) : (
                      <span className="text-slate-600 font-semibold">Unknown</span>
                    )}
                    {d.deal_number && (
                      <span className="text-xs font-mono bg-[#111213] border border-[#252628] text-slate-400 px-2 py-0.5 rounded-md">
                        #{d.deal_number}
                      </span>
                    )}
                    {vehicle && <span className="text-slate-300 text-sm">{vehicle}</span>}
                    {(d.trim_level || d.vin) && (
                      <span className="text-slate-600 text-xs hidden lg:inline">
                        {[d.trim_level, d.vin].filter(Boolean).join(" · ")}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-4 flex-shrink-0 ml-4">
                    {d.deal_date && (
                      <span className="text-slate-600 text-xs hidden sm:inline">
                        {fmtDate(d.deal_date)}
                      </span>
                    )}
                    <DealStatusSelect
                      value={d.deal_status}
                      days={d.pending_payment_days}
                      bucket={d.pending_payment_bucket}
                      onChange={(status) => handleStatusChange(d, status)}
                    />
                    <button
                      onClick={() => setModal({ open: true, deal: d })}
                      className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                    >
                      Edit
                    </button>
                  </div>
                </div>

                {/* Stats bar — 2-col grid on mobile, horizontal flex on desktop */}
                <div className="border-t border-[#1e1f20] grid grid-cols-2 md:flex">
                  <StatCell label="Broker" value={d.broker_id ? brokerMap[d.broker_id] : undefined} />
                  <StatCell label="Dealer" value={d.dealer_id ? dealerMap[d.dealer_id] : undefined} />
                  <StatCell label="Broker Fee" value={money(d.broker_fee)} highlight="green" />
                  <StatCell label="Net" value={money(computedNet)} highlight="blue" />
                  <StatCell label="Collect" value={money(d.collect_amount)} />
                  <StatCell label="Doc Fee" value={money(d.doc_fee)} />
                  <StatCell label="Broker %" value={displayPercent ? `${Number(displayPercent).toFixed(2)}%` : undefined} />
                  <div className="px-4 py-2.5 flex flex-col gap-0.5 border-b border-[#1e1f20] even:border-l md:border-b-0 md:border-l-0 md:border-r md:flex-1 md:min-w-[130px]">
                    <span className="text-xs text-slate-600 uppercase tracking-wider">Commission</span>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-slate-300">{money(commission) ?? "—"}</span>
                      <select
                        value={d.broker_commission_status ?? "pending"}
                        onChange={(e) => handleCommissionMethod(d, e.target.value as "paid" | "pending")}
                        onClick={(e) => e.stopPropagation()}
                        className={`text-xs border rounded-md px-1.5 py-0.5 focus:outline-none cursor-pointer ${
                          d.broker_commission_status === "paid"
                            ? "bg-green-500/10 border-green-500/25 text-green-400"
                            : "bg-amber-500/10 border-amber-500/25 text-amber-400"
                        }`}
                        style={{ backgroundColor: "transparent" }}
                      >
                        <option value="pending">NOT PAID</option>
                        <option value="paid">PAID</option>
                      </select>
                    </div>
                  </div>
                </div>
                {/* Referrals — stacked on mobile, horizontal on desktop */}
                {dealReferrals.length > 0 && (
                  <div className="border-t border-[#1e1f20] flex flex-col md:flex-row md:divide-x md:divide-[#1e1f20]">
                    {dealReferrals.map((ref, idx) => (
                      <div key={ref.id} className={`px-4 py-2.5 flex items-center justify-between gap-3 md:flex-col md:items-start md:justify-start md:gap-0.5 md:flex-1 md:min-w-[160px]${idx < dealReferrals.length - 1 ? " border-b border-[#1e1f20] md:border-b-0" : ""}`}>
                        <div className="flex flex-col gap-0.5 min-w-0">
                          <span className="text-xs text-slate-600 uppercase tracking-wider">
                            Referral{ref.amount != null ? ` · $${Number(ref.amount).toLocaleString()}` : ""}
                          </span>
                          <span className="text-sm text-slate-300 truncate">
                            {[ref.first_name, ref.last_name].filter(Boolean).join(" ")}
                          </span>
                        </div>
                        <select
                          value={ref.status}
                          onChange={(e) => handleReferralStatus(ref, e.target.value as ReferralStatus)}
                          onClick={(e) => e.stopPropagation()}
                          className={`text-xs border rounded-md px-1.5 py-0.5 focus:outline-none cursor-pointer shrink-0 ${
                            ref.status === "paid"
                              ? "bg-green-500/10 border-green-500/25 text-green-400"
                              : "bg-amber-500/10 border-amber-500/25 text-amber-400"
                          }`}
                          style={{ backgroundColor: "transparent" }}
                        >
                          <option value="none">NOT PAID</option>
                          <option value="paid">PAID</option>
                        </select>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            );
          })}
          {hasMore && (
            <div className="flex justify-center pt-2">
              <button
                onClick={() => {
                  setLoading(true);
                  setOffset((o) => o + PAGE_SIZE);
                }}
                disabled={loading}
                className="px-6 py-2 bg-[#0b0c0d] border border-[#252628] rounded-lg text-sm text-slate-400 hover:text-white hover:border-slate-500 transition-colors disabled:opacity-50"
              >
                {loading ? "Loading…" : "Load more"}
              </button>
            </div>
          )}
        </div>
      )}

      {modal.open && (
        <DealModal
          deal={modal.deal}
          nextDealNumber={modal.nextNumber}
          clients={allClients}
          brokers={allBrokers}
          dealers={allDealers}
          onSave={handleSave}
          onClose={() => setModal({ open: false })}
        />
      )}
    </div>
  );
}

function money(value?: number) {
  return value != null ? `$${Number(value).toLocaleString(undefined, { maximumFractionDigits: 2 })}` : undefined;
}


function StatCell({ label, value, highlight }: { label: string; value?: string; highlight?: "green" | "blue" }) {
  const valCls = highlight === "green" ? "text-green-400 font-bold text-base"
    : highlight === "blue" ? "text-blue-400 font-bold text-base"
    : "text-slate-300 text-sm";
  return (
    <div className="px-4 py-2.5 flex flex-col gap-0.5 border-b border-[#1e1f20] even:border-l md:border-b-0 md:border-l-0 md:border-r md:flex-1 md:min-w-[90px]">
      <span className="text-xs text-slate-600 uppercase tracking-wider whitespace-nowrap">{label}</span>
      <span className={value ? valCls + " truncate" : "text-slate-600 text-sm"}>{value ?? "—"}</span>
    </div>
  );
}
