"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { deals, clients, brokers, dealers } from "@/lib/api";
import type { Deal, Client, Broker, Dealer, DealStatus, BrokerPaymentMethod } from "@/lib/types";
import DealModal from "@/components/DealModal";
import DealStatusSelect from "@/components/DealStatusSelect";

type FilterTab = "all" | "pending" | "paid";

const statusBorder: Record<string, string> = {
  paid:    "border-l-green-500",
  pending: "border-l-amber-500",
};

export default function DealsPage() {
  const [allDeals, setAllDeals] = useState<Deal[]>([]);
  const [allClients, setAllClients] = useState<Client[]>([]);
  const [clientMap, setClientMap] = useState<Record<string, Client>>({});
  const [allBrokers, setAllBrokers] = useState<Broker[]>([]);
  const [allDealers, setAllDealers] = useState<Dealer[]>([]);
  const [brokerMap, setBrokerMap] = useState<Record<string, string>>({});
  const [dealerMap, setDealerMap] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [tab, setTab] = useState<FilterTab>("all");
  const [modal, setModal] = useState<{ open: boolean; deal?: Deal; nextNumber?: string }>({ open: false });

  useEffect(() => {
    Promise.all([deals.list(), clients.list(), brokers.list(), dealers.list()])
      .then(([d, c, b, dl]) => {
        setAllDeals(d ?? []);
        setAllClients(c ?? []);
        setClientMap(Object.fromEntries((c ?? []).map((x) => [x.id, x])));
        setAllBrokers(b ?? []);
        setAllDealers(dl ?? []);
        setBrokerMap(Object.fromEntries((b ?? []).map((x) => [x.id, `${x.first_name} ${x.last_name}`])));
        setDealerMap(Object.fromEntries((dl ?? []).map((x) => [x.id, x.name])));
      })
      .finally(() => setLoading(false));
  }, []);

  async function handleSave(data: object): Promise<Deal> {
    if (modal.deal) {
      const updated = await deals.update(modal.deal.id, data);
      setAllDeals((prev) => prev.map((d) => (d.id === updated.id ? updated : d)));
      return updated;
    } else {
      const created = await deals.create(data);
      setAllDeals((prev) => [created, ...prev]);
      return created;
    }
  }

  async function handleStatusChange(deal: Deal, status: DealStatus) {
    const updated = await deals.updateStatus(deal.id, status);
    setAllDeals((prev) => {
      if (updated.deal_status === "cancelled") return prev.filter((d) => d.id !== updated.id);
      return prev.map((d) => (d.id === updated.id ? updated : d));
    });
  }

  async function handleCommissionMethod(deal: Deal, method: BrokerPaymentMethod | "") {
    const status = method ? "paid" : "pending";
    const updated = await deals.updateBrokerCommissionStatus(deal.id, status, method || undefined);
    setAllDeals((prev) => prev.map((d) => (d.id === updated.id ? updated : d)));
  }

  const counts = {
    all: allDeals.length,
    pending: allDeals.filter((d) => d.deal_status === "pending").length,
    paid: allDeals.filter((d) => d.deal_status === "paid").length,
  };

  const filtered = allDeals.filter((d) => {
    if (tab === "pending" && d.deal_status !== "pending") return false;
    if (tab === "paid" && d.deal_status !== "paid") return false;
    if (!search) return true;
    const q = search.toLowerCase();
    const client = clientMap[d.client_id];
    return (
      d.make?.toLowerCase().includes(q) ||
      d.model?.toLowerCase().includes(q) ||
      d.vin?.toLowerCase().includes(q) ||
      d.deal_number?.toLowerCase().includes(q) ||
      client?.first_name.toLowerCase().includes(q) ||
      client?.last_name.toLowerCase().includes(q)
    );
  });

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-white">Deals</h1>
          <p className="text-slate-500 text-xs mt-0.5">{allDeals.length} total</p>
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
          {(["all", "pending", "paid"] as FilterTab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors capitalize ${
                tab === t
                  ? "bg-[#1a1b1d] text-white"
                  : "text-slate-500 hover:text-slate-300"
              }`}
            >
              {t} <span className={`ml-1 ${tab === t ? "text-slate-400" : "text-slate-600"}`}>{counts[t]}</span>
            </button>
          ))}
        </div>
        <div className="relative flex-1">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
          </svg>
          <input
            type="text"
            placeholder="Search by client, car, VIN, deal #…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg pl-9 pr-3 py-2 text-sm text-white placeholder:text-slate-700 focus:outline-none"
          />
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      ) : filtered.length === 0 ? (
        <div className="text-center py-12 text-slate-600 text-sm">No deals found</div>
      ) : (
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-3">
          {filtered.map((d) => {
            const client = clientMap[d.client_id];
            const vehicle = [d.make, d.model].filter(Boolean).join(" ");
            const borderColor = statusBorder[d.deal_status ?? ""] ?? "border-l-[#252628]";

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
                    {/* Pending days badge */}
                    {d.deal_status === "pending" && d.pending_payment_days != null && d.pending_payment_days > 0 && (
                      <span className={`text-xs font-medium px-2 py-0.5 rounded-full border ${
                        d.pending_payment_days >= 14
                          ? "bg-red-500/10 border-red-500/25 text-red-400"
                          : "bg-amber-500/10 border-amber-500/25 text-amber-400"
                      }`}>
                        {d.pending_payment_days}d pending
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-4 flex-shrink-0 ml-4">
                    {d.deal_date && (
                      <span className="text-slate-600 text-xs hidden sm:inline">
                        {new Date(d.deal_date).toLocaleDateString()}
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

                {/* Stats bar */}
                <div className="px-5 py-3 flex flex-wrap items-end gap-x-6 gap-y-2">
                  <Stat label="Broker" value={d.broker_id ? brokerMap[d.broker_id] : undefined} />
                  <Stat label="Dealer" value={d.dealer_id ? dealerMap[d.dealer_id] : undefined} />
                  <StatDivider />
                  <BigStat label="Revenue" value={money(d.broker_fee)} color="green" />
                  <BigStat label="Net" value={money(d.your_net)} color="blue" />
                  <StatDivider />
                  <Stat label="Collect" value={money(d.collect_amount)} />
                  <Stat label="Doc Fee" value={money(d.doc_fee)} />
                  {d.broker_percent != null && <Stat label="Broker %" value={`${d.broker_percent}%`} />}
                  {d.broker_commission != null && (
                    <div className="flex flex-col gap-0.5">
                      <span className="text-xs text-slate-600 uppercase tracking-wider">Commission</span>
                      <div className="flex items-center gap-2">
                        <span className="text-sm text-slate-300">{money(d.broker_commission)}</span>
                        <select
                          value={d.broker_commission_status === "paid" ? (d.broker_payment_method ?? "paid_unknown") : ""}
                          onChange={(e) => handleCommissionMethod(d, e.target.value as BrokerPaymentMethod | "")}
                          onClick={(e) => e.stopPropagation()}
                          className={`text-xs border rounded-md px-1.5 py-0.5 focus:outline-none cursor-pointer ${
                            d.broker_commission_status === "paid"
                              ? "bg-green-500/10 border-green-500/25 text-green-400"
                              : "bg-amber-500/10 border-amber-500/25 text-amber-400"
                          }`}
                          style={{ backgroundColor: "transparent" }}
                        >
                          <option value="">Not Paid</option>
                          <option value="zelle_gmotion">Zelle GMotion</option>
                          <option value="zelle_personal">Zelle Personal</option>
                          <option value="cash">Cash</option>
                        </select>
                      </div>
                    </div>
                  )}
                  {d.notes && (
                    <>
                      <StatDivider />
                      <Stat label="Notes" value={d.notes} />
                    </>
                  )}
                </div>
              </div>
            );
          })}
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
          onStatusChange={(updated) =>
            setAllDeals((prev) => prev.map((d) => (d.id === updated.id ? updated : d)))
          }
        />
      )}
    </div>
  );
}

function money(value?: number) {
  return value != null ? `$${Number(value).toLocaleString(undefined, { maximumFractionDigits: 2 })}` : undefined;
}

function StatDivider() {
  return <div className="w-px self-stretch bg-[#1e1f20] mx-1" />;
}

function Stat({ label, value }: { label: string; value?: string }) {
  if (!value) return null;
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-slate-600 uppercase tracking-wider">{label}</span>
      <span className="text-sm text-slate-300">{value}</span>
    </div>
  );
}

function BigStat({ label, value, color }: { label: string; value?: string; color: "green" | "blue" }) {
  if (!value) return null;
  const cls = color === "green" ? "text-green-400" : "text-blue-400";
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-slate-600 uppercase tracking-wider">{label}</span>
      <span className={`text-base font-bold ${cls}`}>{value}</span>
    </div>
  );
}
