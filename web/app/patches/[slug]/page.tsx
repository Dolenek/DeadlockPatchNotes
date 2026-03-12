import { PatchHeroesRail } from "@/components/PatchHeroesRail";
import { notFound } from "next/navigation";
import { PatchSectionRenderer } from "@/components/PatchSectionRenderer";
import { TableOfContents, TableOfContentsGroup } from "@/components/TableOfContents";
import { APIError, getPatchBySlug } from "@/lib/api";
import { PatchSection, PatchTimelineBlock } from "@/lib/types";
import { entryAnchor, formatDisplayDate, formatUpdateLabel, sectionAnchor, timelineBlockAnchor } from "@/lib/utils";
import type { PatchHeroesRailBlock } from "@/components/PatchHeroesRail";

type PatchDetailPageProps = {
  params: Promise<{ slug: string }>;
};

type TimelineBlockForDisplay = PatchTimelineBlock & { sections: PatchSection[] };

export default async function PatchDetailPage({ params }: PatchDetailPageProps) {
  const { slug } = await params;

  try {
    const patch = await getPatchBySlug(slug);
    const timeline = buildTimelineForDisplay(patch.timeline, patch.sections, patch.source, patch.publishedAt);
    const tocGroups = buildTimelineTableOfContents(timeline);
    const heroRailBlocks = buildTimelineHeroRail(timeline);

    return (
      <main className="patch-detail-page">
        <section className="patch-hero" style={{ backgroundImage: `url(${patch.heroImageUrl})` }}>
          <div className="patch-hero-overlay">
            <div className="shell patch-hero-content">
              <h1>{patch.title}</h1>
              <div className="patch-meta">
                <time dateTime={patch.publishedAt}>{formatDisplayDate(patch.publishedAt)}</time>
                <a href={patch.source.url} target="_blank" rel="noreferrer">
                  Open Original
                </a>
              </div>
            </div>
          </div>
        </section>

        <section className="patch-detail-content patch-detail-content--timeline">
          <TableOfContents groups={tocGroups} />
          <div className="patch-sections-column">
            {timeline.map((block) => (
              <article className="timeline-block" id={timelineBlockAnchor(block.id)} key={block.id}>
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
                  {block.sections.map((section) => {
                    const sectionID = `${block.id}-${section.kind}`;
                    return <PatchSectionRenderer section={{ ...section, id: sectionID }} key={sectionID} />;
                  })}
                </div>
              </article>
            ))}
          </div>
          <PatchHeroesRail blocks={heroRailBlocks} />
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
): TimelineBlockForDisplay[] {
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

function buildTimelineTableOfContents(timeline: Array<PatchTimelineBlock & { sections: PatchSection[] }>): TableOfContentsGroup[] {
  return timeline.map((block) => ({
    id: timelineBlockAnchor(block.id),
    label: formatUpdateLabel(block.kind, block.releasedAt),
    sections: block.sections.map((section) => {
      const sectionID = `${block.id}-${section.kind}`;
      return {
        id: sectionAnchor(sectionID),
        label: section.title
      };
    })
  }));
}

function buildTimelineHeroRail(timeline: TimelineBlockForDisplay[]): PatchHeroesRailBlock[] {
  return timeline.map((block) => {
    const heroSection = block.sections.find((section) => section.kind === "heroes");
    const sectionID = heroSection ? `${block.id}-${heroSection.kind}` : "";
    const heroes =
      heroSection?.entries.map((entry) => ({
        id: `${block.id}-${entry.id}`,
        label: entry.entityName,
        targetId: entryAnchor(`${sectionID}-${entry.id}`),
        iconUrl: entry.entityIconUrl,
        iconFallbackUrl: entry.entityIconFallbackUrl
      })) ?? [];

    return {
      id: timelineBlockAnchor(block.id),
      label: formatUpdateLabel(block.kind, block.releasedAt),
      heroes
    };
  });
}
