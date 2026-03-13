import Link from "next/link";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { cache, type CSSProperties } from "react";
import { FallbackImage } from "@/components/FallbackImage";
import { JsonLd } from "@/components/JsonLd";
import { APIError, getSpellChanges } from "@/lib/api";
import { getHeroMediaBySlug } from "@/lib/hero-media";
import { SEO_SITE_NAME, buildAbsoluteURL, resolveSocialImageURL, toISODate, truncateDescription } from "@/lib/seo";
import { SpellChangesResponse } from "@/lib/types";
import { buildPatchTimelineHref, formatDisplayDate, formatUpdateLabel } from "@/lib/utils";

type SpellDetailPageProps = {
  params: Promise<{ slug: string }>;
};

const getSpellTimeline = cache(async (slug: string) => getSpellChanges(slug));

function buildSpellDescription(spellName: string) {
  return truncateDescription(
    `Track all Deadlock patch note changes for ${spellName}, including timeline updates across heroes and releases.`
  );
}

function resolveDominantHeroSlug(payload: SpellChangesResponse) {
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

  return dominantHeroSlug;
}

export async function generateMetadata({ params }: SpellDetailPageProps): Promise<Metadata> {
  const { slug } = await params;

  try {
    const payload = await getSpellTimeline(slug);
    const title = `${payload.spell.name} Spell Changes`;
    const description = buildSpellDescription(payload.spell.name);
    const canonicalPath = `/spells/${payload.spell.slug}`;
    const imageURL = resolveSocialImageURL(payload.spell.iconUrl ?? payload.spell.iconFallbackUrl);

    return {
      title,
      description,
      alternates: {
        canonical: canonicalPath,
      },
      keywords: ["deadlock spell changes", payload.spell.name.toLowerCase(), "deadlock patch notes"],
      openGraph: {
        type: "website",
        url: buildAbsoluteURL(canonicalPath),
        title,
        description,
        siteName: SEO_SITE_NAME,
        images: [{ url: imageURL }],
      },
      twitter: {
        card: "summary_large_image",
        title,
        description,
        images: [imageURL],
      },
    };
  } catch (error) {
    if (error instanceof APIError && error.status === 404) {
      return {
        title: "Spell Not Found",
        robots: { index: false, follow: false },
      };
    }
    throw error;
  }
}

export default async function SpellDetailPage({ params }: SpellDetailPageProps) {
  const { slug } = await params;

  try {
    const payload = await getSpellTimeline(slug);
    const dominantHeroSlug = resolveDominantHeroSlug(payload);

    const heroMedia = dominantHeroSlug ? getHeroMediaBySlug(dominantHeroSlug) : null;
    const spellPageStyle = heroMedia?.backgroundImageUrl
      ? ({
          "--spell-background-image": `url(${heroMedia.backgroundImageUrl})`,
        } as CSSProperties)
      : undefined;
    const canonicalURL = buildAbsoluteURL(`/spells/${payload.spell.slug}`);
    const schema = {
      "@context": "https://schema.org",
      "@type": "WebPage",
      name: `${payload.spell.name} Spell Changes`,
      description: buildSpellDescription(payload.spell.name),
      url: canonicalURL,
      mainEntity: {
        "@type": "Thing",
        name: payload.spell.name,
        url: canonicalURL,
      },
      dateModified: toISODate(payload.spell.lastChangedAt),
    };

    return (
      <main className="hero-detail-page spell-detail-page" style={spellPageStyle}>
        <JsonLd data={schema} />

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
