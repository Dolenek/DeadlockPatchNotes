import Link from "next/link";
import type { Metadata } from "next";
import { DecorativeImageLayers } from "@/components/DecorativeImageLayers";
import { JsonLd } from "@/components/JsonLd";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getItems } from "@/lib/api";
import { ItemListResponse } from "@/lib/types";
import { formatCompactDate } from "@/lib/utils";
import { SEO_SITE_NAME, buildAbsoluteURL, resolveSocialImageURL, truncateDescription } from "@/lib/seo";

const ITEMS_TITLE = "Deadlock Item Change History";
const ITEMS_DESCRIPTION = truncateDescription(
  "Track Deadlock item balance updates over time with full timeline links back to each patch note."
);

export const metadata: Metadata = {
  title: ITEMS_TITLE,
  description: ITEMS_DESCRIPTION,
  alternates: {
    canonical: "/items",
  },
  keywords: ["deadlock items", "deadlock item patch notes", "deadlock item changes"],
  openGraph: {
    type: "website",
    url: buildAbsoluteURL("/items"),
    title: ITEMS_TITLE,
    description: ITEMS_DESCRIPTION,
    siteName: SEO_SITE_NAME,
    images: [{ url: resolveSocialImageURL("/Oldgods_header.png") }],
  },
  twitter: {
    card: "summary_large_image",
    title: ITEMS_TITLE,
    description: ITEMS_DESCRIPTION,
    images: [resolveSocialImageURL("/Oldgods_header.png")],
  },
};

export default async function ItemsPage() {
  let payload: ItemListResponse = { items: [] };
  try {
    payload = await getItems();
  } catch (error) {
    if (!(error instanceof APIError) || error.status !== 404) {
      throw error;
    }
  }

  const schema = {
    "@context": "https://schema.org",
    "@type": "ItemList",
    name: "Deadlock Items",
    itemListElement: payload.items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      url: buildAbsoluteURL(`/items/${item.slug}`),
    })),
  };

  return (
    <main className="page-like-patches">
      <JsonLd data={schema} />

      <section className="heroes-masthead heroes-masthead--items-page">
        <DecorativeImageLayers
          className="heroes-masthead-media"
          layers={[
            {
              src: "/Items.png",
              className: "heroes-masthead-media__layer heroes-masthead-media__layer--items",
              quality: 54,
              priority: true,
            },
          ]}
        />
        <div className="shell">
          <p className="eyebrow">Deadlock Items</p>
          <h1>Items</h1>
          <p>Browse item-specific change history across all dated update blocks.</p>
        </div>
      </section>

      <section className="shell heroes-list-section heroes-list-section--items">
        <div className="heroes-grid">
          {payload.items.map((item, index) => (
            <article key={item.slug} className="hero-card">
              <Link href={`/items/${item.slug}`} className="hero-card-link">
                <FallbackImage
                  src={item.iconUrl}
                  fallbackSrc={item.iconFallbackUrl}
                  alt={item.name}
                  className="hero-card-image"
                  loading={index < 6 ? "eager" : "lazy"}
                  fetchPriority={index < 3 ? "high" : "auto"}
                  width={96}
                  height={96}
                />
                <div className="hero-card-copy">
                  <h2>{item.name}</h2>
                  <p>Last change: {item.lastChangedAt ? formatCompactDate(item.lastChangedAt) : "Unknown"}</p>
                </div>
              </Link>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}
