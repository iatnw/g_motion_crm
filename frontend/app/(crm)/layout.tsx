import Sidebar from "@/components/Sidebar";
import AuthGuard from "@/components/AuthGuard";

export default function CRMLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthGuard>
      <div className="flex h-full bg-[#0b0c0d]">
        <Sidebar />
        <main className="flex-1 overflow-y-auto bg-[#0b0c0d] uppercase">{children}</main>
      </div>
    </AuthGuard>
  );
}
