"use client";

import { useCallback, useEffect, useId, useRef, useState } from "react";

export type StatusOption = { value: string; label: string; description: string; tone: "green" | "amber" | "red" | "slate" | "blue" };
export const paymentOptions: StatusOption[] = [
  { value: "pending", label: "Not paid", description: "Payment is still outstanding", tone: "amber" },
  { value: "paid", label: "Paid", description: "Payment has been completed", tone: "green" },
];
export const referralOptions: StatusOption[] = [
  { value: "none", label: "Not paid", description: "No payout recorded", tone: "slate" },
  { value: "pending", label: "Pending", description: "Payout is outstanding", tone: "amber" },
  { value: "paid", label: "Paid", description: "Payout has been completed", tone: "green" },
];
export const dealOptions: StatusOption[] = [
  { value: "pending", label: "Not funded", description: "Deal is awaiting payment", tone: "amber" },
  { value: "paid", label: "Funded", description: "Deal payment is complete", tone: "green" },
  { value: "cancelled", label: "Cancelled", description: "Deal is no longer proceeding", tone: "red" },
];
export const clientOptions: StatusOption[] = [
  { value: "active", label: "Active", description: "Client is currently working with you", tone: "blue" },
  { value: "closed", label: "Closed", description: "Client workflow is complete", tone: "green" },
  { value: "no_deal", label: "No deal", description: "Client did not proceed with a deal", tone: "slate" },
];

export default function StatusSelect({ value, options, label, detail, disabled = false, fullWidth = false, variant = "status", filterIcon = "calendar", onChange }: {
  value: string;
  options: StatusOption[];
  label: string;
  detail?: string;
  disabled?: boolean;
  fullWidth?: boolean;
  variant?: "status" | "filter";
  filterIcon?: "calendar" | "source" | "sort" | "document";
  onChange: (value: string) => void | Promise<void>;
}) {
  const id = useId();
  const trigger = useRef<HTMLButtonElement>(null);
  const popup = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [position, setPosition] = useState({ left: 12, top: 12 });
  const current = options.find((option) => option.value === value) ?? options[0];
  const place = useCallback(() => {
    if (!trigger.current || !popup.current) return;
    const rect = trigger.current.getBoundingClientRect();
    const width = Math.min(280, window.innerWidth - 24);
    const height = popup.current.offsetHeight || 240;
    setPosition({
      left: Math.max(12, Math.min(rect.right - width, window.innerWidth - width - 12)),
      top: Math.max(12, Math.min(rect.bottom + 7 + height <= window.innerHeight - 12 ? rect.bottom + 7 : rect.top - height - 7, window.innerHeight - height - 12)),
    });
  }, []);
  useEffect(() => {
    if (!open) return;
    window.addEventListener("resize", place);
    document.addEventListener("scroll", place, true);
    return () => { window.removeEventListener("resize", place); document.removeEventListener("scroll", place, true); };
  }, [open, place]);
  function close() { popup.current?.hidePopover(); trigger.current?.focus(); }
  async function choose(next: string) {
    if (saving) return;
    if (next === value) { close(); return; }
    setSaving(true);
    setError("");
    try { await onChange(next); close(); }
    catch { setError("Couldn’t update the status. Please try again."); }
    finally {
      setSaving(false);
      requestAnimationFrame(() => {
        if (popup.current?.matches(":popover-open")) popup.current.querySelector<HTMLButtonElement>('[aria-selected="true"]')?.focus();
        else trigger.current?.focus();
      });
    }
  }
  return (
    <div className={`crm-status-control ${variant === "filter" ? "crm-status-filter" : ""} ${fullWidth ? "crm-status-full" : ""}`}>
      <button ref={trigger} type="button" popoverTarget={id} disabled={disabled || saving}
        className={`crm-status-trigger crm-status-${current.tone}`} aria-label={`${label}: ${current.label}`} aria-haspopup="listbox" aria-expanded={open} aria-controls={`${id}-options`}
        onClick={(event) => { event.stopPropagation(); place(); }}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown" || event.key === "ArrowUp") {
            event.preventDefault(); popup.current?.showPopover(); place();
          }
        }}>
        {variant === "filter" ? (
          <svg aria-hidden="true" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
            {filterIcon === "sort" ? <><path d="M8 4v16m-3-3 3 3 3-3M14 5h6M14 10h4M14 15h2" /></> : filterIcon === "source" ? <><circle cx="8" cy="7" r="3" /><path d="M2 20v-2a6 6 0 0 1 12 0v2M16 10h6m-3-3 3 3-3 3" /></> : filterIcon === "document" ? <><path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9zM14 3v6h6M8 13h8M8 17h5" /></> : <><rect x="3" y="5" width="18" height="16" rx="2" /><path d="M16 3v4M8 3v4M3 11h18" /></>}
          </svg>
        ) : <span className="crm-status-dot" aria-hidden="true" />}
        <span className="crm-status-label">{saving ? "Saving…" : current.label}</span>
        {detail && <span className="crm-status-detail">{detail}</span>}
        <svg className="crm-status-chevron" aria-hidden="true" width="13" height="13" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.7"><path d="m5 7.5 5 5 5-5" /></svg>
      </button>
      <div ref={popup} id={id} popover="auto" className="crm-status-popup" style={{ left: position.left, top: position.top }}
        onClick={(event) => event.stopPropagation()}
        onToggle={(event) => {
          const visible = event.newState === "open"; setOpen(visible);
          if (visible) { place(); setError(""); popup.current?.querySelector<HTMLButtonElement>('[aria-selected="true"]')?.focus(); }
        }}
        onKeyDown={(event) => {
          event.stopPropagation();
          if (event.key === "Escape") { event.preventDefault(); close(); return; }
          const controls = Array.from(popup.current?.querySelectorAll<HTMLButtonElement>('[role="option"]') ?? []);
          const index = controls.indexOf(document.activeElement as HTMLButtonElement);
          let next = index;
          if (event.key === "ArrowDown") next = (index + 1) % controls.length;
          else if (event.key === "ArrowUp") next = (index - 1 + controls.length) % controls.length;
          else if (event.key === "Home") next = 0;
          else if (event.key === "End") next = controls.length - 1;
          else return;
          event.preventDefault(); controls[next]?.focus();
        }}>
        <p className="crm-status-popup-title">{label}</p>
        <div role="listbox" id={`${id}-options`} aria-label={label} aria-busy={saving}>
          {options.map((option) => (
            <button key={option.value} type="button" role="option" aria-selected={value === option.value} disabled={saving}
              className="crm-status-option" onClick={() => void choose(option.value)}>
              <span className={`crm-status-dot crm-status-${option.tone}`} aria-hidden="true" />
              <span><strong>{option.label}</strong><small>{option.description}</small></span>
              {value === option.value && <span className="crm-status-check" aria-hidden="true">✓</span>}
            </button>
          ))}
        </div>
        {error && <p className="crm-status-error" role="alert">{error}</p>}
      </div>
    </div>
  );
}
