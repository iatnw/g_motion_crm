"use client";

import { useEffect, useState } from "react";
import { referrals as referralsApi, deals as dealsApi } from "@/lib/api";
import type { Referral, Deal, ReferralStatus } from "@/lib/types";
import { fmtDate } from "@/lib/utils";
import Modal from "@/components/Modal";
import StatusSelect, { referralOptions } from "@/components/StatusSelect";
import useConfirmDialog from "@/components/useConfirmDialog";

interface ReferralForm {
  deal_id: string;
  first_name: string;
  last_name: string;
  phone: string;
  amount: string;
  status: ReferralStatus;
}

const emptyForm: ReferralForm = {
  deal_id: "",
  first_name: "",
  last_name: "",
  phone: "",
  amount: "",
  status: "none",
};

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

export default function ReferralsPage() {
  const { confirmAction, confirmationDialog } = useConfirmDialog();
  const [list, setList] = useState<Referral[]>([]);
  const [deals, setDeals] = useState<Deal[]>([]);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [loadError, setLoadError] = useState("");
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<Referral | null>(null);
  const [modal, setModal] = useState(false);


  const [form, setForm] = useState<ReferralForm>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([referralsApi.list(), dealsApi.list()])
      .then(([refs, ds]) => {
        setList(refs ?? []);
        setDeals(ds ?? []);
      })
      .catch(() =>
        setLoadError("Couldn’t load your referrals. Please reload the page."),
      )
      .finally(() => setLoading(false));
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
    if (!form.first_name) {
      setError("First name is required");
      return;
    }
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
        const deal = deals.find(
          (d) => d.id === (form.deal_id || updated.deal_id),
        );
        setList((prev) =>
          prev.map((r) =>
            r.id === editing.id
              ? {
                  ...updated,
                  deal_number: deal?.deal_number ?? updated.deal_number,
                }
              : r,
          ),
        );
      } else {
        const ref = await referralsApi.create({
          deal_id: form.deal_id || undefined,
          first_name: form.first_name,
          last_name: form.last_name,
          phone: form.phone || undefined,
          amount: form.amount ? parseFloat(form.amount) : undefined,
        });
        const deal = deals.find((d) => d.id === form.deal_id);
        setList((prev) => [
          { ...ref, deal_number: deal?.deal_number },
          ...prev,
        ]);
      }
      setModal(false);
    } catch (err: unknown) {
      setError(
        err instanceof Error
          ? err.message
          : editing
            ? "Failed to update referral"
            : "Failed to create referral",
      );
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(ref: Referral) {
    if (!(await confirmAction(`The referral for ${ref.first_name} ${ref.last_name} will be permanently removed.`))) return;
    try {
      await referralsApi.delete(ref.id);
      setList((prev) => prev.filter((r) => r.id !== ref.id));
    } catch {
      setLoadError("Failed to delete referral");
    }
  }

  async function handleStatusChange(ref: Referral, status: ReferralStatus) {
    const updated = await referralsApi.updateStatus(ref.id, status);
    setList((previous) => previous.map((r) => r.id === ref.id ? { ...updated, deal_number: r.deal_number } : r));
  }

  const filtered = list.filter(
    (ref) =>
      (!statusFilter || ref.status === statusFilter) &&
      [ref.first_name, ref.last_name, ref.phone, ref.deal_number]
        .filter(Boolean)
        .join(" ")
        .toLowerCase()
        .includes(search.trim().toLowerCase()),
  );
  const currency = (value: number) =>
    new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      maximumFractionDigits: 0,
    }).format(value);
  return (
    <div className="p-6">
      <div className="flex flex-wrap items-center justify-between gap-4 mb-6">
        <div>
          <h1 className="text-xl font-semibold text-white">Referrals</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            {list.length} referrals · Track introductions and payouts
          </p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition-colors"
        >
          + Create Referral
        </button>
      </div>

      <div className="mb-6 grid grid-cols-1 gap-3 sm:grid-cols-3">
        {[
          { label: "All referrals", value: list.length },
          {
            label: "Pending payouts",
            value: currency(
              list
                .filter((ref) => ref.status === "pending")
                .reduce((sum, ref) => sum + Number(ref.amount ?? 0), 0),
            ),
          },
          {
            label: "Paid out",
            value: currency(
              list
                .filter((ref) => ref.status === "paid")
                .reduce((sum, ref) => sum + Number(ref.amount ?? 0), 0),
            ),
          },
        ].map((item) => (
          <div
            key={item.label}
            className="rounded-xl border border-[#252628] bg-[#0f1011] p-4"
          >
            <p className="text-xs text-slate-400">{item.label}</p>
            <p className="mt-2 text-2xl font-semibold text-white">
              {loading ? "—" : item.value}
            </p>
          </div>
        ))}
      </div>
      <div className="mb-5 flex flex-wrap gap-3">
        <input
          type="search"
          aria-label="Search referrals"
          placeholder="Search name, phone, or deal number…"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          className="min-w-0 flex-1 rounded-lg border border-[#252628] bg-[#0b0c0d] px-4 py-2.5 text-sm text-slate-200 placeholder:text-slate-500"
        />
        <StatusSelect label="Filter referrals by status" value={statusFilter}
          options={[{ value: "", label: "All statuses", description: "Show every referral", tone: "blue" }, ...referralOptions]}
          onChange={setStatusFilter} />
      </div>
      {loadError && (
        <p role="alert" className="mb-4 text-sm text-red-400">
          {loadError}{" "}
          <button
            onClick={() => window.location.reload()}
            className="underline"
          >
            Reload
          </button>
        </p>
      )}
      {loading ? (
        <p className="text-slate-500 text-sm">Loading…</p>
      ) : filtered.length === 0 ? (
        <p className="text-slate-500 text-sm">
          {list.length
            ? "No matching referrals. Try another search or status."
            : "No referrals yet. Create one to track your first introduction."}
        </p>
      ) : (
        <div className="bg-[#0f1011] rounded-2xl border border-[#252628] crm-table-scroll">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[#252628]">
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Name
                </th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Phone
                </th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Deal
                </th>
                <th className="text-right px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Commission
                </th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Date
                </th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-[#252628]">
              {filtered.map((ref) => (
                <tr
                  key={ref.id}
                  className="hover:bg-white/[0.02] transition-colors"
                >
                  <td className="px-4 py-3 text-slate-200 font-medium">
                    {ref.first_name} {ref.last_name}
                  </td>
                  <td className="px-4 py-3 text-slate-400">
                    {ref.phone ?? "—"}
                  </td>
                  <td className="px-4 py-3 text-slate-400">
                    {ref.deal_number ?? "—"}
                  </td>
                  <td className="px-4 py-3 text-right text-slate-300">
                    {ref.amount != null
                      ? `$${Number(ref.amount).toLocaleString()}`
                      : "—"}
                  </td>
                  <td className="px-4 py-3">
                    <StatusSelect label="Referral status" value={ref.status} options={referralOptions}
                      onChange={(status) => handleStatusChange(ref, status as ReferralStatus)} />
                  </td>
                  <td className="px-4 py-3 text-slate-500">
                    {ref.paid_date ? fmtDate(ref.paid_date) : "—"}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-3">
                      <button
                        onClick={() => openEdit(ref)}
                        className="text-slate-600 hover:text-blue-400 text-xs transition-colors"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(ref)}
                        className="text-slate-600 hover:text-red-400 text-xs transition-colors"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {modal && (
        <Modal title={editing ? "Edit Referral" : "Create Referral"} description="Track a referral and its payout, with or without a deal." onClose={() => setModal(false)} busy={saving}>
            <form onSubmit={handleSubmit} className="crm-modal-form">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                  Deal (optional)
                </label>
                <select
                  value={form.deal_id}
                  onChange={(e) =>
                    setForm((p) => ({ ...p, deal_id: e.target.value }))
                  }
                  className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2.5 text-sm text-white focus:outline-none"
                >
                  <option value="">No deal</option>
                  {deals.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.deal_number ?? d.id.slice(0, 8)} — {d.make ?? ""}{" "}
                      {d.model ?? ""}
                    </option>
                  ))}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                    First Name
                  </label>
                  <input
                    type="text"
                    value={form.first_name}
                    onChange={(e) =>
                      setForm((p) => ({ ...p, first_name: e.target.value }))
                    }
                    required
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
                    placeholder="First name"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                    Last Name
                  </label>
                  <input
                    type="text"
                    value={form.last_name}
                    onChange={(e) =>
                      setForm((p) => ({ ...p, last_name: e.target.value }))
                    }
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
                    placeholder="Optional"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                    Phone
                  </label>
                  <input
                    type="text"
                    value={form.phone}
                    onChange={(e) =>
                      setForm((p) => ({ ...p, phone: e.target.value }))
                    }
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
                    placeholder="Optional"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                    Commission ($)
                  </label>
                  <input
                    type="number"
                    value={form.amount}
                    onChange={(e) =>
                      setForm((p) => ({ ...p, amount: e.target.value }))
                    }
                    onWheel={(e) => e.currentTarget.blur()}
                    className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white focus:outline-none"
                    placeholder="0"
                  />
                </div>
              </div>
              {editing && (
                <div>
                  <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                    Status
                  </label>
                  <StatusSelect label="Referral status" value={form.status} options={referralOptions} fullWidth
                    onChange={(status) => setForm((previous) => ({ ...previous, status: status as ReferralStatus }))} />
                </div>
              )}
              {error && <p className="text-red-400 text-sm">{error}</p>}
              <div className="crm-modal-footer">
                <button
                  type="button"
                  onClick={() => setModal(false)}
                  disabled={saving}
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
        </Modal>
      )}
      {confirmationDialog}
    </div>
  );
}
