import type { Broker, Client, Deal, Dealer } from "./types";

export type DealCoverData = {
  date: string; dealNumber: string; salesperson: string; vehicle: string; vin: string;
  clientName: string; brokerFee: string; downPayment: string; vendor: string; vendorAddress: string;
};

export function dealCoverDefaults(deal: Deal, client?: Client, broker?: Broker, dealer?: Dealer): DealCoverData {
  return {
    date: deal.deal_date?.slice(0,10) ?? "", dealNumber: deal.deal_number ?? "",
    salesperson: [broker?.first_name,broker?.last_name].filter(Boolean).join(" "),
    vehicle: [deal.make,deal.model,deal.trim_level].filter(Boolean).join(" "), vin: deal.vin ?? "",
    clientName: [client?.first_name,client?.last_name].filter(Boolean).join(" "),
    brokerFee: deal.broker_fee?.toString() ?? "", downPayment: deal.down_payment?.toString() ?? "",
    vendor: dealer?.name ?? "", vendorAddress: dealer?.address ?? "",
  };
}

export function applyDealCoverDetails(previous: DealCoverData, current: DealCoverData, edited: ReadonlySet<string>): DealCoverData {
  const next={...previous};
  for (const field of Object.keys(current) as (keyof DealCoverData)[]) {
    if (!edited.has(field)) next[field]=current[field];
  }
  return next;
}

// Top-left positions on the supplied US Letter scan (612 × 792 points).
export const dealCoverBoxes: Record<keyof DealCoverData,[number,number,number,number]> = {
  date:[111,162,542,185], dealNumber:[253,198,549,221], salesperson:[242,234,548,257],
  vehicle:[168,270,544,294], vin:[87,306,545,330], clientName:[245,342,552,366],
  brokerFee:[383,378,552,402], downPayment:[431,414,560,438],
  vendor:[165,450,568,474], vendorAddress:[33,486,559,510],
};

export async function buildDealCover(template: ArrayBuffer | Uint8Array, fontBytes: ArrayBuffer | Uint8Array, data: DealCoverData): Promise<Uint8Array> {
  const {PDFDocument,rgb}=await import("pdf-lib");
  const {default:fontkit}=await import("@pdf-lib/fontkit");
  const pdf=await PDFDocument.load(template);
  const page=pdf.getPages()[0];
  if(pdf.getPageCount()!==1 || Math.abs(page.getWidth()-612)>1 || Math.abs(page.getHeight()-792)>1) throw new Error("The deal cover template has changed. Check its layout before printing.");
  pdf.registerFontkit(fontkit);
  const font=await pdf.embedFont(fontBytes,{subset:false});
  const supported=new Set(font.getCharacterSet());
  for (const key of Object.keys(dealCoverBoxes) as (keyof DealCoverData)[]) {
    let value=data[key].trim().replace(/\s+/g," ").toUpperCase();
    if (key==="date") value=/^\d{4}-\d{2}-\d{2}$/.test(value) ? `${value.slice(5,7)}/${value.slice(8,10)}/${value.slice(0,4)}` : "";
    if ((key==="brokerFee" || key==="downPayment") && value) {
      const amount=Number(value);
      if(!Number.isFinite(amount) || amount<0) throw new Error("Enter valid, non-negative broker fee and down payment amounts.");
      value=amount.toLocaleString("en-US",{style:"currency",currency:"USD"});
    }
    if(!value) continue;
    const label=key.replace(/([a-z])([A-Z])/g,"$1 $2");
    if([...value].some(char=>!supported.has(char.codePointAt(0)!))) throw new Error(`The ${label} contains characters this PDF font cannot print. Please review that field.`);
    const [left,,right,bottom]=dealCoverBoxes[key];
    const size=Math.min(16,(right-left-2)/font.widthOfTextAtSize(value,1));
    if(size<10) throw new Error(`The ${label} is too long to fit at 10-point type. Please shorten it before printing.`);
    page.drawText(value,{x:left+1,y:page.getHeight()-bottom+3,size,font,color:rgb(0,0,0)});
  }
  pdf.setTitle("G Motion Final Sales Sheet");
  pdf.setAuthor("G Motion LLC");
  return pdf.save();
}
