import type { Client, Deal } from "./types";

export type AgreementData = {
  consumer: string; phone: string; address: string; city: string; state: string; zip: string;
  year: string; make: string; model: string; vin: string; trim: string;
  odometer: string; color: string; transmission: string; license: string; options: string;
  purchasePrice: string; consumerFee: string; expiryDate: string; executionDate: string;
  dealerFee: "" | "yes" | "no";
};

const clientAgreementKeys = ["consumer", "phone", "address", "city", "state", "zip"] as const;

function clientAgreementData(client?: Client) {
  return {
    consumer: [client?.first_name, client?.last_name].filter(Boolean).join(" "),
    phone: client?.phone ?? "", address: client?.address ?? "", city: client?.city ?? "", state: client?.state ?? "", zip: client?.zip ?? "",
  };
}

export function applyAgreementClient(data: AgreementData, client: Client, edited: ReadonlySet<string>): AgreementData {
  const current = clientAgreementData(client);
  const next = { ...data };
  for (const key of clientAgreementKeys) {
    if (!edited.has(key)) next[key] = current[key];
  }
  return next;
}

export function agreementDefaults(deal: Deal, client?: Client): AgreementData {
  const today = new Date();
  const date = `${today.getFullYear()}-${String(today.getMonth()+1).padStart(2,"0")}-${String(today.getDate()).padStart(2,"0")}`;
  return {
    ...clientAgreementData(client),
    year: "", make: deal.make ?? "", model: deal.model ?? "", vin: deal.vin ?? "", trim: deal.trim_level ?? "",
    odometer: "", color: "", transmission: "", license: "", options: "",
    // The agreement consumer fee uses the deal's broker fee, as configured by the owner.
    purchasePrice: "", consumerFee: deal.broker_fee?.toString() ?? "", expiryDate: "", executionDate: date, dealerFee: "",
  };
}

export function agreementDate(value: string): string {
  return /^\d{4}-\d{2}-\d{2}$/.test(value) ? `${value.slice(5,7)}/${value.slice(8,10)}/${value.slice(0,4)}` : "";
}

// Top-left PDF coordinates, measured against the supplied 601 × 776 template.
export const agreementBoxes: Partial<Record<keyof AgreementData, [number, number, number, number]>> = {
  consumer: [35, 119, 443, 131], phone: [450,119,582,131],
  address: [35,143,581,155], city: [35,168,330,180], state: [335,168,443,180], zip: [450,168,581,180],
  year: [35,218,70,230], make: [75,218,201,230], model: [410,218,581,230],
  odometer: [35,242,137,254], vin: [144,242,356,254], color: [35,266,110,278], trim: [116,266,183,278],
  transmission: [262,266,328,278], license: [455,266,581,278], options: [70,290,581,301],
  purchasePrice: [162,349,269,361], expiryDate: [445,374,581,386], consumerFee: [214,399,361,411],
  executionDate: [156,672,290,684],
};

export async function buildBrokerAgreement(template: ArrayBuffer | Uint8Array, fontBytes: ArrayBuffer | Uint8Array, data: AgreementData): Promise<Uint8Array> {
  const { PDFDocument, rgb } = await import("pdf-lib");
  const { default: fontkit } = await import("@pdf-lib/fontkit");
  const pdf = await PDFDocument.load(template);
  const page = pdf.getPages()[0];
  if (pdf.getPageCount() !== 1 || Math.abs(page.getWidth()-601)>1 || Math.abs(page.getHeight()-776)>1) {
    throw new Error("The agreement template has changed. Please check its field layout before printing.");
  }
  pdf.registerFontkit(fontkit);
  const font = await pdf.embedFont(fontBytes, { subset: false });
  const supported = new Set(font.getCharacterSet());
  for (const [key, box] of Object.entries(agreementBoxes)) {
    const field = key as keyof AgreementData;
    let value = data[field].trim().replace(/\s+/g," ").toUpperCase();
    if (field === "expiryDate" || field === "executionDate") value = agreementDate(data[field]);
    if (field === "purchasePrice" || field === "consumerFee") {
      if (value) {
        const amount = Number(value);
        if (!Number.isFinite(amount) || amount < 0) throw new Error("Enter a valid, non-negative purchase price and consumer fee.");
        value = amount.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
      }
    }
    if (!value) continue;
    if ([...value].some((character) => !supported.has(character.codePointAt(0)!))) {
      throw new Error(`The ${field} field contains characters this PDF font cannot print. Please review that field.`);
    }
    const [left, top, right, bottom] = box;
    const size = 10;
    if (font.widthOfTextAtSize(value, size) > right-left-2) {
      throw new Error(`The ${field} field is too long to fit at 10-point type on the agreement. Please shorten it before printing.`);
    }
    // The scan has preprinted phone parentheses; remove only that blank entry area.
    if (field === "phone") page.drawRectangle({ x:left-1, y:page.getHeight()-bottom-1, width:right-left+1, height:bottom-top+2, color:rgb(1,1,1) });
    page.drawText(value, { x:left+1, y:page.getHeight()-bottom+2, size, font, color:rgb(0,0,0) });
  }
  if (data.dealerFee) {
    page.drawText("X", { x:data.dealerFee === "yes" ? 38 : 256, y:page.getHeight()-457+2, size:10, font, color:rgb(0,0,0) });
  }
  // No signatures are synthesized. Preserve the original agreement and its terms.
  pdf.setTitle("Motor Vehicle Brokering Agreement");
  pdf.setAuthor("G Motion LLC");
  return pdf.save();
}
