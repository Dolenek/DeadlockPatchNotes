import type { Metadata } from "next";
import Link from "next/link";
import Image from "next/image";
import { JsonLd } from "@/components/JsonLd";
import { APIError, getPatches } from "@/lib/api";
import { PatchListResponse } from "@/lib/types";
import { formatDisplayDate, formatUpdateLabel } from "@/lib/utils";
import {
  SEO_DEFAULT_DESCRIPTION,
  SEO_SITE_NAME,
  buildAbsoluteURL,
  resolveSocialImageURL,
  truncateDescription,
} from "@/lib/seo";

const HOME_TITLE = "Latest Deadlock Patch Notes, Hero, Item, and Spell Changes";
const HOME_DESCRIPTION = truncateDescription(
  "Track every Deadlock patch note in one place. Browse full patch timelines, hero ability changes, item balance updates, and spell history."
);

export const metadata: Metadata = {
  title: HOME_TITLE,
  description: HOME_DESCRIPTION,
  alternates: {
    canonical: "/",
  },
  keywords: [
    "deadlock patch notes",
    "latest deadlock update",
    "deadlock hero changes",
    "deadlock item balance",
    "deadlock spell updates",
  ],
  openGraph: {
    type: "website",
    url: buildAbsoluteURL("/"),
    title: HOME_TITLE,
    description: HOME_DESCRIPTION,
    siteName: SEO_SITE_NAME,
    images: [
      {
        url: resolveSocialImageURL(),
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: HOME_TITLE,
    description: HOME_DESCRIPTION,
    images: [resolveSocialImageURL()],
  },
};

const HUB_IMAGE_SIZES = "(max-width: 720px) 100vw, (max-width: 1100px) 50vw, 25vw";
const LAST_UPDATE_IMAGE_FALLBACK = "/random-card-images/new_ss_01.jpg";

function resolveLatestPatchImage(rawImageURL: string | undefined) {
  const normalized = rawImageURL?.trim() ?? "";
  if (normalized === "") {
    return LAST_UPDATE_IMAGE_FALLBACK;
  }

  if (normalized.startsWith("/")) {
    return normalized;
  }

  try {
    const parsed = new URL(normalized);
    if (parsed.protocol === "http:" || parsed.protocol === "https:") {
      return parsed.toString();
    }
  } catch {
    // fall back to local image when API data is malformed
  }

  return LAST_UPDATE_IMAGE_FALLBACK;
}

export default async function HomePage() {
  let patchList: PatchListResponse = {
    patches: [],
    pagination: { page: 1, pageSize: 1, totalItems: 0, totalPages: 1 },
  };

  try {
    patchList = await getPatches(1, 1);
  } catch (error) {
    if (!(error instanceof APIError) || error.status !== 404) {
      throw error;
    }
  }

  const latestPatch = patchList.patches[0];
  const latestPatchImage = resolveLatestPatchImage(latestPatch?.imageUrl);
  const followUpTimeline = (latestPatch?.releaseTimeline ?? []).filter(
    (timelineBlock, index) => !(index === 0 && timelineBlock.releaseType === "initial")
  );

  const websiteURL = buildAbsoluteURL("/");
  const patchesURL = buildAbsoluteURL("/patches");
  const schema = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "WebSite",
        name: SEO_SITE_NAME,
        url: websiteURL,
        description: SEO_DEFAULT_DESCRIPTION,
      },
      {
        "@type": "CollectionPage",
        name: "Deadlock Patch Notes Archive",
        url: patchesURL,
        isPartOf: {
          "@type": "WebSite",
          url: websiteURL,
          name: SEO_SITE_NAME,
        },
      },
    ],
  };

  return (
    <main className="page-like-patches">
      <JsonLd data={schema} />

      <section className="patch-list-masthead patch-list-masthead--home">
        <div className="shell">
          <p className="eyebrow">Community Archive</p>
          <h1>
            Deadlock <span className="home-title-nowrap">Patch Notes</span>
          </h1>
          <p>
            Browse complete Deadlock update history with timeline-level patch notes, hero adjustments, item balance changes,
            and spell updates.
          </p>

          <div className="home-hub-grid" role="navigation" aria-label="Browse sections">
            <article className="home-hub-card home-hub-card--last-update">
              {latestPatch ? (
                <Link href={`/patches/${latestPatch.slug}`} className="home-hub-card-link">
                  <div className="home-hub-card-media">
                    <Image
                      src={latestPatchImage}
                      alt={`${latestPatch.title} cover image`}
                      fill
                      sizes={HUB_IMAGE_SIZES}
                      quality={68}
                      className="home-hub-card-image"
                      priority
                    />
                  </div>
                  <div className="home-hub-last-content">
                    <p className="home-hub-label">Last Update</p>
                    <div className="home-hub-last-row">
                      <h2>{latestPatch.title}</h2>
                      <time dateTime={latestPatch.publishedAt}>{formatDisplayDate(latestPatch.publishedAt)}</time>
                    </div>
                    {followUpTimeline.length > 0 ? (
                      <ul className="home-hub-last-followups">
                        {followUpTimeline.map((timelineBlock) => (
                          <li key={timelineBlock.id}>{formatUpdateLabel(timelineBlock.releaseType, timelineBlock.releasedAt)}</li>
                        ))}
                      </ul>
                    ) : null}
                  </div>
                </Link>
              ) : (
                <Link href="/patches" className="home-hub-card-link home-hub-card-link--fallback">
                  <div className="home-hub-last-content">
                    <p className="home-hub-label">Last Update</p>
                    <div className="home-hub-last-row">
                      <h2>Update Unavailable</h2>
                    </div>
                    <p className="home-hub-empty">
                      Latest patch data is temporarily unavailable. Open the patch archive to retry.
                    </p>
                  </div>
                </Link>
              )}
            </article>

            <article className="home-hub-card home-hub-card--feature home-hub-card--heroes">
              <Link href="/heroes" className="home-hub-card-link">
                <div className="home-hub-card-media">
                  <Image
                    src="/header_heroes.png"
                    alt="Heroes update history"
                    fill
                    sizes={HUB_IMAGE_SIZES}
                    quality={74}
                    className="home-hub-card-image"
                  />
                </div>
                <div className="home-hub-feature-copy">
                  <p className="home-hub-label">Explore</p>
                  <h2>Hero Change History</h2>
                </div>
              </Link>
            </article>

            <article className="home-hub-card home-hub-card--feature home-hub-card--items">
              <Link href="/items" className="home-hub-card-link">
                <div className="home-hub-card-media">
                  <Image
                    src="/Items.png"
                    alt="Item update history"
                    fill
                    sizes={HUB_IMAGE_SIZES}
                    quality={74}
                    className="home-hub-card-image"
                  />
                </div>
                <div className="home-hub-feature-copy">
                  <p className="home-hub-label">Explore</p>
                  <h2>Item Change History</h2>
                </div>
              </Link>
            </article>

            <article className="home-hub-card home-hub-card--feature home-hub-card--spells">
              <Link href="/spells" className="home-hub-card-link">
                <div className="home-hub-card-media">
                  <Image
                    src="/rem_helper.png"
                    alt="Spell update history"
                    fill
                    sizes={HUB_IMAGE_SIZES}
                    quality={74}
                    className="home-hub-card-image"
                  />
                </div>
                <div className="home-hub-feature-copy">
                  <p className="home-hub-label">Explore</p>
                  <h2>Spell Change History</h2>
                </div>
              </Link>
            </article>
          </div>
        </div>
      </section>
    </main>
  );
}
