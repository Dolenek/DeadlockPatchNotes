#!/usr/bin/env node
import fs from "node:fs/promises";
import path from "node:path";
import crypto from "node:crypto";
import { fileURLToPath } from "node:url";

const PATCH_SLUG = "2026-03-06-update";
const PATCH_STEAM_TITLE = "Gameplay Update - 03-06-2026";
const PATCH_STEAM_GID = "1826362059925616";
const SOURCE_URL = "https://store.steampowered.com/news/app/1422450/view/519740319207522795";
const HERO_IMAGE_URL =
  "https://clan.akamai.steamstatic.com/images/45164767/1a200778c94a048c5b2580a1e1a36071679ff19e.png";

const HERO_ALIAS = new Map([
  ["doorman", "The Doorman"],
  ["vindcita", "Vindicta"],
]);

const ITEM_ALIAS = new Map([
  // Legacy/renamed item names that still appear in patch text.
  ["backstabber", "stalker"],
]);

const CARD_TYPE_NAMES = new Set(["spades", "diamond", "hearts", "clubs", "joker"]);

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT = path.resolve(__dirname, "..");

const WEB_PUBLIC_DIR = path.join(ROOT, "web", "public");
const ASSET_PREFIX = `/assets/patches/${PATCH_SLUG}`;
const FIXTURE_PATH = path.join(ROOT, "api", "internal", "patches", "data", `${PATCH_SLUG}.json`);
const MANIFEST_PATH = path.join(WEB_PUBLIC_DIR, ASSET_PREFIX, "manifest.json");

