import path from "node:path";
import {
  ASSET_PREFIX,
  HERO_IMAGE_URL,
  PATCH_SLUG,
  PATCH_STEAM_GID,
  PATCH_STEAM_TITLE,
  SOURCE_URL,
} from "./config.mjs";
import { CARD_TYPE_NAMES, collectItemsByName, createHeroState, ensureGroup, resolveHeroDisplayName } from "./lookups.mjs";
import { extractNewItemName, getCoreSections, parseBullet } from "./text_parse.mjs";
import { abilityMatch, hashToken, norm, normalizeHeroLine, slugify, stripAbilityPrefix } from "./utils.mjs";

export function findSteamPatchItem(steamPayload) {
  return (
    steamPayload.appnews.newsitems.find((item) => item.gid === PATCH_STEAM_GID) ||
    steamPayload.appnews.newsitems.find((item) => item.title === PATCH_STEAM_TITLE)
  );
}

function buildGeneralEntry(generalSection) {
  return {
    id: "general-gameplay",
    entityName: "Core Gameplay",
    changes: generalSection.lines.map((line, index) => ({
      id: `general-${index + 1}`,
      text: line.replace(/^-\s*/, ""),
    })),
  };
}

function registerItemIcon(assetsRegistry, itemName, itemImage) {
  if (!itemImage) {
    return null;
  }

  const ext = path.extname(new URL(itemImage).pathname) || ".png";
  return assetsRegistry.register(itemImage, `${ASSET_PREFIX}/items/${slugify(itemName)}-${hashToken(itemImage)}${ext}`);
}

function buildItemEntries(itemsSection, itemsLookup, assetsRegistry) {
  const itemEntries = [];
  const itemState = new Map();

  for (const line of itemsSection.lines) {
    const parsed = parseBullet(line);
    if (!parsed) {
      continue;
    }

    const parsedItem = normalizeParsedItem(parsed);
    const key = norm(parsedItem.itemName);

    if (!itemState.has(key)) {
      const entry = createItemEntry(parsedItem.itemName, parsedItem.changeText, itemsLookup, assetsRegistry);
      itemState.set(key, entry);
      itemEntries.push(entry);
    }

    const target = itemState.get(key);
    target.changes.push({
      id: `${target.id}-${target.changes.length + 1}`,
      text: parsedItem.changeText || `${parsedItem.itemName} updated.`,
    });
  }

  return itemEntries;
}

function normalizeParsedItem(parsed) {
  let itemName = parsed.prefix;
  let changeText = parsed.text;

  if (norm(parsed.prefix) === "added new t1 spirit item") {
    const extracted = extractNewItemName(parsed.text);
    if (extracted) {
      itemName = extracted;
      changeText = "Added as a new T1 Spirit Item.";
    }
  }

  return { itemName, changeText };
}

function createItemEntry(itemName, changeText, itemsLookup, assetsRegistry) {
  const assetItem = itemsLookup.resolve(itemName, changeText);
  const itemImage = assetItem?.shop_image || assetItem?.shop_image_webp || assetItem?.image;
  const iconAsset = registerItemIcon(assetsRegistry, itemName, itemImage);

  return {
    id: slugify(itemName),
    entityName: itemName,
    entityIconUrl: iconAsset?.publicPath,
    entityIconFallbackUrl: itemImage,
    changes: [],
  };
}

function collectHeroMentions(heroesSection, heroesLookup) {
  const mentions = [];

  for (const line of heroesSection.lines) {
    const parsed = parseBullet(line);
    const heroName = parsed ? parsed.prefix : extractHeroHeading(line);
    if (!heroName) {
      continue;
    }

    const heroAsset = heroesLookup.resolve(heroName);
    if (!heroAsset) {
      continue;
    }

    mentions.push({
      displayName: resolveHeroDisplayName(heroName, heroAsset),
      asset: heroAsset,
    });
  }

  return mentions;
}

async function loadHeroAbilitiesByName(heroMentions, fetchJson) {
  const heroAbilitiesByName = new Map();

  for (const mention of heroMentions) {
    if (heroAbilitiesByName.has(mention.displayName)) {
      continue;
    }

    const abilityItems = await fetchJson(
      `https://assets.deadlock-api.com/v2/items/by-hero-id/${encodeURIComponent(String(mention.asset.id))}`,
    );

    const abilities = abilityItems
      .filter((item) => item.type === "ability")
      .map((item) => ({ ...item, _norm: norm(item.name) }))
      .sort((a, b) => b._norm.length - a._norm.length);

    heroAbilitiesByName.set(mention.displayName, abilities);
  }

  return heroAbilitiesByName;
}

