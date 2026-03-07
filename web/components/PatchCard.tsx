import Link from "next/link";
import { PatchSummary } from "@/lib/types";
import { formatDisplayDate } from "@/lib/utils";

type PatchCardProps = {
  patch: PatchSummary;
};

export function PatchCard({ patch }: PatchCardProps) {
  return (
    <article className="patch-card">
      <Link href={`/patches/${patch.slug}`} className="patch-card-link" aria-label={patch.title}>
        <div className="patch-card-image-wrap">
          <img src={patch.coverImageUrl} alt="" loading="lazy" className="patch-card-image" />
          <span className="card-corner-mark" aria-hidden>
            ↗
          </span>
        </div>
        <div className="patch-card-body">
          <div className="card-meta">
            <span>{patch.category}</span>
            <time dateTime={patch.publishedAt}>{formatDisplayDate(patch.publishedAt)}</time>
          </div>
          <h2>{patch.title}</h2>
          <p>{patch.excerpt}</p>
        </div>
      </Link>
    </article>
  );
}
