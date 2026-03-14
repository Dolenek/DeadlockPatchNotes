import fs from "node:fs";

function normalizeRuleKey(value) {
  return String(value || "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, " ")
    .trim();
}

const rulesPath = new URL("../../api/internal/structuredparse/rules.json", import.meta.url);
const rawRules = JSON.parse(fs.readFileSync(rulesPath, "utf8"));

const heroAliases = new Map(
  Object.entries(rawRules.heroAliases || {}).map(([key, value]) => [normalizeRuleKey(key), String(value || "")]),
);
const heroCanonicalNames = new Map(
  Object.entries(rawRules.heroCanonicalNames || {}).map(([key, value]) => [normalizeRuleKey(key), String(value || "")]),
);
const heroAbilityAliases = new Map(
  Object.entries(rawRules.heroAbilityAliases || {}).map(([heroKey, aliases]) => [
    normalizeRuleKey(heroKey),
    new Map(
      Object.entries(aliases || {}).map(([canonical, values]) => [
        normalizeRuleKey(canonical),
        Array.isArray(values) ? values.map((value) => normalizeRuleKey(value)) : [],
      ]),
    ),
  ]),
);

export const CARD_TYPE_NAMES = new Set((rawRules.cardTypeNames || []).map((value) => normalizeRuleKey(value)));

export function canonicalHeroKey(name) {
  let key = normalizeRuleKey(name);
  const alias = heroAliases.get(key);
  if (alias) {
    key = normalizeRuleKey(alias);
  }
  const canonicalName = heroCanonicalNames.get(key);
  if (canonicalName) {
    return normalizeRuleKey(canonicalName);
  }
  return key;
}

export function canonicalHeroDisplayName(name) {
  const trimmed = String(name || "").trim();
  if (!trimmed) {
    return "";
  }
  let key = normalizeRuleKey(trimmed);
  const alias = heroAliases.get(key);
  if (alias) {
    key = normalizeRuleKey(alias);
  }
  return heroCanonicalNames.get(key) || trimmed;
}

export function resolveHeroAlias(name) {
  const key = normalizeRuleKey(name);
  return heroAliases.get(key) || key;
}

export function applyAbilityAlias(normalizedText, heroName = "") {
  const aliases = heroAbilityAliases.get(canonicalHeroKey(heroName));
  if (!aliases) {
    return normalizedText;
  }

  for (const [canonical, values] of aliases.entries()) {
    for (const alias of values) {
      if (normalizedText === alias) {
        return canonical;
      }
      if (normalizedText.startsWith(`${alias} `)) {
        return `${canonical}${normalizedText.slice(alias.length)}`;
      }
    }
  }

  return normalizedText;
}
