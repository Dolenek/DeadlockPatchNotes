import { notFound } from "next/navigation";
import { PatchSectionRenderer } from "@/components/PatchSectionRenderer";
import { TableOfContents } from "@/components/TableOfContents";
import { APIError, getPatchBySlug } from "@/lib/api";
import { formatDisplayDate } from "@/lib/utils";

type PatchDetailPageProps = {
  params: Promise<{ slug: string }>;
};

export default async function PatchDetailPage({ params }: PatchDetailPageProps) {
  const { slug } = await params;

  try {
    const patch = await getPatchBySlug(slug);

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

        <section className="shell patch-detail-content">
          <TableOfContents sections={patch.sections} />
          <div className="patch-sections-column">
            {patch.sections.map((section) => (
              <PatchSectionRenderer section={section} key={section.id} />
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