function registerHeroIcon(assetsRegistry, displayName, imageUrl) {
  if (!imageUrl) {
    return null;
  }
  const ext = path.extname(new URL(imageUrl).pathname) || ".png";
  return assetsRegistry.register(imageUrl, `${ASSET_PREFIX}/heroes/${slugify(displayName)}${ext}`);
}

function registerAbilityIcon(assetsRegistry, displayName, ability) {
  if (!ability.image) {
    return null;
  }
  const ext = path.extname(new URL(ability.image).pathname) || ".png";
  return assetsRegistry.register(
    ability.image,
    `${ASSET_PREFIX}/abilities/${slugify(displayName)}-${slugify(ability.name)}${ext}`,
  );
}

function finalizeHeroEntries(heroEntries) {
  for (const heroState of heroEntries) {
    heroState.groups = [...heroState.groupMap.values()].filter((group) => group.changes.length > 0);
    delete heroState.groupMap;
  }
}

async function buildHeroEntries(heroesSection, heroesLookup, assetsRegistry, fetchJson) {
  const heroEntries = [];
  const heroStateMap = new Map();
  const heroMentions = collectHeroMentions(heroesSection, heroesLookup);
  const heroAbilitiesByName = await loadHeroAbilitiesByName(heroMentions, fetchJson);

  let currentHero = null;
  let currentSpecialGroup = null;

  for (const line of heroesSection.lines) {
    const parsed = parseBullet(line);
    if (!parsed) {
      const headingState = getOrCreateHeroStateFromLine(line, heroesLookup, heroStateMap, heroEntries, assetsRegistry);
      if (headingState) {
        currentHero = headingState.displayName;
        currentSpecialGroup = null;
        continue;
      }

      if (!currentHero) {
        continue;
      }

      currentSpecialGroup = applyHeroPlainFollowupChange(heroStateMap.get(currentHero), line, currentSpecialGroup);
      continue;
    }

    const state = getOrCreateHeroState(parsed, heroesLookup, heroStateMap, heroEntries, assetsRegistry);
    if (state) {
      currentHero = state.displayName;
      currentSpecialGroup = null;
      applyHeroBulletChange(state.heroState, parsed, heroAbilitiesByName.get(state.displayName) || [], assetsRegistry);
      continue;
    }

    if (!currentHero) {
      continue;
    }

    currentSpecialGroup = applyHeroFollowupChange(
      heroStateMap.get(currentHero),
      parsed,
      currentSpecialGroup,
      heroAbilitiesByName.get(currentHero) || [],
      assetsRegistry,
    );
  }

  finalizeHeroEntries(heroEntries);
  return heroEntries;
}

function getOrCreateHeroState(parsed, heroesLookup, heroStateMap, heroEntries, assetsRegistry) {
  return getOrCreateHeroStateByName(parsed.prefix, heroesLookup, heroStateMap, heroEntries, assetsRegistry);
}

function getOrCreateHeroStateFromLine(line, heroesLookup, heroStateMap, heroEntries, assetsRegistry) {
  const heroName = extractHeroHeading(line);
  if (!heroName) {
    return null;
  }
  return getOrCreateHeroStateByName(heroName, heroesLookup, heroStateMap, heroEntries, assetsRegistry);
}

function getOrCreateHeroStateByName(heroName, heroesLookup, heroStateMap, heroEntries, assetsRegistry) {
  const resolvedHero = heroesLookup.resolve(heroName);
  if (!resolvedHero) {
    return null;
  }

  const displayName = resolveHeroDisplayName(heroName, resolvedHero);
  if (!heroStateMap.has(displayName)) {
    const imageUrl = resolvedHero.images?.icon_image_small;
    const iconAsset = registerHeroIcon(assetsRegistry, displayName, imageUrl);
    const heroState = createHeroState(displayName, resolvedHero, iconAsset);
    heroStateMap.set(displayName, heroState);
    heroEntries.push(heroState);
  }

  return {
    displayName,
    heroState: heroStateMap.get(displayName),
  };
}

