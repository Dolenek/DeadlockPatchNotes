import Link from "next/link";
import { notFound } from "next/navigation";
import { FallbackImage } from "@/components/FallbackImage";
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
      <main className="hero-detail-page">
        <section className="hero-detail-hero">
          <div className="shell hero-detail-head">
            <FallbackImage
              src={payload.item.iconUrl}
              fallbackSrc={payload.item.iconFallbackUrl}
              alt={payload.item.name}
              className="hero-detail-image"
            />
            <div className="hero-detail-copy">
              <p className="eyebrow">Item Timeline</p>
              <h1>{payload.item.name}</h1>
              <p>
                Most recent change:{" "}
                {payload.item.lastChangedAt ? formatDisplayDate(payload.item.lastChangedAt) : "Unknown"}
              </p>
            </div>
          </div>
        </section>

        <section className="shell hero-timeline-section">
          {payload.items.map((block) => (
            <article key={block.id} className="hero-timeline-block">
              <header className="hero-timeline-header">
                <div>
                  <p className="eyebrow">{formatUpdateLabel(block.kind, block.releasedAt)}</p>
                  <h2>{block.patch.title}</h2>
                </div>
                <div className="hero-timeline-meta">
                  <time dateTime={block.releasedAt}>{formatDisplayDate(block.releasedAt)}</time>
                  <Link href={`/patches/${block.patch.slug}`}>Open Patch</Link>
                </div>
              </header>

              <section className="hero-skill-group">
                <h3>{payload.item.name}</h3>
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