function norm(value) {
  return String(value || "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, " ")
    .trim();
}

function slugify(value) {
  return String(value || "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");
}

function hashToken(value, length = 8) {
  return crypto.createHash("sha1").update(String(value || "")).digest("hex").slice(0, length);
}

function escapeRegex(source) {
  return source.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

async function fetchJson(url) {
  const response = await fetch(url, {
    headers: {
      "User-Agent": "deadlockpatchnotes-fixture-generator/1.0",
    },
  });
  if (!response.ok) {
    throw new Error(`Request failed: ${url} (${response.status})`);
  }
  return response.json();
}

async function fetchBuffer(url) {
  const response = await fetch(url, {
    headers: {
      "User-Agent": "deadlockpatchnotes-fixture-generator/1.0",
    },
  });
  if (!response.ok) {
    throw new Error(`Asset download failed: ${url} (${response.status})`);
  }
  const bytes = await response.arrayBuffer();
  return Buffer.from(bytes);
}

function cleanSteamContent(contents) {
  return contents
    .replace(/\\r/g, "")
    .replace(/\[p\]\[\/p\]/g, "\n")
    .replace(/\[p\]/g, "")
    .replace(/\[\/p\]/g, "\n")
    .replace(/\[b\]|\[\/b\]|\[u\]|\[\/u\]/g, "")
    .replace(/\\\[/g, "[")
    .replace(/\\\]/g, "]");
}

function splitSections(lines) {
  const sections = [];
  let current = null;

  for (const line of lines) {
    const match = line.match(/^[\[\]]\s+(.+?)\s+\]$/);
    if (match) {
      current = { name: match[1], lines: [] };
      sections.push(current);
      continue;
    }

    if (!current) {
      continue;
    }
    current.lines.push(line);
  }

  return sections;
}

function parseBullet(line) {
  const match = line.match(/^-\s*([^:]+):\s*(.*)$/);
  if (!match) {
    return null;
  }
  return {
    prefix: match[1].trim(),
    text: match[2].trim(),
  };
}

function createAssetRegistry() {
  const byUrl = new Map();

  return {
    register(url, relPath) {
      if (!url) {
        return null;
      }
      const cleanRelPath = relPath.replace(/^\//, "");
      if (!byUrl.has(url)) {
        byUrl.set(url, {
          url,
          relPath: cleanRelPath,
          publicPath: `/${cleanRelPath}`,
        });
      }
      return byUrl.get(url);
    },
    entries() {
      return [...byUrl.values()];
    },
  };
}

function heroLookupFromAssets(heroList) {
  const byNormalized = new Map();
  for (const hero of heroList) {
    byNormalized.set(norm(hero.name), hero);

    const noThe = hero.name.replace(/^The\s+/i, "");
    byNormalized.set(norm(noThe), hero);
  }

  return {
    resolve(prefix) {
      const key = norm(prefix);
      const aliased = HERO_ALIAS.get(key);
      if (aliased) {
        return byNormalized.get(norm(aliased)) || null;
      }
      return byNormalized.get(key) || null;
    },
  };
}

function stripAbilityPrefix(text, abilityName) {
  const pattern = new RegExp(`^${escapeRegex(abilityName)}(?:\\s+|$)`, "i");
  const stripped = text.replace(pattern, "").trim();
  return stripped || text;
}

function normalizeHeroLine(raw) {
  return raw.replace(/^Should Charge\b/i, "Shoulder Charge");
}

function createHeroState(heroName, heroAsset, iconInfo) {
  return {
    id: slugify(heroName),
    entityName: heroName,
    entityIconUrl: iconInfo?.publicPath,
    entityIconFallbackUrl: heroAsset?.images?.icon_image_small,
    changes: [],
    groups: [],
    groupMap: new Map(),
    heroAsset,
  };
}

function ensureGroup(heroState, key, title, iconInfo, iconFallback) {
  if (!heroState.groupMap.has(key)) {
    const group = {
      id: `${heroState.id}-${slugify(title)}`,
      title,
      changes: [],
    };

    if (iconInfo?.publicPath) {
      group.iconUrl = iconInfo.publicPath;
    }
    if (iconFallback) {
      group.iconFallbackUrl = iconFallback;
    }

    heroState.groupMap.set(key, group);
  }
  return heroState.groupMap.get(key);
}

function abilityMatch(text, abilities) {
  const normalizedText = norm(text);
  for (const ability of abilities) {
    const abilityNorm = ability._norm;
    if (normalizedText === abilityNorm || normalizedText.startsWith(`${abilityNorm} `)) {
      return ability;
    }
  }
  return null;
}

function collectItemsByName(items) {
  const byName = new Map();
  for (const item of items) {
    byName.set(norm(item.name), item);
  }
  return {
    resolve(name, changeText) {
      const key = norm(name);
      const direct = byName.get(key);
      if (direct) {
        return direct;
      }

      const aliased = ITEM_ALIAS.get(key);
      if (aliased && byName.get(norm(aliased))) {
        return byName.get(norm(aliased));
      }

      // Handle explicit rename bullets, e.g. "Backstabber: Renamed to Stalker".
      const rename = String(changeText || "").match(/^Renamed to\s+(.+?)[.]*$/i);
      if (rename) {
        const renamedTarget = byName.get(norm(rename[1]));
        if (renamedTarget) {
          return renamedTarget;
        }
      }

      return null;
    },
  };
}

function extractNewItemName(text) {
  const cleaned = text.replace(/[.]+$/, "").trim();
  return cleaned || null;
}

async function downloadAssets(registry) {
  for (const asset of registry.entries()) {
    const outputPath = path.join(WEB_PUBLIC_DIR, asset.relPath);
    await fs.mkdir(path.dirname(outputPath), { recursive: true });
    try {
      const bytes = await fetchBuffer(asset.url);
      await fs.writeFile(outputPath, bytes);
      process.stdout.write(`downloaded ${asset.relPath}\n`);
    } catch (error) {
      process.stdout.write(`warn: ${asset.url} -> ${String(error.message)}\n`);
    }
  }
}

async function main() {
  const steamPayload = await fetchJson(
    "https://api.steampowered.com/ISteamNews/GetNewsForApp/v2/?appid=1422450&count=120&maxlength=0&format=json",
  );

  const steamItem =
    steamPayload.appnews.newsitems.find((item) => item.gid === PATCH_STEAM_GID) ||
    steamPayload.appnews.newsitems.find((item) => item.title === PATCH_STEAM_TITLE);

  if (!steamItem) {
    throw new Error(`Could not locate Steam patch item ${PATCH_STEAM_TITLE}`);
  }

  const cleaned = cleanSteamContent(steamItem.contents);
  const lines = cleaned
    .split(/\n+/)
    .map((line) => line.trim())
    .filter(Boolean);

  const sections = splitSections(lines);
  const sectionByName = new Map(sections.map((section) => [section.name.toLowerCase(), section]));

  const generalSection = sectionByName.get("general");
  const itemsSection = sectionByName.get("items");
  const heroesSection = sectionByName.get("heroes");

  if (!generalSection || !itemsSection || !heroesSection) {
    throw new Error("Expected General, Items, and Heroes sections in source patch");
  }

  const assetsRegistry = createAssetRegistry();

  const allHeroes = await fetchJson("https://assets.deadlock-api.com/v2/heroes");
  const allItems = await fetchJson("https://assets.deadlock-api.com/v2/items");
  const heroesLookup = heroLookupFromAssets(allHeroes);
  const itemsLookup = collectItemsByName(allItems);

  const heroById = new Map(allHeroes.map((hero) => [hero.id, hero]));

  const generalEntry = {
    id: "general-gameplay",
    entityName: "Core Gameplay",
    changes: generalSection.lines.map((line, index) => ({
      id: `general-${index + 1}`,
      text: line.replace(/^-\s*/, ""),
    })),
  };

  const itemEntries = [];
  const itemState = new Map();

  for (const line of itemsSection.lines) {
    const parsed = parseBullet(line);
    if (!parsed) {
      continue;
    }

    let itemName = parsed.prefix;
    let changeText = parsed.text;

    if (norm(parsed.prefix) === "added new t1 spirit item") {
      const extracted = extractNewItemName(parsed.text);
      if (extracted) {
        itemName = extracted;
        changeText = "Added as a new T1 Spirit Item.";
      }
    }

    const key = norm(itemName);

    if (!itemState.has(key)) {
      const assetItem = itemsLookup.resolve(itemName, changeText);
      const itemImage = assetItem?.shop_image || assetItem?.shop_image_webp || assetItem?.image;
      const ext = path.extname(new URL(itemImage || "https://dummy.local/icon.png").pathname) || ".png";
      const iconAsset = itemImage
        ? assetsRegistry.register(
            itemImage,
            `${ASSET_PREFIX}/items/${slugify(itemName)}-${hashToken(itemImage)}${ext}`,
          )
        : null;

      itemState.set(key, {
        id: slugify(itemName),
        entityName: itemName,
        entityIconUrl: iconAsset?.publicPath,
        entityIconFallbackUrl: itemImage,
        changes: [],
      });
      itemEntries.push(itemState.get(key));
    }

    const target = itemState.get(key);
    target.changes.push({
      id: `${target.id}-${target.changes.length + 1}`,
      text: changeText || `${itemName} updated.`,
    });
  }

  const heroEntries = [];
  const heroStateMap = new Map();
  const heroAbilitiesByName = new Map();

  const heroMentions = [];
  for (const line of heroesSection.lines) {
    const parsed = parseBullet(line);
    if (!parsed) {
      continue;
    }

    const heroAsset = heroesLookup.resolve(parsed.prefix);
    if (heroAsset) {
      const heroDisplayName = norm(parsed.prefix) === "doorman" ? "Doorman" : heroAsset.name;
      heroMentions.push({
        displayName: heroDisplayName,
        asset: heroAsset,
      });
    }
  }

  for (const mention of heroMentions) {
    if (heroAbilitiesByName.has(mention.displayName)) {
      continue;
    }

    const abilityItems = await fetchJson(
      `https://assets.deadlock-api.com/v2/items/by-hero-id/${encodeURIComponent(String(mention.asset.id))}`,
    );

    const abilities = abilityItems
      .filter((item) => item.type === "ability")
      .map((item) => ({
        ...item,
        _norm: norm(item.name),
      }))
      .sort((a, b) => b._norm.length - a._norm.length);

    heroAbilitiesByName.set(mention.displayName, abilities);
  }

  let currentHero = null;
  let currentSpecialGroup = null;

  for (const line of heroesSection.lines) {
    const parsed = parseBullet(line);
    if (!parsed) {
      continue;
    }

    const resolvedHero = heroesLookup.resolve(parsed.prefix);

    if (resolvedHero) {
      const displayName = norm(parsed.prefix) === "doorman" ? "Doorman" : resolvedHero.name;
      currentHero = displayName;
      currentSpecialGroup = null;

      if (!heroStateMap.has(displayName)) {
        const imageUrl = resolvedHero.images?.icon_image_small;
        const ext = path.extname(new URL(imageUrl || "https://dummy.local/icon.png").pathname) || ".png";
        const iconAsset = imageUrl
          ? assetsRegistry.register(imageUrl, `${ASSET_PREFIX}/heroes/${slugify(displayName)}${ext}`)
          : null;
        const heroState = createHeroState(displayName, resolvedHero, iconAsset);
        heroStateMap.set(displayName, heroState);
        heroEntries.push(heroState);
      }

      const heroState = heroStateMap.get(displayName);
      const normalizedText = normalizeHeroLine(parsed.text);
      const abilities = heroAbilitiesByName.get(displayName) || [];

      if (/^Talents\s+/i.test(normalizedText)) {
        const group = ensureGroup(heroState, "talents", "Talents", null, null);
        group.changes.push({
          id: `${group.id}-${group.changes.length + 1}`,
          text: normalizedText.replace(/^Talents\s+/i, ""),
        });
        continue;
      }

      const matchedAbility = abilityMatch(normalizedText, abilities);
      if (matchedAbility) {
        const abilityIcon = matchedAbility.image
          ? assetsRegistry.register(
              matchedAbility.image,
              `${ASSET_PREFIX}/abilities/${slugify(displayName)}-${slugify(matchedAbility.name)}${
                path.extname(new URL(matchedAbility.image).pathname) || ".png"
              }`,
            )
          : null;
        const group = ensureGroup(
          heroState,
          `ability-${slugify(matchedAbility.name)}`,
          matchedAbility.name,
          abilityIcon,
          matchedAbility.image,
        );

        group.changes.push({
          id: `${group.id}-${group.changes.length + 1}`,
          text: stripAbilityPrefix(normalizedText, matchedAbility.name),
        });
        continue;
      }

      heroState.changes.push({
        id: `${heroState.id}-general-${heroState.changes.length + 1}`,
        text: normalizedText,
      });

      continue;
    }

    if (!currentHero) {
      continue;
    }

    const heroState = heroStateMap.get(currentHero);
    const prefixKey = norm(parsed.prefix);

    if (prefixKey === "card types") {
      currentSpecialGroup = "Card Types";
      ensureGroup(heroState, "card-types", "Card Types", null, null);
      continue;
    }

    if (currentSpecialGroup === "Card Types" && CARD_TYPE_NAMES.has(prefixKey)) {
      const group = ensureGroup(heroState, "card-types", "Card Types", null, null);
      group.changes.push({
        id: `${group.id}-${group.changes.length + 1}`,
        text: `${parsed.prefix}: ${parsed.text}`,
      });
      continue;
    }

    const text = parsed.text ? `${parsed.prefix}: ${parsed.text}` : parsed.prefix;
    heroState.changes.push({
      id: `${heroState.id}-general-${heroState.changes.length + 1}`,
      text,
    });
  }

  for (const heroState of heroEntries) {
    heroState.groups = [...heroState.groupMap.values()].filter((group) => group.changes.length > 0);
    delete heroState.groupMap;
    delete heroState.heroAsset;
  }

  const detail = {
    id: steamItem.gid,
    slug: PATCH_SLUG,
    title: steamItem.title,
    publishedAt: new Date(steamItem.date * 1000).toISOString(),
    category: "Regular Update",
    source: {
      type: "steam-news",
      url: SOURCE_URL,
    },
    heroImageUrl: HERO_IMAGE_URL,
    intro:
      "Comprehensive balance and systems update covering map flow, item economy, and broad hero tuning across nearly the entire roster.",
    sections: [
      {
        id: "general",
        title: "General",
        kind: "general",
        entries: [generalEntry],
      },
      {
        id: "items",
        title: "Items",
        kind: "items",
        entries: itemEntries,
      },
      {
        id: "heroes",
        title: "Heroes",
        kind: "heroes",
        entries: heroEntries,
      },
    ],
  };

  const patchAssetsDir = path.join(WEB_PUBLIC_DIR, ASSET_PREFIX.replace(/^\//, ""));
  await fs.rm(patchAssetsDir, { recursive: true, force: true });

  await downloadAssets(assetsRegistry);

  await fs.mkdir(path.dirname(FIXTURE_PATH), { recursive: true });
  await fs.writeFile(FIXTURE_PATH, `${JSON.stringify(detail, null, 2)}\n`);

  const manifest = {
    generatedAt: new Date().toISOString(),
    patchSlug: PATCH_SLUG,
    assetCount: assetsRegistry.entries().length,
    assets: assetsRegistry.entries().map((entry) => ({
      url: entry.url,
      localPath: entry.publicPath,
    })),
  };

  await fs.mkdir(path.dirname(MANIFEST_PATH), { recursive: true });
  await fs.writeFile(MANIFEST_PATH, `${JSON.stringify(manifest, null, 2)}\n`);

  process.stdout.write(`wrote fixture ${path.relative(ROOT, FIXTURE_PATH)}\n`);
  process.stdout.write(`wrote manifest ${path.relative(ROOT, MANIFEST_PATH)}\n`);
  process.stdout.write(`general lines: ${generalEntry.changes.length}\n`);
  process.stdout.write(`item changes: ${itemEntries.reduce((sum, entry) => sum + entry.changes.length, 0)}\n`);
  process.stdout.write(
    `hero changes: ${heroEntries.reduce((sum, entry) => {
      const groupCount = entry.groups.reduce((inner, group) => inner + group.changes.length, 0);
      return sum + entry.changes.length + groupCount;
    }, 0)}\n`,
  );
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
