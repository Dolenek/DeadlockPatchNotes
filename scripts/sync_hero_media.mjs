#!/usr/bin/env node
import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { createAssetRegistry, downloadAssets, fetchJson } from "./patch_fixture/assets.mjs";
import { slugify } from "./patch_fixture/utils.mjs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT = path.resolve(__dirname, "..");
const WEB_PUBLIC_DIR = path.join(ROOT, "web", "public");
const HERO_ASSET_DIR = path.join(WEB_PUBLIC_DIR, "assets", "heroes");
const HERO_MANIFEST_PATH = path.join(ROOT, "web", "lib", "hero-media-manifest.json");

const HERO_LIST_URL = "https://assets.deadlock-api.com/v2/heroes";
const HERO_BY_NAME_BASE_URL = "https://assets.deadlock-api.com/v2/heroes/by-name/";
const REQUEST_HEADERS = {
  "User-Agent": "deadlockpatchnotes-hero-media-sync/1.0",
};
const RETRYABLE_STATUS = new Set([429, 500, 502, 503, 504]);

const HERO_NAME_OVERRIDES = new Map([
  ["Doorman", "The Doorman"],
]);
const HERO_SLUG_ALIASES = new Map([
  ["The Doorman", ["doorman"]],
]);

function isInGameHero(hero) {
  return (
    hero?.player_selectable === true &&
    hero?.disabled !== true &&
    hero?.in_development !== true &&
    hero?.needs_testing !== true &&
    hero?.assigned_players_only !== true &&
    hero?.prerelease_only !== true &&
    hero?.limited_testing !== true
  );
}

function extractMediaURLs(heroPayload) {
  const images = heroPayload?.images || {};
  return {
    backgroundURL:
      images.background_image_webp ||
      images.background_image ||
      images.Background_Image_Webp ||
      images.Background_Image,
    nameImageURL: images.name_image || images.Name_Image,
  };
}

function mediaScore(heroPayload) {
  const media = extractMediaURLs(heroPayload);
  let score = 0;
  if (media.backgroundURL) {
    score += 1;
  }
  if (media.nameImageURL) {
    score += 1;
  }
  return score;
}

function collectManifestSlugs(heroName) {
  const primarySlug = slugify(heroName);
  const aliases = HERO_SLUG_ALIASES.get(heroName) ?? [];
  const all = [primarySlug, ...aliases].filter(Boolean);
  return [...new Set(all)];
}

function mediaExtFromURL(url, fallbackExt) {
  if (!url) {
    return fallbackExt;
  }

  try {
    const { pathname } = new URL(url);
    const ext = path.extname(pathname);
    return ext || fallbackExt;
  } catch {
    return fallbackExt;
  }
}

async function fetchHeroByName(name) {
  const target = `${HERO_BY_NAME_BASE_URL}${encodeURIComponent(name)}`;
  const maxAttempts = 3;

  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    try {
      const response = await fetch(target, { headers: REQUEST_HEADERS });
      if (response.status === 404) {
        return null;
      }
      if (response.ok) {
        return response.json();
      }
      if (!RETRYABLE_STATUS.has(response.status) || attempt >= maxAttempts) {
        throw new Error(`Request failed: ${target} (${response.status})`);
      }
    } catch (error) {
      if (attempt >= maxAttempts) {
        throw error;
      }
    }

    await new Promise((resolve) => setTimeout(resolve, 500 * attempt));
  }

  return null;
}

async function resolveHeroAtlas(heroName, seedPayload) {
  let bestPayload = seedPayload ?? null;
  let bestScore = mediaScore(bestPayload);
  const candidates = [heroName];
  const override = HERO_NAME_OVERRIDES.get(heroName);
  if (override) {
    candidates.push(override);
  }

  if (!heroName.toLowerCase().startsWith("the ")) {
    candidates.push(`The ${heroName}`);
  }

  const uniqueCandidates = [...new Set(candidates)];
  for (const candidate of uniqueCandidates) {
    const payload = await fetchHeroByName(candidate);
    if (payload) {
      const payloadScore = mediaScore(payload);
      if (payloadScore > bestScore) {
        bestPayload = payload;
        bestScore = payloadScore;
      }
      if (bestScore >= 2) {
        break;
      }
    }
  }

  return bestPayload;
}

