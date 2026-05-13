"use client";

import { useState, useEffect } from "react";
import type { Deal, Broker, Dealer, Client, Referral } from "@/lib/types";
import { deals as dealsApi, referrals as referralsApi } from "@/lib/api";

interface Props {
  deal?: Deal;
  clientId?: string;
  nextDealNumber?: string;
  clients?: Client[];
  brokers: Broker[];
  dealers: Dealer[];
  onSave: (data: object) => Promise<Deal>;
  onClose: () => void;
  onStatusChange?: (deal: Deal) => void;
}

type ReferralDraft = { first_name: string; last_name: string; phone: string; amount: string };

export default function DealModal({ deal, clientId, nextDealNumber, clients, brokers, dealers, onSave, onClose, onStatusChange }: Props) {
  const [currentDeal, setCurrentDeal] = useState<Deal | undefined>(deal);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [onClose]);
  const [form, setForm] = useState({
    client_id: clientId ?? deal?.client_id ?? "",
    deal_date: deal?.deal_date?.slice(0, 10) ?? formatDateInput(),
    deal_number: deal?.deal_number ?? nextDealNumber ?? "",
    vin: deal?.vin ?? "",
    make: deal?.make ?? "",
    model: deal?.model ?? "",
    trim_level: deal?.trim_level ?? "",
    broker_id: deal?.broker_id ?? "",
    dealer_id: deal?.dealer_id ?? "",
    down_payment: deal?.down_payment?.toString() ?? "",
    down_payment_method: deal?.down_payment_method ?? "",
    monthly_payment: deal?.monthly_payment?.toString() ?? "",
    broker_fee: deal?.broker_fee?.toString() ?? "",
    broker_percent: deal?.broker_percent?.toString() ?? "",
    broker_commission: deal?.broker_commission?.toString() ?? "",
    your_net: deal?.your_net?.toString() ?? "",
    broker_fee_status: deal?.broker_fee_status ?? "pending",
    broker_commission_status: deal?.broker_commission_status ?? "pending",
    deal_status: deal?.deal_status ?? "pending",
    collect_amount: deal?.collect_amount?.toString() ?? "",
    doc_fee: deal?.doc_fee?.toString() ?? "",
    broker_payment_method: deal?.broker_payment_method ?? "",
    notes: deal?.notes ?? "",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const [existingReferrals, setExistingReferrals] = useState<Referral[]>([]);
  const [referralAmounts, setReferralAmounts] = useState<Record<string, string>>({});
  const [newReferrals, setNewReferrals] = useState<ReferralDraft[]>([]);
  const [allReferralContacts, setAllReferralContacts] = useState<Referral[]>([]);

  // Auto-correct commission/net on load if stored values don't match the current formula
  useEffect(() => {
    const base = commissionBase(form.broker_fee, form.collect_amount, form.doc_fee);
    const pct = parseFloat(form.broker_percent);
    if (base > 0 && !isNaN(pct) && pct > 0) {
      const correct = base * pct / 100;
      const stored = parseFloat(form.broker_commission) || 0;
      if (Math.abs(correct - stored) > 0.02) {
        setForm((prev) => ({
          ...prev,
          broker_commission: correct.toFixed(2),
          your_net: (base - correct).toFixed(2),
        }));
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (deal) {
      referralsApi.listByDeal(deal.id).then((refs) => {
        setExistingReferrals(refs);
        const amounts: Record<string, string> = {};
        for (const r of refs) amounts[r.id] = r.amount != null ? String(r.amount) : "";
        setReferralAmounts(amounts);
      }).catch(() => {});
    }
    referralsApi.list().then((data) => {
      // Deduplicate by full name, keeping latest entry per person
      const seen = new Map<string, Referral>();
      for (const r of (data ?? [])) {
        const key = `${r.first_name.trim().toLowerCase()} ${r.last_name.trim().toLowerCase()}`;
        if (!seen.has(key)) seen.set(key, r);
      }
      setAllReferralContacts(Array.from(seen.values()));
    }).catch(() => {});
  }, [deal]);

  function set(field: string, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  async function toggleReferralStatus(ref: Referral) {
    const next = ref.status === "paid" ? "pending" : "paid";
    try {
      const updated = await referralsApi.updateStatus(ref.id, next);
      setExistingReferrals((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
    } catch {
      setError("Failed to update referral status");
    }
  }

  async function saveReferralAmount(ref: Referral) {
    const raw = referralAmounts[ref.id] ?? "";
    const amount = raw ? Number(raw) : undefined;
    if (amount === ref.amount || (amount == null && ref.amount == null)) return;
    try {
      const updated = await referralsApi.update(ref.id, { amount });
      setExistingReferrals((prev) => prev.map((r) => r.id === ref.id ? { ...r, amount: updated.amount } : r));
    } catch {
      setError("Failed to update referral amount");
    }
  }

  async function deleteReferral(id: string) {
    try {
      await referralsApi.delete(id);
      setExistingReferrals((prev) => prev.filter((r) => r.id !== id));
    } catch {
      setError("Failed to delete referral");
    }
  }

  function addReferralRow() {
    setNewReferrals((prev) => [...prev, { first_name: "", last_name: "", phone: "", amount: "" }]);
  }

  function updateReferralRow(i: number, field: keyof ReferralDraft, value: string) {
    setNewReferrals((prev) => prev.map((r, idx) => idx === i ? { ...r, [field]: value } : r));
  }

  function removeReferralRow(i: number) {
    setNewReferrals((prev) => prev.filter((_, idx) => idx !== i));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.client_id) {
      setError("Client is required");
      return;
    }
    setSaving(true);
    setError("");
    try {
      const numericFields = new Set([
        "down_payment",
        "monthly_payment",
        "broker_fee",
        "broker_percent",
        "broker_commission",
        "your_net",
        "collect_amount",
        "doc_fee",
      ]);
      const payload: Record<string, string | number> = {};
      for (const [k, v] of Object.entries(form)) {
        if (v === "") continue;
        payload[k] = numericFields.has(k) ? Number(v) : v;
      }
      const savedDeal = await onSave(payload);

      const validNewReferrals = newReferrals.filter((r) => r.first_name.trim());
      await Promise.all(
        validNewReferrals.map((r) =>
          referralsApi.create({
            deal_id: savedDeal.id,
            first_name: r.first_name,
            last_name: r.last_name || undefined,
            phone: r.phone || undefined,
            amount: r.amount ? Number(r.amount) : undefined,
          })
        )
      );

      onClose();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  const showClientSelector = !clientId && !!clients;

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-[#0b0c0d] rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto border border-[#252628] shadow-2xl" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-[#252628]">
          <h2 className="text-base font-semibold text-white">
            {deal ? "Edit Deal" : "New Deal"}
          </h2>
          <button onClick={onClose} className="text-slate-600 hover:text-slate-300 text-xl transition-colors">✕</button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* Client selector */}
          {showClientSelector && (
            <section>
              <div>
                <FieldLabel>Client *</FieldLabel>
                <DarkSelect
                  value={form.client_id}
                  onChange={(e) => set("client_id", e.target.value)}
                  required
                >
                  <option value="">Select a client…</option>
                  {clients!.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.first_name} {c.last_name}{c.phone ? ` — ${c.phone}` : ""}
                    </option>
                  ))}
                </DarkSelect>
              </div>
            </section>
          )}

          {/* Vehicle */}
          <section>
            <SectionLabel>Vehicle</SectionLabel>
            <div className="grid grid-cols-2 gap-3">
              <Field label="Make" value={form.make} onChange={(v) => set("make", v)} />
              <Field label="Model" value={form.model} onChange={(v) => set("model", v)} />
              <Field label="Trim" value={form.trim_level} onChange={(v) => set("trim_level", v)} />
              <Field label="VIN" value={form.vin} onChange={(v) => set("vin", v)} />
            </div>
          </section>

          {/* Deal Info */}
          <section>
            <SectionLabel>Deal Info</SectionLabel>
            <div className="grid grid-cols-2 gap-3">
              <Field label="Deal Number" value={form.deal_number} onChange={(v) => set("deal_number", v)} readOnly={!deal} />
              <Field label="Deal Date" type="date" value={form.deal_date} onChange={(v) => set("deal_date", v)} />
              <div>
                <FieldLabel>Broker</FieldLabel>
                <DarkSelect value={form.broker_id} onChange={(e) => set("broker_id", e.target.value)}>
                  <option value="">None</option>
                  {brokers.map((b) => (
                    <option key={b.id} value={b.id}>{b.first_name} {b.last_name}</option>
                  ))}
                </DarkSelect>
              </div>
              <div>
                <FieldLabel>Dealer</FieldLabel>
                <DarkSelect value={form.dealer_id} onChange={(e) => set("dealer_id", e.target.value)}>
                  <option value="">None</option>
                  {dealers.map((d) => (
                    <option key={d.id} value={d.id}>{d.name}</option>
                  ))}
                </DarkSelect>
              </div>
            </div>
          </section>

          {/* Financials */}
          <section>
            <SectionLabel>Financials</SectionLabel>
            <div className="grid grid-cols-2 gap-3">
              <Field label="Down Payment ($)" type="number" value={form.down_payment} onChange={(v) => set("down_payment", v)} />
              <div>
                <FieldLabel>Down Payment Method</FieldLabel>
                <DarkSelect value={form.down_payment_method} onChange={(e) => set("down_payment_method", e.target.value)}>
                  <option value="">None</option>
                  <option value="cash">Cash</option>
                  <option value="card">Card</option>
                </DarkSelect>
              </div>
              <Field label="Monthly Payment ($)" type="number" value={form.monthly_payment} onChange={(v) => set("monthly_payment", v)} />
              <Field label="Doc Fee ($)" type="number" value={form.doc_fee} onChange={(v) => {
                const base = commissionBase(form.broker_fee, form.collect_amount, v);
                const pct = parseFloat(form.broker_percent);
                setForm((prev) => {
                  const next: Partial<typeof prev> = { doc_fee: v };
                  if (!isNaN(pct) && pct > 0 && base > 0) {
                    const commission = base * pct / 100;
                    next.broker_commission = commission.toFixed(2);
                    next.your_net = (base - commission).toFixed(2);
                  } else {
                    const com = parseFloat(prev.broker_commission);
                    if (!isNaN(com) && base > 0) next.your_net = (base - com).toFixed(2);
                  }
                  return { ...prev, ...next };
                });
              }} />
              <Field label="Broker Fee ($)" type="number" value={form.broker_fee} onChange={(v) => {
                const base = commissionBase(v, form.collect_amount, form.doc_fee);
                const pct = parseFloat(form.broker_percent);
                const com = parseFloat(form.broker_commission);
                setForm((prev) => {
                  const next: Partial<typeof prev> = { broker_fee: v };
                  if (form.broker_percent && !isNaN(pct) && base > 0) {
                    const commission = base * pct / 100;
                    next.broker_commission = commission.toFixed(2);
                    next.your_net = (base - commission).toFixed(2);
                  } else if (form.broker_commission && !isNaN(com) && base > 0) {
                    next.broker_percent = (com / base * 100).toFixed(2);
                    next.your_net = (base - com).toFixed(2);
                  }
                  return { ...prev, ...next };
                });
              }} />
              <Field label="Collect ($)" type="number" value={form.collect_amount} onChange={(v) => {
                const base = commissionBase(form.broker_fee, v, form.doc_fee);
                const pct = parseFloat(form.broker_percent);
                const com = parseFloat(form.broker_commission);
                setForm((prev) => {
                  const next: Partial<typeof prev> = { collect_amount: v };
                  if (form.broker_percent && !isNaN(pct) && base > 0) {
                    const commission = base * pct / 100;
                    next.broker_commission = commission.toFixed(2);
                    next.your_net = (base - commission).toFixed(2);
                  } else if (form.broker_commission && !isNaN(com) && base > 0) {
                    next.broker_percent = (com / base * 100).toFixed(2);
                    next.your_net = (base - com).toFixed(2);
                  }
                  return { ...prev, ...next };
                });
              }} />
              <Field label="Broker %" type="number" value={form.broker_percent} onChange={(v) => {
                const base = commissionBase(form.broker_fee, form.collect_amount, form.doc_fee);
                const pct = parseFloat(v);
                setForm((prev) => {
                  if (!v || isNaN(pct) || base === 0) return { ...prev, broker_percent: v };
                  const commission = base * pct / 100;
                  return { ...prev, broker_percent: v, broker_commission: commission.toFixed(2), your_net: (base - commission).toFixed(2) };
                });
              }} />
              <Field label="Broker Commission ($)" type="number" value={form.broker_commission} onChange={(v) => {
                const base = commissionBase(form.broker_fee, form.collect_amount, form.doc_fee);
                const com = parseFloat(v);
                setForm((prev) => {
                  if (!v || isNaN(com)) return { ...prev, broker_commission: v };
                  const next: Partial<typeof prev> = { broker_commission: v, your_net: base > 0 ? (base - com).toFixed(2) : prev.your_net };
                  if (base > 0) next.broker_percent = (com / base * 100).toFixed(2);
                  return { ...prev, ...next };
                });
              }} />
              <Field label="Your Net ($)" type="number" value={form.your_net} onChange={(v) => set("your_net", v)} />
              {currentDeal && (
                <div>
                  <FieldLabel>Deal Status</FieldLabel>
                  <DarkSelect value={form.deal_status} onChange={(e) => set("deal_status", e.target.value)}>
                    <option value="pending">NOT PAID</option>
                    <option value="paid">PAID</option>
                    <option value="cancelled">Cancelled</option>
                  </DarkSelect>
                </div>
              )}
              <div>
                <FieldLabel>Broker Payment Type</FieldLabel>
                <DarkSelect value={form.broker_payment_method} onChange={(e) => set("broker_payment_method", e.target.value)}>
                  <option value="">None</option>
                  <option value="zelle_gmotion">Zelle G Motion</option>
                  <option value="zelle_personal">Zelle Personal</option>
                  <option value="cash">Cash</option>
                </DarkSelect>
              </div>
            </div>
            <div className="mt-3">
              <FieldLabel>Notes</FieldLabel>
              <textarea
                value={form.notes}
                onChange={(e) => set("notes", e.target.value)}
                rows={3}
                className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white placeholder:text-slate-700 focus:outline-none transition-colors resize-none"
              />
            </div>
          </section>


          {/* Referrals */}
          <section>
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Referrals</h3>
              <button
                type="button"
                onClick={addReferralRow}
                className="text-xs text-blue-400 hover:text-blue-300 transition-colors font-medium"
              >
                + Add
              </button>
            </div>

            {existingReferrals.length === 0 && newReferrals.length === 0 && (
              <p className="text-xs text-slate-600 py-1">No referrals. Click + Add to add one.</p>
            )}

            {existingReferrals.length > 0 && (
              <div className="space-y-2 mb-2">
                {existingReferrals.map((ref) => (
                  <div key={ref.id} className="flex items-center gap-2 bg-[#111213] border border-[#252628] rounded-lg px-3 py-2">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-slate-200 truncate">{ref.first_name} {ref.last_name}</p>
                      {ref.phone && <p className="text-xs text-slate-600">{ref.phone}</p>}
                    </div>
                    <input
                      type="number"
                      value={referralAmounts[ref.id] ?? ""}
                      onChange={(e) => setReferralAmounts((prev) => ({ ...prev, [ref.id]: e.target.value }))}
                      onBlur={() => saveReferralAmount(ref)}
                      placeholder="$0"
                      className="w-20 bg-[#0b0c0d] border border-[#252628] rounded-md px-2 py-0.5 text-sm text-slate-300 focus:outline-none shrink-0"
                    />
                    <button
                      type="button"
                      onClick={() => toggleReferralStatus(ref)}
                      className={`shrink-0 text-xs px-2 py-0.5 rounded-full border font-medium transition-colors ${
                        ref.status === "paid"
                          ? "bg-green-500/10 border-green-500/25 text-green-400"
                          : "bg-[#0a0a1e] border-[#1e1e45] text-slate-500 hover:text-slate-300"
                      }`}
                    >
                      {ref.status === "paid" ? "Paid" : "Pending"}
                    </button>
                    <button
                      type="button"
                      onClick={() => deleteReferral(ref.id)}
                      className="shrink-0 text-slate-600 hover:text-red-400 transition-colors text-sm"
                    >
                      ✕
                    </button>
                  </div>
                ))}
              </div>
            )}

            {newReferrals.length > 0 && (
              <div className="space-y-2">
                {newReferrals.map((ref, i) => (
                  <div key={i} className="grid grid-cols-2 gap-2 bg-[#111213] border border-[#252628] rounded-lg p-3">
                    {allReferralContacts.length > 0 && (
                      <div className="col-span-2">
                        <FieldLabel>Existing Referral</FieldLabel>
                        <select
                          defaultValue=""
                          onChange={(e) => {
                            const contact = allReferralContacts.find((c) => c.id === e.target.value);
                            if (contact) {
                              setNewReferrals((prev) => prev.map((r, idx) => idx === i
                                ? { ...r, first_name: contact.first_name, last_name: contact.last_name, phone: contact.phone ?? "" }
                                : r
                              ));
                            }
                          }}
                          className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white focus:outline-none"
                        >
                          <option value="">— Select existing or fill in below —</option>
                          {allReferralContacts.map((c) => (
                            <option key={c.id} value={c.id}>
                              {c.first_name} {c.last_name}{c.phone ? ` · ${c.phone}` : ""}
                            </option>
                          ))}
                        </select>
                      </div>
                    )}
                    <div>
                      <FieldLabel>First Name</FieldLabel>
                      <input
                        type="text"
                        value={ref.first_name}
                        onChange={(e) => updateReferralRow(i, "first_name", e.target.value)}
                        placeholder="First name"
                        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white placeholder:text-slate-700 focus:outline-none"
                      />
                    </div>
                    <div>
                      <FieldLabel>Last Name</FieldLabel>
                      <input
                        type="text"
                        value={ref.last_name}
                        onChange={(e) => updateReferralRow(i, "last_name", e.target.value)}
                        placeholder="Last name"
                        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white placeholder:text-slate-700 focus:outline-none"
                      />
                    </div>
                    <div>
                      <FieldLabel>Phone</FieldLabel>
                      <input
                        type="text"
                        value={ref.phone}
                        onChange={(e) => updateReferralRow(i, "phone", e.target.value)}
                        placeholder="Phone"
                        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white placeholder:text-slate-700 focus:outline-none"
                      />
                    </div>
                    <div>
                      <FieldLabel>Commission ($)</FieldLabel>
                      <input
                        type="number"
                        value={ref.amount}
                        onChange={(e) => updateReferralRow(i, "amount", e.target.value)}
                        placeholder="0"
                        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white placeholder:text-slate-700 focus:outline-none"
                      />
                    </div>
                    <div className="col-span-2 flex justify-end">
                      <button
                        type="button"
                        onClick={() => removeReferralRow(i)}
                        className="text-xs text-slate-600 hover:text-red-400 transition-colors"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <div className="flex justify-end gap-3 pt-1">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-[#1e1e45] rounded-lg text-sm text-slate-400 hover:text-white hover:border-slate-500 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-all glow-blue-sm"
            >
              {saving ? "Saving…" : "Save Deal"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3">{children}</h3>
  );
}

function FieldLabel({ children }: { children: React.ReactNode }) {
  return (
    <label className="block text-xs font-medium text-slate-500 mb-1.5 uppercase tracking-wider">{children}</label>
  );
}

function DarkSelect({ children, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      {...props}
      className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white focus:outline-none transition-colors"
    >
      {children}
    </select>
  );
}

function Field({
  label,
  value,
  onChange,
  type = "text",
  readOnly,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  readOnly?: boolean;
}) {
  return (
    <div>
      <FieldLabel>{label}</FieldLabel>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        readOnly={readOnly}
        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white placeholder:text-slate-700 focus:outline-none transition-colors read-only:text-slate-500"
      />
    </div>
  );
}

function formatDateInput(date = new Date()) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function commissionBase(brokerFee: string, collectAmount: string, docFee: string) {
  const fee = parseFloat(brokerFee) || 0;
  const collect = parseFloat(collectAmount) || 0;
  const doc = parseFloat(docFee) || 0;
  return fee > 0 ? fee - doc : collect;
}

