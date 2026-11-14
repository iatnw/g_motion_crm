"use client";

import { useEffect, useId, useState, useRef } from "react";
import Modal from "@/components/Modal";
import VinField from "@/components/VinField";
import { clients } from "@/lib/api";
import { applyAgreementClient, agreementDefaults, type AgreementData } from "@/lib/broker-agreement";
import type { Client, Deal } from "@/lib/types";

const sections: { title: string; fields: { key: keyof AgreementData; label: string; type?: string }[] }[] = [
  { title: "Consumer", fields: [
    {key:"consumer",label:"Full name"},{key:"phone",label:"Telephone"},{key:"address",label:"Street address"},
    {key:"city",label:"City"},{key:"state",label:"State"},{key:"zip",label:"ZIP"},
  ]},
  { title: "Vehicle", fields: [
    {key:"year",label:"Year"},{key:"make",label:"Make"},{key:"model",label:"Model"},{key:"vin",label:"VIN"},{key:"trim",label:"Trim"},
    {key:"odometer",label:"Odometer reading"},{key:"color",label:"Color"},{key:"transmission",label:"Transmission"},{key:"license",label:"License plate"},{key:"options",label:"Options"},
  ]},
  { title: "Agreement details — review before printing", fields: [
    {key:"purchasePrice",label:"Vehicle purchase price (excluding tax and license)",type:"number"},
    {key:"consumerFee",label:"Fee the consumer is obligated to pay (from broker fee)",type:"number"},
    {key:"expiryDate",label:"Agreement expires",type:"date"},{key:"executionDate",label:"Agreement execution date",type:"date"},
  ]},
];

type Props = {deal:Deal; client?:Client; onClose:()=>void};

export default function BrokerAgreementModal(props: Props) {
  return <AgreementForm key={props.deal.id} {...props} />;
}

