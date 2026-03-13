import Link from "next/link";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { cache, type CSSProperties } from "react";
import { FallbackImage } from "@/components/FallbackImage";
import { JsonLd } from "@/components/JsonLd";
import { APIError, getHeroChanges } from "@/lib/api";
import { getHeroMediaBySlug } from "@/lib/hero-media";
import { SEO_SITE_NAME, buildAbsoluteURL, resolveSocialImageURL, toISODate, truncateDescription } from "@/lib/seo";
import { buildPatchTimelineHref, formatDisplayDate, formatUpdateLabel, normalizeLookupKey, slugifyLookup } from "@/lib/utils";

type HeroDetailPageProps = {
  params: Promise<{ slug: string }>;
};

const getHeroTimeline = cache(async (slug: string) => getHeroChanges(slug));

function buildHeroDescription(heroName: string) {
  return truncateDescription(`Track all patch note changes for ${heroName} in Deadlock, including general balance and ability updates.`);
}

export async function generateMetadata({ params }: HeroDetailPageProps): Promise<Metadata> {
  const { slug } = await params;

  try {
    const payload = await getHeroTimeline(slug);
    const title = `${payload.hero.name} Hero Changes`;
    const description = buildHeroDescription(payload.hero.name);
    const canonicalPath = `/heroes/${payload.hero.slug}`;
    const imageURL = resolveSocialImageURL(payload.hero.iconUrl ?? payload.hero.iconFallbackUrl);

    return {
      title,
      description,
      alternates: {
        canonical: canonicalPath,
      },
      keywords: ["deadlock hero changes", payload.hero.name.toLowerCase(), "deadlock patch notes"],
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
        title: "Hero Not Found",
        robots: { index: false, follow: false },
      };
    }
    throw error;
  }
}

export default async function HeroDetailPage({ params }: HeroDetailPageProps) {
  const { slug } = await params;

  try {
    const payload = await getHeroTimeline(slug);
    const heroMedia = getHeroMediaBySlug(payload.hero.slug);
    const heroPageStyle = heroMedia?.backgroundImageUrl
      ? ({
          "--hero-background-image": `url(${heroMedia.backgroundImageUrl})`,
        } as CSSProperties)
      : undefined;
    const canonicalURL = buildAbsoluteURL(`/heroes/${payload.hero.slug}`);
    const schema = {
      "@context": "https://schema.org",
      "@type": "WebPage",
      name: `${payload.hero.name} Hero Changes`,
      description: buildHeroDescription(payload.hero.name),
      url: canonicalURL,
      mainEntity: {
        "@type": "Thing",
        name: payload.hero.name,
        url: canonicalURL,
      },
      dateModified: toISODate(payload.hero.lastChangedAt),
    };

    return (
      <main className="hero-detail-page hero-detail-page--hero" style={heroPageStyle}>
        <JsonLd data={schema} />

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
                <>
                  <img src={heroMedia.nameImageUrl} alt={payload.hero.name} className="hero-detail-name-image" />
                  <h1 className="visually-hidden">{payload.hero.name}</h1>
                </>
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
                  <h2>
                    <Link
                      href={buildPatchTimelineHref(block.patchRef.slug, block.id, payload.hero.slug)}
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

              {block.skills.map((skill) => {
                const skillSlug = slugifyLookup(skill.title);
                const skillNameNormalized = normalizeLookupKey(skill.title);
                const isMetaSkill = skillNameNormalized === "talents" || skillNameNormalized === "card types";
                const shouldLinkSkill = !isMetaSkill && skillSlug !== "entry";

                return (
                  <section className="hero-skill-group" key={skill.id}>
                    <header className="hero-skill-header">
                      <FallbackImage
                        src={skill.iconUrl}
                        fallbackSrc={skill.iconFallbackUrl}
                        alt={skill.title}
                        className="hero-skill-image"
                      />
                      <h3>
                        {shouldLinkSkill ? (
                          <Link href={`/spells/${skillSlug}`} className="hero-skill-title-link">
                            {skill.title}
                          </Link>
                        ) : (
                          skill.title
                        )}
                      </h3>
                    </header>
                    <ul>
                      {skill.changes.map((change) => (
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
