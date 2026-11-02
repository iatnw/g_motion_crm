"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState, useSyncExternalStore } from "react";
import { auth } from "@/lib/api";

const links = [
  {
    href: "/dashboard",
    label: "Dashboard",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <rect x="3" y="3" width="7" height="8" rx="1" />
        <rect x="14" y="3" width="7" height="5" rx="1" />
        <rect x="14" y="12" width="7" height="9" rx="1" />
        <rect x="3" y="15" width="7" height="6" rx="1" />
      </svg>
    ),
  },
  {
    href: "/clients",
    label: "Clients",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
        <circle cx="12" cy="7" r="4" />
      </svg>
    ),
  },
  {
    href: "/deals",
    label: "Deals",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <path d="M8 21h8M12 17v4" />
      </svg>
    ),
  },
  {
    href: "/brokers",
    label: "Brokers",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
      </svg>
    ),
  },
  {
    href: "/referrals",
    label: "Referrals",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M23 11l-4 4-2-2" />
      </svg>
    ),
  },
  {
    href: "/dealers",
    label: "Dealers",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
        <polyline points="9 22 9 12 15 12 15 22" />
      </svg>
    ),
  },
  {
    href: "/settings",
    label: "Settings",
    icon: (
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
      </svg>
    ),
  },
];

function subscribeTheme(callback: () => void) {
  window.addEventListener("crm-theme-change", callback);
  window.addEventListener("storage", callback);
  return () => {
    window.removeEventListener("crm-theme-change", callback);
    window.removeEventListener("storage", callback);
  };
}
function readTheme() {
  return document.documentElement.getAttribute("data-theme") === "light"
    ? "light"
    : "dark";
}

export default function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  const router = useRouter();
  const theme = useSyncExternalStore(subscribeTheme, readTheme, () => "dark");
  const [signingOut, setSigningOut] = useState(false);
  const [error, setError] = useState("");
  function toggleTheme() {
    const next = theme === "dark" ? "light" : "dark";
    document.documentElement.setAttribute("data-theme", next);
    try {
      localStorage.setItem("theme", next);
    } catch {
      /* The current theme still works when storage is unavailable. */
    }
    window.dispatchEvent(new Event("crm-theme-change"));
  }
  async function handleLogout() {
    setSigningOut(true);
    setError("");
    try {
      await auth.logout();
      onNavigate?.();
      router.replace("/login");
    } catch {
      setError("Couldn’t sign out. Please try again.");
      setSigningOut(false);
    }
  }
  function navLink(href: string) {
    const item = links.find((link) => link.href === href)!;
    const active = pathname === href || pathname.startsWith(href + "/");
    return (
      <Link
        key={href}
        href={href}
        onClick={onNavigate}
        aria-current={active ? "page" : undefined}
        className={`crm-nav-link ${active ? "crm-nav-active" : ""}`}
      >
        <span aria-hidden="true">{item.icon}</span>
        {item.label}
        {active && (
          <span
            aria-hidden="true"
            className="ml-auto h-1.5 w-1.5 rounded-full bg-blue-400"
          />
        )}
      </Link>
    );
  }
  return (
    <aside className="crm-sidebar flex h-full w-60 flex-col">
      <Link
        href="/dashboard"
        onClick={onNavigate}
        className="flex items-center gap-3 border-b border-[#252628] px-5 py-6"
      >
        <span
          aria-hidden="true"
          className="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-600 text-sm font-bold text-white"
        >
          GM
        </span>
        <span>
          <span className="block text-sm font-semibold tracking-tight text-white">
            G Motion
          </span>
          <span className="mt-0.5 block text-[11px] text-slate-400">
            Auto finance workspace
          </span>
        </span>
      </Link>
      <nav
        aria-label="Main navigation"
        className="flex-1 space-y-6 overflow-y-auto px-3 py-6"
      >
        <div>
          <p className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-widest text-slate-400">
            Workspace
          </p>
          <div className="space-y-1">
            {["/dashboard", "/deals", "/clients"].map(navLink)}
          </div>
        </div>
        <div>
          <p className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-widest text-slate-400">
            Network
          </p>
          <div className="space-y-1">
            {["/brokers", "/dealers", "/referrals"].map(navLink)}
          </div>
        </div>
      </nav>
      <div className="space-y-1 border-t border-[#252628] p-3">
        {navLink("/settings")}
        <button onClick={toggleTheme} className="crm-nav-link w-full">
          <span aria-hidden="true">{theme === "dark" ? "☀" : "☾"}</span>
          {theme === "dark" ? "Light appearance" : "Dark appearance"}
        </button>
        <button
          onClick={() => void handleLogout()}
          disabled={signingOut}
          className="crm-nav-link w-full disabled:opacity-50"
        >
          <span aria-hidden="true">↪</span>
          {signingOut ? "Signing out…" : "Sign out"}
        </button>
        {error && (
          <p role="alert" className="px-3 py-2 text-xs text-red-400">
            {error}
          </p>
        )}
        <p className="px-3 pt-3 text-[10px] text-slate-400">Owner workspace</p>
      </div>
    </aside>
  );
}
