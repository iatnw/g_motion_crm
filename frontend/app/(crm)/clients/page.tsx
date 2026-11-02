"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { clients } from "@/lib/api";
import type { Client, ClientStatus } from "@/lib/types";
import {
  ClientAvatar,
  ClientStatusBadge,
  clientSourceLabels,
  clientStatusLabels,
} from "@/components/ClientIdentity";
import { fmtDate } from "@/lib/utils";

const pageSize = 20;
type StatusFilter = "all" | ClientStatus;

export default function ClientsPage() {
  const [all, setAll] = useState<Client[]>([]);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<StatusFilter>("all");
  const [source, setSource] = useState("");
  const [sort, setSort] = useState("newest");
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setAll((await clients.list()) ?? []);
    } catch {
      setError("We couldn’t load your clients. Please try again.");
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    void load();
  }, [load]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    const phoneQuery = q.replace(/\D/g, "");
    return all
      .filter((c) => {
        const fullName = `${c.first_name} ${c.last_name}`.toLowerCase();
        const matches =
          !q ||
          fullName.includes(q) ||
          c.email?.toLowerCase().includes(q) ||
          c.phone?.toLowerCase().includes(q) ||
          (/^[\d\s()+.\-]+$/.test(q) &&
            phoneQuery.length > 0 &&
            c.phone?.replace(/\D/g, "").includes(phoneQuery));
        return (
          matches &&
          (status === "all" || c.status === status) &&
          (!source || c.source === source)
        );
      })
      .sort((a, b) =>
        sort === "name"
          ? `${a.first_name} ${a.last_name}`.localeCompare(
              `${b.first_name} ${b.last_name}`,
            )
          : sort === "oldest"
            ? Date.parse(a.created_at) - Date.parse(b.created_at)
            : Date.parse(b.created_at) - Date.parse(a.created_at),
      );
  }, [all, search, status, source, sort]);
  const pages = Math.max(1, Math.ceil(filtered.length / pageSize));
  const currentPage = Math.min(page, pages);
  const visible = filtered.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize,
  );
  const hasFilters = !!search || status !== "all" || !!source;
  function clearFilters() {
    setSearch("");
    setStatus("all");
    setSource("");
    setPage(1);
  }
  const control =
    "rounded-lg border border-[#252628] bg-[#0b0c0d] px-3 py-2.5 text-sm text-slate-300 focus-visible:outline-2 focus-visible:outline-blue-400";

  return (
    <div className="mx-auto max-w-7xl p-4 normal-case sm:p-6 lg:p-8">
      <div className="mb-7 flex flex-wrap items-center justify-between gap-4">
        <div>
          <p className="mb-1 text-xs font-medium uppercase tracking-widest text-blue-400">
            Your network
          </p>
          <h1 className="text-2xl font-semibold tracking-tight text-white">
            Clients
          </h1>
          <p className="mt-1 text-sm text-slate-400">
            Keep every relationship, contact, and next step in view.
          </p>
        </div>
        <Link
          href="/clients/new"
          className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400"
        >
          <span aria-hidden="true">＋</span> New client
        </Link>
      </div>

      <div
        className="mb-6 grid grid-cols-2 gap-3 lg:grid-cols-4"
        aria-label="Client status filters"
      >
        {(["all", "active", "closed", "no_deal"] as StatusFilter[]).map(
          (value) => (
            <button
              key={value}
              aria-pressed={status === value}
              onClick={() => {
                setStatus(value);
                setPage(1);
              }}
              className={`rounded-xl border p-4 text-left transition-colors focus-visible:outline-2 focus-visible:outline-blue-400 ${status === value ? "border-blue-500/50 bg-blue-500/10" : "border-[#252628] bg-[#0f1011] hover:border-blue-500/30"}`}
            >
              <span className="text-xs font-medium text-slate-400">
                {value === "all" ? "All clients" : clientStatusLabels[value]}
              </span>
              <span className="mt-2 block text-2xl font-semibold tabular-nums text-white">
                {loading
                  ? "—"
                  : value === "all"
                    ? all.length
                    : all.filter((c) => c.status === value).length}
              </span>
            </button>
          ),
        )}
      </div>

      <section
        aria-label="Client directory"
        className="overflow-hidden rounded-xl border border-[#252628] bg-[#0f1011]"
      >
        <div className="flex flex-col gap-3 border-b border-[#252628] p-4 lg:flex-row">
          <div className="relative flex-1">
            <svg
              aria-hidden="true"
              className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
              width="17"
              height="17"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.7"
            >
              <circle cx="10.5" cy="10.5" r="7" />
              <path d="m16 16 4.5 4.5" />
            </svg>
            <input
              type="search"
              aria-label="Search clients"
              placeholder="Search name, phone, or email"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              className={`${control} w-full pl-10 placeholder:text-slate-500`}
            />
          </div>
          <div className="flex flex-wrap gap-2">
            <select
              aria-label="Filter by source"
              value={source}
              onChange={(e) => {
                setSource(e.target.value);
                setPage(1);
              }}
              className={`${control} flex-1`}
            >
              <option value="">All sources</option>
              {Object.entries(clientSourceLabels).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
            <select
              aria-label="Sort clients"
              value={sort}
              onChange={(e) => {
                setSort(e.target.value);
                setPage(1);
              }}
              className={`${control} flex-1`}
            >
              <option value="newest">Newest first</option>
              <option value="oldest">Oldest first</option>
              <option value="name">Name A–Z</option>
            </select>
          </div>
        </div>
        <div
          className="flex items-center justify-between gap-3 px-4 py-3 text-xs text-slate-400"
          aria-live="polite"
        >
          <span>
            {loading
              ? "Loading directory…"
              : `${filtered.length} ${filtered.length === 1 ? "client" : "clients"}${hasFilters ? " match your filters" : " in your directory"}`}
          </span>
          {hasFilters && (
            <button
              onClick={clearFilters}
              className="font-medium text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-blue-400"
            >
              Clear filters
            </button>
          )}
        </div>
        {loading ? (
          <div role="status" className="space-y-3 px-4 pb-5">
            {[0, 1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-16 animate-pulse rounded-lg bg-[#1a1b1d]"
              />
            ))}
            <span className="sr-only">Loading clients</span>
          </div>
        ) : error ? (
          <div role="alert" className="px-6 py-12 text-center">
            <p className="text-sm text-red-400">{error}</p>
            <button
              onClick={() => void load()}
              className="mt-4 text-sm text-blue-400 hover:underline"
            >
              Try again
            </button>
          </div>
        ) : filtered.length === 0 ? (
          <div className="px-6 py-14 text-center">
            <div aria-hidden="true" className="mb-3 text-3xl text-slate-500">
              ◎
            </div>
            <h2 className="font-medium text-white">
              {all.length
                ? "No matching clients"
                : "Your client directory starts here"}
            </h2>
            <p className="mt-2 text-sm text-slate-400">
              {all.length
                ? "Try a different name or clear your filters."
                : "Add your first client to keep their details and deals together."}
            </p>
            {all.length ? (
              <button
                onClick={clearFilters}
                className="mt-4 text-sm font-medium text-blue-400 hover:underline"
              >
                Reset filters
              </button>
            ) : (
              <Link
                href="/clients/new"
                className="mt-4 inline-block text-sm font-medium text-blue-400 hover:underline"
              >
                Add a client →
              </Link>
            )}
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full text-left text-sm">
                <thead className="border-y border-[#252628] bg-[#0b0c0d] text-xs text-slate-400">
                  <tr>
                    {["Client", "Contact", "Status", "Source", "Added", ""].map(
                      (title, i) => (
                        <th
                          key={i}
                          scope="col"
                          className="px-5 py-3 font-medium"
                        >
                          {title || (
                            <span className="sr-only">Open profile</span>
                          )}
                        </th>
                      ),
                    )}
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#252628]">
                  {visible.map((c) => (
                    <tr
                      key={c.id}
                      className="group transition-colors hover:bg-blue-500/5"
                    >
                      <td className="px-5 py-4">
                        <Link
                          href={`/clients/${c.id}`}
                          className="flex items-center gap-3 rounded-lg focus-visible:outline-2 focus-visible:outline-blue-400"
                        >
                          <ClientAvatar client={c} />
                          <span>
                            <span className="block font-medium text-slate-200 group-hover:text-blue-400">
                              {c.first_name} {c.last_name}
                            </span>
                            <span className="mt-0.5 block text-xs text-slate-400">
                              {[c.city, c.state].filter(Boolean).join(", ") ||
                                "Location not added"}
                            </span>
                          </span>
                        </Link>
                      </td>
                      <td className="px-5 py-4">
                        <ContactLinks client={c} />
                      </td>
                      <td className="px-5 py-4">
                        <ClientStatusBadge status={c.status} />
                      </td>
                      <td className="px-5 py-4 text-slate-400">
                        {c.source ? clientSourceLabels[c.source] : "—"}
                      </td>
                      <td className="whitespace-nowrap px-5 py-4 text-slate-400">
                        {fmtDate(c.created_at)}
                      </td>
                      <td className="px-5 py-4">
                        <Link
                          aria-label={`Open ${c.first_name} ${c.last_name}'s profile`}
                          href={`/clients/${c.id}`}
                          className="rounded p-2 text-slate-400 hover:text-blue-400 focus-visible:outline-2 focus-visible:outline-blue-400"
                        >
                          ↗
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="divide-y divide-[#252628] md:hidden">
              {visible.map((c) => (
                <article key={c.id} className="p-4">
                  <div className="mb-3 flex items-center justify-between gap-3">
                    <Link
                      href={`/clients/${c.id}`}
                      className="flex min-w-0 items-center gap-3"
                    >
                      <ClientAvatar client={c} />
                      <span className="min-w-0">
                        <span className="block truncate font-medium text-slate-200">
                          {c.first_name} {c.last_name}
                        </span>
                        <span className="block text-xs text-slate-400">
                          {c.source
                            ? clientSourceLabels[c.source]
                            : "No source added"}
                        </span>
                      </span>
                    </Link>
                    <ClientStatusBadge status={c.status} />
                  </div>
                  <ContactLinks client={c} />
                  <div className="mt-3 flex justify-between text-xs text-slate-400">
                    <span>Added {fmtDate(c.created_at)}</span>
                    <Link href={`/clients/${c.id}`} className="text-blue-400">
                      View profile →
                    </Link>
                  </div>
                </article>
              ))}
            </div>
            <div className="flex flex-wrap items-center justify-between gap-3 border-t border-[#252628] px-4 py-4 text-xs text-slate-400">
              <span>
                Showing {(currentPage - 1) * pageSize + 1}–
                {Math.min(currentPage * pageSize, filtered.length)} of{" "}
                {filtered.length}
              </span>
              <div className="flex items-center gap-3">
                <button
                  aria-label="Previous page"
                  disabled={currentPage === 1}
                  onClick={() => setPage(currentPage - 1)}
                  className="rounded-lg border border-[#252628] px-3 py-2 hover:text-white disabled:cursor-not-allowed disabled:opacity-40"
                >
                  ← Previous
                </button>
                <span>
                  {currentPage} / {pages}
                </span>
                <button
                  aria-label="Next page"
                  disabled={currentPage === pages}
                  onClick={() => setPage(currentPage + 1)}
                  className="rounded-lg border border-[#252628] px-3 py-2 hover:text-white disabled:cursor-not-allowed disabled:opacity-40"
                >
                  Next →
                </button>
              </div>
            </div>
          </>
        )}
      </section>
    </div>
  );
}

function ContactLinks({ client }: { client: Client }) {
  return (
    <div className="space-y-1 text-sm">
      {client.email && (
        <a
          className="block break-all text-slate-300 hover:text-blue-400 hover:underline"
          href={`mailto:${client.email}`}
        >
          {client.email}
        </a>
      )}
      {client.phone && (
        <a
          className="block text-slate-400 hover:text-blue-400 hover:underline"
          href={`tel:${client.phone.replace(/[^+\d]/g, "")}`}
        >
          {client.phone}
        </a>
      )}
      {!client.email && !client.phone && (
        <span className="text-slate-400">No contact details</span>
      )}
    </div>
  );
}
