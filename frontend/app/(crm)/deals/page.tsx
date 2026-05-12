"use client";

import { Fragment, useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { deals, clients, brokers, dealers, referrals } from "@/lib/api";
import type { Deal, Client, Broker, Dealer, DealShow, DealStatus, Referral } from "@/lib/types";
import DealModal from "@/components/DealModal";
import DealStatusSelect from "@/components/DealStatusSelect";

type ExpandedData = { deal: DealShow; refs: Referral[] };

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
  const [modal, setModal] = useState<{ open: boolean; deal?: Deal; nextNumber?: string }>({ open: false });
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [expandedData, setExpandedData] = useState<Record<string, ExpandedData>>({});
  const [expandedLoading, setExpandedLoading] = useState<string | null>(null);

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
      if (updated.deal_status === "cancelled") {
        return prev.filter((d) => d.id !== updated.id);
      }
      return prev.map((d) => (d.id === updated.id ? updated : d));
    });
  }

  const toggleExpand = useCallback(async (id: string) => {
    if (expandedId === id) {
      setExpandedId(null);
      return;
    }
    setExpandedId(id);
    if (expandedData[id]) return;
    setExpandedLoading(id);
    try {
      const [data, refs] = await Promise.all([deals.getShow(id), referrals.listByDeal(id)]);
      setExpandedData((prev) => ({ ...prev, [id]: { deal: data, refs } }));
    } catch {
      setExpandedId(null);
    } finally {
      setExpandedLoading(null);
    }
  }, [expandedId, expandedData]);

  const filtered = allDeals.filter((d) => {
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

      {/* Search */}
      <div className="relative mb-5">
        <svg className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
        </svg>
        <input
          type="text"
          placeholder="Search by client, car, VIN, deal #…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg pl-9 pr-3 py-2 text-sm text-white placeholder:text-slate-700 focus:outline-none transition-colors"
        />
      </div>

      {loading ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      ) : (
        <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] overflow-hidden">
          {filtered.length === 0 ? (
            <div className="text-center py-12 text-slate-600 text-sm">No deals found</div>
          ) : (
            <table className="w-full text-sm">
              <thead className="border-b border-[#252628]">
                <tr>
                  <th className="w-6 px-4 py-3" />
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Client</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Deal #</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Vehicle</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Broker</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Dealer</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Revenue</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Status</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Date</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1e1f20]">
                {filtered.map((d) => {
                  const client = clientMap[d.client_id];
                  const isExpanded = expandedId === d.id;
                  const isLoading = expandedLoading === d.id;
                  const data = expandedData[d.id];
                  return (
                    <Fragment key={d.id}>
                      <tr
                        className={`transition-colors group cursor-pointer ${isExpanded ? "bg-blue-500/5" : "hover:bg-blue-500/5"}`}
                        onClick={() => toggleExpand(d.id)}
                      >
                        <td className="px-4 py-3.5 text-slate-600">
                          <svg
                            width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"
                            className={`transition-transform duration-200 ${isExpanded ? "rotate-90" : ""}`}
                          >
                            <path d="m9 18 6-6-6-6" />
                          </svg>
                        </td>
                        <td className="px-4 py-3.5" onClick={(e) => e.stopPropagation()}>
                          {client ? (
                            <Link
                              href={`/clients/${d.client_id}`}
                              className="text-blue-400 hover:text-blue-300 font-medium transition-colors"
                            >
                              {client.first_name} {client.last_name}
                            </Link>
                          ) : (
                            <span className="text-slate-600">Unknown</span>
                          )}
                        </td>
                        <td className="px-4 py-3.5 text-slate-400">
                          {d.deal_number ? `#${d.deal_number}` : "–"}
                        </td>
                        <td className="px-4 py-3.5 text-slate-300">
                          {[d.make, d.model].filter(Boolean).join(" ") || "–"}
                        </td>
                        <td className="px-4 py-3.5 text-slate-400">
                          {d.broker_id ? brokerMap[d.broker_id] ?? "–" : "–"}
                        </td>
                        <td className="px-4 py-3.5 text-slate-400">
                          {d.dealer_id ? dealerMap[d.dealer_id] ?? "–" : "–"}
                        </td>
                        <td className="px-4 py-3.5">
                          {d.broker_fee != null ? (
                            <span className="font-semibold text-green-400">{money(d.broker_fee)}</span>
                          ) : (
                            <span className="text-slate-600">–</span>
                          )}
                        </td>
                        <td className="px-4 py-3.5" onClick={(e) => e.stopPropagation()}>
                          <DealStatusSelect
                            value={d.deal_status}
                            days={d.pending_payment_days}
                            bucket={d.pending_payment_bucket}
                            onChange={(status) => handleStatusChange(d, status)}
                          />
                        </td>
                        <td className="px-4 py-3.5 text-slate-600">
                          {d.deal_date ? new Date(d.deal_date).toLocaleDateString() : "–"}
                        </td>
                        <td className="px-4 py-3.5 text-right" onClick={(e) => e.stopPropagation()}>
                          <button
                            onClick={() => setModal({ open: true, deal: d })}
                            className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                          >
                            Edit
                          </button>
                        </td>
                      </tr>

                      {isExpanded && (
                        <tr className="bg-[#08080f]">
                          <td colSpan={10} className="px-6 py-5">
                            {isLoading || !data ? (
                              <div className="text-center py-4 text-slate-600 text-sm">Loading…</div>
                            ) : (
                              <div className="space-y-4">
                                {/* Financial grid */}
                                <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-2">
                                  <InfoCell label="Trim" value={data.deal.trim} />
                                  <InfoCell label="VIN" value={data.deal.vin} />
                                  <InfoCell label="Collect" value={money(data.deal.collect_amount)} />
                                  <InfoCell label="Doc Fee" value={money(data.deal.doc_fee)} />
                                  <InfoCell label="Referral Total" value={money(data.deal.referral_total)} />
                                  <InfoCell label="Broker %" value={data.deal.broker_percent != null ? `${data.deal.broker_percent}%` : undefined} />
                                  <InfoCell label="Broker Profit" value={money(data.deal.broker_profit)} />
                                  <InfoCell label="Net" value={money(data.deal.company_net)} highlight />
                                </div>
                                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                                  <InfoCell label="Payment Type" value={labelize(data.deal.broker_payment_type)} />
                                  <InfoCell label="Check Status" value={data.deal.check_status} />
                                  <InfoCell label="Broker Paid" value={data.deal.broker_paid_status} />
                                  <InfoCell label="Notes" value={data.deal.notes} />
                                </div>

                                {/* Referrals */}
                                {data.refs.length > 0 && (
                                  <div>
                                    <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">Referrals</p>
                                    <div className="flex flex-wrap gap-2">
                                      {data.refs.map((ref) => (
                                        <div key={ref.id} className="flex items-center gap-3 bg-[#111213] border border-[#252628] rounded-lg px-3 py-2">
                                          <div>
                                            <p className="text-sm text-slate-200">{ref.first_name} {ref.last_name}</p>
                                            {ref.phone && <p className="text-xs text-slate-600">{ref.phone}</p>}
                                          </div>
                                          {ref.amount != null && (
                                            <span className="text-sm font-semibold text-slate-300">${Number(ref.amount).toLocaleString()}</span>
                                          )}
                                          <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${
                                            ref.status === "paid" ? "bg-green-500/10 border-green-500/25 text-green-400"
                                            : ref.status === "pending" ? "bg-yellow-500/10 border-yellow-500/25 text-yellow-400"
                                            : "bg-[#0a0a1e] border-[#1e1e45] text-slate-500"
                                          }`}>
                                            {ref.status}
                                          </span>
                                        </div>
                                      ))}
                                    </div>
                                  </div>
                                )}
                              </div>
                            )}
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
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

function labelize(value?: string) {
  return value ? value.replaceAll("_", " ") : undefined;
}

function InfoCell({ label, value, highlight }: { label: string; value?: string; highlight?: boolean }) {
  return (
    <div className="bg-[#0b0c0d] border border-[#1e1f20] rounded-lg px-3 py-2.5">
      <p className="text-xs text-slate-600 uppercase tracking-wider">{label}</p>
      <p className={`text-sm mt-0.5 font-medium ${highlight ? "text-green-400" : "text-slate-300"}`}>{value || "–"}</p>
    </div>
  );
}
