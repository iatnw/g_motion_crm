"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { brokers, clients, dealers as dealersApi, deals as dealsApi, referrals as referralsApi } from "@/lib/api";
import type { Broker, Client, Deal, Dealer, Referral } from "@/lib/types";
import { fmtDate, groupReferralsByDeal } from "@/lib/utils";
import DealModal from "@/components/DealModal";

const PAGE_SIZE = 20;

const statusBorder: Record<string, string> = {
  paid: "border-l-green-500",
  pending: "border-l-amber-500",
  cancelled: "border-l-red-500",
};

export default function DealerDealsPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const [dealer, setDealer] = useState<Dealer | null>(null);
  const [deals, setDeals] = useState<Deal[]>([]);
  const [clientMap, setClientMap] = useState<Record<string, Client>>({});
  const [brokerMap, setBrokerMap] = useState<Record<string, string>>({});
  const [referralMap, setReferralMap] = useState<Record<string, Referral[]>>({});
  const [allClients, setAllClients] = useState<Client[]>([]);
  const [allBrokers, setAllBrokers] = useState<Broker[]>([]);
  const [allDealers, setAllDealers] = useState<Dealer[]>([]);
  const [loading, setLoading] = useState(true);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [modal, setModal] = useState<{ open: boolean; deal?: Deal }>({ open: false });

  useEffect(() => {
    Promise.all([dealersApi.get(id), clients.list(), brokers.list(), dealersApi.list(), referralsApi.list()])
      .then(([d, cs, bs, ds, refs]) => {
        setDealer(d);
        setAllClients(cs ?? []);
        setAllBrokers(bs ?? []);
        setAllDealers(ds ?? []);
        setClientMap(Object.fromEntries((cs ?? []).map((c) => [c.id, c])));
        setBrokerMap(Object.fromEntries((bs ?? []).map((b) => [b.id, `${b.first_name} ${b.last_name}`])));
        setReferralMap(groupReferralsByDeal(refs ?? []));
      });
  }, [id]);

  useEffect(() => {
    dealsApi.listByDealerPaged(id, PAGE_SIZE, offset).then((res) => {
      setDeals((prev) => offset === 0 ? res.data : [...prev, ...res.data]);
      setHasMore(res.has_more);
    }).finally(() => setLoading(false));
  }, [id, offset]);

  async function handleSave(data: object): Promise<Deal> {
    if (!modal.deal) throw new Error("A dealer deal must already exist");
    const updated = await dealsApi.update(modal.deal.id, data);
    setDeals((prev) => prev.map((d) => d.id === updated.id ? updated : d));
    return updated;
  }

  const money = (value?: number) => value != null
    ? `$${Number(value).toLocaleString(undefined, { maximumFractionDigits: 2 })}`
    : "—";

  return (
    <div className="p-6">
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => router.back()} className="text-slate-500 hover:text-slate-300 transition-colors text-sm">
          ← Back
        </button>
        <div>
          <h1 className="text-xl font-semibold text-white">{dealer?.name ?? "…"}</h1>
          <p className="text-slate-500 text-xs mt-0.5">All deals</p>
        </div>
      </div>

      {loading && deals.length === 0 ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      ) : deals.length === 0 ? (
        <div className="text-center py-12 text-slate-600 text-sm">No deals for this dealer</div>
      ) : (
        <div className="space-y-3">
          {deals.map((deal) => {
            const client = clientMap[deal.client_id];
            const vehicle = [deal.make, deal.model].filter(Boolean).join(" ");
            const referrals = referralMap[deal.id] ?? [];
            const referralTotal = referrals.reduce((sum, referral) => sum + (referral.amount ?? 0), 0);
            const base = (deal.broker_fee ?? 0) + (deal.collect_amount ?? 0) - (deal.doc_fee ?? 0);
            const commission = deal.broker_commission || (deal.broker_percent != null && base > 0 ? base * deal.broker_percent / 100 : 0);
            const net = base - commission - referralTotal;
            return (
              <div key={deal.id} className={`bg-[#0b0c0d] border border-[#252628] border-l-4 ${statusBorder[deal.deal_status ?? ""] ?? "border-l-[#252628]"} rounded-xl`}>
                <div className="flex items-center justify-between px-5 py-3.5 border-b border-[#1e1f20]">
                  <div className="flex items-center gap-4 min-w-0 flex-wrap">
                    <span className="text-white font-semibold">{client ? `${client.first_name} ${client.last_name}` : "Unknown"}</span>
                    {deal.deal_number && <span className="text-xs font-mono bg-[#111213] border border-[#252628] text-slate-400 px-2 py-0.5 rounded-md">#{deal.deal_number}</span>}
                    {vehicle && <span className="text-slate-300 text-sm">{vehicle}</span>}
                  </div>
                  <div className="flex items-center gap-4 flex-shrink-0 ml-4">
                    {deal.deal_date && <span className="text-slate-600 text-xs hidden sm:inline">{fmtDate(deal.deal_date)}</span>}
                    <StatusBadge status={deal.deal_status} />
                    <button onClick={() => setModal({ open: true, deal })} className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors">Edit</button>
                  </div>
                </div>
                <div className="flex divide-x divide-[#1e1f20]">
                  <Cell label="Broker" value={deal.broker_id ? brokerMap[deal.broker_id] : undefined} />
                  <Cell label="Broker Fee" value={money(deal.broker_fee)} green />
                  <Cell label="Net" value={money(net)} blue />
                  <Cell label="Doc Fee" value={money(deal.doc_fee)} />
                  <Cell label="Broker %" value={deal.broker_percent != null ? `${deal.broker_percent}%` : undefined} />
                  <Cell label="Commission" value={money(commission)} />
                </div>
              </div>
            );
          })}
          {hasMore && (
            <div className="flex justify-center pt-2">
              <button onClick={() => { setLoading(true); setOffset((value) => value + PAGE_SIZE); }} disabled={loading} className="px-6 py-2 bg-[#0b0c0d] border border-[#252628] rounded-lg text-sm text-slate-400 hover:text-white hover:border-slate-500 transition-colors disabled:opacity-50">
                {loading ? "Loading…" : "Load more"}
              </button>
            </div>
          )}
        </div>
      )}

      {modal.open && modal.deal && (
        <DealModal deal={modal.deal} clients={allClients} brokers={allBrokers} dealers={allDealers} onSave={handleSave} onClose={() => setModal({ open: false })} />
      )}
    </div>
  );
}

function StatusBadge({ status }: { status?: Deal["deal_status"] }) {
  const style = status === "paid"
    ? "bg-green-500/10 border-green-500/25 text-green-400"
    : status === "cancelled"
      ? "bg-red-500/10 border-red-500/25 text-red-400"
      : "bg-amber-500/10 border-amber-500/25 text-amber-400";
  return <span className={`text-xs font-medium px-2 py-1 rounded-md border ${style}`}>{status === "paid" ? "FUNDED" : status === "cancelled" ? "CANCELED" : "NOT FUNDED"}</span>;
}

function Cell({ label, value, green, blue }: { label: string; value?: string; green?: boolean; blue?: boolean }) {
  const cls = green ? "text-green-400 font-bold text-base" : blue ? "text-blue-400 font-bold text-base" : "text-slate-300 text-sm";
  return (
    <div className="flex-1 px-4 py-2.5 flex flex-col gap-0.5 min-w-0">
      <span className="text-xs text-slate-600 uppercase tracking-wider">{label}</span>
      <span className={value ? `${cls} truncate` : "text-slate-600 text-sm"}>{value ?? "—"}</span>
    </div>
  );
}
