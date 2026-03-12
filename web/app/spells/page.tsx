import Link from "next/link";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getSpells } from "@/lib/api";
import { SpellListResponse } from "@/lib/types";
import { formatCompactDate } from "@/lib/utils";

export default async function SpellsPage() {
  let payload: SpellListResponse = { items: [] };
  try {
    payload = await getSpells();
  } catch (error) {
    if (!(error instanceof APIError) || error.status !== 404) {
      throw error;
    }
  }

  return (
    <main className="page-like-patches">
      <section className="heroes-masthead">
        <div className="shell">
          <p className="eyebrow">Deadlock Spells</p>
          <h1>Spells</h1>
          <p>Browse spell-specific change history across all dated update blocks.</p>
        </div>
      </section>

      <section className="shell heroes-list-section">
        <div className="heroes-grid">
          {payload.items.map((spell) => (
            <article key={spell.slug} className="hero-card">
              <Link href={`/spells/${spell.slug}`} className="hero-card-link">
                <FallbackImage
                  src={spell.iconUrl}
                  fallbackSrc={spell.iconFallbackUrl}
                  alt={spell.name}
                  className="hero-card-image"
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
