"use client";

import { useId, useState } from "react";
import type { Client } from "@/lib/types";

interface Props {
  initial?: Client;
  onSave: (data: object) => Promise<void>;
  onCancel: () => void;
}

export default function ClientForm({ initial, onSave, onCancel }: Props) {
  const notesId = useId();
  const [form, setForm] = useState({
    first_name: initial?.first_name ?? "",
    last_name: initial?.last_name ?? "",
    phone: initial?.phone ?? "",
    email: initial?.email ?? "",
    source: initial?.source ?? "",
    status: initial?.status ?? "active",
    notes: initial?.notes ?? "",
    date_of_birth: initial?.date_of_birth?.slice(0, 10) ?? "",
    address: initial?.address ?? "",
    city: initial?.city ?? "",
    state: initial?.state ?? "",
    zip: initial?.zip ?? "",
    ssn: "",
    housing_status: initial?.housing_status ?? "",
    housing_duration: initial?.housing_duration ?? "",
    housing_monthly_payment: initial?.housing_monthly_payment?.toString() ?? "",
    employer_name: initial?.employer_name ?? "",
    employer_address: initial?.employer_address ?? "",
    employer_city: initial?.employer_city ?? "",
    employer_state: initial?.employer_state ?? "",
    employer_zip: initial?.employer_zip ?? "",
    occupation: initial?.occupation ?? "",
    employer_phone: initial?.employer_phone ?? "",
    employment_status: initial?.employment_status ?? "",
    employment_duration: initial?.employment_duration ?? "",
    gross_monthly_income: initial?.gross_monthly_income?.toString() ?? "",
    other_monthly_income: initial?.other_monthly_income?.toString() ?? "",
    other_income_source: initial?.other_income_source ?? "",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  function set(field: string, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      const numericFields = new Set([
        "housing_monthly_payment",
        "gross_monthly_income",
        "other_monthly_income",
      ]);
      const payload: Record<string, string | number | null> = {};
      for (const [k, v] of Object.entries(form)) {
        if (numericFields.has(k)) {
          payload[k] = v ? parseFloat(v) : null;
        } else {
          payload[k] = v || null;
        }
      }
      payload.first_name = form.first_name;
      payload.last_name = form.last_name;
      payload.phone = form.phone;
      payload.email = form.email;
      await onSave(payload);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-5xl space-y-5">
      {/* Personal */}
      <Section title="Personal Information">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field
            label="First Name *"
            value={form.first_name}
            onChange={(v) => set("first_name", v)}
            required
          />
          <Field
            label="Last Name *"
            value={form.last_name}
            onChange={(v) => set("last_name", v)}
            required
          />
          <Field
            label="Phone"
            value={form.phone}
            onChange={(v) => set("phone", v)}
          />
          <Field
            label="Email"
            type="email"
            value={form.email}
            onChange={(v) => set("email", v)}
          />
          <Field
            label="Date of Birth"
            type="date"
            value={form.date_of_birth}
            onChange={(v) => set("date_of_birth", v)}
          />
          <Field
            label="SSN"
            value={form.ssn}
            onChange={(v) => set("ssn", v)}
            placeholder={
              initial ? "Leave blank to keep existing SSN" : "123456789"
            }
          />
          <SelectField
            label="Client status"
            value={form.status}
            onChange={(v) => set("status", v)}
            options={[
              { value: "active", label: "Active" },
              { value: "closed", label: "Closed" },
              { value: "no_deal", label: "No deal" },
            ]}
          />
          <SelectField
            label="Source"
            value={form.source}
            onChange={(v) => set("source", v)}
            options={[
              { value: "", label: "None" },
              { value: "instagram", label: "Instagram" },
              { value: "telegram", label: "Telegram" },
              { value: "referall", label: "Referral" },
              { value: "organic", label: "Organic" },
              { value: "repeat_client", label: "Repeat Client" },
            ]}
          />
        </div>
        <div className="mt-3">
          <label htmlFor={notesId} className="block text-xs font-medium text-slate-500 mb-1.5 uppercase tracking-wider">
            Notes
          </label>
          <textarea
            id={notesId}
            value={form.notes}
            onChange={(e) => set("notes", e.target.value)}
            rows={2}
            className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-2 text-sm text-white placeholder:text-slate-700 focus-visible:outline-2 focus-visible:outline-blue-400 transition-colors resize-none"
          />
        </div>
      </Section>

      {/* Address */}
      <Section title="Address">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="col-span-2">
            <Field
              label="Street Address"
              value={form.address}
              onChange={(v) => set("address", v)}
            />
          </div>
          <Field
            label="City"
            value={form.city}
            onChange={(v) => set("city", v)}
          />
          <Field
            label="State"
            value={form.state}
            onChange={(v) => set("state", v)}
          />
          <Field label="ZIP" value={form.zip} onChange={(v) => set("zip", v)} />
          <SelectField
            label="Housing Status"
            value={form.housing_status}
            onChange={(v) => set("housing_status", v)}
            options={[
              { value: "", label: "None" },
              { value: "own", label: "Own" },
              { value: "rent", label: "Rent" },
              { value: "family", label: "Family" },
            ]}
          />
          <Field
            label="Housing Duration"
            value={form.housing_duration}
            onChange={(v) => set("housing_duration", v)}
            placeholder="e.g. 3 years"
          />
          <Field
            label="Monthly Payment ($)"
            type="number"
            value={form.housing_monthly_payment}
            onChange={(v) => set("housing_monthly_payment", v)}
          />
        </div>
      </Section>

      {/* Employment */}
      <Section title="Employment">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field
            label="Employer Name"
            value={form.employer_name}
            onChange={(v) => set("employer_name", v)}
          />
          <Field
            label="Occupation"
            value={form.occupation}
            onChange={(v) => set("occupation", v)}
          />
          <Field
            label="Employer Phone"
            value={form.employer_phone}
            onChange={(v) => set("employer_phone", v)}
          />
          <SelectField
            label="Employment Status"
            value={form.employment_status}
            onChange={(v) => set("employment_status", v)}
            options={[
              { value: "", label: "None" },
              { value: "full_time", label: "Full Time" },
              { value: "part_time", label: "Part Time" },
              { value: "self_employed", label: "Self Employed" },
            ]}
          />
          <Field
            label="Employment Duration"
            value={form.employment_duration}
            onChange={(v) => set("employment_duration", v)}
          />
          <Field
            label="Employer Address"
            value={form.employer_address}
            onChange={(v) => set("employer_address", v)}
          />
          <Field
            label="Employer City"
            value={form.employer_city}
            onChange={(v) => set("employer_city", v)}
          />
          <Field
            label="Employer State"
            value={form.employer_state}
            onChange={(v) => set("employer_state", v)}
          />
          <Field
            label="Employer ZIP"
            value={form.employer_zip}
            onChange={(v) => set("employer_zip", v)}
          />
          <Field
            label="Gross Monthly Income ($)"
            type="number"
            value={form.gross_monthly_income}
            onChange={(v) => set("gross_monthly_income", v)}
          />
          <Field
            label="Other Monthly Income ($)"
            type="number"
            value={form.other_monthly_income}
            onChange={(v) => set("other_monthly_income", v)}
          />
          <Field
            label="Other Income Source"
            value={form.other_income_source}
            onChange={(v) => set("other_income_source", v)}
          />
        </div>
      </Section>

      {error && (
        <p role="alert" className="text-red-400 text-sm">
          {error}
        </p>
      )}

      <div className="flex justify-end gap-3 pt-1">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 border border-[#1e1e45] rounded-lg text-sm text-slate-400 hover:text-white hover:border-slate-500 transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={saving}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-all glow-blue-sm"
        >
          {saving ? "Saving…" : initial ? "Save Changes" : "Create Client"}
        </button>
      </div>
    </form>
  );
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="bg-[#0b0c0d] rounded-xl border border-[#252628] p-5">
      <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-4">
        {title}
      </h3>
      {children}
    </div>
  );
}

function Field({
  label,
  value,
  onChange,
  type = "text",
  required,
  placeholder,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  required?: boolean;
  placeholder?: string;
}) {
  const inputId = useId();
  return (
    <div>
      <label
        htmlFor={inputId}
        className="block text-xs font-medium text-slate-400 mb-1.5"
      >
        {label}
      </label>
      <input
        id={inputId}
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onWheel={type === "number" ? (e) => e.currentTarget.blur() : undefined}
        required={required}
        placeholder={placeholder}
        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white placeholder:text-slate-700 focus-visible:outline-2 focus-visible:outline-blue-400 transition-colors"
      />
    </div>
  );
}

function SelectField({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: { value: string; label: string }[];
}) {
  const inputId = useId();
  return (
    <div>
      <label
        htmlFor={inputId}
        className="block text-xs font-medium text-slate-400 mb-1.5"
      >
        {label}
      </label>
      <select
        id={inputId}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full bg-[#0b0c0d] border border-[#252628] rounded-lg px-3 py-1.5 text-sm text-white focus-visible:outline-2 focus-visible:outline-blue-400 transition-colors"
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    </div>
  );
}
