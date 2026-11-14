export type DecodedVehicle = {
  vin: string;
  make: string;
  model: string;
  trim_level: string;
  model_year: string;
};

export type VehicleField = "make" | "model" | "trim_level";
export const vehicleFields: VehicleField[] = ["make", "model", "trim_level"];

// Never apply a result for an older VIN or overwrite fields the user edited.
export function applyVehicleLookup<T extends { vin: string; make: string; model: string; trim_level: string }>(
  form: T, vehicle: DecodedVehicle, manualFields: ReadonlySet<string>,
): T {
  if (form.vin.toUpperCase().replace(/\s/g, "") !== vehicle.vin) return form;
  const next = { ...form };
  for (const field of vehicleFields) {
    if (!manualFields.has(field)) next[field] = vehicle[field];
  }
  return next;
}
