import { Pagination } from "@/components/Pagination";
import type { Metadata } from "next";
import { DecorativeImageLayers } from "@/components/DecorativeImageLayers";
import { JsonLd } from "@/components/JsonLd";
import { PatchCard } from "@/components/PatchCard";
import { getPatches } from "@/lib/api";
import { clampPage } from "@/lib/utils";
import { SEO_SITE_NAME, buildAbsoluteURL, resolveSocialImageURL, truncateDescription } from "@/lib/seo";

type PatchesPageProps = {
  searchParams: Promise<{ page?: string | string[] }>;
};

export async function generateMetadata({ searchParams }: PatchesPageProps): Promise<Metadata> {
  const resolvedParams = await searchParams;
  const page = clampPage(resolvedParams.page);
  const isFirstPage = page <= 1;
  const title = isFirstPage ? "Deadlock Patch Notes Archive" : `Deadlock Patch Notes Archive - Page ${page}`;
  const description = truncateDescription(
    isFirstPage
      ? "Read full Deadlock patch notes with historical update timelines and cross-linked hero, item, and spell changes."
      : `Deadlock patch notes archive page ${page}. Browse older Deadlock balance updates and timeline entries.`
  );

  return {
    title,
    description,
    alternates: {
      canonical: isFirstPage ? "/patches" : `/patches?page=${page}`,
    },
    robots: {
      index: isFirstPage,
      follow: true,
    },
    keywords: ["deadlock patch notes", "deadlock patch history", "deadlock updates"],
    openGraph: {
      type: "website",
      url: buildAbsoluteURL("/patches", isFirstPage ? undefined : { page }),
      title,
      description,
      siteName: SEO_SITE_NAME,
      images: [
        {
          url: resolveSocialImageURL("/Oldgods_header.png"),
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title,
      description,
      images: [resolveSocialImageURL("/Oldgods_header.png")],
    },
  };
}

export default async function PatchesPage({ searchParams }: PatchesPageProps) {
  const resolvedParams = await searchParams;
  const page = clampPage(resolvedParams.page);
  const patchList = await getPatches(page, 12);
  const canonicalURL = buildAbsoluteURL("/patches", page > 1 ? { page } : undefined);
  const schema = {
    "@context": "https://schema.org",
    "@type": "CollectionPage",
    name: page > 1 ? `Deadlock Patch Notes - Page ${page}` : "Deadlock Patch Notes Archive",
    url: canonicalURL,
    isPartOf: {
      "@type": "WebSite",
      name: SEO_SITE_NAME,
      url: buildAbsoluteURL("/"),
    },
  };

  return (
    <main className="page-like-patches">
      <JsonLd data={schema} />

      <section className="patch-list-masthead">
        <DecorativeImageLayers
          className="patch-list-masthead-media"
          layers={[
            {
              src: "/Oldgods_header.png",
              className: "patch-list-masthead-media__layer patch-list-masthead-media__layer--oldgods",
              quality: 56,
              priority: true,
            },
          ]}
        />
        <div className="shell">
          <p className="eyebrow">Deadlock Updates</p>
          <h1>Patch Notes</h1>
        </div>
      </section>

      <section className="shell patch-list-section">
        <div className="patch-grid">
          {patchList.patches.map((patch, index) => (
            <PatchCard key={patch.id} patch={patch} index={index} />
          ))}
        </div>

        <Pagination page={patchList.pagination.page} totalPages={patchList.pagination.totalPages} />
      </section>
    </main>
  );
}
