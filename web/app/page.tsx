import type { Metadata } from "next";
import Link from "next/link";
import { JsonLd } from "@/components/JsonLd";
import { PatchCard } from "@/components/PatchCard";
import { APIError, getPatches } from "@/lib/api";
import { PatchListResponse } from "@/lib/types";
import {
  SEO_DEFAULT_DESCRIPTION,
  SEO_SITE_NAME,
  buildAbsoluteURL,
  resolveSocialImageURL,
  truncateDescription,
} from "@/lib/seo";

const HOME_TITLE = "Latest Deadlock Patch Notes, Hero, Item, and Spell Changes";
const HOME_DESCRIPTION = truncateDescription(
  "Track every Deadlock patch note in one place. Browse full patch timelines, hero ability changes, item balance updates, and spell history."
);

export const metadata: Metadata = {
  title: HOME_TITLE,
  description: HOME_DESCRIPTION,
  alternates: {
    canonical: "/",
  },
  keywords: [
    "deadlock patch notes",
    "latest deadlock update",
    "deadlock hero changes",
    "deadlock item balance",
    "deadlock spell updates",
  ],
  openGraph: {
    type: "website",
    url: buildAbsoluteURL("/"),
    title: HOME_TITLE,
    description: HOME_DESCRIPTION,
    siteName: SEO_SITE_NAME,
    images: [
      {
        url: resolveSocialImageURL(),
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: HOME_TITLE,
    description: HOME_DESCRIPTION,
    images: [resolveSocialImageURL()],
  },
};

export default async function HomePage() {
  let patchList: PatchListResponse = {
    patches: [],
    pagination: { page: 1, pageSize: 6, totalItems: 0, totalPages: 1 },
  };

  try {
    patchList = await getPatches(1, 6);
  } catch (error) {
    if (!(error instanceof APIError) || error.status !== 404) {
      throw error;
    }
  }

  const websiteURL = buildAbsoluteURL("/");
  const patchesURL = buildAbsoluteURL("/patches");
  const schema = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "WebSite",
        name: SEO_SITE_NAME,
        url: websiteURL,
        description: SEO_DEFAULT_DESCRIPTION,
      },
      {
        "@type": "CollectionPage",
        name: "Deadlock Patch Notes Archive",
        url: patchesURL,
        isPartOf: {
          "@type": "WebSite",
          url: websiteURL,
          name: SEO_SITE_NAME,
        },
      },
    ],
  };

  return (
    <main className="page-like-patches">
      <JsonLd data={schema} />

      <section className="patch-list-masthead">
        <div className="shell">
          <p className="eyebrow">Community Archive</p>
          <h1>Deadlock Patch Notes</h1>
          <p>
            Browse complete Deadlock update history with timeline-level patch notes, hero adjustments, item balance changes,
            and spell updates.
          </p>
          <div className="home-section-links" role="navigation" aria-label="Browse sections">
            <Link href="/patches">Patch Notes Archive</Link>
            <Link href="/heroes">Hero Timelines</Link>
            <Link href="/items">Item Timelines</Link>
            <Link href="/spells">Spell Timelines</Link>
          </div>
        </div>
      </section>

      <section className="shell patch-list-section">
        <div className="home-section-head">
          <h2>Latest Updates</h2>
          <Link href="/patches" className="home-view-all">
            View full archive
          </Link>
        </div>

        <div className="patch-grid">
          {patchList.patches.length > 0 ? (
            patchList.patches.map((patch, index) => <PatchCard key={patch.id} patch={patch} index={index} />)
          ) : (
            <p>Latest updates are temporarily unavailable. Open the full archive to retry.</p>
          )}
        </div>
      </section>
    </main>
  );
}
