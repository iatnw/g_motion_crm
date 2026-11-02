"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { auth } from "@/lib/api";

export default function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [verified, setVerified] = useState(false);

  useEffect(() => {
    auth.me().then(() => setVerified(true)).catch(() => router.replace("/login"));
  }, [router]);

  if (!verified) return <div role="status" className="flex h-full items-center justify-center bg-[#0b0c0d] text-sm text-slate-400">Opening your workspace…</div>;

  return <>{children}</>;
}
