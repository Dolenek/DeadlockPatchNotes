import Link from "next/link";
import { notFound } from "next/navigation";
import type { CSSProperties } from "react";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getSpellChanges } from "@/lib/api";
import { getHeroMediaBySlug } from "@/lib/hero-media";
import { buildPatchTimelineHref, formatDisplayDate, formatUpdateLabel } from "@/lib/utils";

type SpellDetailPageProps = {
  params: Promise<{ slug: string }>;
};

export default async function SpellDetailPage({ params }: SpellDetailPageProps) {
  const { slug } = await params;

  try {
    const payload = await getSpellChanges(slug);
    const slugScores = new Map<string, number>();
    for (const block of payload.timeline) {
      for (const entry of block.entries) {
        const heroSlug = String(entry.heroSlug || "").trim().toLowerCase();
        if (!heroSlug) {
          continue;
        }
        slugScores.set(heroSlug, (slugScores.get(heroSlug) ?? 0) + 1);
      }
    }

    let dominantHeroSlug = "";
    let dominantScore = -1;
    for (const [heroSlug, score] of slugScores.entries()) {
      if (score > dominantScore) {
        dominantHeroSlug = heroSlug;
        dominantScore = score;
      }
    }

    const heroMedia = dominantHeroSlug ? getHeroMediaBySlug(dominantHeroSlug) : null;
    const spellPageStyle = heroMedia?.backgroundImageUrl
      ? ({
          "--spell-background-image": `url(${heroMedia.backgroundImageUrl})`,
        } as CSSProperties)
      : undefined;

    return (
      <main className="hero-detail-page spell-detail-page" style={spellPageStyle}>
        <section className="hero-detail-hero spell-detail-hero">
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
          {payload.timeline.map((block) => (
            <article key={block.id} className="hero-timeline-block">
              <header className="hero-timeline-header">
                <div>
                  <p className="eyebrow">{block.patchRef.title}</p>
                  <h2>
                    <Link
                      href={buildPatchTimelineHref(block.patchRef.slug, block.id, payload.spell.slug)}
                      className="hero-timeline-title-link"
                    >
                      {formatUpdateLabel(block.releaseType, block.releasedAt)}
                    </Link>
                  </h2>
                </div>
                <div className="hero-timeline-meta">
                  <time dateTime={block.releasedAt}>{formatDisplayDate(block.releasedAt)}</time>
                  <Link href={`/patches/${block.patchRef.slug}`}>Open Patch</Link>
                </div>
              </header>

              {block.entries.map((entry) => {
                const heroSlug = String(entry.heroSlug ?? "").trim().toLowerCase();

                return (
                  <section className="hero-skill-group" key={entry.id}>
                    <header className="hero-skill-header">
                      <FallbackImage
                        src={entry.heroIconUrl}
                        fallbackSrc={entry.heroIconFallbackUrl}
                        alt={entry.heroName ?? payload.spell.name}
                        className="hero-skill-image"
                      />
                      <h3>
                        {heroSlug ? (
                          <Link href={`/heroes/${heroSlug}`} className="hero-skill-title-link">
                            {entry.heroName ?? payload.spell.name}
                          </Link>
                        ) : (
                          entry.heroName ?? payload.spell.name
                        )}
                      </h3>
                    </header>
                    <ul>
                      {entry.changes.map((change) => (
                        <li key={change.id}>{change.text}</li>
                      ))}
                    </ul>
                  </section>
                );
              })}
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
