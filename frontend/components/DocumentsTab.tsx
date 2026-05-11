"use client";

import { useEffect, useState, useCallback } from "react";
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
  const [docs, setDocs] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
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
      alert("Failed to get link");
    }
  }

  async function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setError("");
    try {
      let doc: Document;
      if (clientId) {
        doc = await documents.uploadForClient(clientId, file, selectedType);
      } else {
        doc = await documents.uploadForDeal(dealId!, file, selectedType);
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
    if (!confirm("Delete this document?")) return;
    try {
      await documents.delete(id);
      setDocs((prev) => prev.filter((d) => d.id !== id));
    } catch {
      alert("Failed to delete");
    }
  }

  return (
    <>
    <div className="space-y-4">
      {/* Upload */}
      <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] p-4">
        <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3">Upload Document</h3>
        <div className="flex gap-3 items-center flex-wrap">
          <select
            value={selectedType}
            onChange={(e) => setSelectedType(e.target.value as DocumentType)}
            className="bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white focus:outline-none transition-colors"
          >
            {Object.entries(docTypeLabels).map(([v, l]) => (
              <option key={v} value={v}>{l}</option>
            ))}
          </select>
          <label className={`cursor-pointer px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${uploading ? "bg-slate-800 text-slate-600 cursor-not-allowed" : "bg-blue-600 hover:bg-blue-500 text-white glow-blue-sm"}`}>
            {uploading ? "Uploading…" : "Choose File"}
            <input
              type="file"
              className="hidden"
              disabled={uploading}
              onChange={handleUpload}
              accept="image/*,application/pdf,.doc,.docx"
            />
          </label>
        </div>
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
                      const a = window.document.createElement("a");
                      a.href = url;
                      a.download = doc.file_name;
                      a.click();
                    } catch {
                      alert("Failed to get download link");
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
        <div
          className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4"
          onClick={() => setPreview(null)}
        >
          <div className="relative max-w-4xl max-h-full" onClick={(e) => e.stopPropagation()}>
            <button
              onClick={() => setPreview(null)}
              className="absolute -top-10 right-0 text-slate-400 hover:text-white text-sm transition-colors"
            >
              ✕ Close
            </button>
            <img
              src={preview.url}
              alt={preview.name}
              className="max-w-full max-h-[85vh] rounded-xl object-contain"
            />
            <p className="text-center text-slate-500 text-xs mt-2">{preview.name}</p>
          </div>
        </div>
      )}
    </>
  );
}
