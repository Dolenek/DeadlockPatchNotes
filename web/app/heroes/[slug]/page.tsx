import Link from "next/link";
import { notFound } from "next/navigation";
import type { CSSProperties } from "react";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getHeroChanges } from "@/lib/api";
import { getHeroMediaBySlug } from "@/lib/hero-media";
import { formatDisplayDate, formatUpdateLabel } from "@/lib/utils";

type HeroDetailPageProps = {
  params: Promise<{ slug: string }>;
};

export default async function HeroDetailPage({ params }: HeroDetailPageProps) {
  const { slug } = await params;

  try {
    const payload = await getHeroChanges(slug);
    const heroMedia = getHeroMediaBySlug(payload.hero.slug);
    const heroPageStyle = heroMedia?.backgroundImageUrl
      ? ({
          "--hero-background-image": `url(${heroMedia.backgroundImageUrl})`,
        } as CSSProperties)
      : undefined;

    return (
      <main className="hero-detail-page hero-detail-page--hero" style={heroPageStyle}>
        <section className="hero-detail-hero hero-detail-hero--hero-page">
          <div className="shell hero-detail-head">
            <FallbackImage
              src={payload.hero.iconUrl}
              fallbackSrc={payload.hero.iconFallbackUrl}
              alt={payload.hero.name}
              className="hero-detail-image"
            />
            <div className="hero-detail-copy hero-detail-copy--hero-page">
              {heroMedia?.nameImageUrl ? (
                <img src={heroMedia.nameImageUrl} alt={payload.hero.name} className="hero-detail-name-image" />
              ) : (
                <h1>{payload.hero.name}</h1>
              )}
              <p>
                Most recent change:{" "}
                {payload.hero.lastChangedAt ? formatDisplayDate(payload.hero.lastChangedAt) : "Unknown"}
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
                  <h2>{formatUpdateLabel(block.releaseType, block.releasedAt)}</h2>
                </div>
                <div className="hero-timeline-meta">
                  <time dateTime={block.releasedAt}>{formatDisplayDate(block.releasedAt)}</time>
                  <Link href={`/patches/${block.patchRef.slug}`}>Open Patch</Link>
                </div>
              </header>

              {Array.isArray(block.generalChanges) && block.generalChanges.length > 0 ? (
                <section className="hero-skill-group">
                  <h3>General</h3>
                  <ul>
                    {block.generalChanges.map((change) => (
                      <li key={change.id}>{change.text}</li>
                    ))}
                  </ul>
                </section>
              ) : null}

              {block.skills.map((skill) => (
                <section className="hero-skill-group" key={skill.id}>
                  <header className="hero-skill-header">
                    <FallbackImage
                      src={skill.iconUrl}
                      fallbackSrc={skill.iconFallbackUrl}
                      alt={skill.title}
                      className="hero-skill-image"
                    />
                    <h3>{skill.title}</h3>
                  </header>
                  <ul>
                    {skill.changes.map((change) => (
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
