import assert from "node:assert/strict";
import test from "node:test";
import { normalizeVin, vinChanged, vinError } from "../lib/vin.ts";

test("accepts published VIN examples including X check digit", () => {
  for (const vin of ["1M8GDM9AXKP042788", "1HGCM82633A004352", "5GZCZ43D13S812715"]) {
    assert.equal(vinError(vin), null, vin);
  }
});
test("permits an empty VIN for deals without a vehicle yet", () => {
  assert.equal(vinError(""), null);
  assert.equal(vinError("  "), null);
});
test("normalizes casing and pasted whitespace without guessing characters", () => {
  assert.equal(normalizeVin(" 1hg cm82633a004352\n"), "1HGCM82633A004352");
  assert.equal(vinError(" 1hg cm82633a004352\n"), null);
  assert.equal(normalizeVin("i-o-q"), "I-O-Q");
});
test("rejects forbidden letters, symbols, short and overlong input", () => {
  for (const letter of ["I", "O", "Q"]) assert.match(vinError("1HGCM82633A00435" + letter), /never use/);
  assert.match(vinError("1HGCM82633A00435!"), /symbols/);
  assert.match(vinError("1HG"), /entered 3/);
  assert.match(vinError("1HGCM82633A0043521"), /entered 18/);
});
test("catches wrong serial digit or check digit", () => {
  assert.match(vinError("1HGCM82633A004353"), /check digit/);
  assert.match(vinError("1HGCM82643A004352"), /check digit/);
});
test("legacy comparison ignores formatting but detects a replacement", () => {
  assert.equal(vinChanged(" old123 ", "OLD123"), false);
  assert.equal(vinChanged("OLD124", "OLD123"), true);
});
