import Link from "next/link";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getHeroes } from "@/lib/api";
import { HeroListResponse } from "@/lib/types";
import { formatCompactDate } from "@/lib/utils";

export default async function HeroesPage() {
  let payload: HeroListResponse = { heroes: [] };
  try {
    payload = await getHeroes();
  } catch (error) {
    if (!(error instanceof APIError) || error.status !== 404) {
      throw error;
    }
  }

  return (
    <main className="page-like-patches">
      <section className="heroes-masthead">
        <div className="shell">
          <p className="eyebrow">Deadlock Heroes</p>
          <h1>Heroes</h1>
          <p>Browse hero-specific change history across all dated update blocks.</p>
        </div>
      </section>

      <section className="shell heroes-list-section">
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
