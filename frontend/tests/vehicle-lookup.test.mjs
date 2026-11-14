import assert from "node:assert/strict";
import test from "node:test";
import { applyVehicleLookup } from "../lib/vehicle-lookup.ts";
const vehicle = { vin: "1HGCM82633A004352", make: "HONDA", model: "Accord", trim_level: "EX", model_year: "2003" };
const form = { vin: vehicle.vin, make: "Old make", model: "Old model", trim_level: "Old trim", notes: "Keep notes" };
test("updates vehicle fields when a replacement VIN is decoded", () => {
  assert.deepEqual(applyVehicleLookup(form,vehicle,new Set()), { ...form, make:"HONDA",model:"Accord",trim_level:"EX" });
  assert.equal(form.make,"Old make");
});
test("ignores a lookup for a VIN the user has since replaced", () => {
  const changed = {...form,vin:"OTHER"};
  assert.equal(applyVehicleLookup(changed,vehicle,new Set()),changed);
});
test("preserves manual edits made before or during lookup", () => {
  assert.deepEqual(applyVehicleLookup(form,vehicle,new Set(["make","trim_level"])), {...form,model:"Accord"});
});
test("does not retain stale trim from a previous vehicle when trim is unavailable", () => {
  assert.equal(applyVehicleLookup(form,{...vehicle,trim_level:""},new Set()).trim_level,"");
});
