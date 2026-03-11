import Link from "next/link";
import { notFound } from "next/navigation";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getHeroChanges } from "@/lib/api";
import { formatDisplayDate, formatUpdateLabel } from "@/lib/utils";

type HeroDetailPageProps = {
  params: Promise<{ slug: string }>;
};

export default async function HeroDetailPage({ params }: HeroDetailPageProps) {
  const { slug } = await params;

  try {
    const payload = await getHeroChanges(slug);

    return (
      <main className="hero-detail-page">
        <section className="hero-detail-hero">
          <div className="shell hero-detail-head">
            <FallbackImage
              src={payload.hero.iconUrl}
              fallbackSrc={payload.hero.iconFallbackUrl}
              alt={payload.hero.name}
              className="hero-detail-image"
            />
            <div className="hero-detail-copy">
              <p className="eyebrow">Hero Timeline</p>
              <h1>{payload.hero.name}</h1>
              <p>
                Most recent change:{" "}
                {payload.hero.lastChangedAt ? formatDisplayDate(payload.hero.lastChangedAt) : "Unknown"}
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
