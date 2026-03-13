import type { MetadataRoute } from "next";
import { getHeroes, getItems, getPatches, getSpells } from "@/lib/api";
import type { HeroSummary, ItemSummary, PatchSummary, SpellSummary } from "@/lib/types";
import { buildAbsoluteURL, toISODate } from "@/lib/seo";

export const revalidate = 1800;

const PATCH_PAGE_LIMIT = 50;

function createCoreRoutes(): MetadataRoute.Sitemap {
  const now = new Date();
  return [
    {
      url: buildAbsoluteURL("/"),
      lastModified: now,
      changeFrequency: "daily",
      priority: 1,
    },
    {
      url: buildAbsoluteURL("/patches"),
      lastModified: now,
      changeFrequency: "hourly",
      priority: 0.95,
    },
    {
      url: buildAbsoluteURL("/heroes"),
      lastModified: now,
      changeFrequency: "daily",
      priority: 0.85,
    },
    {
      url: buildAbsoluteURL("/items"),
      lastModified: now,
      changeFrequency: "daily",
      priority: 0.85,
    },
    {
      url: buildAbsoluteURL("/spells"),
      lastModified: now,
      changeFrequency: "daily",
      priority: 0.85,
    },
  ];
}

async function collectAllPatchSummaries() {
  const firstPage = await getPatches(1, PATCH_PAGE_LIMIT);
  const summaries: PatchSummary[] = [...firstPage.patches];
  const totalPages = Math.max(1, firstPage.pagination.totalPages);

  for (let page = 2; page <= totalPages; page += 1) {
    const nextPage = await getPatches(page, PATCH_PAGE_LIMIT);
    summaries.push(...nextPage.patches);
  }

  return summaries;
}

function createPatchEntries(patches: PatchSummary[]): MetadataRoute.Sitemap {
  return patches.map((patch) => ({
    url: buildAbsoluteURL(`/patches/${patch.slug}`),
    lastModified: toISODate(patch.publishedAt) ?? new Date(),
    changeFrequency: "weekly",
    priority: 0.9,
  }));
}

function createEntityEntries(
  entityType: "heroes" | "items" | "spells",
  entries: Array<HeroSummary | ItemSummary | SpellSummary>
): MetadataRoute.Sitemap {
  return entries.map((entry) => ({
    url: buildAbsoluteURL(`/${entityType}/${entry.slug}`),
    lastModified: toISODate(entry.lastChangedAt) ?? new Date(),
    changeFrequency: "weekly",
    priority: 0.8,
  }));
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const entries = createCoreRoutes();

  try {
    const patches = await collectAllPatchSummaries();
    entries.push(...createPatchEntries(patches));
  } catch (error) {
    console.error("Failed to collect patch URLs for sitemap", error);
  }

  try {
    const [heroesPayload, itemsPayload, spellsPayload] = await Promise.all([getHeroes(), getItems(), getSpells()]);
    entries.push(...createEntityEntries("heroes", heroesPayload.heroes));
    entries.push(...createEntityEntries("items", itemsPayload.items));
    entries.push(...createEntityEntries("spells", spellsPayload.spells));
  } catch (error) {
    console.error("Failed to collect entity URLs for sitemap", error);
  }

  return entries;
}
