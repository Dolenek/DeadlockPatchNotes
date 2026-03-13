import Link from "next/link";
import Image from "next/image";
import { PatchSummary } from "@/lib/types";
import { formatDisplayDate, formatUpdateLabel, timelineBlockAnchor } from "@/lib/utils";

type PatchCardProps = {
  patch: PatchSummary;
  index: number;
};

const STOCK_CARD_IMAGES = [
  "/random-card-images/04_underground.jpg",
  "/random-card-images/archmother_game.jpg",
  "/random-card-images/hidden_king_game.jpg",
  "/random-card-images/new_ss_01.jpg",
  "/random-card-images/new_ss_02.jpg",
  "/random-card-images/new_ss_03.jpg",
  "/random-card-images/new_ss_04.jpg",
  "/random-card-images/new_ss_spawn_02.jpg",
  "/random-card-images/shot_0009.jpg",
  "/random-card-images/shot_0010.jpg",
  "/random-card-images/shot_0016.jpg",
  "/random-card-images/shot_0017.jpg",
  "/random-card-images/shot_0018.jpg",
  "/random-card-images/shot_0019.jpg",
  "/random-card-images/shot_0023.jpg",
  "/random-card-images/shot_0025.jpg",
  "/random-card-images/shot_0026.jpg",
  "/random-card-images/shot_0027.jpg"
];

function resolveFallbackCardImage(slug: string) {
  let hash = 0;
  for (let i = 0; i < slug.length; i += 1) {
    hash = (hash * 31 + slug.charCodeAt(i)) >>> 0;
  }
  return STOCK_CARD_IMAGES[hash % STOCK_CARD_IMAGES.length];
}

function resolveCardImageURL(rawCoverImageURL: string | undefined, slug: string) {
  const trimmed = rawCoverImageURL?.trim() ?? "";
  if (trimmed !== "") {
    if (trimmed.startsWith("/")) {
      return { src: trimmed, fallback: false };
    }

    try {
      const parsed = new URL(trimmed);
      if (parsed.protocol === "http:" || parsed.protocol === "https:") {
        return { src: parsed.toString(), fallback: false };
      }
    } catch {
      // fall back to local random images
    }
  }

  return { src: resolveFallbackCardImage(slug), fallback: true };
}

const CARD_IMAGE_SIZES = "(max-width: 720px) 100vw, (max-width: 1024px) 50vw, 33vw";

export function PatchCard({ patch, index }: PatchCardProps) {
  const { src: imageUrl, fallback: isFallbackImage } = resolveCardImageURL(patch.imageUrl, patch.slug);
  const imageClassName = isFallbackImage ? "patch-card-image patch-card-image--fallback" : "patch-card-image";
  const prioritizeImage = index < 3;
  const followUpTimeline = (patch.releaseTimeline ?? []).filter(
    (block, index) => !(index === 0 && block.releaseType === "initial")
  );

  return (
    <article className="patch-card">
      <Link href={`/patches/${patch.slug}`} className="patch-card-link" aria-label={patch.title}>
        <div className="patch-card-image-wrap">
          <Image
            src={imageUrl}
            alt={`${patch.title} cover image`}
            fill
            sizes={CARD_IMAGE_SIZES}
            quality={68}
            className={imageClassName}
            priority={prioritizeImage}
            loading={prioritizeImage ? undefined : "lazy"}
          />
          <span className="card-corner-mark" aria-hidden>
            ↗
          </span>
        </div>
      </Link>
      <div className="patch-card-body">
        <div className="card-meta">
          <h2>
            <Link href={`/patches/${patch.slug}`} className="patch-card-title-link">
              {patch.title}
            </Link>
          </h2>
          <time dateTime={patch.publishedAt}>{formatDisplayDate(patch.publishedAt)}</time>
        </div>

        {followUpTimeline.length > 0 ? (
          <ul className="patch-card-followups">
            {followUpTimeline.map((block) => (
              <li key={block.id}>
                <Link href={`/patches/${patch.slug}#${timelineBlockAnchor(block.id)}`}>
                  {formatUpdateLabel(block.releaseType, block.releasedAt)}
                </Link>
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </article>
  );
}
