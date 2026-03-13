import Link from "next/link";
import type { Metadata } from "next";
import { JsonLd } from "@/components/JsonLd";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getHeroes } from "@/lib/api";
import { HeroListResponse } from "@/lib/types";
import { formatCompactDate } from "@/lib/utils";
import { SEO_SITE_NAME, buildAbsoluteURL, resolveSocialImageURL, truncateDescription } from "@/lib/seo";

const HEROES_TITLE = "Deadlock Hero Change History";
const HEROES_DESCRIPTION = truncateDescription(
  "Browse all Deadlock heroes and track timeline-based balance changes, ability updates, and latest patch impact."
);

export const metadata: Metadata = {
  title: HEROES_TITLE,
  description: HEROES_DESCRIPTION,
  alternates: {
    canonical: "/heroes",
  },
  keywords: ["deadlock heroes", "deadlock hero patch notes", "deadlock hero changes"],
  openGraph: {
    type: "website",
    url: buildAbsoluteURL("/heroes"),
    title: HEROES_TITLE,
    description: HEROES_DESCRIPTION,
    siteName: SEO_SITE_NAME,
    images: [{ url: resolveSocialImageURL("/Oldgods_header.png") }],
  },
  twitter: {
    card: "summary_large_image",
    title: HEROES_TITLE,
    description: HEROES_DESCRIPTION,
    images: [resolveSocialImageURL("/Oldgods_header.png")],
  },
};

export default async function HeroesPage() {
  let payload: HeroListResponse = { heroes: [] };
  try {
    payload = await getHeroes();
  } catch (error) {
    if (!(error instanceof APIError) || error.status !== 404) {
      throw error;
    }
  }

  const schema = {
    "@context": "https://schema.org",
    "@type": "ItemList",
    name: "Deadlock Heroes",
    itemListElement: payload.heroes.map((hero, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: hero.name,
      url: buildAbsoluteURL(`/heroes/${hero.slug}`),
    })),
  };

  return (
    <main className="page-like-patches">
      <JsonLd data={schema} />

      <section className="heroes-masthead heroes-masthead--heroes-page">
        <div className="shell">
          <p className="eyebrow">Deadlock Heroes</p>
          <h1>Heroes</h1>
          <p>Browse hero-specific change history across all dated update blocks.</p>
        </div>
      </section>

      <section className="shell heroes-list-section heroes-list-section--heroes">
        <div className="heroes-grid">
          {payload.heroes.map((hero) => (
            <article key={hero.slug} className="hero-card">
              <Link href={`/heroes/${hero.slug}`} className="hero-card-link">
                <FallbackImage
                  src={hero.iconUrl}
                  fallbackSrc={hero.iconFallbackUrl}
                  alt={hero.name}
                  className="hero-card-image"
                />
                <div className="hero-card-copy">
                  <h2>{hero.name}</h2>
                  <p>Last change: {hero.lastChangedAt ? formatCompactDate(hero.lastChangedAt) : "Unknown"}</p>
                </div>
              </Link>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}
