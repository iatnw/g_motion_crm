"use client";

import { Fragment, useEffect, useState } from "react";
import Link from "next/link";
import { deals, clients, brokers, dealers } from "@/lib/api";
import type { Deal, Client, Broker, Dealer, DealStatus } from "@/lib/types";
import DealModal from "@/components/DealModal";
import DealStatusSelect from "@/components/DealStatusSelect";

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
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Client</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Deal #</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Vehicle</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Broker</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Dealer</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Revenue</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Net</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Status</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Date</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {filtered.map((d) => {
                  const client = clientMap[d.client_id];
                  const details: { label: string; value: string }[] = [];
                  if (d.trim_level)      details.push({ label: "Trim", value: d.trim_level });
                  if (d.vin)             details.push({ label: "VIN", value: d.vin });
                  if (d.collect_amount != null) details.push({ label: "Collect", value: money(d.collect_amount)! });
                  if (d.doc_fee != null) details.push({ label: "Doc Fee", value: money(d.doc_fee)! });
                  if (d.broker_percent != null) details.push({ label: "Broker %", value: `${d.broker_percent}%` });
                  if (d.broker_commission != null) details.push({ label: "Broker Commission", value: money(d.broker_commission)! });
                  if (d.notes)           details.push({ label: "Notes", value: d.notes });

                  return (
                    <Fragment key={d.id}>
                      {/* Primary row */}
                      <tr className="hover:bg-blue-500/5 transition-colors border-t border-[#1e1f20] first:border-t-0">
                        <td className="px-4 pt-3 pb-1">
                          {client ? (
                            <Link href={`/clients/${d.client_id}`} className="text-blue-400 hover:text-blue-300 font-medium transition-colors">
                              {client.first_name} {client.last_name}
                            </Link>
                          ) : (
                            <span className="text-slate-600">Unknown</span>
                          )}
                        </td>
                        <td className="px-4 pt-3 pb-1 text-slate-400">
                          {d.deal_number ? `#${d.deal_number}` : "–"}
                        </td>
                        <td className="px-4 pt-3 pb-1 text-slate-300">
                          {[d.make, d.model].filter(Boolean).join(" ") || "–"}
                        </td>
                        <td className="px-4 pt-3 pb-1 text-slate-400">
                          {d.broker_id ? brokerMap[d.broker_id] ?? "–" : "–"}
                        </td>
                        <td className="px-4 pt-3 pb-1 text-slate-400">
                          {d.dealer_id ? dealerMap[d.dealer_id] ?? "–" : "–"}
                        </td>
                        <td className="px-4 pt-3 pb-1">
                          {d.broker_fee != null
                            ? <span className="font-semibold text-green-400">{money(d.broker_fee)}</span>
                            : <span className="text-slate-600">–</span>}
                        </td>
                        <td className="px-4 pt-3 pb-1">
                          {d.your_net != null
                            ? <span className="font-semibold text-blue-400">{money(d.your_net)}</span>
                            : <span className="text-slate-600">–</span>}
                        </td>
                        <td className="px-4 pt-3 pb-1">
                          <DealStatusSelect
                            value={d.deal_status}
                            days={d.pending_payment_days}
                            bucket={d.pending_payment_bucket}
                            onChange={(status) => handleStatusChange(d, status)}
                          />
                        </td>
                        <td className="px-4 pt-3 pb-1 text-slate-600">
                          {d.deal_date ? new Date(d.deal_date).toLocaleDateString() : "–"}
                        </td>
                        <td className="px-4 pt-3 pb-1 text-right">
                          <button
                            onClick={() => setModal({ open: true, deal: d })}
                            className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                          >
                            Edit
                          </button>
                        </td>
                      </tr>

                      {/* Details row */}
                      <tr className="hover:bg-blue-500/5 transition-colors">
                        <td colSpan={10} className="px-4 pb-3 pt-0">
                          {details.length > 0 ? (
                            <div className="flex flex-wrap gap-1.5 mt-1">
                              {details.map(({ label, value }) => (
                                <span key={label} className="inline-flex items-center gap-1 bg-[#111213] border border-[#252628] rounded-md px-2 py-0.5">
                                  <span className="text-slate-600 text-xs">{label}:</span>
                                  <span className="text-slate-300 text-xs font-medium">{value}</span>
                                </span>
                              ))}
                            </div>
                          ) : (
                            <div className="h-1" />
                          )}
                        </td>
                      </tr>
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
