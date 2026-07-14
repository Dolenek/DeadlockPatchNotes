const ALLOWED_IMAGE_HOSTS = new Set([
  "assets-bucket.deadlock-api.com",
  "assets.deadlock-api.com",
  "clan.akamai.steamstatic.com",
  "clan.fastly.steamstatic.com",
  "shared.fastly.steamstatic.com",
]);

const REDIRECT_STATUSES = new Set([301, 302, 303, 307, 308]);
const MAX_REDIRECTS = 5;
export const MAX_IMAGE_BYTES = 8 * 1024 * 1024;

const ALLOWED_IMAGE_TYPES = new Set([
  "image/avif",
  "image/gif",
  "image/jpeg",
  "image/png",
  "image/webp",
]);

export class ImageProxyValidationError extends Error {
  constructor(message: string, readonly status: 413 | 415) {
    super(message);
    this.name = "ImageProxyValidationError";
  }
}

export const IMAGE_PROXY_REQUEST_HEADERS = {
  "User-Agent": "deadlockpatchnotes-web-image-proxy/1.0",
};

export function parseAllowedImageURL(raw: string | null) {
  if (!raw) {
    return null;
  }

  try {
    const parsed = new URL(raw);
    if (
      parsed.protocol !== "https:" ||
      parsed.port !== "" ||
      parsed.username !== "" ||
      parsed.password !== "" ||
      !ALLOWED_IMAGE_HOSTS.has(parsed.hostname)
    ) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

export async function fetchAllowedImage(initialTarget: URL, fetcher: typeof fetch = fetch, signal?: AbortSignal) {
  let target = initialTarget;

  for (let redirectCount = 0; redirectCount <= MAX_REDIRECTS; redirectCount += 1) {
    const response = await fetcher(target.toString(), {
      headers: IMAGE_PROXY_REQUEST_HEADERS,
      redirect: "manual",
      signal,
    });

    if (!REDIRECT_STATUSES.has(response.status)) {
      return response;
    }

    if (redirectCount === MAX_REDIRECTS) {
      throw new Error("too many image redirects");
    }

    const location = response.headers.get("location");
    const redirectedTarget = location ? parseAllowedImageURL(new URL(location, target).toString()) : null;
    if (!redirectedTarget) {
      throw new Error("image redirect target is not allowed");
    }

    await response.body?.cancel().catch(() => undefined);
    target = redirectedTarget;
  }

  throw new Error("too many image redirects");
}

async function validateImageResponse(response: Response, maxBytes: number) {
  const contentType = response.headers.get("content-type")?.split(";", 1)[0].trim().toLowerCase() ?? "";
  if (!ALLOWED_IMAGE_TYPES.has(contentType)) {
    await response.body?.cancel().catch(() => undefined);
    throw new ImageProxyValidationError("upstream response is not an allowed raster image", 415);
  }

  const declaredLength = Number.parseInt(response.headers.get("content-length") ?? "", 10);
  if (Number.isFinite(declaredLength) && declaredLength > maxBytes) {
    await response.body?.cancel().catch(() => undefined);
    throw new ImageProxyValidationError("upstream image is too large", 413);
  }
  if (!response.body) {
    throw new Error("upstream image has no body");
  }

  return { contentType, reader: response.body.getReader() };
}

export async function readAllowedImageResponse(response: Response, maxBytes = MAX_IMAGE_BYTES) {
  const { contentType, reader } = await validateImageResponse(response, maxBytes);
  const chunks: Uint8Array[] = [];
  let totalBytes = 0;
  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    totalBytes += value.byteLength;
    if (totalBytes > maxBytes) {
      await reader.cancel().catch(() => undefined);
      throw new ImageProxyValidationError("upstream image is too large", 413);
    }
    chunks.push(value);
  }

  const bytes = new Uint8Array(totalBytes);
  let offset = 0;
  for (const chunk of chunks) {
    bytes.set(chunk, offset);
    offset += chunk.byteLength;
  }
  if (!hasExpectedImageSignature(bytes, contentType)) {
    throw new ImageProxyValidationError("upstream image signature does not match its content type", 415);
  }
  return { bytes, contentType };
}

function hasExpectedImageSignature(bytes: Uint8Array, contentType: string) {
  if (contentType === "image/png") {
    return startsWith(bytes, [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
  }
  if (contentType === "image/jpeg") {
    return startsWith(bytes, [0xff, 0xd8, 0xff]);
  }
  if (contentType === "image/gif") {
    return startsWithText(bytes, "GIF87a") || startsWithText(bytes, "GIF89a");
  }
  if (contentType === "image/webp") {
    return startsWithText(bytes, "RIFF") && textAt(bytes, 8, 4) === "WEBP";
  }
  return contentType === "image/avif" && textAt(bytes, 4, 4) === "ftyp" && ["avif", "avis"].includes(textAt(bytes, 8, 4));
}

function startsWith(bytes: Uint8Array, signature: number[]) {
  return bytes.length >= signature.length && signature.every((value, index) => bytes[index] === value);
}

function startsWithText(bytes: Uint8Array, signature: string) {
  return textAt(bytes, 0, signature.length) === signature;
}

function textAt(bytes: Uint8Array, offset: number, length: number) {
  if (bytes.length < offset + length) {
    return "";
  }
  return String.fromCharCode(...bytes.subarray(offset, offset + length));
}
