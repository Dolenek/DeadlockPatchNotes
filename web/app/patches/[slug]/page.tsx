import type { Metadata } from "next";
import { cache } from "react";
import { PatchHeroesRail } from "@/components/PatchHeroesRail";
import { JsonLd } from "@/components/JsonLd";
import { notFound } from "next/navigation";
import { PatchSectionRenderer } from "@/components/PatchSectionRenderer";
import { TableOfContents, TableOfContentsGroup } from "@/components/TableOfContents";
import { APIError, getPatchBySlug } from "@/lib/api";
import { PatchDetail, PatchSection, PatchTimelineBlock } from "@/lib/types";
import {
  SEO_SITE_NAME,
  buildAbsoluteURL,
  resolveSocialImageURL,
  toISODate,
  truncateDescription,
} from "@/lib/seo";
import { entryAnchor, formatDisplayDate, formatUpdateLabel, sectionAnchor, timelineBlockAnchor } from "@/lib/utils";
import type { PatchHeroesRailBlock } from "@/components/PatchHeroesRail";

type PatchDetailPageProps = {
  params: Promise<{ slug: string }>;
};

type TimelineBlockForDisplay = PatchTimelineBlock & { sections: PatchSection[] };

const getPatch = cache(async (slug: string) => getPatchBySlug(slug));

function buildPatchDescription(patch: PatchDetail) {
  if (patch.intro.trim() !== "") {
    return truncateDescription(patch.intro);
  }
  return truncateDescription(`Read the full Deadlock patch timeline and balance changes for ${patch.title}.`);
}

function resolvePatchModifiedAt(patch: PatchDetail) {
  const timeline = patch.releaseTimeline ?? [];
  const latestTimelineDate = timeline.reduce<string | undefined>((latest, block) => {
    if (!latest) {
      return block.releasedAt;
    }
    return block.releasedAt > latest ? block.releasedAt : latest;
  }, undefined);

  return toISODate(latestTimelineDate ?? patch.publishedAt);
}

export async function generateMetadata({ params }: PatchDetailPageProps): Promise<Metadata> {
  const { slug } = await params;

  try {
    const patch = await getPatch(slug);
    const title = patch.title;
    const description = buildPatchDescription(patch);
    const canonicalPath = `/patches/${patch.slug}`;
    const publishedTime = toISODate(patch.publishedAt);
    const modifiedTime = resolvePatchModifiedAt(patch);
    const imageURL = resolveSocialImageURL(patch.imageUrl);

    return {
      title,
      description,
      alternates: {
        canonical: canonicalPath,
      },
      keywords: ["deadlock patch notes", "deadlock update", patch.title.toLowerCase()],
      openGraph: {
        type: "article",
        url: buildAbsoluteURL(canonicalPath),
        title,
        description,
        siteName: SEO_SITE_NAME,
        images: [{ url: imageURL }],
        publishedTime,
        modifiedTime,
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
        title: "Patch Not Found",
        robots: { index: false, follow: false },
      };
    }
    throw error;
  }
}

export default async function PatchDetailPage({ params }: PatchDetailPageProps) {
  const { slug } = await params;

  try {
    const patch = await getPatch(slug);
    const timeline = buildTimelineForDisplay(patch.releaseTimeline, patch.sections, patch.source, patch.publishedAt);
    const tocGroups = buildTimelineTableOfContents(timeline);
    const heroRailBlocks = buildTimelineHeroRail(timeline);
    const patchURL = buildAbsoluteURL(`/patches/${patch.slug}`);
    const patchSchema = {
      "@context": "https://schema.org",
      "@type": "Article",
      headline: patch.title,
      description: buildPatchDescription(patch),
      datePublished: toISODate(patch.publishedAt),
      dateModified: resolvePatchModifiedAt(patch),
      image: [resolveSocialImageURL(patch.imageUrl)],
      author: {
        "@type": "Organization",
        name: "Valve",
      },
      publisher: {
        "@type": "Organization",
        name: SEO_SITE_NAME,
        url: buildAbsoluteURL("/"),
      },
      mainEntityOfPage: patchURL,
      url: patchURL,
    };

    return (
      <main className="patch-detail-page">
        <JsonLd data={patchSchema} />

        <section className="patch-hero">
          <div className="shell patch-hero-content">
            <h1>{patch.title}</h1>
            <div className="patch-meta">
              <time dateTime={patch.publishedAt}>{formatDisplayDate(patch.publishedAt)}</time>
              <a href={patch.source.url} target="_blank" rel="noreferrer">
                Open Original
              </a>
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
                    <p className="eyebrow">{formatUpdateLabel(block.releaseType, block.releasedAt)}</p>
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
      releaseType: "initial",
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
    label: formatUpdateLabel(block.releaseType, block.releasedAt),
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
      label: formatUpdateLabel(block.releaseType, block.releasedAt),
      heroes
    };
  });
}
