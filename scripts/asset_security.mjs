import path from "node:path";

export const MAX_ASSET_BYTES = 8 * 1024 * 1024;
export const MAX_JSON_BYTES = 10 * 1024 * 1024;
export const ALLOWED_DOWNLOAD_HOSTS = new Set([
  "api.steampowered.com",
  "assets-bucket.deadlock-api.com",
  "assets.deadlock-api.com",
  "clan.akamai.steamstatic.com",
  "clan.fastly.steamstatic.com",
  "shared.fastly.steamstatic.com",
]);

const REDIRECT_STATUSES = new Set([301, 302, 303, 307, 308]);
const RASTER_CONTENT_TYPES = new Set(["image/avif", "image/gif", "image/jpeg", "image/png", "image/webp"]);

export function parseAllowedDownloadURL(rawURL, allowedHosts = ALLOWED_DOWNLOAD_HOSTS) {
  try {
    const parsed = new URL(String(rawURL ?? "").trim());
    if (
      parsed.protocol !== "https:" ||
      parsed.port !== "" ||
      parsed.username !== "" ||
      parsed.password !== "" ||
      !allowedHosts.has(parsed.hostname)
    ) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

export function resolveContainedPath(rootDirectory, untrustedPath, label = "Path") {
  const root = path.resolve(rootDirectory);
  const normalizedInput = String(untrustedPath ?? "").replace(/^[/\\]+/, "");
  const candidate = path.resolve(root, normalizedInput);
  const relative = path.relative(root, candidate);
  if (
    normalizedInput === "" ||
    relative === "" ||
    relative === ".." ||
    relative.startsWith(`..${path.sep}`) ||
    path.isAbsolute(relative)
  ) {
    throw new Error(`${label} must stay inside ${root}: ${untrustedPath}`);
  }
  return candidate;
}

export async function fetchAllowedURL(
  rawURL,
  { fetcher = fetch, headers = {}, allowedHosts = ALLOWED_DOWNLOAD_HOSTS, timeoutMs = 20_000, maxRedirects = 3 } = {},
) {
  let target = parseAllowedDownloadURL(rawURL, allowedHosts);
  if (!target) {
    throw new Error(`Download URL is not allowed: ${rawURL}`);
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);
  timeout.unref?.();
  for (let redirectCount = 0; redirectCount <= maxRedirects; redirectCount += 1) {
    const response = await fetcher(target.toString(), {
      headers,
      redirect: "manual",
      signal: controller.signal,
    });
    if (!REDIRECT_STATUSES.has(response.status)) {
      return response;
    }
    if (redirectCount === maxRedirects) {
      throw new Error(`Too many redirects while downloading ${rawURL}`);
    }

    const location = response.headers.get("location");
    const redirected = location
      ? parseAllowedDownloadURL(new URL(location, target).toString(), allowedHosts)
      : null;
    await response.body?.cancel().catch(() => undefined);
    if (!redirected) {
      throw new Error(`Redirect target is not allowed while downloading ${rawURL}`);
    }
    target = redirected;
  }
  throw new Error(`Too many redirects while downloading ${rawURL}`);
}

export async function readLimitedResponse(response, maxBytes) {
  const declaredLength = Number.parseInt(response.headers.get("content-length") ?? "", 10);
  if (Number.isFinite(declaredLength) && declaredLength > maxBytes) {
    await response.body?.cancel().catch(() => undefined);
    throw new Error(`Response exceeds ${maxBytes} bytes`);
  }
  if (!response.body) {
    throw new Error("Response has no body");
  }

  const reader = response.body.getReader();
  const chunks = [];
  let totalBytes = 0;
  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    totalBytes += value.byteLength;
    if (totalBytes > maxBytes) {
      await reader.cancel().catch(() => undefined);
      throw new Error(`Response exceeds ${maxBytes} bytes`);
    }
    chunks.push(Buffer.from(value));
  }
  return Buffer.concat(chunks, totalBytes);
}

export function validateImageBytes(bytes, rawContentType) {
  const contentType = String(rawContentType ?? "").split(";", 1)[0].trim().toLowerCase();
  if (contentType === "image/svg+xml") {
    if (!isSafeSVG(bytes)) {
      throw new Error("Downloaded SVG contains active or external content");
    }
    return contentType;
  }
  if (!RASTER_CONTENT_TYPES.has(contentType) || !hasRasterSignature(bytes, contentType)) {
    throw new Error(`Downloaded asset does not match an allowed image type: ${contentType || "missing"}`);
  }
  return contentType;
}

function isSafeSVG(bytes) {
  const svg = bytes.toString("utf8");
  if (!/<svg(?:\s|>)/i.test(svg)) {
    return false;
  }
  return ![
    /<\s*(?:script|foreignObject|iframe|object|embed)\b/i,
    /<\s*!\s*(?:doctype|entity)\b/i,
    /<\?xml-stylesheet\b/i,
    /\son[a-z]+\s*=/i,
    /(?:href|xlink:href)\s*=\s*["']\s*(?:https?:|\/\/|data:|javascript:)/i,
    /url\(\s*["']?\s*(?:https?:|\/\/|data:|javascript:)/i,
    /@import\b/i,
  ].some((pattern) => pattern.test(svg));
}

function hasRasterSignature(bytes, contentType) {
  if (contentType === "image/png") {
    return bytes.subarray(0, 8).equals(Buffer.from("89504e470d0a1a0a", "hex"));
  }
  if (contentType === "image/jpeg") {
    return bytes.length >= 3 && bytes[0] === 0xff && bytes[1] === 0xd8 && bytes[2] === 0xff;
  }
  if (contentType === "image/gif") {
    return ["GIF87a", "GIF89a"].includes(bytes.subarray(0, 6).toString("ascii"));
  }
  if (contentType === "image/webp") {
    return bytes.subarray(0, 4).toString("ascii") === "RIFF" && bytes.subarray(8, 12).toString("ascii") === "WEBP";
  }
  return (
    contentType === "image/avif" &&
    bytes.subarray(4, 8).toString("ascii") === "ftyp" &&
    ["avif", "avis"].includes(bytes.subarray(8, 12).toString("ascii"))
  );
}
