"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { auth, setToken } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const [pendingToken, setPendingToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await auth.login(password);
      if (res.totp_required) {
        setToken(res.token);
        setPendingToken(res.token);
      } else {
        setToken(res.token);
        router.push("/clients");
      }
    } catch {
      setError("Invalid password");
    } finally {
      setLoading(false);
    }
  }

  async function handleTOTP(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await auth.verifyTOTP(totpCode);
      setToken(res.token);
      router.push("/clients");
    } catch {
      setError("Invalid code");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-full flex items-center justify-center bg-[#0b0c0d]" style={{ backgroundImage: "radial-gradient(ellipse at 50% 0%, rgba(59,130,246,0.08) 0%, transparent 60%)" }}>
      <div className="w-full max-w-sm px-4">
        {/* Logo mark */}
        <div className="flex flex-col items-center mb-8">
          <div className="w-12 h-12 rounded-2xl bg-blue-600 flex items-center justify-center mb-4 glow-blue">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">G Motion CRM</h1>
          <p className="text-slate-500 text-sm mt-1">Auto Finance Management</p>
        </div>

        <div className="bg-[#0b0c0d] rounded-2xl p-8 border border-[#252628] border-glow">
          {!pendingToken ? (
            <form onSubmit={handleLogin} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                  Password
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  autoFocus
                  className="w-full rounded-lg bg-[#0b0c0d] border border-[#252628] text-white px-3 py-2.5 text-sm focus:outline-none transition-colors placeholder:text-slate-700"
                  placeholder="Enter your password"
                />
              </div>
              {error && <p className="text-red-400 text-sm">{error}</p>}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 px-4 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-medium rounded-lg text-sm transition-all glow-blue-sm hover:glow-blue"
              >
                {loading ? "Signing in…" : "Sign in"}
              </button>
            </form>
          ) : (
            <form onSubmit={handleTOTP} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                  Authenticator Code
                </label>
                <input
                  type="text"
                  inputMode="numeric"
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value)}
                  required
                  autoFocus
                  maxLength={6}
                  className="w-full rounded-lg bg-[#0b0c0d] border border-[#252628] text-white px-3 py-2.5 text-sm text-center tracking-widest text-lg focus:outline-none transition-colors"
                  placeholder="000000"
                />
              </div>
              {error && <p className="text-red-400 text-sm">{error}</p>}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 px-4 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-medium rounded-lg text-sm transition-all glow-blue-sm"
              >
                {loading ? "Verifying…" : "Verify"}
              </button>
              <button
                type="button"
                onClick={() => setPendingToken(null)}
                className="w-full text-slate-500 text-xs hover:text-slate-300 transition-colors"
              >
                Back to password
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
