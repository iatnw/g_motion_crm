"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { clients } from "@/lib/api";
import type { Client, ClientStatus } from "@/lib/types";

const statusLabel: Record<ClientStatus, string> = {
  active: "Active",
  closed: "Closed",
  no_deal: "No Deal",
};

const statusStyle: Record<ClientStatus, string> = {
  active: "bg-green-500/10 text-green-400 border border-green-500/25",
  closed: "bg-slate-500/10 text-slate-400 border border-slate-500/25",
  no_deal: "bg-red-500/10 text-red-400 border border-red-500/25",
};

function isNew(createdAt: string) {
  return Date.now() - new Date(createdAt).getTime() < 48 * 60 * 60 * 1000;
}

export default function ClientsPage() {
  const [all, setAll] = useState<Client[]>([]);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<ClientStatus | "all">("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    clients
      .list()
      .then((data) => setAll(data ?? []))
      .catch(() => setError("Failed to load clients"))
      .finally(() => setLoading(false));
  }, []);

  const filtered = all.filter((c) => {
    const q = search.toLowerCase();
    const matchSearch =
      !q ||
      c.first_name.toLowerCase().includes(q) ||
      c.last_name.toLowerCase().includes(q) ||
      c.phone.includes(q) ||
      c.email.toLowerCase().includes(q);
    const matchFilter = filter === "all" || c.status === filter;
    return matchSearch && matchFilter;
  });

  return (
    <div className="p-6 max-w-6xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-white">Clients</h1>
          <p className="text-slate-500 text-xs mt-0.5">{all.length} total</p>
        </div>
        <Link
          href="/clients/new"
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition-all glow-blue-sm hover:glow-blue"
        >
          + New Client
        </Link>
      </div>

      {/* Filters */}
      <div className="flex gap-3 mb-5">
        <div className="relative flex-1">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
          </svg>
          <input
            type="text"
            placeholder="Search by name, phone, email…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg pl-9 pr-3 py-2 text-sm text-white placeholder:text-slate-700 focus:outline-none transition-colors"
          />
        </div>
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value as ClientStatus | "all")}
          className="bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-slate-300 focus:outline-none transition-colors"
        >
          <option value="all">All Statuses</option>
          <option value="active">Active</option>
          <option value="closed">Closed</option>
          <option value="no_deal">No Deal</option>
        </select>
      </div>

      {loading && (
        <div className="text-center py-12 text-slate-600 text-sm">Loading…</div>
      )}
      {error && <p className="text-red-400 text-sm">{error}</p>}

      {!loading && !error && (
        <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] overflow-hidden">
          {filtered.length === 0 ? (
            <div className="text-center py-12 text-slate-600 text-sm">No clients found</div>
          ) : (
            <table className="w-full text-sm">
              <thead className="border-b border-[#252628]">
                <tr>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Name</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Phone</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Email</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Status</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-slate-600 uppercase tracking-wider">Added</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1e1f20]">
                {filtered.map((c) => (
                  <tr key={c.id} className="hover:bg-blue-500/5 transition-colors cursor-pointer group">
                    <td className="px-4 py-3.5">
                      <Link href={`/clients/${c.id}`} className="flex items-center gap-2">
                        <span className="font-medium text-slate-200 group-hover:text-white transition-colors">
                          {c.first_name} {c.last_name}
                        </span>
                        {isNew(c.created_at) && (
                          <span className="text-xs bg-blue-500/15 text-blue-400 border border-blue-500/25 px-1.5 py-0.5 rounded font-medium">
                            New
                          </span>
                        )}
                      </Link>
                    </td>
                    <td className="px-4 py-3.5 text-slate-400">{c.phone}</td>
                    <td className="px-4 py-3.5 text-slate-400">{c.email}</td>
                    <td className="px-4 py-3.5">
                      <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${statusStyle[c.status]}`}>
                        {statusLabel[c.status]}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-slate-600">
                      {new Date(c.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
}
