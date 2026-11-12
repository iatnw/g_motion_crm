"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { auth, AUTH_EXPIRED_EVENT } from "@/lib/api";

export default function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [verified, setVerified] = useState(false);

  useEffect(() => {
    let active = true;
    let expired = false;
    function signOut() {
      if (!active) return;
      expired = true;
      setVerified(false);
      router.replace("/login");
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, signOut);
    auth.me().then(() => { if (active && !expired) setVerified(true); }).catch(signOut);
    return () => {
      active = false;
      window.removeEventListener(AUTH_EXPIRED_EVENT, signOut);
    };
  }, [router]);

  if (!verified) return <div role="status" className="flex h-full items-center justify-center bg-[#0b0c0d] text-sm text-slate-400">Opening your workspace…</div>;

  return <>{children}</>;
}
