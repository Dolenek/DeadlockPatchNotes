import crypto from "node:crypto";

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

export function abilityMatch(text, abilities) {
  const normalizedText = norm(text);
  for (const ability of abilities) {
    const abilityNorm = ability._norm;
    if (normalizedText === abilityNorm || normalizedText.startsWith(`${abilityNorm} `)) {
      return ability;
    }
  }
  return null;
}
