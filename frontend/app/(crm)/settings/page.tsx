"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { auth } from "@/lib/api";

export default function SettingsPage() {
  const router = useRouter();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleChangePassword(e: React.FormEvent) {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }
    setLoading(true);
    setError("");
    try {
      await auth.setPassword(newPassword, currentPassword);
      router.replace("/login");
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : "Failed to change password",
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold text-white">Settings</h1>
      <p className="mb-7 mt-2 text-sm text-slate-400">
        Manage your account and keep your workspace secure.
      </p>

      <div className="max-w-xl bg-[#0f1011] rounded-2xl border border-[#252628] p-6">
        <h2 className="text-base font-semibold text-white">Change password</h2>
        <p className="mb-6 mt-2 text-sm leading-relaxed text-slate-400">
          Changing your password signs out every device. You’ll sign in again
          with your new password.
        </p>
        <form onSubmit={handleChangePassword} className="space-y-4">
          <div>
            <label
              htmlFor="current-password"
              className="block text-xs font-medium text-slate-400 mb-1.5"
            >
              Current Password
            </label>
            <input
              type="password"
              id="current-password"
              autoComplete="current-password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              required
              className="w-full rounded-lg bg-[#0b0c0d] border border-[#252628] text-white px-3 py-2.5 text-sm focus:outline-none transition-colors"
              placeholder="Enter current password"
            />
          </div>
          <div>
            <label
              htmlFor="new-password"
              className="block text-xs font-medium text-slate-400 mb-1.5"
            >
              New Password
            </label>
            <input
              type="password"
              id="new-password"
              autoComplete="new-password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              minLength={12}
              className="w-full rounded-lg bg-[#0b0c0d] border border-[#252628] text-white px-3 py-2.5 text-sm focus:outline-none transition-colors"
              placeholder="At least 12 characters"
            />
          </div>
          <div>
            <label
              htmlFor="confirm-password"
              className="block text-xs font-medium text-slate-400 mb-1.5"
            >
              Confirm New Password
            </label>
            <input
              type="password"
              id="confirm-password"
              autoComplete="new-password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              className="w-full rounded-lg bg-[#0b0c0d] border border-[#252628] text-white px-3 py-2.5 text-sm focus:outline-none transition-colors"
              placeholder="Repeat new password"
            />
          </div>
          {error && (
            <p role="alert" className="text-red-400 text-sm">
              {error}
            </p>
          )}
          <button
            type="submit"
            disabled={loading}
            className="w-full py-2.5 px-4 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-medium rounded-lg text-sm transition-all"
          >
            {loading ? "Saving…" : "Change Password"}
          </button>
        </form>
      </div>
    </div>
  );
}
