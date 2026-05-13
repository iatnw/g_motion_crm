"use client";

import { useEffect, useState } from "react";
import { dealers as dealersApi } from "@/lib/api";
import type { Dealer } from "@/lib/types";

interface DealerForm {
  name: string;
  contact_person: string;
  contact_person_email: string;
  phone: string;
  address: string;
  notes: string;
}

const empty: DealerForm = { name: "", contact_person: "", contact_person_email: "", phone: "", address: "", notes: "" };

export default function DealersPage() {
  const [list, setList] = useState<Dealer[]>([]);
  const [loading, setLoading] = useState(true);
  const [modal, setModal] = useState<{ open: boolean; dealer?: Dealer }>({ open: false });

  useEffect(() => {
    if (!modal.open) return;
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') setModal({ open: false }); };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [modal.open]);
  const [form, setForm] = useState<DealerForm>(empty);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    dealersApi.list().then((data) => setList(data ?? [])).finally(() => setLoading(false));
  }, []);

  function openNew() {
    setForm(empty);
    setError("");
    setModal({ open: true });
  }

  function openEdit(dealer: Dealer) {
    setForm({
      name: dealer.name,
      contact_person: dealer.contact_person ?? "",
      contact_person_email: dealer.contact_person_email ?? "",
      phone: dealer.phone ?? "",
      address: dealer.address ?? "",
      notes: dealer.notes ?? "",
    });
    setError("");
    setModal({ open: true, dealer });
  }

  async function handleDelete(dealer: Dealer) {
    if (!confirm(`Delete ${dealer.name}? This cannot be undone.`)) return;
    try {
      await dealersApi.delete(dealer.id);
      setList((prev) => prev.filter((d) => d.id !== dealer.id));
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to delete dealer");
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      const payload = {
        name: form.name,
        contact_person: form.contact_person || null,
        contact_person_email: form.contact_person_email || null,
        phone: form.phone || null,
        address: form.address || null,
        notes: form.notes || null,
      };
      if (modal.dealer) {
        const updated = await dealersApi.update(modal.dealer.id, payload);
        setList((prev) => prev.map((d) => (d.id === updated.id ? updated : d)));
      } else {
        const created = await dealersApi.create(payload);
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
          <h1 className="text-xl font-semibold text-white">Dealers</h1>
          <p className="text-slate-500 text-xs mt-0.5">{list.length} total</p>
        </div>
        <button
          onClick={openNew}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition-all glow-blue-sm hover:glow-blue"
        >
          + New Dealer
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      ) : list.length === 0 ? (
        <div className="text-center py-12 text-slate-600 text-sm">No dealers yet</div>
      ) : (
        <div className="space-y-4">
          {groupByBrand(list).map(({ brand, dealers }) => (
            <div key={brand} className="bg-[#0b0c0d] rounded-xl border border-[#252628] overflow-hidden">
              <div className="px-4 py-2 bg-[#111213] border-b border-[#252628]">
                <span className="text-xs font-semibold text-slate-500 uppercase tracking-widest">{brand}</span>
              </div>
              <table className="w-full text-sm">
                <tbody className="divide-y divide-[#1e1f20]">
                  {dealers.map((d) => (
                    <tr key={d.id} className="hover:bg-blue-500/5 transition-colors">
                      <td className="px-4 py-3.5 font-medium text-slate-200">{d.name}</td>
                      <td className="px-4 py-3.5 text-slate-400">{d.contact_person ?? "–"}</td>
                      <td className="px-4 py-3.5 text-slate-400">{d.phone ?? "–"}</td>
                      <td className="px-4 py-3.5 text-slate-400 truncate max-w-xs">{d.address ?? "–"}</td>
                      <td className="px-4 py-3.5 text-right">
                        <div className="flex items-center justify-end gap-3">
                          <button onClick={() => openEdit(d)} className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors">Edit</button>
                          <button onClick={() => handleDelete(d)} className="text-slate-600 hover:text-red-400 text-sm font-medium transition-colors">Delete</button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ))}
        </div>
      )}

      {modal.open && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4" onClick={() => setModal({ open: false })}>
          <div className="bg-[#0b0c0d] rounded-2xl w-full max-w-md border border-[#252628] shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between px-6 py-4 border-b border-[#252628]">
              <h2 className="text-base font-semibold text-white">
                {modal.dealer ? "Edit Dealer" : "New Dealer"}
              </h2>
              <button onClick={() => setModal({ open: false })} className="text-slate-600 hover:text-slate-300 text-xl transition-colors">✕</button>
            </div>
            <form onSubmit={handleSave} className="p-6 space-y-4">
              <FormField label="Dealership Name *" value={form.name} onChange={(v) => setForm((p) => ({ ...p, name: v }))} required />
              <FormField label="Contact Person" value={form.contact_person} onChange={(v) => setForm((p) => ({ ...p, contact_person: v }))} />
              <FormField label="Contact Email" value={form.contact_person_email} onChange={(v) => setForm((p) => ({ ...p, contact_person_email: v }))} />
              <FormField label="Phone" value={form.phone} onChange={(v) => setForm((p) => ({ ...p, phone: v }))} />
              <FormField label="Address" value={form.address} onChange={(v) => setForm((p) => ({ ...p, address: v }))} />
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

const BRANDS: { label: string; keywords: string[] }[] = [
  { label: "Acura",          keywords: ["acura"] },
  { label: "Audi",           keywords: ["audi"] },
  { label: "BMW",            keywords: ["bmw"] },
  { label: "Buick",          keywords: ["buick"] },
  { label: "Cadillac",       keywords: ["cadillac"] },
  { label: "Chevrolet",      keywords: ["chevrolet", "chevy"] },
  { label: "Chrysler",       keywords: ["chrysler"] },
  { label: "Dodge",          keywords: ["dodge"] },
  { label: "Ford",           keywords: ["ford"] },
  { label: "Genesis",        keywords: ["genesis"] },
  { label: "GMC",            keywords: ["gmc"] },
  { label: "Honda",          keywords: ["honda"] },
  { label: "Hyundai",        keywords: ["hyundai"] },
  { label: "Infiniti",       keywords: ["infiniti"] },
  { label: "Jeep",           keywords: ["jeep"] },
  { label: "Kia",            keywords: ["kia"] },
  { label: "Lexus",          keywords: ["lexus"] },
  { label: "Lincoln",        keywords: ["lincoln"] },
  { label: "Mazda",          keywords: ["mazda"] },
  { label: "Mercedes-Benz",  keywords: ["mercedes", "mercedes-benz", "benz"] },
  { label: "Nissan",         keywords: ["nissan"] },
  { label: "RAM",            keywords: ["ram"] },
  { label: "Subaru",         keywords: ["subaru"] },
  { label: "Tesla",          keywords: ["tesla"] },
  { label: "Toyota",         keywords: ["toyota"] },
  { label: "Volkswagen",     keywords: ["volkswagen", "vw"] },
  { label: "Volvo",          keywords: ["volvo"] },
];

function extractBrand(name: string): string {
  const lower = name.toLowerCase();
  for (const b of BRANDS) {
    if (b.keywords.some((kw) => lower.includes(kw))) return b.label;
  }
  return "Other";
}

function groupByBrand(dealers: Dealer[]): { brand: string; dealers: Dealer[] }[] {
  const map = new Map<string, Dealer[]>();
  const sorted = [...dealers].sort((a, b) => a.name.localeCompare(b.name));
  for (const d of sorted) {
    const brand = extractBrand(d.name);
    if (!map.has(brand)) map.set(brand, []);
    map.get(brand)!.push(d);
  }
  return Array.from(map.entries())
    .sort(([a], [b]) => a === "Other" ? 1 : b === "Other" ? -1 : a.localeCompare(b))
    .map(([brand, dealers]) => ({ brand, dealers }));
}

function FormField({
  label, value, onChange, required,
}: {
  label: string; value: string; onChange: (v: string) => void; required?: boolean;
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-slate-500 mb-1.5 uppercase tracking-wider">{label}</label>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white placeholder:text-slate-700 focus:outline-none transition-colors"
      />
    </div>
  );
}
