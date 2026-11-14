"use client";

import { useId } from "react";
import { normalizeVin, vinError } from "@/lib/vin";

export default function VinField({ value, onChange, warningHint = "You can still save this deal and correct the VIN later." }: {
  value: string;
  onChange: (value: string) => void;
  warningHint?: string;
}) {
  const id = useId();
  const normalized = normalizeVin(value);
  const problem = vinError(value);
  return (
    <div className="crm-deal-field">
      <div className="flex items-center justify-between gap-2">
        <label htmlFor={id} className="block text-xs font-medium text-slate-500 mb-1.5 uppercase tracking-wider">VIN</label>
        <span className="text-xs text-slate-500 tabular-nums">{normalized.length}/17</span>
      </div>
      <div className="crm-deal-field-control">
        <input
          id={id}
          value={value}
          onChange={(event) => onChange(normalizeVin(event.target.value))}
          autoCapitalize="characters"
          autoComplete="off"
          spellCheck={false}
          aria-invalid={!!problem}
          aria-describedby={`${id}-help`}
          className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white focus:outline-none transition-colors font-mono aria-invalid:border-amber-500"
        />
      </div>
      <p id={`${id}-help`} aria-live="polite" className={`mt-1.5 text-xs leading-relaxed ${problem ? "text-red-400" : normalized ? "text-green-400" : "text-slate-500"}`}>
        {problem
          ? `${problem} ${warningHint}`
          : normalized ? "VIN format and check digit pass. Confirm it matches the vehicle." : "Optional. Enter the 17-character VIN from the vehicle or paperwork."}
      </p>
    </div>
  );
}
