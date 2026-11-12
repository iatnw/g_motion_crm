import assert from "node:assert/strict";
import test from "node:test";
import { fillMissingCommission } from "../lib/deal-finances.ts";

const base = {
  broker_fee: "1000", collect_amount: "200", doc_fee: "100",
  broker_percent: "", broker_commission: "", your_net: "",
};

test("percentage-only legacy deal includes collections, document fees and referrals", () => {
  const result = fillMissingCommission({ ...base, broker_percent: "25" }, 125);
  assert.equal(result.broker_commission, "275.00");
  assert.equal(result.your_net, "700.00");
});

test("commission-only legacy deal reconstructs the percentage", () => {
  const result = fillMissingCommission({ ...base, broker_commission: "275" }, 125);
  assert.equal(result.broker_percent, "25.00");
  assert.equal(result.your_net, "700.00");
});

test("populated commission fields and manually adjusted net are preserved", () => {
  const form = { ...base, broker_percent: "20", broker_commission: "275", your_net: "650" };
  assert.deepEqual(fillMissingCommission(form, 125), form);
});

test("empty or zero commission does not create financial values", () => {
  for (const form of [base, { ...base, broker_percent: "0", broker_commission: "0" }]) {
    assert.deepEqual(fillMissingCommission(form, 0), form);
  }
});

test("zero and negative commission bases do not cause division by zero", () => {
  for (const doc_fee of ["1200", "1300"]) {
    const form = { ...base, doc_fee, broker_commission: "100" };
    assert.deepEqual(fillMissingCommission(form, 0), form);
  }
});

test("currency results round to cents and retain a negative net", () => {
  const form = { ...base, broker_fee: "100.01", collect_amount: "", doc_fee: "", broker_percent: "33.33" };
  const result = fillMissingCommission(form, 80);
  assert.equal(result.broker_commission, "33.33");
  assert.equal(result.your_net, "-13.32");
  assert.equal(form.broker_commission, "");
});
