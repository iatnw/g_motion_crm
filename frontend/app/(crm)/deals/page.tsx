"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  deals,
  clients,
  brokers,
  dealers,
  referrals as referralsApi,
} from "@/lib/api";
import type {
  Deal,
  Client,
  Broker,
  Dealer,
  DealStatus,
  Referral,
  ReferralStatus,
} from "@/lib/types";
import DealCoverModal from "@/components/DealCoverModal";
import BrokerAgreementModal from "@/components/BrokerAgreementModal";
import DealModal from "@/components/DealModal";
import StatusSelect, { paymentOptions, referralOptions, type StatusOption } from "@/components/StatusSelect";
import PageHeading from "@/components/PageHeading";
import DealStatusSelect from "@/components/DealStatusSelect";
import { fmtDate, groupReferralsByDeal } from "@/lib/utils";
import { vinError } from "@/lib/vin";

type FilterTab = "all" | "new" | "pending" | "paid" | "cancelled";
type LeaseFilter = "" | "any" | "3" | "6" | "12" | "ended";
const PAGE_SIZE = 20;
const filterTabs: Array<{ value: FilterTab; label: string }> = [
  { value: "all", label: "All" },
  { value: "new", label: "New" },
  { value: "pending", label: "Pending" },
  { value: "paid", label: "Paid" },
  { value: "cancelled", label: "Canceled" },
];
const leaseFilterOptions: StatusOption[] = [
  { value: "", label: "All lease terms", description: "Deals with or without a lease", tone: "slate" },
  { value: "any", label: "Has a lease", description: "All deals with a lease term", tone: "blue" },
  { value: "3", label: "Ends within 3 months", description: "Active leases ending in 1–3 months", tone: "amber" },
  { value: "6", label: "Ends within 6 months", description: "Active leases ending in 1–6 months", tone: "amber" },
  { value: "12", label: "Ends within 12 months", description: "Active leases ending in 1–12 months", tone: "blue" },
  { value: "ended", label: "Lease ended", description: "Leases that have reached their end", tone: "slate" },
];

const statusBorder: Record<string, string> = {
  paid: "border-l-green-500",
  pending: "border-l-amber-500",
};

function belongsToTab(deal: Deal, tab: FilterTab) {
  if (tab === "cancelled") return deal.deal_status === "cancelled";
  if (deal.deal_status === "cancelled") return false;
  const hasVIN = !!deal.vin?.trim();
  if (tab === "new") return !hasVIN;
  if (!hasVIN) return false;
  return tab === "all" || deal.deal_status === tab;
}

