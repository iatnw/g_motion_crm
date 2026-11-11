"use client";

import { useEffect, useId, useRef, type ReactNode } from "react";

export default function Modal({ title, description, children, onClose, size = "medium", busy = false }: {
  title: string;
  description?: string;
  children: ReactNode;
  onClose: () => void;
  size?: "medium" | "large" | "wide";
  busy?: boolean;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  const descriptionId = useId();
  useEffect(() => {
    const element = dialog.current;
    element?.showModal();
    return () => element?.close();
  }, []);
  return (
    <dialog ref={dialog} className={`crm-modal crm-modal-${size}`} aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined} aria-busy={busy}
      onKeyDown={(event) => {
        if (event.key !== "Tab") return;
        event.stopPropagation();
        const controls = Array.from(event.currentTarget.querySelectorAll<HTMLElement>(
          'button:not(:disabled), input:not(:disabled), select:not(:disabled), textarea:not(:disabled), a[href], [tabindex="0"]',
        )).filter((element) => element.getClientRects().length > 0);
        const first = controls[0];
        const last = controls[controls.length - 1];
        if (event.shiftKey && document.activeElement === first) {
          event.preventDefault();
          last?.focus();
        } else if (!event.shiftKey && document.activeElement === last) {
          event.preventDefault();
          first?.focus();
        }
      }}
      onCancel={(event) => { event.preventDefault(); if (!busy) onClose(); }}>
      <div className="crm-modal-frame">
        <header className="crm-modal-header">
          <div>
            <h2 id={titleId}>{title}</h2>
            {description && <p id={descriptionId}>{description}</p>}
          </div>
          <button type="button" className="crm-modal-close" onClick={onClose} disabled={busy} aria-label="Close dialog">
            <svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round"><path d="m6 6 12 12M18 6 6 18" /></svg>
          </button>
        </header>
        <div className="crm-modal-body">{children}</div>
      </div>
    </dialog>
  );
}
