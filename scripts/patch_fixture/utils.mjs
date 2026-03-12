import crypto from "node:crypto";

const HERO_ABILITY_ALIAS = new Map([
  [
    "bebop",
    new Map([
      ["hook", "grapple arm"],
      ["hyperbeam", "hyper beam"],
      ["uppercut", "exploding uppercut"],
    ]),
  ],
]);

export function norm(value) {
  return String(value || "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, " ")
    .trim();
}

export function slugify(value) {
  return String(value || "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");
}

export function hashToken(value, length = 8) {
  return crypto.createHash("sha1").update(String(value || "")).digest("hex").slice(0, length);
}

export function escapeRegex(source) {
  return source.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

export function stripAbilityPrefix(text, abilityName) {
  const pattern = new RegExp(`^${escapeRegex(abilityName)}(?:\\s+|$)`, "i");
  const stripped = text.replace(pattern, "").trim();
  return stripped || text;
}

export function normalizeHeroLine(raw) {
  return raw.replace(/^Should Charge\b/i, "Shoulder Charge");
}

function heroNormKey(name) {
  const key = norm(name);
  return key.startsWith("the ") ? key.slice(4) : key;
}

function applyAbilityAlias(normalizedText, heroName) {
  const aliases = HERO_ABILITY_ALIAS.get(heroNormKey(heroName));
  if (!aliases) {
    return normalizedText;
  }

  for (const [alias, canonical] of aliases.entries()) {
    if (normalizedText === alias) {
      return canonical;
    }
    if (normalizedText.startsWith(`${alias} `)) {
      return `${canonical}${normalizedText.slice(alias.length)}`;
    }
  }

  return normalizedText;
}

export function abilityMatch(text, abilities, heroName = "") {
  const normalizedText = applyAbilityAlias(norm(text), heroName);
  for (const ability of abilities) {
    const abilityNorm = ability._norm;
    if (normalizedText === abilityNorm || normalizedText.startsWith(`${abilityNorm} `)) {
      return ability;
    }
  }
  return null;
}
