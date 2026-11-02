"use client";

import { useEffect, useRef, useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import Sidebar from "@/components/Sidebar";

const pages = [
  {
    href: "/dashboard",
    label: "Dashboard",
    description: "Revenue, collections, and performance",
    group: "Workspace",
  },
  {
    href: "/deals",
    label: "Deals",
    description: "Vehicles, payments, cash deals, and leases",
    group: "Workspace",
  },
  {
    href: "/clients",
    label: "Clients",
    description: "Contacts, profiles, and documents",
    group: "Workspace",
  },
  {
    href: "/brokers",
    label: "Brokers",
    description: "Partners and commissions",
    group: "Network",
  },
  {
    href: "/dealers",
    label: "Dealers",
    description: "Dealerships and contacts",
    group: "Network",
  },
  {
    href: "/referrals",
    label: "Referrals",
    description: "Referrers and payouts",
    group: "Network",
  },
  {
    href: "/settings",
    label: "Settings",
    description: "Manage your account",
    group: "Account",
  },
  {
    href: "/clients/new",
    label: "New client",
    description: "Create a client profile",
    group: "Quick actions",
  },
  {
    href: "/deals?new=1",
    label: "New deal",
    description: "Start a new deal",
    group: "Quick actions",
  },
];

export default function WorkspaceShell({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const section = pages.find(
    (page) => pathname === page.href || pathname.startsWith(page.href + "/"),
  );
  const mobile = useRef<HTMLDialogElement>(null);
  const switcher = useRef<HTMLDialogElement>(null);
  const [query, setQuery] = useState("");
  function openSwitcher() {
    setQuery("");
    switcher.current?.showModal();
  }
  function closeNavigation() {
    mobile.current?.close();
  }
  useEffect(() => {
    const handleKey = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        if (switcher.current?.open) switcher.current.close();
        else openSwitcher();
      }
    };
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  }, []);
  const results = pages.filter((page) =>
    `${page.label} ${page.description} ${page.group}`
      .toLowerCase()
      .includes(query.trim().toLowerCase()),
  );
  return (
    <div className="crm-shell flex h-full">
      <a href="#workspace-main" className="crm-skip-link">
        Skip to content
      </a>
      <div className="hidden h-full shrink-0 md:block">
        <Sidebar />
      </div>
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="crm-topbar flex h-[72px] shrink-0 items-center justify-between gap-4 border-b border-[#252628] px-4 sm:px-6 lg:px-8">
          <div className="flex min-w-0 items-center gap-3">
            <button
              onClick={() => mobile.current?.showModal()}
              aria-label="Open navigation"
              aria-haspopup="dialog"
              className="crm-icon-button md:hidden"
            >
              <svg
                aria-hidden="true"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.8"
              >
                <path d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            <div className="flex items-center gap-2 text-sm">
              <span className="hidden text-slate-400 sm:inline">Workspace</span>
              <span
                aria-hidden="true"
                className="hidden text-slate-500 sm:inline"
              >
                /
              </span>
              <span className="truncate font-medium text-slate-200">
                {section?.label ?? "G Motion"}
              </span>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={openSwitcher}
              aria-haspopup="dialog"
              className="flex items-center gap-2 rounded-lg border border-[#252628] bg-[#0b0c0d] px-3 py-2 text-sm text-slate-400 transition-colors hover:border-blue-500/40 hover:text-slate-200"
            >
              <svg
                aria-hidden="true"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.7"
              >
                <circle cx="10.5" cy="10.5" r="7" />
                <path d="m16 16 4.5 4.5" />
              </svg>
              <span className="hidden sm:inline">Go to a page…</span>
              <span className="sm:hidden">Go to</span>
              <kbd className="hidden rounded border border-[#252628] px-1.5 text-[10px] lg:inline">
                ⌘ / Ctrl K
              </kbd>
            </button>
            <span
              aria-label="Owner workspace"
              className="hidden h-8 w-8 items-center justify-center rounded-full border border-blue-500/20 bg-blue-500/10 text-xs font-semibold text-blue-400 sm:flex"
            >
              GM
            </span>
          </div>
        </header>
        <main
          id="workspace-main"
          tabIndex={-1}
          className="crm-content min-w-0 flex-1 overflow-y-auto"
        >
          {children}
        </main>
      </div>
      <dialog
        ref={mobile}
        aria-label="Workspace navigation"
        className="crm-mobile-navigation"
        onClick={(event) => {
          if (event.target === event.currentTarget) closeNavigation();
        }}
      >
        <button
          onClick={closeNavigation}
          aria-label="Close navigation"
          className="crm-icon-button absolute right-3 top-3"
        >
          ×
        </button>
        <Sidebar onNavigate={closeNavigation} />
      </dialog>
      <dialog
        ref={switcher}
        aria-labelledby="page-switcher-title"
        className="crm-page-switcher"
        onClick={(event) => {
          if (event.target === event.currentTarget) switcher.current?.close();
        }}
      >
        <div className="flex items-center justify-between border-b border-[#252628] px-5 py-4">
          <h2
            id="page-switcher-title"
            className="text-sm font-semibold text-slate-200"
          >
            Go to a page
          </h2>
          <button
            onClick={() => switcher.current?.close()}
            aria-label="Close page switcher"
            className="crm-icon-button"
          >
            ×
          </button>
        </div>
        <div className="p-4">
          <input
            autoFocus
            aria-label="Find a page or action"
            placeholder="Find a page or action…"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            className="w-full rounded-lg border border-[#252628] bg-[#0b0c0d] px-3 py-3 text-sm text-slate-200"
          />
        </div>
        <div className="max-h-[55vh] overflow-y-auto px-3 pb-3">
          {results.length ? (
            results.map((page) => (
              <Link
                key={page.href}
                href={page.href}
                onClick={() => switcher.current?.close()}
                className="flex items-center justify-between gap-3 rounded-lg px-3 py-3 hover:bg-blue-500/10"
              >
                <span>
                  <span className="block text-sm font-medium text-slate-200">
                    {page.label}
                  </span>
                  <span className="mt-1 block text-xs text-slate-400">
                    {page.description}
                  </span>
                </span>
                <span className="text-[10px] text-slate-400">{page.group}</span>
              </Link>
            ))
          ) : (
            <p className="px-3 py-6 text-sm text-slate-400">
              No matching pages. Try “deals” or “clients”.
            </p>
          )}
        </div>
        <p className="border-t border-[#252628] px-5 py-3 text-[11px] text-slate-400">
          Tab to move · Enter to open · Esc to close
        </p>
      </dialog>
    </div>
  );
}
