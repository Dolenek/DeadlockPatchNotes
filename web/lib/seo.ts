type CanonicalQueryValue = string | number | boolean;

export type CanonicalQuery = Record<string, CanonicalQueryValue | undefined>;

const DEFAULT_SITE_URL = "https://www.deadlockpatchnotes.com";

export const SEO_SITE_NAME = "Deadlock Patch Notes";
export const SEO_DEFAULT_DESCRIPTION =
  "Deadlock patch notes archive with timeline-based updates, hero changes, spell updates, and item balance history.";
export const SEO_DEFAULT_IMAGE_PATH = "/Oldgods_header.png";

function resolveSiteURL() {
  const candidate = (process.env.SITE_URL ?? DEFAULT_SITE_URL).trim();
  const fallback = candidate === "" ? DEFAULT_SITE_URL : candidate;

  let parsed: URL;
  try {
    parsed = new URL(fallback);
  } catch {
    throw new Error(`Invalid SITE_URL: ${fallback}`);
  }

  parsed.pathname = "";
  parsed.search = "";
  parsed.hash = "";
  return parsed.toString().replace(/\/$/, "");
}

export const SEO_BASE_URL = resolveSiteURL();
export const SEO_METADATA_BASE_URL = new URL(`${SEO_BASE_URL}/`);

function normalizePathname(pathname: string) {
  const trimmed = pathname.trim();
  if (trimmed === "" || trimmed === "/") {
    return "/";
  }
  return trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
}

export function buildCanonicalPath(pathname: string, query?: CanonicalQuery) {
  const normalizedPathname = normalizePathname(pathname);
  const params = new URLSearchParams();

  for (const [key, value] of Object.entries(query ?? {})) {
    if (value === undefined) {
      continue;
    }
    params.set(key, String(value));
  }

  const search = params.toString();
  return search === "" ? normalizedPathname : `${normalizedPathname}?${search}`;
}

export function buildAbsoluteURL(pathname: string, query?: CanonicalQuery) {
  const canonicalPath = buildCanonicalPath(pathname, query);
  return new URL(canonicalPath, SEO_METADATA_BASE_URL).toString();
}

export function resolveSocialImageURL(candidate?: string) {
  const trimmed = String(candidate ?? "").trim();
  if (trimmed === "") {
    return buildAbsoluteURL(SEO_DEFAULT_IMAGE_PATH);
  }

  if (trimmed.startsWith("/")) {
    return buildAbsoluteURL(trimmed);
  }

  try {
    const parsed = new URL(trimmed);
    if (parsed.protocol === "http:" || parsed.protocol === "https:") {
      return parsed.toString();
    }
  } catch {
    // fall through to default image
  }

  return buildAbsoluteURL(SEO_DEFAULT_IMAGE_PATH);
}

export function truncateDescription(raw: string, maxLength = 160) {
  const normalized = raw.replace(/\s+/g, " ").trim();
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return `${normalized.slice(0, Math.max(0, maxLength - 1)).trimEnd()}…`;
}

export function toISODate(raw?: string) {
  const value = String(raw ?? "").trim();
  if (value === "") {
    return undefined;
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.valueOf())) {
    return undefined;
  }
  return parsed.toISOString();
}