function buildManifestAssetEntry(heroName, primarySlug, heroPayload, assetsRegistry) {
  const { backgroundURL, nameImageURL } = extractMediaURLs(heroPayload);
  const backgroundExt = mediaExtFromURL(backgroundURL, ".webp");
  const nameImageExt = mediaExtFromURL(nameImageURL, ".svg");
  const backgroundAsset = assetsRegistry.register(
    backgroundURL,
    `assets/heroes/${primarySlug}/background${backgroundExt}`,
  );
  const nameImageAsset = assetsRegistry.register(
    nameImageURL,
    `assets/heroes/${primarySlug}/name${nameImageExt}`,
  );

  return {
    name: heroName,
    sourceName: String(heroPayload?.name || heroName),
    backgroundImageUrl: backgroundAsset?.publicPath,
    nameImageUrl: nameImageAsset?.publicPath,
  };
}

async function buildHeroMediaManifest() {
  const roster = await fetchJson(HERO_LIST_URL);
  const allHeroes = Array.isArray(roster) ? roster : [];
  const inGameHeroes = allHeroes.filter(isInGameHero);
  const sortedHeroes = [...inGameHeroes].sort((a, b) =>
    String(a?.name || "").localeCompare(String(b?.name || ""), "en"),
  );

  const assetsRegistry = createAssetRegistry();
  const heroes = {};
  const missingPayload = [];
  const missingMedia = [];
  let completeMediaCount = 0;

  for (const hero of sortedHeroes) {
    const heroName = String(hero?.name || "").trim();
    if (!heroName) {
      continue;
    }

    const manifestSlugs = collectManifestSlugs(heroName);
    const primarySlug = manifestSlugs[0];
    if (!primarySlug) {
      process.stdout.write(`warn: could not build slug for hero "${heroName}"\n`);
      continue;
    }

    const heroPayload = await resolveHeroAtlas(heroName, hero);
    if (!heroPayload) {
      missingPayload.push(heroName);
      process.stdout.write(`warn: no assets payload found for hero "${heroName}"\n`);
      continue;
    }

    const entry = buildManifestAssetEntry(heroName, primarySlug, heroPayload, assetsRegistry);
    if (entry.backgroundImageUrl && entry.nameImageUrl) {
      completeMediaCount += 1;
    } else {
      missingMedia.push(heroName);
      process.stdout.write(`warn: missing media URLs for hero "${heroName}"\n`);
    }

    for (const slug of manifestSlugs) {
      if (heroes[slug]) {
        process.stdout.write(`warn: duplicate slug "${slug}" while processing "${heroName}"\n`);
      }
      heroes[slug] = entry;
    }
  }

  return {
    rosterCount: allHeroes.length,
    inGameRosterCount: sortedHeroes.length,
    manifestKeyCount: Object.keys(heroes).length,
    completeMediaCount,
    missingPayload,
    missingMedia,
    assetsRegistry,
    manifest: {
      generatedAt: new Date().toISOString(),
      heroes,
    },
  };
}

async function writeOutputs(manifest, assetsRegistry) {
  await fs.rm(HERO_ASSET_DIR, { recursive: true, force: true });
  await downloadAssets(assetsRegistry, WEB_PUBLIC_DIR);
  await fs.mkdir(path.dirname(HERO_MANIFEST_PATH), { recursive: true });
  await fs.writeFile(HERO_MANIFEST_PATH, `${JSON.stringify(manifest, null, 2)}\n`);
}

async function main() {
  const {
    rosterCount,
    inGameRosterCount,
    manifestKeyCount,
    completeMediaCount,
    missingPayload,
    missingMedia,
    assetsRegistry,
    manifest,
  } = await buildHeroMediaManifest();
  await writeOutputs(manifest, assetsRegistry);

  process.stdout.write(`heroes in full roster: ${rosterCount}\n`);
  process.stdout.write(`heroes marked in-game: ${inGameRosterCount}\n`);
  process.stdout.write(`manifest hero keys: ${manifestKeyCount}\n`);
  process.stdout.write(`heroes with complete media: ${completeMediaCount}\n`);
  process.stdout.write(`assets downloaded: ${assetsRegistry.entries().length}\n`);
  process.stdout.write(`wrote hero media manifest web/lib/hero-media-manifest.json\n`);
  if (missingPayload.length > 0) {
    process.stdout.write(`missing payload heroes: ${missingPayload.join(", ")}\n`);
  }
  if (missingMedia.length > 0) {
    process.stdout.write(`heroes missing media URLs: ${missingMedia.join(", ")}\n`);
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
