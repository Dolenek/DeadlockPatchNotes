import { notFound } from "next/navigation";
import { PatchSectionRenderer } from "@/components/PatchSectionRenderer";
import { APIError, getPatchBySlug } from "@/lib/api";
import { PatchSection, PatchTimelineBlock } from "@/lib/types";
import { formatDisplayDate, formatUpdateLabel } from "@/lib/utils";

type PatchDetailPageProps = {
  params: Promise<{ slug: string }>;
};

export default async function PatchDetailPage({ params }: PatchDetailPageProps) {
  const { slug } = await params;

  try {
    const patch = await getPatchBySlug(slug);
    const timeline = buildTimelineForDisplay(patch.timeline, patch.sections, patch.source, patch.publishedAt);

    return (
      <main className="patch-detail-page">
        <section className="patch-hero" style={{ backgroundImage: `url(${patch.heroImageUrl})` }}>
          <div className="patch-hero-overlay">
            <div className="shell patch-hero-content">
              <p className="eyebrow">{patch.category}</p>
              <h1>{patch.title}</h1>
              <div className="patch-meta">
                <time dateTime={patch.publishedAt}>{formatDisplayDate(patch.publishedAt)}</time>
                <span>Source: {patch.source.type}</span>
                <a href={patch.source.url} target="_blank" rel="noreferrer">
                  Open Original
                </a>
              </div>
              <p className="patch-intro">{patch.intro}</p>
            </div>
          </div>
        </section>

        <section className="shell patch-detail-content patch-detail-content--timeline">
          <div className="patch-sections-column">
            {timeline.map((block) => (
              <article className="timeline-block" key={block.id}>
                <header className="timeline-block-header">
                  <div className="timeline-block-heading">
                    <p className="eyebrow">{formatUpdateLabel(block.kind, block.releasedAt)}</p>
                    <h2>{block.title}</h2>
                  </div>
                  <div className="timeline-block-meta">
                    <time dateTime={block.releasedAt}>{formatDisplayDate(block.releasedAt)}</time>
                    <a href={block.source.url} target="_blank" rel="noreferrer">
                      Open Source
                    </a>
                  </div>
                </header>
                <div className="timeline-block-sections">
                  {block.sections.map((section) => (
                    <PatchSectionRenderer
                      section={{
                        ...section,
                        id: `${block.id}-${section.kind}`
                      }}
                      key={`${block.id}-${section.kind}`}
                    />
                  ))}
                </div>
              </article>
            ))}
          </div>
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

function buildTimelineForDisplay(
  timeline: PatchTimelineBlock[] | undefined,
  fallbackSections: PatchSection[],
  fallbackSource: { type: string; url: string },
  fallbackPublishedAt: string
) {
  if (Array.isArray(timeline) && timeline.length > 0) {
    return timeline.map((block) => ({
      ...block,
      sections: Array.isArray(block.sections) && block.sections.length > 0 ? block.sections : fallbackSections
    }));
  }

  return [
    {
      id: "fallback-initial",
      kind: "initial",
      title: "Initial Update",
      releasedAt: fallbackPublishedAt,
      source: fallbackSource,
      changes: [],
      sections: fallbackSections
    }
  ];
}
