import Link from "next/link";
import { notFound } from "next/navigation";
import { ItemAbstractPattern } from "@/components/ItemAbstractPattern";
import { FallbackImage } from "@/components/FallbackImage";
import { ItemGradientFromImage } from "@/components/ItemGradientFromImage";
import { APIError, getItemChanges } from "@/lib/api";
import { formatDisplayDate, formatUpdateLabel } from "@/lib/utils";

type ItemDetailPageProps = {
  params: Promise<{ slug: string }>;
};

export default async function ItemDetailPage({ params }: ItemDetailPageProps) {
  const { slug } = await params;

  try {
    const payload = await getItemChanges(slug);

    return (
      <main className="hero-detail-page item-detail-page" id="item-detail-page">
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
                  <h2>{block.patchRef.title}</h2>
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
