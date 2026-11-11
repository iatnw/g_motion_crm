import type { StatusOption } from "@/components/StatusSelect";
import type { Client, ClientSource, ClientStatus } from "@/lib/types";

export const clientStatusLabels: Record<ClientStatus, string> = {
  active: "Active",
  closed: "Closed",
  no_deal: "No deal",
};
export const clientSourceLabels: Record<ClientSource, string> = {
  instagram: "Instagram",
  telegram: "Telegram",
  referall: "Referral",
  organic: "Organic",
  repeat_client: "Repeat client",
};

const sourceDescriptions: Record<ClientSource, string> = {
  instagram: "Introduced through Instagram",
  telegram: "Introduced through Telegram",
  referall: "Introduced by another person",
  organic: "Found your business directly",
  repeat_client: "Returning for another deal",
};
export const clientSourceOptions: StatusOption[] = Object.entries(clientSourceLabels).map(([value, label]) => ({
  value, label, description: sourceDescriptions[value as ClientSource], tone: "blue",
}));

export function ClientAvatar({
  client,
  large = false,
}: {
  client: Pick<Client, "first_name" | "last_name">;
  large?: boolean;
}) {
  return (
    <span
      aria-hidden="true"
      className={`flex shrink-0 items-center justify-center rounded-xl border border-blue-500/20 bg-blue-500/10 font-semibold text-blue-400 ${large ? "h-14 w-14 text-lg" : "h-10 w-10 text-xs"}`}
    >
      {client.first_name.charAt(0)}
      {client.last_name.charAt(0)}
    </span>
  );
}

export function ClientStatusBadge({ status }: { status: ClientStatus }) {
  const colors =
    status === "active"
      ? "border-green-500/20 bg-green-500/10 text-green-400"
      : status === "closed"
        ? "border-blue-500/20 bg-blue-500/10 text-blue-400"
        : "border-[#252628] bg-[#1a1b1d] text-slate-400";
  return (
    <span
      className={`inline-flex items-center gap-1.5 whitespace-nowrap rounded-full border px-2.5 py-1 text-xs font-medium ${colors}`}
    >
      <span
        aria-hidden="true"
        className="h-1.5 w-1.5 rounded-full bg-current"
      />
      {clientStatusLabels[status] ?? status}
    </span>
  );
}
