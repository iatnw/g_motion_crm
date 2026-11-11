"use client";

import { useEffect, useState, useCallback } from "react";
import Modal from "@/components/Modal";
import StatusSelect, { type StatusOption } from "@/components/StatusSelect";
import useConfirmDialog from "@/components/useConfirmDialog";
import { documents } from "@/lib/api";
import type { Document, DocumentType } from "@/lib/types";

const docTypeLabels: Record<DocumentType, string> = {
  credit_app_pdf: "Credit App PDF",
  driver_license: "Driver License",
  insurance: "Insurance Card",
  broker_agreement: "Broker Agreement",
  dealer_docs: "Dealer Docs",
  check_copy: "Check Copy",
  other: "Other",
};

const docTypeDescriptions: Record<DocumentType, string> = {
  credit_app_pdf: "Credit application and finance intake",
  driver_license: "Client identification document",
  insurance: "Vehicle insurance documentation",
  broker_agreement: "Signed agreement with a broker",
  dealer_docs: "Dealership paperwork and records",
  check_copy: "Copy of a payment check",
  other: "Any supporting document",
};
const docTypeOptions: StatusOption[] = Object.entries(docTypeLabels).map(([value, label]) => ({
  value, label, description: docTypeDescriptions[value as DocumentType], tone: "blue",
}));

const docTypeStyle: Record<DocumentType, string> = {
  credit_app_pdf: "bg-purple-500/10 text-purple-400 border border-purple-500/25",
  driver_license: "bg-blue-500/10 text-blue-400 border border-blue-500/25",
  insurance: "bg-teal-500/10 text-teal-400 border border-teal-500/25",
  broker_agreement: "bg-orange-500/10 text-orange-400 border border-orange-500/25",
  dealer_docs: "bg-yellow-500/10 text-yellow-400 border border-yellow-500/25",
  check_copy: "bg-pink-500/10 text-pink-400 border border-pink-500/25",
  other: "bg-slate-500/10 text-slate-200 border border-slate-500/25",
};

function formatBytes(b?: number) {
  if (!b) return "";
  if (b < 1024) return `${b} B`;
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
  return `${(b / (1024 * 1024)).toFixed(1)} MB`;
}

function isImage(fileName: string) {
  return /\.(jpe?g|png|webp|gif|heic|heif)$/i.test(fileName);
}

