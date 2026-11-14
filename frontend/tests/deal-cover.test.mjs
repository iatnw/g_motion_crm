import assert from "node:assert/strict";
import test from "node:test";
import {readFile} from "node:fs/promises";
import {PDFDocument} from "pdf-lib";
import {dealCoverDefaults,applyDealCoverDetails,buildDealCover} from "../lib/deal-cover.ts";

const deal={id:"sample",client_id:"client",deal_date:"2026-11-14",deal_number:"GM-1234",make:"HONDA",model:"Accord",trim_level:"EX-V6",vin:"1HGCM82633A004352",broker_fee:5000,down_payment:2000,broker_commission:750};
const client={first_name:"MARIA",last_name:"EXAMPLE"};
const broker={first_name:"ALEX",last_name:"BROKER"};
const dealer={name:"EXAMPLE MOTORS",address:"123 EXAMPLE AVENUE"};
const template=await readFile(new URL("../public/templates/deal-cover.pdf",import.meta.url));
const font=await readFile(new URL("../public/fonts/NotoSans.ttf",import.meta.url));

test("maps every cover field to the appropriate linked record",()=>{
 assert.deepEqual(dealCoverDefaults(deal,client,broker,dealer),{date:"2026-11-14",dealNumber:"GM-1234",salesperson:"ALEX BROKER",vehicle:"HONDA Accord EX-V6",vin:deal.vin,clientName:"MARIA EXAMPLE",brokerFee:"5000",downPayment:"2000",vendor:"EXAMPLE MOTORS",vendorAddress:"123 EXAMPLE AVENUE"});
});
test("missing values remain blank while zero amounts are retained",()=>{
 const result=dealCoverDefaults({id:"sample",client_id:"client",broker_fee:0,down_payment:0});
 assert.equal(result.brokerFee,"0");assert.equal(result.downPayment,"0");
 for(const field of ["date","dealNumber","salesperson","vehicle","vin","clientName","vendor","vendorAddress"])assert.equal(result[field],"");
});
test("fresh records update defaults without overwriting manual edits",()=>{
 const initial={...dealCoverDefaults(deal),salesperson:"MANUAL PERSON",brokerFee:"100"};
 const fresh=dealCoverDefaults(deal,client,broker,dealer);
 const result=applyDealCoverDetails(initial,fresh,new Set(["salesperson","brokerFee"]));
 assert.equal(result.salesperson,"MANUAL PERSON");assert.equal(result.brokerFee,"100");assert.equal(result.vendor,dealer.name);assert.equal(result.clientName,"MARIA EXAMPLE");
});
test("produces a single US Letter PDF from the supplied cover",async()=>{
 const bytes=await buildDealCover(template,font,dealCoverDefaults(deal,client,broker,dealer));
 const pdf=await PDFDocument.load(bytes);
 assert.equal(pdf.getPageCount(),1);assert.deepEqual(pdf.getPages()[0].getSize(),{width:612,height:792});
});
test("rejects negative amounts and text that would require type smaller than 10 points",async()=>{
 const data=dealCoverDefaults(deal,client,broker,dealer);
 await assert.rejects(buildDealCover(template,font,{...data,downPayment:"-1"}),/non-negative/);
 await assert.rejects(buildDealCover(template,font,{...data,clientName:"VERY LONG NAME ".repeat(100)}),/10-point/);
});
test("invalid VINs remain printable with the existing nonblocking warning policy",async()=>{
 const data={...dealCoverDefaults(deal,client,broker,dealer),vin:"INCORRECT"};
 const pdf=await PDFDocument.load(await buildDealCover(template,font,data));assert.equal(pdf.getPageCount(),1);
});
