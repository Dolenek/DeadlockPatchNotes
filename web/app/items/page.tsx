import Link from "next/link";
import { FallbackImage } from "@/components/FallbackImage";
import { APIError, getItems } from "@/lib/api";
import { ItemListResponse } from "@/lib/types";
import { formatCompactDate } from "@/lib/utils";

export default async function ItemsPage() {
  let payload: ItemListResponse = { items: [] };
  try {
    payload = await getItems();
  } catch (error) {
    if (!(error instanceof APIError) || error.status !== 404) {
      throw error;
    }
  }

  return (
    <main>
      <section className="heroes-masthead">
        <div className="shell">
          <p className="eyebrow">Deadlock Items</p>
          <h1>Items</h1>
          <p>Browse item-specific change history across all dated update blocks.</p>
        </div>
      </section>

      <section className="shell heroes-list-section">
        <div className="heroes-grid">
          {payload.items.map((item) => (
            <article key={item.slug} className="hero-card">
              <Link href={`/items/${item.slug}`} className="hero-card-link">
                <FallbackImage
                  src={item.iconUrl}
                  fallbackSrc={item.iconFallbackUrl}
                  alt={item.name}
                  className="hero-card-image"
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
