"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { brokers as brokersApi } from "@/lib/api";
import type { Broker } from "@/lib/types";

interface BrokerForm {
  first_name: string;
  last_name: string;
  phone: string;
  email: string;
  notes: string;
}

const empty: BrokerForm = { first_name: "", last_name: "", phone: "", email: "", notes: "" };

export default function BrokersPage() {
  const [list, setList] = useState<Broker[]>([]);
  const [loading, setLoading] = useState(true);
  const [modal, setModal] = useState<{ open: boolean; broker?: Broker }>({ open: false });

  useEffect(() => {
    if (!modal.open) return;
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') setModal({ open: false }); };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [modal.open]);
  const [form, setForm] = useState<BrokerForm>(empty);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    brokersApi.list().then((data) => setList(data ?? [])).finally(() => setLoading(false));
  }, []);

  function openNew() {
    setForm(empty);
    setError("");
    setModal({ open: true });
  }

  function openEdit(broker: Broker) {
    setForm({
      first_name: broker.first_name,
      last_name: broker.last_name,
      phone: broker.phone ?? "",
      email: broker.email ?? "",
      notes: broker.notes ?? "",
    });
    setError("");
    setModal({ open: true, broker });
  }

  async function handleDelete(broker: Broker) {
    if (!confirm(`Delete ${broker.first_name} ${broker.last_name}? This cannot be undone.`)) return;
    try {
      await brokersApi.delete(broker.id);
      setList((prev) => prev.filter((b) => b.id !== broker.id));
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to delete broker");
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      const payload = {
        first_name: form.first_name,
        last_name: form.last_name,
        phone: form.phone || null,
        email: form.email || null,
        notes: form.notes || null,
      };
      if (modal.broker) {
        const updated = await brokersApi.update(modal.broker.id, payload);
        setList((prev) => prev.map((b) => (b.id === updated.id ? updated : b)));
      } else {
        const created = await brokersApi.create(payload);
        setList((prev) => [created, ...prev]);
      }
      setModal({ open: false });
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-white">Brokers</h1>
          <p className="text-slate-500 text-xs mt-0.5">{list.length} total</p>
        </div>
        <button
          onClick={openNew}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition-all glow-blue-sm hover:glow-blue"
        >
          + New Broker
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      ) : (
        <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] overflow-hidden">
          {list.length === 0 ? (
            <div className="text-center py-12 text-slate-600 text-sm">No brokers yet</div>
          ) : (
            <table className="w-full text-sm">
              <thead className="border-b border-[#252628]">
                <tr>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Name</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Phone</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Email</th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1e1f20]">
                {list.map((b) => (
                  <tr key={b.id} className="hover:bg-blue-500/5 transition-colors group">
                    <td className="px-4 py-3.5 font-medium">
                      <Link href={`/brokers/${b.id}`} className="text-slate-200 hover:text-blue-400 transition-colors">
                        {b.first_name} {b.last_name}
                      </Link>
                    </td>
                    <td className="px-4 py-3.5 text-slate-400">{b.phone ?? "–"}</td>
                    <td className="px-4 py-3.5 text-slate-400">{b.email ?? "–"}</td>
                    <td className="px-4 py-3.5 text-right">
                      <div className="flex items-center justify-end gap-3">
                        <button
                          onClick={() => openEdit(b)}
                          className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDelete(b)}
                          className="text-slate-600 hover:text-red-400 text-sm font-medium transition-colors"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {modal.open && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-[#0b0c0d] rounded-2xl w-full max-w-md border border-[#252628] shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between px-6 py-4 border-b border-[#252628]">
              <h2 className="text-base font-semibold text-white">
                {modal.broker ? "Edit Broker" : "New Broker"}
              </h2>
              <button onClick={() => setModal({ open: false })} className="text-slate-600 hover:text-slate-300 text-xl transition-colors">✕</button>
            </div>
            <form onSubmit={handleSave} className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <FormField label="First Name *" value={form.first_name} onChange={(v) => setForm((p) => ({ ...p, first_name: v }))} required />
                <FormField label="Last Name *" value={form.last_name} onChange={(v) => setForm((p) => ({ ...p, last_name: v }))} required />
                <FormField label="Phone" value={form.phone} onChange={(v) => setForm((p) => ({ ...p, phone: v }))} />
                <FormField label="Email" type="email" value={form.email} onChange={(v) => setForm((p) => ({ ...p, email: v }))} />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-500 mb-1.5 uppercase tracking-wider">Notes</label>
                <textarea
                  value={form.notes}
                  onChange={(e) => setForm((p) => ({ ...p, notes: e.target.value }))}
                  rows={2}
                  className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white placeholder:text-slate-700 focus:outline-none transition-colors resize-none"
                />
              </div>
              {error && <p className="text-red-400 text-sm">{error}</p>}
              <div className="flex justify-end gap-3 pt-1">
                <button
                  type="button"
                  onClick={() => setModal({ open: false })}
                  className="px-4 py-2 border border-[#1e1e45] rounded-lg text-sm text-slate-400 hover:text-white hover:border-slate-500 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-all glow-blue-sm"
                >
                  {saving ? "Saving…" : "Save"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

function FormField({
  label, value, onChange, type = "text", required,
}: {
  label: string; value: string; onChange: (v: string) => void; type?: string; required?: boolean;
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-slate-500 mb-1.5 uppercase tracking-wider">{label}</label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white placeholder:text-slate-700 focus:outline-none transition-colors"
      />
    </div>
  );
}
