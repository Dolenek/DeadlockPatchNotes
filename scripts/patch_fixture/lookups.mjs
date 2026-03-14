import { CARD_TYPE_NAMES as PARSER_CARD_TYPE_NAMES, canonicalHeroDisplayName, resolveHeroAlias } from "./parser_rules.mjs";
import { norm, slugify } from "./utils.mjs";

const ITEM_ALIAS = new Map([
  // Legacy/renamed item names that still appear in patch text.
  ["backstabber", "stalker"],
]);

export const CARD_TYPE_NAMES = PARSER_CARD_TYPE_NAMES;

export function heroLookupFromAssets(heroList) {
  const byNormalized = new Map();
  for (const hero of heroList) {
    byNormalized.set(norm(hero.name), hero);
    byNormalized.set(norm(hero.name.replace(/^The\s+/i, "")), hero);
  }

  return {
    resolve(prefix) {
      const key = resolveHeroAlias(prefix);
      return byNormalized.get(norm(key)) || byNormalized.get(norm(prefix)) || null;
    },
  };
}

export function collectItemsByName(items) {
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

export function createHeroState(heroName, heroAsset, iconInfo) {
  return {
    id: slugify(heroName),
    entityName: heroName,
    entityIconUrl: iconInfo?.publicPath,
    entityIconFallbackUrl: heroAsset?.images?.icon_image_small,
    changes: [],
    groups: [],
    groupMap: new Map(),
  };
}

export function ensureGroup(heroState, key, title, iconInfo, iconFallback) {
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

export function resolveHeroDisplayName(prefix, resolvedHero) {
  const displayName = canonicalHeroDisplayName(prefix);
  if (displayName && norm(displayName) === norm(resolvedHero?.name)) {
    return displayName;
  }
  return canonicalHeroDisplayName(resolvedHero?.name);
}
