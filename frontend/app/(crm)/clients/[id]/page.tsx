"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { clients, deals, brokers, dealers } from "@/lib/api";
import type { Client, Deal, Broker, Dealer } from "@/lib/types";
import DealModal from "@/components/DealModal";
import DocumentsTab from "@/components/DocumentsTab";
import ClientForm from "@/components/ClientForm";
import DealStatusSelect from "@/components/DealStatusSelect";

type Tab = "profile" | "documents" | "deals";

export default function ClientDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [client, setClient] = useState<Client | null>(null);
  const [clientDeals, setClientDeals] = useState<Deal[]>([]);
  const [allBrokers, setAllBrokers] = useState<Broker[]>([]);
  const [allDealers, setAllDealers] = useState<Dealer[]>([]);
  const [tab, setTab] = useState<Tab>("profile");
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [dealModal, setDealModal] = useState<{ open: boolean; deal?: Deal }>({ open: false });
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    if (!id || id === "new") return;
    try {
      const [c, d, b, dl] = await Promise.all([
        clients.get(id),
        deals.listByClient(id),
        brokers.list(),
        dealers.list(),
      ]);
      setClient(c);
      setClientDeals(d);
      setAllBrokers(b);
      setAllDealers(dl);
    } catch {
      setError("Failed to load client");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    if (id === "new") {
      setLoading(false);
      setEditing(true);
    } else {
      load();
    }
  }, [id, load]);

  async function handleSaveClient(data: object) {
    try {
      if (id === "new") {
        const created = await clients.create(data);
        router.push(`/clients/${created.id}`);
      } else {
        const updated = await clients.update(id, data);
        setClient(updated);
        setEditing(false);
      }
    } catch (e: unknown) {
      throw e;
    }
  }

  async function handleSaveDeal(data: object) {
    try {
      if (dealModal.deal) {
        const updated = await deals.update(dealModal.deal.id, { ...data, client_id: id });
        setClientDeals((prev) => prev.map((d) => (d.id === updated.id ? updated : d)));
      } else {
        const created = await deals.create({ ...data, client_id: id });
        setClientDeals((prev) => [created, ...prev]);
      }
      setDealModal({ open: false });
    } catch (e: unknown) {
      throw e;
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-slate-600 text-sm">
        Loading…
      </div>
    );
  }

  if (error) {
    return <div className="p-6 text-red-400 text-sm">{error}</div>;
  }

  if (id === "new" || editing) {
    return (
      <div className="p-6 max-w-3xl mx-auto">
        <div className="flex items-center gap-3 mb-6">
          <button
            onClick={() => id === "new" ? router.push("/clients") : setEditing(false)}
            className="text-slate-500 hover:text-slate-300 text-sm transition-colors"
          >
            ← Back
          </button>
          <h1 className="text-xl font-semibold text-white">
            {id === "new" ? "New Client" : "Edit Client"}
          </h1>
        </div>
        <ClientForm
          initial={client ?? undefined}
          onSave={handleSaveClient}
          onCancel={() => id === "new" ? router.push("/clients") : setEditing(false)}
        />
      </div>
    );
  }

  if (!client) return null;

  return (
    <div className="p-6 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <Link href="/clients" className="text-slate-500 hover:text-slate-300 text-sm transition-colors">
            ← Clients
          </Link>
          <h1 className="text-xl font-semibold text-white mt-1">
            {client.first_name} {client.last_name}
          </h1>
          <p className="text-slate-500 text-sm mt-0.5">{client.email} · {client.phone}</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setEditing(true)}
            className="px-3 py-1.5 border border-[#1e1e45] rounded-lg text-sm text-slate-400 hover:text-white hover:border-slate-500 transition-colors"
          >
            Edit
          </button>
          <button
            onClick={() => setDealModal({ open: true })}
            className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition-all glow-blue-sm"
          >
            + New Deal
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-[#252628] mb-6">
        <nav className="flex gap-0">
          {(["profile", "documents", "deals"] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-all capitalize -mb-px ${
                tab === t
                  ? "border-blue-500 text-blue-400"
                  : "border-transparent text-slate-500 hover:text-slate-300"
              }`}
            >
              {t}
              {t === "deals" && clientDeals.length > 0 && (
                <span className="ml-1.5 text-xs bg-blue-500/15 text-blue-400 border border-blue-500/25 px-1.5 py-0.5 rounded-full">
                  {clientDeals.length}
                </span>
              )}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab content */}
      {tab === "profile" && <ProfileTab client={client} />}
      {tab === "documents" && <DocumentsTab clientId={id} />}
      {tab === "deals" && (
        <DealsTab
          deals={clientDeals}
          brokers={allBrokers}
          dealers={allDealers}
          onEdit={(deal) => setDealModal({ open: true, deal })}
          onStatusChange={(updated) =>
            setClientDeals((prev) => prev.map((d) => (d.id === updated.id ? updated : d)))
          }
        />
      )}

      {dealModal.open && (
        <DealModal
          deal={dealModal.deal}
          clientId={id}
          brokers={allBrokers}
          dealers={allDealers}
          onSave={handleSaveDeal}
          onClose={() => setDealModal({ open: false })}
          onStatusChange={(updated) =>
            setClientDeals((prev) => prev.map((d) => (d.id === updated.id ? updated : d)))
          }
        />
      )}
    </div>
  );
}

function ProfileTab({ client }: { client: Client }) {
  const sections = [
    {
      title: "Personal Information",
      fields: [
        { label: "Full Name", value: `${client.first_name} ${client.last_name}` },
        { label: "Email", value: client.email },
        { label: "Phone", value: client.phone },
        { label: "Date of Birth", value: client.date_of_birth },
        { label: "SSN", value: client.ssn ? "•••-••-" + client.ssn.slice(-4) : undefined },
        { label: "Source", value: client.source },
        { label: "Notes", value: client.notes },
      ],
    },
    {
      title: "Address",
      fields: [
        { label: "Address", value: client.address },
        { label: "City", value: client.city },
        { label: "State", value: client.state },
        { label: "ZIP", value: client.zip },
        { label: "Housing Status", value: client.housing_status },
        { label: "Housing Duration", value: client.housing_duration },
        { label: "Monthly Payment", value: client.housing_monthly_payment ? `$${client.housing_monthly_payment}` : undefined },
      ],
    },
    {
      title: "Employment",
      fields: [
        { label: "Employer", value: client.employer_name },
        { label: "Employer Address", value: client.employer_address },
        { label: "Employer City", value: client.employer_city },
        { label: "Employer State", value: client.employer_state },
        { label: "Employer ZIP", value: client.employer_zip },
        { label: "Occupation", value: client.occupation },
        { label: "Employer Phone", value: client.employer_phone },
        { label: "Employment Status", value: client.employment_status },
        { label: "Employment Duration", value: client.employment_duration },
        { label: "Gross Monthly Income", value: client.gross_monthly_income ? `$${client.gross_monthly_income}` : undefined },
        { label: "Other Monthly Income", value: client.other_monthly_income ? `$${client.other_monthly_income}` : undefined },
        { label: "Other Income Source", value: client.other_income_source },
      ],
    },
  ];

  return (
    <div className="space-y-4">
      {sections.map((section) => {
        const visible = section.fields.filter((f) => f.value);
        if (!visible.length) return null;
        return (
          <div key={section.title} className="bg-[#0b0c0d] rounded-xl border border-[#252628] p-5">
            <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-4">{section.title}</h3>
            <dl className="grid grid-cols-2 gap-x-6 gap-y-4">
              {visible.map(({ label, value }) => (
                <div key={label}>
                  <dt className="text-xs text-slate-600 font-medium uppercase tracking-wider">{label}</dt>
                  <dd className="text-sm text-slate-200 mt-0.5">{String(value)}</dd>
                </div>
              ))}
            </dl>
          </div>
        );
      })}
    </div>
  );
}

function DealsTab({
  deals: dealList,
  brokers: allBrokers,
  dealers: allDealers,
  onEdit,
  onStatusChange,
}: {
  deals: Deal[];
  brokers: Broker[];
  dealers: Dealer[];
  onEdit: (deal: Deal) => void;
  onStatusChange: (deal: Deal) => void;
}) {
  const brokerMap = Object.fromEntries(allBrokers.map((b) => [b.id, `${b.first_name} ${b.last_name}`]));
  const dealerMap = Object.fromEntries(allDealers.map((d) => [d.id, d.name]));

  async function toggleStatus(deal: Deal, field: "broker_fee_status" | "broker_commission_status") {
    try {
      let updated: Deal;
      if (field === "broker_fee_status") {
        const next = deal.broker_fee_status === "paid" ? "pending" : "paid";
        updated = await deals.updateBrokerFeeStatus(deal.id, next);
      } else {
        const next = deal.broker_commission_status === "paid" ? "pending" : "paid";
        updated = await deals.updateBrokerCommissionStatus(deal.id, next);
      }
      onStatusChange(updated);
    } catch {}
  }

  async function updateDealStatus(deal: Deal, status: Deal["deal_status"]) {
    if (!status) return;
    try {
      const updated = await deals.updateStatus(deal.id, status);
      onStatusChange(updated);
    } catch {}
  }

  if (!dealList.length) {
    return (
      <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] py-12 text-center text-slate-600 text-sm">
        No deals yet
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {dealList.map((deal) => (
        <div key={deal.id} className="bg-[#0b0c0d] rounded-xl border border-[#252628] p-5 hover:border-blue-500/20 transition-colors">
          <div className="flex items-start justify-between mb-4">
            <div>
              <p className="font-medium text-slate-200">
                {[deal.make, deal.model, deal.trim_level].filter(Boolean).join(" ") || "Deal"}
              </p>
              <p className="text-slate-600 text-xs mt-0.5">
                {deal.deal_number && `#${deal.deal_number} · `}
                {deal.deal_date && new Date(deal.deal_date).toLocaleDateString()}
                {deal.vin && ` · VIN: ${deal.vin}`}
              </p>
            </div>
            <button
              onClick={() => onEdit(deal)}
              className="text-sm text-slate-500 hover:text-blue-400 font-medium transition-colors"
            >
              Edit
            </button>
          </div>

          <div className="grid grid-cols-3 gap-4 text-sm mb-4">
            {deal.broker_id && (
              <div>
                <span className="text-slate-600 text-xs uppercase tracking-wider">Broker</span>
                <p className="text-slate-300 mt-0.5">{brokerMap[deal.broker_id] ?? "–"}</p>
              </div>
            )}
            {deal.dealer_id && (
              <div>
                <span className="text-slate-600 text-xs uppercase tracking-wider">Dealer</span>
                <p className="text-slate-300 mt-0.5">{dealerMap[deal.dealer_id] ?? "–"}</p>
              </div>
            )}
            {deal.down_payment && (
              <div>
                <span className="text-slate-600 text-xs uppercase tracking-wider">Down Payment</span>
                <p className="text-slate-300 mt-0.5">${deal.down_payment} {deal.down_payment_method && `(${deal.down_payment_method})`}</p>
              </div>
            )}
            {deal.monthly_payment && (
              <div>
                <span className="text-slate-600 text-xs uppercase tracking-wider">Monthly</span>
                <p className="text-slate-300 mt-0.5">${deal.monthly_payment}</p>
              </div>
            )}
            {deal.collect_amount != null && (
              <div>
                <span className="text-slate-600 text-xs uppercase tracking-wider">Collect</span>
                <p className="font-semibold text-green-400 mt-0.5">${deal.collect_amount}</p>
              </div>
            )}
            {deal.doc_fee != null && (
              <div>
                <span className="text-slate-600 text-xs uppercase tracking-wider">Doc Fee</span>
                <p className="text-slate-300 mt-0.5">${deal.doc_fee}</p>
              </div>
            )}
            <div>
              <span className="text-slate-600 text-xs uppercase tracking-wider">Deal Status</span>
              <div className="mt-1">
                <DealStatusSelect
                  value={deal.deal_status}
                  days={deal.pending_payment_days}
                  bucket={deal.pending_payment_bucket}
                  onChange={(status) => updateDealStatus(deal, status)}
                />
              </div>
            </div>
          </div>

          <div className="flex gap-2 flex-wrap">
            <StatusButton
              label="Broker Fee"
              amount={deal.broker_fee}
              status={deal.broker_fee_status}
              paidDate={deal.broker_fee_paid_date}
              onClick={() => toggleStatus(deal, "broker_fee_status")}
            />
            <StatusButton
              label="Commission"
              amount={deal.broker_commission}
              status={deal.broker_commission_status}
              paidDate={deal.broker_commission_paid_date}
              onClick={() => toggleStatus(deal, "broker_commission_status")}
            />
          </div>
          {deal.pending_payment_days != null && (
            <p className="text-xs text-slate-500 mt-3">
              Pending {deal.pending_payment_days} days · {deal.pending_payment_bucket}
            </p>
          )}
        </div>
      ))}
    </div>
  );
}

function StatusButton({
  label, amount, status, paidDate, onClick,
}: {
  label: string; amount?: number; status?: string; paidDate?: string; onClick: () => void;
}) {
  const paid = status === "paid";
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium border transition-all ${
        paid
          ? "bg-green-500/10 border-green-500/25 text-green-400 hover:bg-green-500/15 glow-green-sm"
          : "bg-[#0a0a1e] border-[#1e1e45] text-slate-500 hover:text-slate-300 hover:border-slate-600"
      }`}
    >
      <span>{paid ? "✓" : "○"}</span>
      <span>{label}{amount ? ` $${amount}` : ""}</span>
      {paid && paidDate && (
        <span className="text-green-600 opacity-80">{new Date(paidDate).toLocaleDateString()}</span>
      )}
    </button>
  );
}
