"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { deals, clients, brokers, dealers } from "@/lib/api";
import type { Deal, Client, Broker, Dealer, DealShow } from "@/lib/types";
import DealModal from "@/components/DealModal";

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
  const [show, setShow] = useState<{ open: boolean; loading?: boolean; deal?: DealShow }>({ open: false });

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

  async function handleSave(data: object) {
    if (modal.deal) {
      const updated = await deals.update(modal.deal.id, data);
      setAllDeals((prev) => prev.map((d) => (d.id === updated.id ? updated : d)));
    } else {
      const created = await deals.create(data);
      setAllDeals((prev) => [created, ...prev]);
    }
    setModal({ open: false });
  }

  const openShow = useCallback(async (id: string) => {
    setShow({ open: true, loading: true });
    try {
      const data = await deals.getShow(id);
      setShow({ open: true, deal: data });
    } catch {
      setShow({ open: false });
      alert("Failed to load deal");
    }
  }, []);

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
    <div className="p-6 max-w-6xl mx-auto">
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
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Client</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Vehicle</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Broker</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Dealer</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Revenue</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Status</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Date</th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1e1f20]">
                {filtered.map((d) => {
                  const client = clientMap[d.client_id];
                  return (
                    <tr key={d.id} className="hover:bg-blue-500/5 transition-colors group">
                      <td className="px-4 py-3.5">
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
                      <td className="px-4 py-3.5 text-slate-300">
                        {[d.make, d.model].filter(Boolean).join(" ") || "–"}
                        {d.deal_number && (
                          <span className="text-slate-600 ml-1">#{d.deal_number}</span>
                        )}
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
                      <td className="px-4 py-3.5">
                        <StatusPill status={d.deal_status} bucket={d.pending_payment_bucket} days={d.pending_payment_days} />
                      </td>
                      <td className="px-4 py-3.5 text-slate-600">
                        {d.deal_date ? new Date(d.deal_date).toLocaleDateString() : "–"}
                      </td>
                      <td className="px-4 py-3.5 text-right">
                        <div className="flex justify-end gap-3">
                          <button
                            onClick={() => openShow(d.id)}
                            className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                          >
                            View
                          </button>
                          <button
                            onClick={() => setModal({ open: true, deal: d })}
                            className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                          >
                            Edit
                          </button>
                        </div>
                      </td>
                    </tr>
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

      {show.open && (
        <DealShowModal
          loading={!!show.loading}
          deal={show.deal}
          onClose={() => setShow({ open: false })}
        />
      )}
    </div>
  );
}

function money(value?: number) {
  return `$${Number(value ?? 0).toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
}

function StatusPill({ status, bucket, days }: { status?: string; bucket?: string; days?: number }) {
  const color =
    status === "paid"
      ? "bg-green-500/10 text-green-400 border-green-500/25"
      : status === "pending"
        ? "bg-yellow-500/10 text-yellow-400 border-yellow-500/25"
        : status === "cancelled"
          ? "bg-red-500/10 text-red-400 border-red-500/25"
          : "bg-[#0a0a1e] text-slate-500 border-[#1e1e45]";
  return (
    <span className={`inline-flex items-center px-2 py-1 rounded-md text-xs border ${color}`}>
      {status ?? "none"}{days != null ? ` · ${days}d ${bucket ?? ""}` : ""}
    </span>
  );
}

function DealShowModal({ loading, deal, onClose }: { loading: boolean; deal?: DealShow; onClose: () => void }) {
  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-[#0b0c0d] rounded-2xl w-full max-w-4xl max-h-[90vh] overflow-y-auto border border-[#252628] shadow-2xl">
        <div className="flex items-center justify-between px-6 py-4 border-b border-[#252628]">
          <div>
            <h2 className="text-base font-semibold text-white">DealShow</h2>
            {deal && <p className="text-xs text-slate-500 mt-0.5">{deal.client_name}</p>}
          </div>
          <button onClick={onClose} className="text-slate-600 hover:text-slate-300 text-xl transition-colors">✕</button>
        </div>
        {loading || !deal ? (
          <div className="py-16 text-center text-sm text-slate-600">Loading...</div>
        ) : (
          <div className="p-6 space-y-5">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <ShowField label="Data" value={deal.date} />
              <ShowField label="Name" value={deal.client_name} />
              <ShowField label="Make" value={deal.make} />
              <ShowField label="Model" value={deal.model} />
              <ShowField label="Trim" value={deal.trim} />
              <ShowField label="VIN" value={deal.vin} />
              <ShowField label="Dealer" value={deal.dealer_name} />
              <ShowField label="Broker" value={deal.broker_name} />
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <ShowField label="Broker fee" value={money(deal.broker_fee)} />
              <ShowField label="Collect" value={money(deal.collect_amount)} />
              <ShowField label="Doc fee" value={money(deal.doc_fee)} />
              <ShowField label="Referral money" value={money(deal.referral_total)} />
              <ShowField label="Broker percent" value={deal.broker_percent == null ? undefined : `${deal.broker_percent}%`} />
              <ShowField label="Broker profit" value={money(deal.broker_profit)} />
              <ShowField label="Net" value={money(deal.company_net)} highlight />
              <ShowField label="Payment type" value={labelize(deal.broker_payment_type)} />
            </div>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
              <ShowField label="Check status" value={deal.check_status} />
              <ShowField label="Broker paid/not paid" value={deal.broker_paid_status} />
              <ShowField label="Notes" value={deal.notes} />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function ShowField({ label, value, highlight }: { label: string; value?: string; highlight?: boolean }) {
  return (
    <div className="bg-[#08080f] border border-[#252628] rounded-lg p-3 min-h-[72px]">
      <p className="text-xs text-slate-600 uppercase tracking-wider">{label}</p>
      <p className={`text-sm mt-1 break-words ${highlight ? "text-green-400 font-semibold" : "text-slate-200"}`}>{value || "–"}</p>
    </div>
  );
}

function labelize(value?: string) {
  return value ? value.replaceAll("_", " ") : undefined;
}
