import type { Metadata } from "next";
import { Cinzel, JetBrains_Mono } from "next/font/google";
import localFont from "next/font/local";
import { IntentPrefetchManager } from "@/components/IntentPrefetchManager";
import { TopNav } from "@/components/TopNav";
import {
  SEO_DEFAULT_DESCRIPTION,
  SEO_METADATA_BASE_URL,
  SEO_SITE_NAME,
  buildAbsoluteURL,
  resolveSocialImageURL,
} from "@/lib/seo";
import "./globals.css";

const barlow = localFont({
  src: "./fonts/BarlowGX.woff2",
  weight: "100 900",
  style: "normal",
  variable: "--font-barlow",
  display: "swap",
});

const cinzel = Cinzel({
  subsets: ["latin"],
  weight: ["600", "700", "800"],
  variable: "--font-cinzel",
  display: "swap",
});

const jetBrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["500"],
  variable: "--font-jetbrains-mono",
  display: "swap",
  preload: false,
});

export const metadata: Metadata = {
  metadataBase: SEO_METADATA_BASE_URL,
  title: {
    default: SEO_SITE_NAME,
    template: `%s | ${SEO_SITE_NAME}`,
  },
  description: SEO_DEFAULT_DESCRIPTION,
  alternates: {
    canonical: "/",
  },
  openGraph: {
    type: "website",
    url: buildAbsoluteURL("/"),
    title: SEO_SITE_NAME,
    description: SEO_DEFAULT_DESCRIPTION,
    siteName: SEO_SITE_NAME,
    locale: "en_US",
    images: [
      {
        url: resolveSocialImageURL(),
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: SEO_SITE_NAME,
    description: SEO_DEFAULT_DESCRIPTION,
    images: [resolveSocialImageURL()],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-snippet": -1,
      "max-image-preview": "large",
      "max-video-preview": -1,
    },
  },
  keywords: [
    "deadlock patch notes",
    "deadlock updates",
    "deadlock hero changes",
    "deadlock spell changes",
    "deadlock item changes",
  ],
  icons: {
    icon: "/deadlock_logo.webp",
    shortcut: "/deadlock_logo.webp",
    apple: "/deadlock_logo.webp",
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${barlow.variable} ${cinzel.variable} ${jetBrainsMono.variable}`}>
      <body>
        <IntentPrefetchManager />
        <div className="site-root">
          <div className="site-background" aria-hidden>
            <div className="site-background__layer site-background__layer--base" />
            <div className="site-background__layer site-background__layer--dark" />
          </div>
          <div className="site-content">
            <TopNav />
            {children}
            <footer className="site-footer">
              <div className="shell">
                <p className="site-footer__text">Disclaimer: This is not an official product by Valve. It is fan-made.</p>
              </div>
            </footer>
          </div>
        </div>
      </body>
    </html>
  );
}
