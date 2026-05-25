"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { brokers as brokersApi, deals as dealsApi, clients, dealers, referrals as referralsApi } from "@/lib/api";
import type { Broker, Deal, Client, Dealer, Referral } from "@/lib/types";
import { fmtDate, groupReferralsByDeal } from "@/lib/utils";
import DealModal from "@/components/DealModal";

const PAGE_SIZE = 20;

const statusBorder: Record<string, string> = {
  paid:    "border-l-green-500",
  pending: "border-l-amber-500",
};

export default function BrokerDealsPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [broker, setBroker] = useState<Broker | null>(null);
  const [deals, setDeals] = useState<Deal[]>([]);
  const [clientMap, setClientMap] = useState<Record<string, Client>>({});
  const [dealerMap, setDealerMap] = useState<Record<string, string>>({});
  const [referralMap, setReferralMap] = useState<Record<string, Referral[]>>({});
  const [allBrokers, setAllBrokers] = useState<Broker[]>([]);
  const [allClients, setAllClients] = useState<Client[]>([]);
  const [allDealers, setAllDealers] = useState<Dealer[]>([]);
  const [loading, setLoading] = useState(true);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [modal, setModal] = useState<{ open: boolean; deal?: Deal }>({ open: false });

  useEffect(() => {
    Promise.all([
      brokersApi.get(id),
      brokersApi.list(),
      clients.list(),
      dealers.list(),
      referralsApi.list(),
    ]).then(([b, bs, cs, ds, refs]) => {
      setBroker(b);
      setAllBrokers(bs ?? []);
      setAllClients(cs ?? []);
      setAllDealers(ds ?? []);
      setClientMap(Object.fromEntries((cs ?? []).map((c) => [c.id, c])));
      setDealerMap(Object.fromEntries((ds ?? []).map((d) => [d.id, d.name])));
      setReferralMap(groupReferralsByDeal(refs ?? []));
    });
  }, [id]);

  useEffect(() => {
    setLoading(true);
    dealsApi.listByBrokerPaged(id, PAGE_SIZE, offset).then((res) => {
      setDeals((prev) => offset === 0 ? res.data : [...prev, ...res.data]);
      setHasMore(res.has_more);
    }).finally(() => setLoading(false));
  }, [id, offset]);

  async function handleSave(data: object): Promise<Deal> {
    if (modal.deal) {
      const { deals: dealsApiInner } = await import("@/lib/api");
      const updated = await dealsApiInner.update(modal.deal.id, data);
      setDeals((prev) => prev.map((d) => d.id === updated.id ? updated : d));
      return updated;
    }
    const { deals: dealsApiInner } = await import("@/lib/api");
    const created = await dealsApiInner.create(data);
    setDeals((prev) => [created, ...prev]);
    return created;
  }

  const money = (v?: number) => v != null ? `$${Number(v).toLocaleString(undefined, { maximumFractionDigits: 2 })}` : "—";

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => router.back()} className="text-slate-500 hover:text-slate-300 transition-colors text-sm">
          ← Back
        </button>
        <div>
          <h1 className="text-xl font-semibold text-white">
            {broker ? `${broker.first_name} ${broker.last_name}`.trim() : "…"}
          </h1>
          <p className="text-slate-500 text-xs mt-0.5">All deals</p>
        </div>
      </div>

      {/* Deal list */}
      {loading && deals.length === 0 ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      ) : deals.length === 0 ? (
        <div className="text-center py-12 text-slate-600 text-sm">No deals for this broker</div>
      ) : (
        <div className="space-y-3">
          {deals.map((d) => {
            const client = clientMap[d.client_id];
            const vehicle = [d.make, d.model].filter(Boolean).join(" ");
            const borderColor = statusBorder[d.deal_status ?? ""] ?? "border-l-[#252628]";
            const dealReferrals = referralMap[d.id] ?? [];
            const referralTotal = dealReferrals.reduce((sum, r) => sum + (r.amount ?? 0), 0);
            const base = (d.broker_fee ?? 0) + (d.collect_amount ?? 0) - (d.doc_fee ?? 0);
            const commission = d.broker_commission || (d.broker_percent != null && base > 0 ? base * d.broker_percent / 100 : 0);
            const net = base - commission - referralTotal;

            return (
              <div key={d.id} className={`bg-[#0b0c0d] border border-[#252628] border-l-4 ${borderColor} rounded-xl`}>
                <div className="flex items-center justify-between px-5 py-3.5 border-b border-[#1e1f20]">
                  <div className="flex items-center gap-4 min-w-0 flex-wrap">
                    <span className="text-white font-semibold">{client ? `${client.first_name} ${client.last_name}` : "Unknown"}</span>
                    {d.deal_number && (
                      <span className="text-xs font-mono bg-[#111213] border border-[#252628] text-slate-400 px-2 py-0.5 rounded-md">#{d.deal_number}</span>
                    )}
                    {vehicle && <span className="text-slate-300 text-sm">{vehicle}</span>}
                  </div>
                  <div className="flex items-center gap-4 flex-shrink-0 ml-4">
                    {d.deal_date && <span className="text-slate-600 text-xs hidden sm:inline">{fmtDate(d.deal_date)}</span>}
                    <span className={`text-xs font-medium px-2 py-1 rounded-md border ${d.deal_status === "paid" ? "bg-green-500/10 border-green-500/25 text-green-400" : "bg-amber-500/10 border-amber-500/25 text-amber-400"}`}>
                      {d.deal_status === "paid" ? "FUNDED" : "NOT FUNDED"}
                    </span>
                    <button onClick={() => setModal({ open: true, deal: d })} className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors">Edit</button>
                  </div>
                </div>
                <div className="flex divide-x divide-[#1e1f20]">
                  <Cell label="Dealer" value={d.dealer_id ? dealerMap[d.dealer_id] : undefined} />
                  <Cell label="Broker Fee" value={money(d.broker_fee)} green />
                  <Cell label="Net" value={money(net)} blue />
                  <Cell label="Doc Fee" value={money(d.doc_fee)} />
                  <Cell label="Broker %" value={d.broker_percent != null ? `${d.broker_percent}%` : undefined} />
                  <Cell label="Commission" value={money(commission)} />
                </div>
              </div>
            );
          })}

          {/* Pagination */}
          {hasMore && (
            <div className="flex justify-center pt-2">
              <button
                onClick={() => setOffset((o) => o + PAGE_SIZE)}
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

function Cell({ label, value, green, blue }: { label: string; value?: string; green?: boolean; blue?: boolean }) {
  const cls = green ? "text-green-400 font-bold text-base"
    : blue ? "text-blue-400 font-bold text-base"
    : "text-slate-300 text-sm";
  return (
    <div className="flex-1 px-4 py-2.5 flex flex-col gap-0.5 min-w-0">
      <span className="text-xs text-slate-600 uppercase tracking-wider">{label}</span>
      <span className={value ? cls + " truncate" : "text-slate-600 text-sm"}>{value ?? "—"}</span>
    </div>
  );
}
