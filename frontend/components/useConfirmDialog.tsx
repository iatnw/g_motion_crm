"use client";

import { useEffect, useRef, useState } from "react";
import Modal from "@/components/Modal";

export default function useConfirmDialog() {
  const [message, setMessage] = useState<string | null>(null);
  const [title, setTitle] = useState("Delete this record?");
  const [actionLabel, setActionLabel] = useState("Delete record");
  const resolve = useRef<((confirmed: boolean) => void) | null>(null);
  useEffect(() => () => { resolve.current?.(false); }, []);
  function finish(confirmed: boolean) {
    resolve.current?.(confirmed);
    resolve.current = null;
    setMessage(null);
  }
  function confirmAction(description: string, title = "Delete this record?", actionLabel = "Delete record"): Promise<boolean> {
    setTitle(title);
    setActionLabel(actionLabel);
    resolve.current?.(false);
    return new Promise((done) => {
      resolve.current = done;
      setMessage(description);
    });
  }
  const confirmationDialog = message === null ? null : (
    <Modal title={title} description={message} onClose={() => finish(false)}>
      <div className="crm-confirm-content">
        <div className="crm-confirm-symbol" aria-hidden="true">!</div>
        <p>Review the details before continuing.</p>
      </div>
      <div className="crm-modal-footer crm-confirm-footer">
        <button type="button" autoFocus className="crm-secondary-button" onClick={() => finish(false)}>Cancel</button>
        <button type="button" className="crm-danger-button" onClick={() => finish(true)}>{actionLabel}</button>
      </div>
    </Modal>
  );
  return { confirmAction, confirmationDialog };
}
