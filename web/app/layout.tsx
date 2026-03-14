import type { Metadata } from "next";
import { TopNav } from "@/components/TopNav";
import {
  SEO_DEFAULT_DESCRIPTION,
  SEO_METADATA_BASE_URL,
  SEO_SITE_NAME,
  buildAbsoluteURL,
  resolveSocialImageURL,
} from "@/lib/seo";
import "./globals.css";

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
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://assets-bucket.deadlock-api.com" crossOrigin="" />
        <link rel="preconnect" href="https://assets.deadlock-api.com" crossOrigin="" />
        <link rel="preconnect" href="https://clan.akamai.steamstatic.com" crossOrigin="" />
        <link rel="dns-prefetch" href="//assets-bucket.deadlock-api.com" />
        <link rel="dns-prefetch" href="//assets.deadlock-api.com" />
        <link rel="dns-prefetch" href="//clan.akamai.steamstatic.com" />
      </head>
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
