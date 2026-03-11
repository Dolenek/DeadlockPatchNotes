import Link from "next/link";
import { notFound } from "next/navigation";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getSpellChanges } from "@/lib/api";
import { formatDisplayDate, formatUpdateLabel } from "@/lib/utils";

type SpellDetailPageProps = {
  params: Promise<{ slug: string }>;
};

export default async function SpellDetailPage({ params }: SpellDetailPageProps) {
  const { slug } = await params;

  try {
    const payload = await getSpellChanges(slug);

    return (
      <main className="hero-detail-page">
        <section className="hero-detail-hero">
          <div className="shell hero-detail-head">
            <FallbackImage
              src={payload.spell.iconUrl}
              fallbackSrc={payload.spell.iconFallbackUrl}
              alt={payload.spell.name}
              className="hero-detail-image"
            />
            <div className="hero-detail-copy">
              <p className="eyebrow">Spell Timeline</p>
              <h1>{payload.spell.name}</h1>
              <p>
                Most recent change:{" "}
                {payload.spell.lastChangedAt ? formatDisplayDate(payload.spell.lastChangedAt) : "Unknown"}
              </p>
            </div>
          </div>
        </section>

        <section className="shell hero-timeline-section">
          {payload.items.map((block) => (
            <article key={block.id} className="hero-timeline-block">
              <header className="hero-timeline-header">
                <div>
                  <p className="eyebrow">{block.patch.title}</p>
                  <h2>{formatUpdateLabel(block.kind, block.releasedAt)}</h2>
                </div>
                <div className="hero-timeline-meta">
                  <time dateTime={block.releasedAt}>{formatDisplayDate(block.releasedAt)}</time>
                  <Link href={`/patches/${block.patch.slug}`}>Open Patch</Link>
                </div>
              </header>

              {block.entries.map((entry) => (
                <section className="hero-skill-group" key={entry.id}>
                  <header className="hero-skill-header">
                    <FallbackImage
                      src={entry.heroIconUrl}
                      fallbackSrc={entry.heroIconFallbackUrl}
                      alt={entry.heroName ?? payload.spell.name}
                      className="hero-skill-image"
                    />
                    <h3>{entry.heroName ?? payload.spell.name}</h3>
                  </header>
                  <ul>
                    {entry.changes.map((change) => (
                      <li key={change.id}>{change.text}</li>
                    ))}
                  </ul>
                </section>
              ))}
            </article>
          ))}
        </section>
      </main>
    );
  } catch (error) {
    if (error instanceof APIError && error.status === 404) {
      notFound();
    }

    throw error;
  }
}