function AgreementForm({deal, client, onClose}: Props) {
  const id = useId();
  const [data,setData] = useState(() => agreementDefaults(deal,client));
  const [busy,setBusy] = useState(false);
  const [clientLoading, setClientLoading] = useState(true);
  const [clientError, setClientError] = useState("");
  const [clientAttempt, setClientAttempt] = useState(0);
  const editedFields = useRef(new Set<string>());
  useEffect(() => {
    let active = true;
    clients.get(deal.client_id).then((profile) => {
      if (!active) return;
      const edited = new Set(editedFields.current);
      setData(previous => applyAgreementClient(previous, profile, edited));
    }).catch(() => {
      if (active) setClientError("Could not load the current client profile. Review or enter the consumer details manually before printing.");
    }).finally(() => {
      if (active) setClientLoading(false);
    });
    return () => { active = false; };
  }, [deal.client_id, clientAttempt]);
  const [error,setError] = useState("");
  const [preview,setPreview] = useState("");
  useEffect(() => () => { if (preview) URL.revokeObjectURL(preview); }, [preview]);
  function update(key:keyof AgreementData,value:string) {
    editedFields.current.add(key);
    setPreview("");
    setData(previous=>({...previous,[key]:value}));
  }
  async function generate(event:React.FormEvent) {
    event.preventDefault();
    if (clientLoading) return;
    setBusy(true); setError("");
    try {
      const [template,font] = await Promise.all([fetch("/templates/broker-agreement.pdf"),fetch("/fonts/NotoSans.ttf")]);
      if (!template.ok || !font.ok) throw new Error("Could not load the agreement template. Please try again.");
      const {buildBrokerAgreement} = await import("@/lib/broker-agreement");
      const bytes = await buildBrokerAgreement(await template.arrayBuffer(),await font.arrayBuffer(),data);
      setPreview(URL.createObjectURL(new Blob([new Uint8Array(bytes)],{type:"application/pdf"})));
    } catch (err) { setError(err instanceof Error ? err.message : "Could not generate agreement"); }
    finally { setBusy(false); }
  }
  return (
    <Modal title="Broker agreement" description={`Review the details for ${deal.deal_number ?? "this deal"}, then generate your original agreement as a printable PDF.`} onClose={onClose} size="large" busy={busy}>
      <form onSubmit={generate} className="crm-modal-form crm-deal-form">
        <p className="text-xs leading-relaxed text-slate-400">Client and vehicle details are prefilled. Other values stay blank until entered. Changes here apply only to this PDF; they do not update the deal or client.</p>
        {clientLoading && <p role="status" className="text-xs text-slate-400">Loading the client’s current name, address, and telephone…</p>}
        {clientError && <div role="alert" className="text-xs text-amber-400">
          <p>{clientError}</p>
          <button type="button" className="mt-1 font-medium underline" onClick={() => {setClientError(""); setClientLoading(true); setPreview(""); setClientAttempt(attempt => attempt + 1);}}>Retry client details</button>
        </div>}
        {sections.map(section=>(
          <section key={section.title}>
            <h3 className="mb-3 text-sm font-semibold text-white">{section.title}</h3>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {section.fields.map(field=>(
                <div key={field.key}>
                  {field.key === "vin" ? (
                    <VinField value={data.vin} onChange={value => update("vin", value)} warningHint="You can still generate and print this agreement." />
                  ) : <>
                  <label htmlFor={`${id}-${field.key}`} className="mb-1.5 block text-xs text-slate-400">{field.label}</label>
                  <input id={`${id}-${field.key}`} type={field.type ?? "text"} value={data[field.key]} onChange={event=>update(field.key,event.target.value)}
                    min={field.type==="number" ? "0" : undefined} step={field.type==="number" ? "0.01" : undefined}
                    className="w-full rounded-lg border border-[#252628] bg-[#0b0c0d] px-3 py-2 text-sm text-white" />
                  </>}
                </div>
              ))}
            </div>
          </section>
        ))}
        <fieldset>
          <legend className="mb-2 text-xs text-slate-400">Does G Motion receive a fee from the selling dealer?</legend>
          <div className="flex flex-wrap gap-4 text-sm text-slate-300">
            {([{value:"",label:"Leave blank"},{value:"yes",label:"Yes"},{value:"no",label:"No"}] as const).map(choice=>(
              <label key={choice.value} className="flex items-center gap-2"><input type="radio" name={`${id}-dealer-fee`} value={choice.value} checked={data.dealerFee===choice.value} onChange={()=>update("dealerFee",choice.value)} />{choice.label}</label>
            ))}
          </div>
        </fieldset>
        <p className="text-xs text-slate-400">The consumer fee is prefilled from the deal’s broker fee and can be adjusted for this PDF. Filled text prints at 10 points. Signatures remain blank for signing.</p>
        {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
        {preview && <section aria-label="Agreement PDF preview">
          <iframe title="Filled broker agreement" src={preview} className="h-[65vh] w-full rounded-lg border border-slate-600 bg-white" />
          <p className="mt-2 text-xs text-slate-400">Review the PDF before signing. Open it in the PDF viewer to print using its Print button or Ctrl/Cmd+P. Choose actual size when printing.</p>
        </section>}
        <div className="crm-modal-footer flex flex-wrap gap-3">
          <button type="button" onClick={onClose} disabled={busy} className="rounded-lg border border-[#252628] px-4 py-2 text-sm text-slate-400">Close</button>
          {preview && <>
            <a href={preview} target="_blank" rel="noopener noreferrer" className="rounded-lg border border-blue-500/40 px-4 py-2 text-sm text-blue-400">Open PDF to print</a>
            <a href={preview} download={`broker-agreement-${deal.deal_number ?? deal.id}.pdf`} className="rounded-lg border border-[#252628] px-4 py-2 text-sm text-slate-300">Download PDF</a>
          </>}
          <button type="submit" disabled={busy || clientLoading} className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white disabled:opacity-50">{busy ? "Preparing PDF…" : clientLoading ? "Loading client…" : preview ? "Regenerate PDF" : "Generate agreement"}</button>
        </div>
      </form>
    </Modal>
  );
}
