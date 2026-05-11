"use client";

import { useEffect, useState } from "react";
import { referrals as referralsApi } from "@/lib/api";
import type { Referral } from "@/lib/types";

export default function ReferralsPage() {
  const [list, setList] = useState<Referral[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    referralsApi.list().then((data) => setList(data ?? [])).finally(() => setLoading(false));
  }, []);

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-white">Referrals</h1>
        <p className="text-sm text-slate-500 mt-0.5">{list.length} total</p>
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
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${
                      ref.status === "paid"
                        ? "bg-green-500/10 border-green-500/25 text-green-400"
                        : ref.status === "pending"
                        ? "bg-[#0a0a1e] border-[#1e1e45] text-slate-400"
                        : "bg-[#111213] border-[#252628] text-slate-600"
                    }`}>
                      {ref.status === "paid" ? "Paid" : ref.status === "pending" ? "Pending" : "None"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-500">
                    {ref.paid_date ? new Date(ref.paid_date).toLocaleDateString() : "—"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