function applyHeroBulletChange(heroState, parsed, abilities, assetsRegistry) {
  const normalizedText = normalizeHeroLine(parsed.text);

  if (/^Talents\s+/i.test(normalizedText)) {
    const group = ensureGroup(heroState, "talents", "Talents", null, null);
    group.changes.push({
      id: `${group.id}-${group.changes.length + 1}`,
      text: normalizedText.replace(/^Talents\s+/i, ""),
    });
    return;
  }

  const matchedAbility = abilityMatch(normalizedText, abilities);
  if (matchedAbility) {
    const abilityIcon = registerAbilityIcon(assetsRegistry, heroState.entityName, matchedAbility);
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
    return;
  }

  heroState.changes.push({
    id: `${heroState.id}-general-${heroState.changes.length + 1}`,
    text: normalizedText,
  });
}

function applyHeroFollowupChange(heroState, parsed, currentSpecialGroup, abilities, assetsRegistry) {
  const prefixKey = norm(parsed.prefix);

  if (prefixKey === "card types") {
    ensureGroup(heroState, "card-types", "Card Types", null, null);
    return "Card Types";
  }

  if (currentSpecialGroup === "Card Types" && CARD_TYPE_NAMES.has(prefixKey)) {
    const group = ensureGroup(heroState, "card-types", "Card Types", null, null);
    group.changes.push({
      id: `${group.id}-${group.changes.length + 1}`,
      text: `${parsed.prefix}: ${parsed.text}`,
    });
    return currentSpecialGroup;
  }

  if (parsed.text) {
    const prefixedAbility = abilityMatch(parsed.prefix, abilities);
    if (prefixedAbility) {
      const abilityIcon = registerAbilityIcon(assetsRegistry, heroState.entityName, prefixedAbility);
      const group = ensureGroup(
        heroState,
        `ability-${slugify(prefixedAbility.name)}`,
        prefixedAbility.name,
        abilityIcon,
        prefixedAbility.image,
      );
      group.changes.push({
        id: `${group.id}-${group.changes.length + 1}`,
        text: parsed.text,
      });
      return currentSpecialGroup;
    }
  }

  const text = parsed.text ? `${parsed.prefix}: ${parsed.text}` : parsed.prefix;
  heroState.changes.push({
    id: `${heroState.id}-general-${heroState.changes.length + 1}`,
    text,
  });

  return currentSpecialGroup;
}

function applyHeroPlainFollowupChange(heroState, line, currentSpecialGroup) {
  const value = extractHeroHeading(line);
  if (!value) {
    return currentSpecialGroup;
  }

  if (norm(value) === "card types") {
    ensureGroup(heroState, "card-types", "Card Types", null, null);
    return "Card Types";
  }

  heroState.changes.push({
    id: `${heroState.id}-general-${heroState.changes.length + 1}`,
    text: value,
  });
  return currentSpecialGroup;
}

function extractHeroHeading(line) {
  const cleaned = String(line || "").replace(/^-\s*/, "").trim();
  return cleaned || null;
}

function countHeroChanges(heroEntries) {
  return heroEntries.reduce((sum, entry) => {
    const groupCount = entry.groups.reduce((inner, group) => inner + group.changes.length, 0);
    return sum + entry.changes.length + groupCount;
  }, 0);
}

function buildDetailObject(steamItem, generalEntry, itemEntries, heroEntries) {
  return {
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
}

function buildStats(generalEntry, itemEntries, heroEntries) {
  return {
    generalLines: generalEntry.changes.length,
    itemChanges: itemEntries.reduce((sum, entry) => sum + entry.changes.length, 0),
    heroChanges: countHeroChanges(heroEntries),
  };
}

export async function buildPatchDetail({ steamItem, allItems, assetsRegistry, fetchJson, heroesLookup }) {
  const { generalSection, itemsSection, heroesSection } = getCoreSections(steamItem);
  const itemsLookup = collectItemsByName(allItems);

  const generalEntry = buildGeneralEntry(generalSection);
  const itemEntries = buildItemEntries(itemsSection, itemsLookup, assetsRegistry);
  const heroEntries = await buildHeroEntries(heroesSection, heroesLookup, assetsRegistry, fetchJson);

  return {
    detail: buildDetailObject(steamItem, generalEntry, itemEntries, heroEntries),
    stats: buildStats(generalEntry, itemEntries, heroEntries),
  };
}
