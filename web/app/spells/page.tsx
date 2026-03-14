import Link from "next/link";
import type { Metadata } from "next";
import { JsonLd } from "@/components/JsonLd";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getSpells } from "@/lib/api";
import { SpellListResponse } from "@/lib/types";
import { formatCompactDate } from "@/lib/utils";
import { SEO_SITE_NAME, buildAbsoluteURL, resolveSocialImageURL, truncateDescription } from "@/lib/seo";

const SPELLS_TITLE = "Deadlock Spell Change History";
const SPELLS_DESCRIPTION = truncateDescription(
  "Browse Deadlock spell timelines and review how each ability has changed across updates and patch cycles."
);

export const metadata: Metadata = {
  title: SPELLS_TITLE,
  description: SPELLS_DESCRIPTION,
  alternates: {
    canonical: "/spells",
  },
  keywords: ["deadlock spells", "deadlock ability patch notes", "deadlock spell changes"],
  openGraph: {
    type: "website",
    url: buildAbsoluteURL("/spells"),
    title: SPELLS_TITLE,
    description: SPELLS_DESCRIPTION,
    siteName: SEO_SITE_NAME,
    images: [{ url: resolveSocialImageURL("/Oldgods_header.png") }],
  },
  twitter: {
    card: "summary_large_image",
    title: SPELLS_TITLE,
    description: SPELLS_DESCRIPTION,
    images: [resolveSocialImageURL("/Oldgods_header.png")],
  },
};

export default async function SpellsPage() {
  let payload: SpellListResponse = { spells: [] };
  try {
    payload = await getSpells();
  } catch (error) {
    if (!(error instanceof APIError) || error.status !== 404) {
      throw error;
    }
  }

  const schema = {
    "@context": "https://schema.org",
    "@type": "ItemList",
    name: "Deadlock Spells",
    itemListElement: payload.spells.map((spell, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: spell.name,
      url: buildAbsoluteURL(`/spells/${spell.slug}`),
    })),
  };

  return (
    <main className="page-like-patches">
      <JsonLd data={schema} />

      <section className="heroes-masthead">
        <div className="shell">
          <p className="eyebrow">Deadlock Spells</p>
          <h1>Spells</h1>
          <p>Browse spell-specific change history across all dated update blocks.</p>
        </div>
      </section>

      <section className="shell heroes-list-section heroes-list-section--spells">
        <div className="heroes-grid">
          {payload.spells.map((spell, index) => (
            <article key={spell.slug} className="hero-card">
              <Link href={`/spells/${spell.slug}`} className="hero-card-link">
                <FallbackImage
                  src={spell.iconUrl}
                  fallbackSrc={spell.iconFallbackUrl}
                  alt={spell.name}
                  className="hero-card-image"
                  loading={index < 6 ? "eager" : "lazy"}
                  fetchPriority={index < 3 ? "high" : "auto"}
                  width={96}
                  height={96}
                />
                <div className="hero-card-copy">
                  <h2>{spell.name}</h2>
                  <p>Last change: {spell.lastChangedAt ? formatCompactDate(spell.lastChangedAt) : "Unknown"}</p>
                </div>
              </Link>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}