export default function DealsPage() {
  return (
    <Suspense
      fallback={
        <div className="p-6 text-sm text-slate-400">Loading deals…</div>
      }
    >
      <DealsWorkspace />
    </Suspense>
  );
}
function DealsWorkspace() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const createRequested = searchParams.get("new");
  const [listError, setListError] = useState("");
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
  const [tab, setTab] = useState<FilterTab>(() => {
    const value = searchParams.get("status");
    return filterTabs.some((tab) => tab.value === value)
      ? (value as FilterTab)
      : "all";
  });
  const [coverDeal, setCoverDeal] = useState<Deal | null>(null);
  const [agreementDeal, setAgreementDeal] = useState<Deal | null>(null);
  const [warningsOnly, setWarningsOnly] = useState(false);
  const [leaseFilter, setLeaseFilter] = useState<LeaseFilter>("");
  const [modal, setModal] = useState<{
    open: boolean;
    deal?: Deal;
    nextNumber?: string;
  }>({ open: false });
  const [referralMap, setReferralMap] = useState<Record<string, Referral[]>>(
    {},
  );
  const [counts, setCounts] = useState({
    all: 0,
    new: 0,
    pending: 0,
    paid: 0,
    cancelled: 0,
  });
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [refreshVersion, setRefreshVersion] = useState(0);

  useEffect(() => {
    if (createRequested !== "1") return;
    let cancelled = false;
    deals
      .nextNumber()
      .catch(() => "")
      .then((nextNumber) => {
        if (!cancelled) setModal({ open: true, nextNumber });
      });
    return () => {
      cancelled = true;
    };
  }, [createRequested]);
  function closeModal() {
    setModal({ open: false });
    if (createRequested) {
      const params = new URLSearchParams(searchParams.toString());
      params.delete("new");
      router.replace(`/deals${params.size ? `?${params}` : ""}`, {
        scroll: false,
      });
    }
  }
  useEffect(() => {
    Promise.all([clients.list(), brokers.list(), dealers.list()])
      .then(([c, b, dl]) => {
        setAllClients(c ?? []);
        setClientMap(Object.fromEntries((c ?? []).map((x) => [x.id, x])));
        setAllBrokers(b ?? []);
        setAllDealers(dl ?? []);
        setBrokerMap(
          Object.fromEntries(
            (b ?? []).map((x) => [x.id, `${x.first_name} ${x.last_name}`]),
          ),
        );
        setDealerMap(Object.fromEntries((dl ?? []).map((x) => [x.id, x.name])));
      })
      .catch(() =>
        setListError(
          "Couldn’t load clients and partners. Please reload the page.",
        ),
      );
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
    deals
      .listPaged({
        limit: PAGE_SIZE,
        offset,
        status:
          tab === "pending" || tab === "paid" || tab === "cancelled"
            ? tab
            : undefined,
        search: querySearch,
        warnings: warningsOnly ? "any" : undefined,
        vin:
          tab === "new"
            ? "missing"
            : tab === "cancelled"
              ? undefined
              : "present",
        leaseFilter:
          leaseFilter === "any" || leaseFilter === "ended"
            ? leaseFilter
            : undefined,
        leaseWithin:
          leaseFilter === "3" || leaseFilter === "6" || leaseFilter === "12"
            ? Number(leaseFilter)
            : undefined,
      })
      .then(async (res) => {
        const refs = await referralsApi.listByDeals(
          res.data.map((deal) => deal.id),
        );
        if (cancelled) return;
        setDealList((prev) =>
          offset === 0 ? res.data : [...prev, ...res.data],
        );
        setReferralMap((prev) => {
          const pageMap = groupReferralsByDeal(refs ?? []);
          return offset === 0 ? pageMap : { ...prev, ...pageMap };
        });
        setHasMore(res.has_more);
        if (res.counts) setCounts(res.counts);
      })
      .catch(() => {
        if (!cancelled)
          setListError("Couldn’t load your deals. Please reload the page.");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [tab, querySearch, offset, refreshVersion, leaseFilter, warningsOnly]);

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

  async function handleCommissionMethod(
    deal: Deal,
    status: "paid" | "pending",
  ) {
    const updated = await deals.updateBrokerCommissionStatus(deal.id, status);
    setDealList((prev) => prev.map((d) => (d.id === updated.id ? updated : d)));
  }

  async function handleReferralStatus(ref: Referral, status: ReferralStatus) {
    const updated = await referralsApi.updateStatus(ref.id, status);
    setReferralMap((prev) => {
      if (!ref.deal_id) return prev;
      return {
        ...prev,
        [ref.deal_id]: (prev[ref.deal_id] ?? []).map((r) =>
          r.id === ref.id ? { ...r, ...updated } : r,
        ),
      };
    });
  }

  return (
    <div className="p-6">
      <PageHeading eyebrow="Deal management" title="Your deals" description={`${counts.all} assigned deals. Every vehicle, payment, and partner in one place.`} actions={
        <button
          onClick={async () => {
            const num = await deals.nextNumber().catch(() => "");
            setModal({ open: true, nextNumber: num });
          }}
          className="crm-primary-button"
        >
          + New Deal
        </button>
      } />

      {/* Filter tabs + search */}
      <div className="crm-filter-toolbar flex flex-col items-stretch gap-3 mb-5 xl:flex-row xl:items-center">
        <div className="crm-segmented-group crm-deal-segments" role="group" aria-label="Filter deals by status">
          {filterTabs.map(({ value, label }) => (
            <button
              key={value}
              aria-pressed={tab === value}
              onClick={() => {
                if (value === tab) return;
                setLoading(true);
                setTab(value);
                setOffset(0);
              }}
              className="crm-filter-text crm-segment-button"
            >
              {label}{" "}
              <span
                className="crm-filter-count"
              >
                {counts[value]}
              </span>
            </button>
          ))}
        </div>
        <div className="relative min-w-0 flex-1">
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.35-4.35" />
          </svg>
          <input
            type="text"
            aria-label="Search deals"
            placeholder="Search client, vehicle, VIN, or deal number…"
            value={search}
            onChange={(e) => {
              setLoading(true);
              setSearch(e.target.value);
            }}
            className="crm-filter-text w-full bg-[#0b0c0d] border border-[#252628] rounded-lg pl-9 pr-3 py-2 text-sm text-white placeholder:text-slate-700 focus:outline-none"
          />
        </div>
        <StatusSelect label="Filter deals by warnings" value={warningsOnly ? "any" : ""}
          options={[
            { value: "", label: "All", description: "Show deals with or without warnings", tone: "slate" },
            { value: "any", label: "Has warnings", description: "VIN format or check digit needs attention", tone: "red" },
          ]} variant="filter" onChange={(value) => {
            setLoading(true);
            setWarningsOnly(value === "any");
            setOffset(0);
          }} />
        <StatusSelect label="Filter deals by lease term" value={leaseFilter} options={leaseFilterOptions} variant="filter"
          onChange={(value) => {
            setLoading(true);
            setLeaseFilter(value as LeaseFilter);
            setOffset(0);
          }} />
      </div>

      {listError && (
        <div
          role="alert"
          className="mb-5 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-red-500/25 bg-red-500/5 p-4 text-sm text-red-400"
        >
          <p>{listError}</p>
          <button
            onClick={() => window.location.reload()}
            className="font-medium underline"
          >
            Reload page
          </button>
        </div>
      )}
      {loading ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      ) : dealList.length === 0 ? (
        <div className="crm-empty-state">
          <span className="crm-empty-symbol" aria-hidden="true">◇</span>
          <h2>{search || tab !== "all" || leaseFilter !== "" || warningsOnly ? "No deals match these filters" : "Your next deal starts here"}</h2>
          <p>{search || tab !== "all" || leaseFilter !== "" || warningsOnly ? "Try a different search or reset your filters to see more deals." : "Create a deal to keep its vehicle, payments, and partners together."}</p>
          <button className="crm-primary-button" onClick={async () => {
            if (search || tab !== "all" || leaseFilter !== "" || warningsOnly) { setLoading(true); setSearch(""); setTab("all"); setLeaseFilter(""); setWarningsOnly(false); setOffset(0); }
            else { const num = await deals.nextNumber().catch(() => ""); setModal({ open: true, nextNumber: num }); }
          }}>{search || tab !== "all" || leaseFilter !== "" || warningsOnly ? "Reset filters" : "+ Create your first deal"}</button>
        </div>
      ) : (
        <div className="space-y-3">
          {dealList.map((d) => {
            const client = clientMap[d.client_id];
            const vehicle = [d.make, d.model].filter(Boolean).join(" ");
            const vinProblem = vinError(d.vin ?? "");
            const borderColor =
              statusBorder[d.deal_status ?? ""] ?? "border-l-[#252628]";
            const dealReferrals = referralMap[d.id] ?? [];
            const referralTotal = dealReferrals.reduce(
              (sum, r) => sum + (r.amount ?? 0),
              0,
            );
            const base =
              (d.broker_fee ?? 0) + (d.collect_amount ?? 0) - (d.doc_fee ?? 0);
            const commission =
              d.broker_commission ||
              (d.broker_percent != null && base > 0
                ? (base * d.broker_percent) / 100
                : 0);
            const displayPercent =
              d.broker_percent ||
              (commission > 0 && base > 0 ? (commission / base) * 100 : null);
            const computedNet = base - commission - referralTotal;

            return (
              <div
                key={d.id}
                className={`crm-deal-card bg-[#0b0c0d] border border-[#252628] border-l-4 ${borderColor} rounded-xl hover:border-[#333436] hover:border-l-4 transition-colors`}
                data-vin-warning={vinProblem ? "true" : undefined}
                aria-describedby={vinProblem ? `vin-warning-${d.id}` : undefined}
              >
                {vinProblem && <span id={`vin-warning-${d.id}`} className="sr-only">VIN needs review. {vinProblem}</span>}
                {/* Vehicle details shrink independently of the fixed action layout. */}
                <div className="crm-deal-header px-4 py-3.5 sm:px-5 border-b border-[#1e1f20]">
                  <div className="crm-deal-identity">
                    <div className="crm-deal-client-line">
                      {client ? (
                        <Link href={`/clients/${d.client_id}`} className="crm-deal-client-name text-white font-semibold hover:text-blue-400 transition-colors">
                          {client.first_name} {client.last_name}
                        </Link>
                      ) : <span className="text-slate-600 font-semibold">Unknown</span>}
                      {d.deal_number && <span className="shrink-0 text-xs font-mono bg-[#111213] border border-[#252628] text-slate-400 px-2 py-0.5 rounded-md">#{d.deal_number}</span>}
                      {vinProblem && (
                        <button type="button" onClick={() => setModal({ open: true, deal: d })}
                          aria-label="VIN warning. Edit this deal to review the VIN."
                          title={vinProblem}
                          className="inline-flex shrink-0 items-center gap-1.5 rounded-md border border-amber-500/25 bg-amber-500/10 px-2 py-0.5 text-xs font-medium text-amber-400 transition-colors hover:bg-amber-500/20">
                          <svg aria-hidden="true" width="12" height="12" viewBox="0 0 16 16" fill="none">
                            <path d="M8 2 15 14H1L8 2Z" stroke="currentColor" strokeLinejoin="round" />
                            <path d="M8 6v4M8 12v.1" stroke="currentColor" strokeLinecap="round" />
                          </svg>
                          VIN warning
                        </button>
                      )}
                      {d.is_cash_deal && <span className="shrink-0 text-xs font-semibold bg-amber-500/10 border border-amber-500/25 text-amber-400 px-2 py-0.5 rounded-md">CASH</span>}
                      {d.deal_date && <span className="shrink-0 text-slate-600 text-xs">{fmtDate(d.deal_date)}</span>}
                    </div>
                    {(vehicle || d.trim_level || d.vin) && <div className="crm-deal-vehicle">
                      {vehicle && <span className="crm-deal-vehicle-name text-slate-300 text-sm">{vehicle}</span>}
                      {d.trim_level && <span className="crm-deal-trim text-slate-500 text-xs" title={d.trim_level}>{d.trim_level}</span>}
                      {d.vin && <span className="crm-deal-vin text-slate-600 text-xs font-mono">{d.vin}</span>}
                    </div>}
                  </div>
                  <div className="crm-deal-controls">
                    <div className="crm-deal-status-control">
                      <DealStatusSelect value={d.deal_status} days={d.pending_payment_days} bucket={d.pending_payment_bucket}
                        onChange={(status) => handleStatusChange(d, status)} />
                    </div>
                    <div className="crm-deal-action-links">
                      <button type="button" onClick={() => setCoverDeal(d)} className="text-sm font-medium text-slate-400 transition-colors hover:text-blue-400">Deal cover</button>
                      <button type="button" onClick={() => setAgreementDeal(d)} className="text-sm font-medium text-slate-400 transition-colors hover:text-blue-400">Broker agreement</button>
                      <button type="button" onClick={() => setModal({ open: true, deal: d })} className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors">Edit</button>
                    </div>
                  </div>
                </div>

                {/* Stats wrap as space runs out; commission reserves room for its control. */}
                <div className="crm-deal-stats border-t border-[#1e1f20]">
                  <StatCell
                    label="Broker"
                    value={d.broker_id ? brokerMap[d.broker_id] : undefined}
                  />
                  <StatCell
                    label="Dealer"
                    value={d.dealer_id ? dealerMap[d.dealer_id] : undefined}
                  />
                  <StatCell
                    label="Broker Fee"
                    value={money(d.broker_fee)}
                    highlight="green"
                  />
                  <StatCell
                    label="Net"
                    value={money(computedNet)}
                    highlight="blue"
                  />
                  <StatCell label="Collect" value={money(d.collect_amount)} />
                  <StatCell label="Doc Fee" value={money(d.doc_fee)} />
                  <StatCell
                    label="Broker %"
                    value={
                      displayPercent
                        ? `${Number(displayPercent).toFixed(2)}%`
                        : undefined
                    }
                  />
                  {d.lease_term_months != null && (
                    <StatCell
                      label="Lease Left"
                      value={
                        d.lease_months_remaining != null &&
                        d.lease_months_remaining > 0
                          ? `${d.lease_months_remaining} mo`
                          : "Ended"
                      }
                    />
                  )}
                  <div className="crm-deal-commission px-4 py-2.5 flex flex-col gap-0.5 border-b border-r border-[#1e1f20]">
                    <span className="text-xs text-slate-600 uppercase tracking-wider">
                      Commission
                    </span>
                    <div className="flex min-w-0 flex-wrap items-center gap-2">
                      <span className="min-w-0 text-sm text-slate-300 tabular-nums [overflow-wrap:anywhere]">
                        {money(commission) ?? "—"}
                      </span>
                      <StatusSelect label="Commission payment status" value={d.broker_commission_status ?? "pending"} options={paymentOptions}
                        onChange={(status) => handleCommissionMethod(d, status as "paid" | "pending")} />
                    </div>
                  </div>
                </div>
                {/* Referrals — stacked on mobile, horizontal on desktop */}
                {dealReferrals.length > 0 && (
                  <div className="border-t border-[#1e1f20] flex flex-col md:flex-row md:divide-x md:divide-[#1e1f20]">
                    {dealReferrals.map((ref, idx) => (
                      <div
                        key={ref.id}
                        className={`px-4 py-2.5 flex items-center justify-between gap-3 md:flex-col md:items-start md:justify-start md:gap-0.5 md:flex-1 md:min-w-[160px]${idx < dealReferrals.length - 1 ? " border-b border-[#1e1f20] md:border-b-0" : ""}`}
                      >
                        <div className="flex flex-col gap-0.5 min-w-0">
                          <span className="text-xs text-slate-600 uppercase tracking-wider">
                            Referral
                            {ref.amount != null
                              ? ` · $${Number(ref.amount).toLocaleString()}`
                              : ""}
                          </span>
                          <span className="text-sm text-slate-300 truncate">
                            {[ref.first_name, ref.last_name]
                              .filter(Boolean)
                              .join(" ")}
                          </span>
                        </div>
                        <StatusSelect label="Referral status" value={ref.status} options={referralOptions}
                          onChange={(status) => handleReferralStatus(ref, status as ReferralStatus)} />
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

      {coverDeal && <DealCoverModal deal={coverDeal} client={clientMap[coverDeal.client_id]} onClose={() => setCoverDeal(null)} />}
      {agreementDeal && <BrokerAgreementModal deal={agreementDeal} client={clientMap[agreementDeal.client_id]} onClose={() => setAgreementDeal(null)} />}
      {modal.open && (
        <DealModal
          deal={modal.deal}
          nextDealNumber={modal.nextNumber}
          clients={allClients}
          brokers={allBrokers}
          dealers={allDealers}
          onSave={handleSave}
          onClose={closeModal}
        />
      )}
    </div>
  );
}

function money(value?: number) {
  return value != null
    ? `$${Number(value).toLocaleString(undefined, { maximumFractionDigits: 2 })}`
    : undefined;
}

function StatCell({
  label,
  value,
  highlight,
}: {
  label: string;
  value?: string;
  highlight?: "green" | "blue";
}) {
  const valCls =
    highlight === "green"
      ? "text-green-400 font-bold text-base"
      : highlight === "blue"
        ? "text-blue-400 font-bold text-base"
        : "text-slate-300 text-sm";
  return (
    <div className="crm-deal-stat px-4 py-2.5 flex flex-col gap-0.5 border-b border-r border-[#1e1f20]">
      <span className="text-xs text-slate-600 uppercase tracking-wider whitespace-nowrap">
        {label}
      </span>
      <span className={value ? valCls + " truncate" : "text-slate-600 text-sm"}>
        {value ?? "—"}
      </span>
    </div>
  );
}
