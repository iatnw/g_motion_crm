const weights = [8, 7, 6, 5, 4, 3, 2, 10, 0, 9, 8, 7, 6, 5, 4, 3, 2];
const letters: Record<string, number> = {
  A: 1, B: 2, C: 3, D: 4, E: 5, F: 6, G: 7, H: 8,
  J: 1, K: 2, L: 3, M: 4, N: 5, P: 7, R: 9,
  S: 2, T: 3, U: 4, V: 5, W: 6, X: 7, Y: 8, Z: 9,
};

// Do not guess substitutions such as O -> 0 or silently truncate a pasted VIN.
export function normalizeVin(value: string): string {
  return value.toUpperCase().replace(/\s/g, "");
}

// Standard U.S. 17-character VIN and position-nine check digit (49 CFR 565.15).
// A passing checksum does not establish vehicle identity or history.
export function vinError(value: string): string | null {
  const vin = normalizeVin(value);
  if (!vin) return null;
  if (/[IOQ]/.test(vin)) return "VINs never use I, O, or Q. Check the number on the vehicle or paperwork.";
  if (/[^A-HJ-NPR-Z0-9]/.test(vin)) return "Use only VIN letters and numbers; remove punctuation or other symbols.";
  if (vin.length !== 17) return `A standard VIN needs 17 characters. You entered ${vin.length}.`;
  const sum = [...vin].reduce((total, char, index) => total + (letters[char] ?? Number(char)) * weights[index], 0);
  const remainder = sum % 11;
  if (vin[8] !== (remainder === 10 ? "X" : String(remainder))) {
    return "The VIN check digit does not match. Recheck all 17 characters against the vehicle or paperwork.";
  }
  return null;
}

export function vinChanged(value: string, original: string): boolean {
  return normalizeVin(value) !== normalizeVin(original);
}
