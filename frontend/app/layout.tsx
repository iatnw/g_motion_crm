import type { Metadata } from "next";
import "./globals.css";
import "./appearance.css";

export const metadata: Metadata = {
  title: "G Motion CRM",
  description: "G Motion Auto Finance CRM",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="h-full" data-theme="dark" suppressHydrationWarning>
      <head>
        {/* Apply saved theme before first paint to avoid flash */}
        <script dangerouslySetInnerHTML={{ __html: `(function(){try{var t=localStorage.getItem('theme');document.documentElement.setAttribute('data-theme',t==='light'?'light':'dark');}catch(e){}})()` }} />
      </head>
      <body className="h-full overflow-hidden font-sans antialiased">{children}</body>
    </html>
  );
}
