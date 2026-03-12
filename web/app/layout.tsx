import type { Metadata } from "next";
import { TopNav } from "@/components/TopNav";
import "./globals.css";

export const metadata: Metadata = {
  title: "Deadlock Patch Notes",
  description: "Deadlock patch notes archive with long-form gameplay updates and balance changes.",
  icons: {
    icon: "/deadlock_logo.webp",
    shortcut: "/deadlock_logo.webp",
    apple: "/deadlock_logo.webp"
  }
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="site-root">
          <div className="site-background" aria-hidden>
            <div className="site-background__layer site-background__layer--base" />
            <div className="site-background__layer site-background__layer--dark" />
            <div className="site-background__layer site-background__layer--darkest" />
            <div className="site-background__void" />
          </div>
          <div className="site-content">
            <TopNav />
            {children}
          </div>
        </div>
      </body>
    </html>
  );
}
