import Link from "next/link";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { cache } from "react";
import { ItemAbstractPattern } from "@/components/ItemAbstractPattern";
import { FallbackImage } from "@/components/FallbackImage";
import { ItemGradientFromImage } from "@/components/ItemGradientFromImage";
import { JsonLd } from "@/components/JsonLd";
import { APIError, getItemChanges } from "@/lib/api";
import { SEO_SITE_NAME, buildAbsoluteURL, resolveSocialImageURL, toISODate, truncateDescription } from "@/lib/seo";
import { buildPatchTimelineHref, formatDisplayDate, formatUpdateLabel } from "@/lib/utils";

type ItemDetailPageProps = {
  params: Promise<{ slug: string }>;
};

const getItemTimeline = cache(async (slug: string) => getItemChanges(slug));

function buildItemDescription(itemName: string) {
  return truncateDescription(`Track all Deadlock patch note changes for ${itemName}, including buffs, nerfs, and balance adjustments.`);
}

export async function generateMetadata({ params }: ItemDetailPageProps): Promise<Metadata> {
  const { slug } = await params;

  try {
    const payload = await getItemTimeline(slug);
    const title = `${payload.item.name} Item Changes`;
    const description = buildItemDescription(payload.item.name);
    const canonicalPath = `/items/${payload.item.slug}`;
    const imageURL = resolveSocialImageURL(payload.item.iconUrl ?? payload.item.iconFallbackUrl);

    return {
      title,
      description,
      alternates: {
        canonical: canonicalPath,
      },
      keywords: ["deadlock item changes", payload.item.name.toLowerCase(), "deadlock patch notes"],
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
        title: "Item Not Found",
        robots: { index: false, follow: false },
      };
    }
    throw error;
  }
}

export default async function ItemDetailPage({ params }: ItemDetailPageProps) {
  const { slug } = await params;

  try {
    const payload = await getItemTimeline(slug);
    const canonicalURL = buildAbsoluteURL(`/items/${payload.item.slug}`);
    const schema = {
      "@context": "https://schema.org",
      "@type": "WebPage",
      name: `${payload.item.name} Item Changes`,
      description: buildItemDescription(payload.item.name),
      url: canonicalURL,
      mainEntity: {
        "@type": "Thing",
        name: payload.item.name,
        url: canonicalURL,
      },
      dateModified: toISODate(payload.item.lastChangedAt),
    };

    return (
      <main className="hero-detail-page item-detail-page" id="item-detail-page">
        <JsonLd data={schema} />

        <ItemGradientFromImage
          targetID="item-detail-page"
          src={payload.item.iconUrl}
          fallbackSrc={payload.item.iconFallbackUrl}
        />
        <ItemAbstractPattern />
        <section className="hero-detail-hero item-detail-hero">
          <div className="shell hero-detail-head item-detail-head">
            <FallbackImage
              src={payload.item.iconUrl}
              fallbackSrc={payload.item.iconFallbackUrl}
              alt={payload.item.name}
              className="hero-detail-image item-detail-image"
              loading="eager"
              fetchPriority="high"
              width={168}
              height={168}
            />
            <div className="hero-detail-copy item-detail-copy">
              <p className="eyebrow">Item Timeline</p>
              <h1>{payload.item.name}</h1>
              <p>
                Most recent change:{" "}
                {payload.item.lastChangedAt ? formatDisplayDate(payload.item.lastChangedAt) : "Unknown"}
              </p>
            </div>
          </div>
        </section>

        <section className="shell hero-timeline-section item-timeline-section">
          {payload.timeline.map((block) => (
            <article key={block.id} className="hero-timeline-block item-timeline-block">
              <header className="hero-timeline-header item-timeline-header">
                <div>
                  <p className="eyebrow">{formatUpdateLabel(block.releaseType, block.releasedAt)}</p>
                  <h2>
                    <Link
                      href={buildPatchTimelineHref(block.patchRef.slug, block.id, payload.item.slug)}
                      className="hero-timeline-title-link"
                    >
                      {block.patchRef.title}
                    </Link>
                  </h2>
                </div>
                <div className="hero-timeline-meta item-timeline-meta">
                  <time dateTime={block.releasedAt}>{formatDisplayDate(block.releasedAt)}</time>
                  <Link href={`/patches/${block.patchRef.slug}`}>Open Patch</Link>
                </div>
              </header>

              <section className="hero-skill-group item-timeline-change-group">
                <ul>
                  {block.changes.map((change) => (
                    <li key={change.id}>{change.text}</li>
                  ))}
                </ul>
              </section>
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