export default function DocumentsTab({ clientId, dealId }: { clientId?: string; dealId?: string }) {
  const { confirmAction, confirmationDialog } = useConfirmDialog();
  const [docs, setDocs] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [uploadPct, setUploadPct] = useState(0);
  const [selectedType, setSelectedType] = useState<DocumentType>("other");
  const [error, setError] = useState("");
  const [preview, setPreview] = useState<{ url: string; name: string } | null>(null);

  const load = useCallback(async () => {
    try {
      const list = clientId
        ? await documents.listByClient(clientId)
        : await documents.listByDeal(dealId!);
      setDocs(list);
    } catch {
      setError("Failed to load documents");
    } finally {
      setLoading(false);
    }
  }, [clientId, dealId]);

  useEffect(() => { load(); }, [load]);

  async function handleOpen(doc: Document) {
    try {
      const { url } = await documents.getURL(doc.id);
      if (isImage(doc.file_name)) {
        setPreview({ url, name: doc.file_name });
      } else {
        window.open(url, "_blank");
      }
    } catch {
      setError("Couldn’t open the document. Please try again.");
    }
  }

  async function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setUploadPct(0);
    setError("");
    try {
      let doc: Document;
      if (clientId) {
        doc = await documents.uploadForClient(clientId, file, selectedType, setUploadPct);
      } else {
        doc = await documents.uploadForDeal(dealId!, file, selectedType, setUploadPct);
      }
      setDocs((prev) => [doc, ...prev]);
    } catch {
      setError("Upload failed");
    } finally {
      setUploading(false);
      e.target.value = "";
    }
  }

  async function handleDelete(id: string) {
    if (!(await confirmAction("This document will be permanently deleted.", "Delete document?", "Delete document"))) return;
    try {
      await documents.delete(id);
      setDocs((prev) => prev.filter((d) => d.id !== id));
    } catch {
      setError("Couldn’t delete the document. Please try again.");
    }
  }

  return (
    <>
    <div className="space-y-4">
      {/* Upload */}
      <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] p-4">
        <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3">Upload Document</h3>
        <div className="flex gap-3 items-center flex-wrap">
          <StatusSelect label="Document type" value={selectedType} options={docTypeOptions}
            variant="filter" filterIcon="document" disabled={uploading}
            onChange={(value) => setSelectedType(value as DocumentType)} />
          <label className={`cursor-pointer px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${uploading ? "bg-slate-800 text-slate-600 cursor-not-allowed" : "bg-blue-600 hover:bg-blue-500 text-white glow-blue-sm"}`}>
            {uploading ? `Uploading… ${uploadPct}%` : "Choose File"}
            <input
              type="file"
              className="hidden"
              disabled={uploading}
              onChange={handleUpload}
              accept="image/*,.dng,.raw,.cr2,.cr3,.arw,.nef,.orf,.raf,application/pdf,.doc,.docx"
            />
          </label>
        </div>
        {uploading && (
          <div className="mt-3 h-1.5 bg-[#1e1f20] rounded-full overflow-hidden">
            <div className="h-full bg-blue-500 transition-all duration-300 rounded-full" style={{ width: `${uploadPct}%` }} />
          </div>
        )}
        {error && <p className="text-red-400 text-xs mt-2">{error}</p>}
      </div>

      {/* List */}
      {loading ? (
        <div className="text-center py-8 text-slate-600 text-sm">Loading…</div>
      ) : docs.length === 0 ? (
        <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] py-12 text-center text-slate-600 text-sm">
          No documents yet
        </div>
      ) : (
        <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] divide-y divide-[#1e1f20]">
          {docs.map((doc) => (
            <div key={doc.id} className="flex items-center justify-between px-4 py-3 hover:bg-blue-500/5 transition-colors group">
              <div className="flex items-center gap-3 min-w-0">
                <div className="w-8 h-8 rounded-lg bg-blue-500/10 border border-blue-500/20 flex items-center justify-center flex-shrink-0">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#60a5fa" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                    <polyline points="14 2 14 8 20 8" />
                  </svg>
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-slate-200 truncate">{doc.file_name}</p>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className={`text-xs px-1.5 py-0.5 rounded font-medium ${docTypeStyle[doc.document_type]}`}>
                      {docTypeLabels[doc.document_type]}
                    </span>
                    {doc.file_size_bytes && (
                      <span className="text-xs text-slate-600">{formatBytes(doc.file_size_bytes)}</span>
                    )}
                    <span className="text-xs text-slate-600">
                      {new Date(doc.uploaded_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-3 ml-3">
                <button
                  onClick={() => handleOpen(doc)}
                  className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                >
                  {isImage(doc.file_name) ? "View" : "Open"}
                </button>
                <button
                  onClick={async () => {
                    try {
                      const { url } = await documents.getURL(doc.id);
                      const blob = await fetch(url).then((r) => r.blob());
                      const objectUrl = URL.createObjectURL(blob);
                      const a = window.document.createElement("a");
                      a.href = objectUrl;
                      a.download = doc.file_name;
                      a.click();
                      URL.revokeObjectURL(objectUrl);
                    } catch {
                      setError("Couldn’t download the document. Please try again.");
                    }
                  }}
                  className="text-slate-500 hover:text-blue-400 text-sm font-medium transition-colors"
                >
                  Download
                </button>
                <button
                  onClick={() => handleDelete(doc.id)}
                  className="text-slate-700 hover:text-red-400 text-sm transition-colors"
                >
                  ✕
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>

      {preview && (
        <Modal title="Document preview" description={preview.name} onClose={() => setPreview(null)} size="wide">
          <div className="crm-document-preview">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={preview.url}
              alt={preview.name}
              className="max-w-full max-h-[65dvh] rounded-lg object-contain bg-white"
            />

          </div>
        </Modal>
      )}
      {confirmationDialog}
    </>
  );
}
