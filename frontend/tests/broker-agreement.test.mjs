import assert from "node:assert/strict";
import test from "node:test";
import { readFile } from "node:fs/promises";
import { PDFDocument } from "pdf-lib";
import { agreementDefaults, agreementDate, applyAgreementClient, buildBrokerAgreement } from "../lib/broker-agreement.ts";

const deal = {id:"demo",client_id:"demo-client",make:"HONDA",model:"Accord",trim_level:"EX-V6",vin:"1HGCM82633A004352",broker_fee:5000,collect_amount:1200,doc_fee:300,notes:"Private internal note"};
const client = {first_name:"Марія",last_name:"Example",phone:"747-555-0100",address:"123 Example Avenue",city:"North Hollywood",state:"CA",zip:"91605"};
const template = await readFile(new URL("../public/templates/broker-agreement.pdf",import.meta.url));
const font = await readFile(new URL("../public/fonts/NotoSans.ttf",import.meta.url));

test("prefills consumer and vehicle details and maps the broker fee to the consumer fee", () => {
  const data = agreementDefaults(deal,client);
  assert.equal(data.consumer,"Марія Example");
  assert.equal(data.address,client.address);
  assert.equal(data.vin,deal.vin);
  assert.equal(data.consumerFee,"5000");
  for (const key of ["purchasePrice","expiryDate","dealerFee","options","year"]) assert.equal(data[key],"");
  assert.ok(!JSON.stringify(data).includes(deal.notes));
  assert.equal(client.first_name,"Марія");
});
test("formats agreement dates without timezone shifts",()=>{
  assert.equal(agreementDate("2026-11-14"),"11/14/2026");
  assert.equal(agreementDate(""),"");
});
test("creates a printable single-page agreement using the original template",async()=>{
  const data={...agreementDefaults(deal,client),purchasePrice:"32500",consumerFee:"450",expiryDate:"2026-12-01",dealerFee:"yes"};
  const bytes=await buildBrokerAgreement(template,font,data);
  const pdf=await PDFDocument.load(bytes);
  assert.equal(pdf.getPageCount(),1);
  assert.deepEqual(pdf.getPages()[0].getSize(),{width:601,height:776});
  assert.equal(pdf.getForm().getFields().length,0);
});
test("rejects overflowing text and negative amounts instead of corrupting the agreement",async()=>{
  const data=agreementDefaults(deal,client);
  await assert.rejects(buildBrokerAgreement(template,font,{...data,consumer:"Very Long Name ".repeat(100)}),/too long/);
  await assert.rejects(buildBrokerAgreement(template,font,{...data,consumerFee:"-10"}),/non-negative/);
});

test("refreshes all consumer details from the linked client without changing vehicle or financial fields", () => {
  const data = {...agreementDefaults(deal),consumerFee:"250",make:"MANUAL MAKE"};
  const updated = applyAgreementClient(data,client,new Set());
  for (const key of ["consumer","phone","address","city","state","zip"]) assert.equal(updated[key],agreementDefaults(deal,client)[key]);
  assert.equal(updated.make,"MANUAL MAKE");
  assert.equal(updated.consumerFee,"250");
});
test("a late profile response preserves manual edits in the agreement", () => {
  const data = {...agreementDefaults(deal),consumer:"CUSTOM NAME",address:"CUSTOM ADDRESS"};
  const updated = applyAgreementClient(data,client,new Set(["consumer","address"]));
  assert.equal(updated.consumer,"CUSTOM NAME");
  assert.equal(updated.address,"CUSTOM ADDRESS");
  assert.equal(updated.city,client.city);
});

test("preserves a zero broker fee and leaves an absent broker fee blank", () => {
  assert.equal(agreementDefaults({...deal,broker_fee:0},client).consumerFee,"0");
  assert.equal(agreementDefaults({...deal,broker_fee:undefined},client).consumerFee,"");
});
test("refuses text that would require shrinking below 10 points", async () => {
  const data={...agreementDefaults(deal,client),trim:"EXTRA LONG TRIM"};
  await assert.rejects(buildBrokerAgreement(template,font,data),/10-point/);
});
