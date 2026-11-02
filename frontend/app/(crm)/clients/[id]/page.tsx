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
import { fmtDate } from "@/lib/utils";
import {
  ClientAvatar,
  ClientStatusBadge,
  clientSourceLabels,
} from "@/components/ClientIdentity";

type Tab = "profile" | "documents" | "deals";

export default function ClientDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [client, setClient] = useState<Client | null>(null);
  const [clientDeals, setClientDeals] = useState<Deal[]>([]);
  const [cosignerDeals, setCosignerDeals] = useState<Deal[]>([]);
  const [allClients, setAllClients] = useState<Client[]>([]);
  const [allBrokers, setAllBrokers] = useState<Broker[]>([]);
  const [allDealers, setAllDealers] = useState<Dealer[]>([]);
  const [tab, setTab] = useState<Tab>("profile");
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [dealModal, setDealModal] = useState<{ open: boolean; deal?: Deal }>({
    open: false,
  });
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    if (!id || id === "new") return;
    try {
      setError("");
      const [c, d, cosigned, allC, b, dl] = await Promise.all([
        clients.get(id),
        deals.listByClient(id),
        deals.listByCosigner(id),
        clients.list(),
        brokers.list(),
        dealers.list(),
      ]);
      setClient(c);
      setEditing(false);
      setClientDeals(d);
      setCosignerDeals(cosigned);
      setAllClients(allC);
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
        setEditing(false);
        router.replace(`/clients/${created.id}`);
      } else {
        const updated = await clients.update(id, data);
        setClient(updated);
        setEditing(false);
      }
    } catch (e: unknown) {
      throw e;
    }
  }

  async function handleSaveDeal(data: object): Promise<Deal> {
    if (dealModal.deal) {
      const updated = await deals.update(dealModal.deal.id, {
        ...data,
        client_id: id,
      });
      setClientDeals((prev) =>
        prev.map((d) => (d.id === updated.id ? updated : d)),
      );
      return updated;
    } else {
      const created = await deals.create({ ...data, client_id: id });
      setClientDeals((prev) => [created, ...prev]);
      return created;
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
      <div className="mx-auto max-w-7xl p-4 normal-case sm:p-6 lg:p-8">
        <div className="flex items-center gap-3 mb-6">
          <button
            onClick={() =>
              id === "new" ? router.push("/clients") : setEditing(false)
            }
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
          onCancel={() =>
            id === "new" ? router.push("/clients") : setEditing(false)
          }
        />
      </div>
    );
  }

  if (!client) return null;

  return (
    <div className="mx-auto max-w-7xl p-4 normal-case sm:p-6 lg:p-8">
      <Link
        href="/clients"
        className="mb-5 inline-flex items-center gap-2 text-sm text-slate-400 hover:text-blue-400"
      >
        ← All clients
      </Link>
      <div className="mb-6 flex flex-wrap items-start justify-between gap-5 rounded-xl border border-[#252628] bg-[#0f1011] p-5 sm:p-6">
        <div className="flex min-w-0 items-start gap-4">
          <ClientAvatar client={client} large />
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-3">
              <h1 className="text-2xl font-semibold tracking-tight text-white">
                {client.first_name} {client.last_name}
              </h1>
              <ClientStatusBadge status={client.status} />
            </div>
            <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm">
              {client.email && (
                <a
                  href={`mailto:${client.email}`}
                  className="break-all text-slate-400 hover:text-blue-400 hover:underline"
                >
                  {client.email}
                </a>
              )}
              {client.phone && (
                <a
                  href={`tel:${client.phone.replace(/[^+\d]/g, "")}`}
                  className="text-slate-400 hover:text-blue-400 hover:underline"
                >
                  {client.phone}
                </a>
              )}
              {!client.email && !client.phone && (
                <span className="text-slate-400">No contact details added</span>
              )}
            </div>
            <p className="mt-2 text-xs text-slate-400">
              Added {fmtDate(client.created_at)}
              {client.source ? ` · ${clientSourceLabels[client.source]}` : ""}
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setEditing(true)}
            className="rounded-lg border border-[#252628] px-4 py-2.5 text-sm font-medium text-slate-300 transition-colors hover:border-blue-500/40 hover:text-blue-400 focus-visible:outline-2 focus-visible:outline-blue-400"
          >
            Edit profile
          </button>
          <button
            onClick={() => setDealModal({ open: true })}
            className="rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-500 focus-visible:outline-2 focus-visible:outline-blue-400"
          >
            ＋ New deal
          </button>
        </div>
      </div>
      <div className="mb-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          { label: "Primary deals", value: clientDeals.length },
          { label: "Cosigned deals", value: cosignerDeals.length },
          {
            label: "Cash deals",
            value: clientDeals.filter((d) => d.is_cash_deal).length,
          },
          {
            label: "Leases",
            value: clientDeals.filter((d) => d.lease_term_months != null)
              .length,
          },
        ].map((item) => (
          <div
            key={item.label}
            className="rounded-xl border border-[#252628] bg-[#0f1011] px-4 py-3"
          >
            <p className="text-xs text-slate-400">{item.label}</p>
            <p className="mt-1 text-xl font-semibold tabular-nums text-white">
              {item.value}
            </p>
          </div>
        ))}
      </div>

      {/* Tabs */}
      <div className="border-b border-[#252628] mb-6">
        <nav className="flex gap-0" aria-label="Client sections">
          {(["profile", "documents", "deals"] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              aria-current={tab === t ? "page" : undefined}
              className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-all capitalize -mb-px ${
                tab === t
                  ? "border-blue-500 text-blue-400"
                  : "border-transparent text-slate-500 hover:text-slate-300"
              }`}
            >
              {t}
              {t === "deals" &&
                clientDeals.length + cosignerDeals.length > 0 && (
                  <span className="ml-1.5 text-xs bg-blue-500/15 text-blue-400 border border-blue-500/25 px-1.5 py-0.5 rounded-full">
                    {clientDeals.length + cosignerDeals.length}
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
          cosignerDeals={cosignerDeals}
          brokers={allBrokers}
          dealers={allDealers}
          onEdit={(deal) => setDealModal({ open: true, deal })}
          onStatusChange={(updated) =>
            setClientDeals((prev) =>
              prev.map((d) => (d.id === updated.id ? updated : d)),
            )
          }
        />
      )}

      {dealModal.open && (
        <DealModal
          deal={dealModal.deal}
          clientId={id}
          clients={allClients}
          brokers={allBrokers}
          dealers={allDealers}
          onSave={handleSaveDeal}
          onClose={() => setDealModal({ open: false })}
        />
      )}
    </div>
  );
}

function money(value?: string | number | null) {
  if (value == null || value === "") return undefined;
  const amount = Number(value);
  return Number.isFinite(amount)
    ? new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: "USD",
        maximumFractionDigits: 2,
      }).format(amount)
    : undefined;
}

function ProfileTab({ client }: { client: Client }) {
  const employment = {
    full_time: "Full time",
    part_time: "Part time",
    self_employed: "Self employed",
  };
  const housing: Record<string, string> = {
    own: "Own",
    rent: "Rent",
    family: "Living with family",
  };
  const sections = [
    {
      title: "Personal details",
      description: "Identity and contact information",
      fields: [
        {
          label: "Full name",
          value: `${client.first_name} ${client.last_name}`,
        },
        { label: "Email", value: client.email },
        { label: "Phone", value: client.phone },
        {
          label: "Date of birth",
          value: client.date_of_birth
            ? fmtDate(client.date_of_birth)
            : undefined,
        },
        {
          label: "Social Security number",
          value: client.has_ssn ? "Stored securely · hidden" : "Not provided",
        },
        {
          label: "Client source",
          value: client.source ? clientSourceLabels[client.source] : undefined,
        },
      ],
    },
    {
      title: "Address & housing",
      description: "Where your client lives",
      fields: [
        { label: "Street address", value: client.address },
        { label: "City", value: client.city },
        { label: "State", value: client.state },
        { label: "ZIP code", value: client.zip },
        {
          label: "Housing",
          value: client.housing_status
            ? (housing[client.housing_status] ?? client.housing_status)
            : undefined,
        },
        { label: "Time at address", value: client.housing_duration },
        {
          label: "Monthly housing payment",
          value: money(client.housing_monthly_payment),
        },
      ],
    },
    {
      title: "Employment",
      description: "Work and employer information",
      fields: [
        { label: "Employer", value: client.employer_name },
        { label: "Occupation", value: client.occupation },
        {
          label: "Employment type",
          value: client.employment_status
            ? employment[client.employment_status]
            : undefined,
        },
        { label: "Time employed", value: client.employment_duration },
        { label: "Employer phone", value: client.employer_phone },
        {
          label: "Employer address",
          value: [
            client.employer_address,
            client.employer_city,
            client.employer_state,
            client.employer_zip,
          ]
            .filter(Boolean)
            .join(", "),
        },
      ],
    },
    {
      title: "Income",
      description: "Monthly financial information",
      fields: [
        {
          label: "Gross monthly income",
          value: money(client.gross_monthly_income),
        },
        {
          label: "Other monthly income",
          value: money(client.other_monthly_income),
        },
        { label: "Other income source", value: client.other_income_source },
      ],
    },
  ];
  return (
    <div className="grid items-start gap-4 lg:grid-cols-2">
      {sections.map((section) => (
        <section
          key={section.title}
          className="h-full rounded-xl border border-[#252628] bg-[#0f1011] p-5 sm:p-6"
        >
          <h2 className="font-semibold text-slate-200">{section.title}</h2>
          <p className="mb-5 mt-1 text-xs text-slate-400">
            {section.description}
          </p>
          <dl className="grid grid-cols-1 gap-x-5 gap-y-5 sm:grid-cols-2">
            {section.fields.map(({ label, value }) => (
              <div key={label} className="min-w-0">
                <dt className="text-xs text-slate-400">{label}</dt>
                <dd
                  className={`mt-1 break-words text-sm ${value != null && value !== "" ? "text-slate-200" : "text-slate-500"}`}
                >
                  {value != null && value !== ""
                    ? String(value)
                    : "Not provided"}
                </dd>
              </div>
            ))}
          </dl>
        </section>
      ))}
      <section className="rounded-xl border border-[#252628] bg-[#0f1011] p-5 sm:p-6 lg:col-span-2">
        <h2 className="font-semibold text-slate-200">Notes</h2>
        <p className="mt-3 whitespace-pre-wrap break-words text-sm leading-relaxed text-slate-400">
          {client.notes ||
            "No notes added yet. Edit the profile to keep useful context here."}
        </p>
      </section>
    </div>
  );
}

function DealsTab({
  deals: dealList,
  cosignerDeals,
  brokers: allBrokers,
  dealers: allDealers,
  onEdit,
  onStatusChange,
}: {
  deals: Deal[];
  cosignerDeals: Deal[];
  brokers: Broker[];
  dealers: Dealer[];
  onEdit: (deal: Deal) => void;
  onStatusChange: (deal: Deal) => void;
}) {
  const [mutationError, setMutationError] = useState("");
  const [busy, setBusy] = useState(false);
  const brokerMap = Object.fromEntries(
    allBrokers.map((b) => [b.id, `${b.first_name} ${b.last_name}`]),
  );
  const dealerMap = Object.fromEntries(allDealers.map((d) => [d.id, d.name]));

  async function toggleStatus(
    deal: Deal,
    field: "broker_fee_status" | "broker_commission_status",
  ) {
    setMutationError("");
    setBusy(true);
    try {
      let updated: Deal;
      if (field === "broker_fee_status") {
        const next = deal.broker_fee_status === "paid" ? "pending" : "paid";
        updated = await deals.updateBrokerFeeStatus(deal.id, next);
      } else {
        const next =
          deal.broker_commission_status === "paid" ? "pending" : "paid";
        updated = await deals.updateBrokerCommissionStatus(deal.id, next);
      }
      onStatusChange(updated);
    } catch {
      setMutationError("Couldn’t update this deal. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  async function updateDealStatus(deal: Deal, status: Deal["deal_status"]) {
    if (!status) return;
    setMutationError("");
    setBusy(true);
    try {
      const updated = await deals.updateStatus(deal.id, status);
      onStatusChange(updated);
    } catch {
      setMutationError("Couldn’t update this deal. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  if (!dealList.length && !cosignerDeals.length) {
    return (
      <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] py-12 text-center text-slate-600 text-sm">
        No deals yet
      </div>
    );
  }

  return (
    <fieldset disabled={busy} className="min-w-0 space-y-4">
      {mutationError && (
        <p role="alert" className="text-sm text-red-400">
          {mutationError}
        </p>
      )}
      {cosignerDeals.length > 0 && (
        <div className="bg-blue-500/10 border border-blue-500/25 rounded-xl px-4 py-3">
          <p className="text-xs font-semibold text-blue-400 uppercase tracking-wider">
            Cosigner on
          </p>
          <div className="mt-2 space-y-1">
            {cosignerDeals.map((deal) => (
              <p key={deal.id} className="text-sm text-slate-300">
                {deal.deal_number ? `#${deal.deal_number} · ` : ""}
                {[deal.make, deal.model, deal.trim_level]
                  .filter(Boolean)
                  .join(" ") || "Deal"}
                {deal.vin ? ` · VIN: ${deal.vin}` : ""}
              </p>
            ))}
          </div>
        </div>
      )}
      {dealList.map((deal) => (
        <div
          key={deal.id}
          className="bg-[#0b0c0d] rounded-xl border border-[#252628] p-5 hover:border-blue-500/20 transition-colors"
        >
          <div className="flex flex-wrap items-start justify-between gap-3 mb-4">
            <div>
              <p className="font-medium text-slate-200">
                {[deal.make, deal.model, deal.trim_level]
                  .filter(Boolean)
                  .join(" ") || "Deal"}
              </p>
              <div className="my-2 flex flex-wrap gap-2 text-xs">
                {deal.is_cash_deal && (
                  <span className="rounded-full border border-green-500/25 bg-green-500/10 px-2 py-0.5 text-green-400">
                    Cash deal
                  </span>
                )}
                {deal.lease_term_months != null && (
                  <span className="rounded-full border border-blue-500/25 bg-blue-500/10 px-2 py-0.5 text-blue-400">
                    {deal.lease_term_months}-month lease
                  </span>
                )}
                {!deal.vin && (
                  <span className="rounded-full border border-[#252628] px-2 py-0.5 text-slate-400">
                    VIN not assigned
                  </span>
                )}
              </div>
              <p className="text-slate-400 text-xs mt-0.5 break-all">
                {deal.deal_number && `#${deal.deal_number} · `}
                {fmtDate(deal.deal_date)}
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

          <div className="grid grid-cols-2 gap-4 text-sm mb-4 sm:grid-cols-3">
            {deal.broker_id && (
              <div>
                <span className="text-slate-400 text-xs uppercase tracking-wider">
                  Broker
                </span>
                <p className="text-slate-300 mt-0.5">
                  {brokerMap[deal.broker_id] ?? "–"}
                </p>
              </div>
            )}
            {deal.dealer_id && (
              <div>
                <span className="text-slate-400 text-xs uppercase tracking-wider">
                  Dealer
                </span>
                <p className="text-slate-300 mt-0.5">
                  {dealerMap[deal.dealer_id] ?? "–"}
                </p>
              </div>
            )}
            {deal.down_payment != null && (
              <div>
                <span className="text-slate-400 text-xs uppercase tracking-wider">
                  Down Payment
                </span>
                <p className="text-slate-300 mt-0.5">
                  {money(deal.down_payment)}{" "}
                  {deal.down_payment_method && `(${deal.down_payment_method})`}
                </p>
              </div>
            )}
            {deal.monthly_payment != null && (
              <div>
                <span className="text-slate-400 text-xs uppercase tracking-wider">
                  Monthly
                </span>
                <p className="text-slate-300 mt-0.5">
                  {money(deal.monthly_payment)}
                </p>
              </div>
            )}
            {deal.collect_amount != null && (
              <div>
                <span className="text-slate-400 text-xs uppercase tracking-wider">
                  Collect
                </span>
                <p className="font-semibold text-green-400 mt-0.5">
                  {money(deal.collect_amount)}
                </p>
              </div>
            )}
            {deal.doc_fee != null && (
              <div>
                <span className="text-slate-400 text-xs uppercase tracking-wider">
                  Doc Fee
                </span>
                <p className="text-slate-300 mt-0.5">{money(deal.doc_fee)}</p>
              </div>
            )}
            <div>
              <span className="text-slate-400 text-xs uppercase tracking-wider">
                Deal Status
              </span>
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
              Pending {deal.pending_payment_days} days ·{" "}
              {deal.pending_payment_bucket}
            </p>
          )}
        </div>
      ))}
    </fieldset>
  );
}

function StatusButton({
  label,
  amount,
  status,
  paidDate,
  onClick,
}: {
  label: string;
  amount?: number;
  status?: string;
  paidDate?: string;
  onClick: () => void;
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
      <span>
        {label}
        {amount != null ? ` ${money(amount)}` : ""}
      </span>
      {paid && paidDate && (
        <span className="text-green-600 opacity-80">{fmtDate(paidDate)}</span>
      )}
    </button>
  );
}
