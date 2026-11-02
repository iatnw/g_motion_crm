"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { auth } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [setupSecret, setSetupSecret] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const [totpRequired, setTotpRequired] = useState(false);
  const [setupRequired, setSetupRequired] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    auth
      .setupRequired()
      .then((res) => setSetupRequired(res.required))
      .catch(() => setSetupRequired(false));
  }, []);

  async function handleSetup(e: React.FormEvent) {
    e.preventDefault();
    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }
    setLoading(true);
    setError("");
    try {
      await auth.setup(setupSecret, password);
      setSetupRequired(false);
      setPassword("");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to set password");
    } finally {
      setLoading(false);
    }
  }

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await auth.login(password);
      if (res.totp_required) {
        setTotpRequired(true);
      } else {
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
      await auth.verifyTOTP(totpCode);
      router.push("/clients");
    } catch {
      setError("Invalid code");
    } finally {
      setLoading(false);
    }
  }

  if (setupRequired === null)
    return (
      <div
        role="status"
        className="flex h-full items-center justify-center text-sm text-slate-400"
      >
        Connecting to your workspace…
      </div>
    );

  return (
    <div
      className="min-h-full flex items-center justify-center bg-[#0b0c0d]"
      style={{
        backgroundImage:
          "radial-gradient(ellipse at 50% 0%, rgba(59,130,246,0.08) 0%, transparent 60%)",
      }}
    >
      <div className="w-full max-w-sm px-4">
        <div className="flex flex-col items-center mb-8">
          <div className="w-12 h-12 rounded-2xl bg-blue-600 flex items-center justify-center mb-4 glow-blue">
            <svg
              width="22"
              height="22"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">
            G Motion CRM
          </h1>
          <p className="text-slate-500 text-sm mt-1">Auto Finance Management</p>
        </div>

        <div className="bg-[#0b0c0d] rounded-2xl p-8 border border-[#252628] border-glow">
          {setupRequired ? (
            <form onSubmit={handleSetup} className="space-y-4">
              <div className="mb-2">
                <p className="text-white text-sm font-medium">
                  Welcome to G Motion CRM
                </p>
                <p className="text-slate-500 text-xs mt-1">
                  Set your password to get started.
                </p>
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                  Setup Secret
                </label>
                <input
                  type="password"
                  value={setupSecret}
                  onChange={(e) => setSetupSecret(e.target.value)}
                  required
                  autoFocus
                  className="w-full rounded-lg bg-[#0b0c0d] border border-[#252628] text-white px-3 py-2.5 text-sm focus:outline-none transition-colors placeholder:text-slate-700"
                  placeholder="From your deployment environment"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                  Password
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  minLength={12}
                  className="w-full rounded-lg bg-[#0b0c0d] border border-[#252628] text-white px-3 py-2.5 text-sm focus:outline-none transition-colors placeholder:text-slate-700"
                  placeholder="At least 12 characters"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1.5 uppercase tracking-wider">
                  Confirm Password
                </label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                  className="w-full rounded-lg bg-[#0b0c0d] border border-[#252628] text-white px-3 py-2.5 text-sm focus:outline-none transition-colors placeholder:text-slate-700"
                  placeholder="Repeat your password"
                />
              </div>
              {error && <p className="text-red-400 text-sm">{error}</p>}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 px-4 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-medium rounded-lg text-sm transition-all glow-blue-sm hover:glow-blue"
              >
                {loading ? "Setting up…" : "Set Password"}
              </button>
            </form>
          ) : !totpRequired ? (
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
                onClick={() => setTotpRequired(false)}
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
