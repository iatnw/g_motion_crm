"use client";

import type { DealStatus } from "@/lib/types";
import StatusSelect, { dealOptions } from "@/components/StatusSelect";

interface Props {
  value?: DealStatus;
  disabled?: boolean;
  days?: number;
  bucket?: string;
  onChange: (status: DealStatus) => void | Promise<void>;
}

export default function DealStatusSelect({ value = "pending", disabled, days, bucket, onChange }: Props) {
  const aging = bucket === "red" ? "Overdue" : bucket === "yellow" ? "Watch" : "On time";
  return <StatusSelect value={value} options={dealOptions} label="Deal status" disabled={disabled}
    detail={value === "pending" && days != null ? `${days}d · ${aging}` : undefined}
    onChange={(status) => onChange(status as DealStatus)} />;
}
