"use client";

import { useEffect, useState } from "react";
import { referrals as referralsApi, deals as dealsApi } from "@/lib/api";
import type { Referral, Deal, ReferralStatus } from "@/lib/types";

interface ReferralForm {
  deal_id: string;
  first_name: string;
  last_name: string;
  phone: string;
  amount: string;
  status: ReferralStatus;
}

const emptyForm: ReferralForm = { deal_id: "", first_name: "", last_name: "", phone: "", amount: "", status: "none" };

function formFromReferral(ref: Referral): ReferralForm {
  return {
    deal_id: ref.deal_id ?? "",
    first_name: ref.first_name,
    last_name: ref.last_name,
    phone: ref.phone ?? "",
    amount: ref.amount != null ? String(ref.amount) : "",
    status: ref.status,
  };
}

const statusStyles: Record<ReferralStatus, string> = {
  paid:    "bg-green-500/10 border-green-500/25 text-green-400",
  pending: "bg-[#0a0a1e] border-[#1e1e45] text-slate-400",
  none:    "bg-[#111213] border-[#252628] text-slate-600",
};

export default function ReferralsPage() {
  const [list, setList] = useState<Referral[]>([]);
  const [deals, setDeals] = useState<Deal[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<Referral | null>(null);
  const [modal, setModal] = useState(false);

  useEffect(() => {
    if (!modal) return;
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') setModal(false); };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [modal]);
  const [form, setForm] = useState<ReferralForm>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([
      referralsApi.list(),
      dealsApi.list(),
    ]).then(([refs, ds]) => {
      setList(refs ?? []);
      setDeals(ds ?? []);
    }).finally(() => setLoading(false));
  }, []);

  function openCreate() {
    setEditing(null);
    setForm(emptyForm);
    setError("");
    setModal(true);
  }

  function openEdit(ref: Referral) {
    setEditing(ref);
    setForm(formFromReferral(ref));
    setError("");
    setModal(true);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.first_name) { setError("First name is required"); return; }
    setSaving(true);
    setError("");
    try {
      if (editing) {
        const payload = {
          deal_id: form.deal_id || undefined,
          first_name: form.first_name,
          last_name: form.last_name,
          phone: form.phone || undefined,
          amount: form.amount ? parseFloat(form.amount) : undefined,
        };
        let updated = await referralsApi.update(editing.id, payload);
        if (form.status !== editing.status) {
          updated = await referralsApi.updateStatus(editing.id, form.status);
        }
        const deal = deals.find((d) => d.id === (form.deal_id || updated.deal_id));
        setList((prev) => prev.map((r) => r.id === editing.id ? { ...updated, deal_number: deal?.deal_number ?? updated.deal_number } : r));
      } else {
        const ref = await referralsApi.create({
          deal_id: form.deal_id || undefined,
          first_name: form.first_name,
          last_name: form.last_name,
          phone: form.phone || undefined,
          amount: form.amount ? parseFloat(form.amount) : undefined,
        });
        const deal = deals.find((d) => d.id === form.deal_id);
        setList((prev) => [{ ...ref, deal_number: deal?.deal_number }, ...prev]);
      }
      setModal(false);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : editing ? "Failed to update referral" : "Failed to create referral");
    } finally {
      setSaving(false);
    }
  }

  async function handleStatusChange(ref: Referral, status: ReferralStatus) {
    try {
      const updated = await referralsApi.updateStatus(ref.id, status);
      setList((prev) => prev.map((r) => r.id === ref.id ? { ...updated, deal_number: r.deal_number } : r));
    } catch {
      alert("Failed to update status");
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-white">Referrals</h1>
          <p className="text-sm text-slate-500 mt-0.5">{list.length} total</p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition-colors"
        >
          + Create Referral
        </button>
      </div>

      {loading ? (
        <p className="text-slate-500 text-sm">Loading…</p>
      ) : list.length === 0 ? (
        <p className="text-slate-500 text-sm">No referrals yet.</p>
      ) : (
        <div className="bg-[#0f1011] rounded-2xl border border-[#252628] overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[#252628]">
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Name</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Phone</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Deal</th>
                <th className="text-right px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Commission</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Date</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-[#252628]">
              {list.map((ref) => (
                <tr key={ref.id} className="hover:bg-white/[0.02] transition-colors">
                  <td className="px-4 py-3 text-slate-200 font-medium">
                    {ref.first_name} {ref.last_name}
                  </td>
                  <td className="px-4 py-3 text-slate-400">{ref.phone ?? "—"}</td>
                  <td className="px-4 py-3 text-slate-400">{ref.deal_number ?? "—"}</td>
                  <td className="px-4 py-3 text-right text-slate-300">
                    {ref.amount != null ? `$${Number(ref.amount).toLocaleString()}` : "—"}
                  </td>
                  <td className="px-4 py-3">
                    <select
                      value={ref.status}
                      onChange={(e) => handleStatusChange(ref, e.target.value as ReferralStatus)}
                      className={`text-xs font-medium px-2 py-0.5 rounded-full border bg-transparent cursor-pointer focus:outline-none ${statusStyles[ref.status]}`}
                    >
                      <option value="none">None</option>
                      <option value="pending">Pending</option>
                      <option value="paid">Paid</option>
                    </select>
                  </td>
                  <td className="px-4 py-3 text-slate-500">
                    {ref.paid_date ? new Date(ref.paid_date).toLocaleDateString() : "—"}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openEdit(ref)}
                      className="text-slate-600 hover:text-blue-400 text-xs transition-colors"
                    >
                      Edit
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {modal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" onClick={() => setModal(false)}>
          <div className="bg-[#0f1011] rounded-2xl border border-[#252628] w-full max-w-md p-6" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-base font-semibold text-white mb-5">
              {editing ? "Edit Referral" : "Create Referral"}
            </h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">Deal (optional)</label>
                <select
                  value={form.deal_id}
                  onChange={(e) => setForm((p) => ({ ...p, deal_id: e.target.value }))}
                  className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2.5 text-sm text-white focus:outline-none"
                >
                  <option value="">No deal</option>
                  {deals.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.deal_number ?? d.id.slice(0, 8)} — {d.make ?? ""} {d.model ?? ""}
                    </option>
                  ))}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">First Name</label>
                  <input
                    type="text"
                    value={form.first_name}
                    onChange={(e) => setForm((p) => ({ ...p, first_name: e.target.value }))}
                    required
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
                    placeholder="First name"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">Last Name</label>
                  <input
                    type="text"
                    value={form.last_name}
                    onChange={(e) => setForm((p) => ({ ...p, last_name: e.target.value }))}
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
                    placeholder="Optional"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">Phone</label>
                  <input
                    type="text"
                    value={form.phone}
                    onChange={(e) => setForm((p) => ({ ...p, phone: e.target.value }))}
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
                    placeholder="Optional"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">Commission ($)</label>
                  <input
                    type="number"
                    value={form.amount}
                    onChange={(e) => setForm((p) => ({ ...p, amount: e.target.value }))}
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
                    placeholder="0"
                  />
                </div>
              </div>
              {editing && (
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">Status</label>
                  <select
                    value={form.status}
                    onChange={(e) => setForm((p) => ({ ...p, status: e.target.value as ReferralStatus }))}
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2.5 text-sm text-white focus:outline-none"
                  >
                    <option value="none">None</option>
                    <option value="pending">Pending</option>
                    <option value="paid">Paid</option>
                  </select>
                </div>
              )}
              {error && <p className="text-red-400 text-sm">{error}</p>}
              <div className="flex justify-end gap-3 pt-1">
                <button
                  type="button"
                  onClick={() => setModal(false)}
                  className="px-4 py-2 border border-[#1e1e45] rounded-lg text-sm text-slate-400 hover:text-white transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
                >
                  {saving ? "Saving…" : editing ? "Save Changes" : "Create"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
