"use client";

import type { DealStatus } from "@/lib/types";

const statusStyles: Record<DealStatus, string> = {
  paid: "bg-green-500/10 text-green-400 border-green-500/25",
  pending: "bg-yellow-500/10 text-yellow-400 border-yellow-500/25",
  cancelled: "bg-red-500/10 text-red-400 border-red-500/25",
};

interface Props {
  value?: DealStatus;
  disabled?: boolean;
  days?: number;
  bucket?: string;
  onChange: (status: DealStatus) => void | Promise<void>;
}

export default function DealStatusSelect({ value = "pending", disabled, days, bucket, onChange }: Props) {
  const color = statusStyles[value] ?? statusStyles.pending;
  const suffix = value === "pending" && days != null ? ` · ${days}d ${bucket ?? ""}` : "";

  return (
    <label>
      <span className="sr-only">Deal status</span>
      <select
        value={value}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value as DealStatus)}
        className={`appearance-none cursor-pointer rounded-md border px-2 py-1 text-xs font-medium outline-none transition-colors disabled:cursor-wait disabled:opacity-60 ${color}`}
      >
        <option value="pending">pending{suffix}</option>
        <option value="paid">paid</option>
        <option value="cancelled">cancelled</option>
      </select>
    </label>
  );
}
