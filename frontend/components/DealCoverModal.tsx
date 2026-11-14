"use client";

import {useEffect,useId,useRef,useState} from "react";
import Modal from "@/components/Modal";
import VinField from "@/components/VinField";
import {clients,brokers,dealers,deals} from "@/lib/api";
import {applyDealCoverDetails,dealCoverDefaults,type DealCoverData} from "@/lib/deal-cover";
import type {Client,Deal} from "@/lib/types";

type Props={deal:Deal;client?:Client;onClose:()=>void};
const fields: {key:keyof DealCoverData;label:string;type?:string}[]=[
  {key:"date",label:"Deal date",type:"date"},{key:"dealNumber",label:"Deal number"},
  {key:"salesperson",label:"Sale person (linked broker)"},{key:"clientName",label:"Client name"},
  {key:"vehicle",label:"Vehicle (make, model and trim)"},{key:"vin",label:"VIN"},
  {key:"brokerFee",label:"Broker fee amount",type:"number"},{key:"downPayment",label:"Down payment amount",type:"number"},
  {key:"vendor",label:"Vendor (linked dealer)"},{key:"vendorAddress",label:"Vendor address"},
];

export default function DealCoverModal(props:Props){return <DealCoverForm key={props.deal.id} {...props}/>;}

function DealCoverForm({deal,client,onClose}:Props){
  const id=useId();
  const [data,setData]=useState(()=>dealCoverDefaults(deal,client));
  const [loading,setLoading]=useState(true);
  const [loadError,setLoadError]=useState("");
  const [attempt,setAttempt]=useState(0);
  const [busy,setBusy]=useState(false);
  const [error,setError]=useState("");
  const [preview,setPreview]=useState("");
  const edited=useRef(new Set<string>());
  useEffect(()=>{
    let active=true;
    deals.get(deal.id).then(async current=>{
      const records=await Promise.allSettled([
        clients.get(current.client_id),
        current.broker_id ? brokers.get(current.broker_id) : Promise.resolve(undefined),
        current.dealer_id ? dealers.get(current.dealer_id) : Promise.resolve(undefined),
      ]);
      if(!active)return;
      const [clientResult,brokerResult,dealerResult]=records;
      const details=dealCoverDefaults(current,
        clientResult.status==="fulfilled" ? clientResult.value : client,
        brokerResult.status==="fulfilled" ? brokerResult.value : undefined,
        dealerResult.status==="fulfilled" ? dealerResult.value : undefined);
      const protectedFields=new Set(edited.current);
      setData(previous=>applyDealCoverDetails(previous,details,protectedFields));
      const missing=records.flatMap((result,index)=>result.status==="rejected" ? [["client","broker","dealer"][index]] : []);
      if(missing.length)setLoadError(`Could not refresh ${missing.join(", ")} details. Review those fields manually or retry.`);
    }).catch(()=>{
      if(active)setLoadError("Could not refresh this deal. Review the prefilled values manually or retry.");
    }).finally(()=>{if(active)setLoading(false);});
    return ()=>{active=false;};
  },[deal.id,client,attempt]);
  useEffect(()=>()=>{if(preview)URL.revokeObjectURL(preview);},[preview]);
  function update(key:keyof DealCoverData,value:string){edited.current.add(key);setPreview("");setData(previous=>({...previous,[key]:value}));}
  async function generate(event:React.FormEvent){
    event.preventDefault();if(loading)return;
    setBusy(true);setError("");
    try{
      const [template,font]=await Promise.all([fetch("/templates/deal-cover.pdf"),fetch("/fonts/NotoSans.ttf")]);
      if(!template.ok || !font.ok)throw new Error("Could not load the deal cover template. Please try again.");
      const {buildDealCover}=await import("@/lib/deal-cover");
      const bytes=await buildDealCover(await template.arrayBuffer(),await font.arrayBuffer(),data);
      setPreview(URL.createObjectURL(new Blob([new Uint8Array(bytes)],{type:"application/pdf"})));
    }catch(err){setError(err instanceof Error ? err.message : "Could not generate the deal cover");}
    finally{setBusy(false);}
  }
  return <Modal title="Deal cover" description={`Review the final sales sheet for ${deal.deal_number ?? "this deal"}, then download or print it.`} onClose={onClose} size="large" busy={busy}>
    <form onSubmit={generate} className="crm-modal-form crm-deal-form">
      <p className="text-xs leading-relaxed text-slate-400">Details are pulled from the current deal, client, broker and dealer records. Changes here apply only to this PDF. Missing information stays blank.</p>
      {loading && <p role="status" className="text-xs text-slate-400">Loading current deal and client details…</p>}
      {loadError && <div role="alert" className="text-xs text-amber-400"><p>{loadError}</p><button type="button" className="mt-1 underline" onClick={()=>{setLoadError("");setLoading(true);setPreview("");setAttempt(value=>value+1);}}>Retry details</button></div>}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        {fields.map(field=><div key={field.key}>
          {field.key==="vin" ? <VinField value={data.vin} onChange={value=>update("vin",value)} warningHint="You can still generate and print this deal cover."/> : <>
            <label htmlFor={`${id}-${field.key}`} className="mb-1.5 block text-xs text-slate-400">{field.label}</label>
            <input id={`${id}-${field.key}`} type={field.type ?? "text"} value={data[field.key]} onChange={event=>update(field.key,event.target.value)} min={field.type==="number" ? "0" : undefined} step={field.type==="number" ? "0.01" : undefined} className="w-full rounded-lg border border-[#252628] bg-[#0b0c0d] px-3 py-2 text-sm text-white"/>
          </>}
        </div>)}
      </div>
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      {preview && <section aria-label="Deal cover PDF preview">
        <iframe title="Filled deal cover" src={preview} className="h-[65vh] w-full rounded-lg border border-slate-600 bg-white"/>
        <p className="mt-2 text-xs text-slate-400">Open the PDF to print using its Print button or Ctrl/Cmd+P. Choose actual size. Filled text is at least 10 points.</p>
      </section>}
      <div className="crm-modal-footer flex flex-wrap gap-3">
        <button type="button" onClick={onClose} disabled={busy} className="rounded-lg border border-[#252628] px-4 py-2 text-sm text-slate-400">Close</button>
        {preview && <>
          <a href={preview} target="_blank" rel="noopener noreferrer" className="rounded-lg border border-blue-500/40 px-4 py-2 text-sm text-blue-400">Open PDF to print</a>
          <a href={preview} download={`deal-cover-${deal.deal_number ?? deal.id}.pdf`} className="rounded-lg border border-[#252628] px-4 py-2 text-sm text-slate-300">Download PDF</a>
        </>}
        <button type="submit" disabled={busy || loading} className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white disabled:opacity-50">{busy ? "Preparing PDF…" : loading ? "Loading details…" : preview ? "Regenerate PDF" : "Generate deal cover"}</button>
      </div>
    </form>
  </Modal>;
}
