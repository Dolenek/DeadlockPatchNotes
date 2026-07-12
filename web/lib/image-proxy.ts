const ALLOWED_IMAGE_HOSTS = new Set([
  "assets-bucket.deadlock-api.com",
  "assets.deadlock-api.com",
  "clan.akamai.steamstatic.com",
  "clan.fastly.steamstatic.com",
  "shared.fastly.steamstatic.com",
]);

const REDIRECT_STATUSES = new Set([301, 302, 303, 307, 308]);
const MAX_REDIRECTS = 5;

export const IMAGE_PROXY_REQUEST_HEADERS = {
  "User-Agent": "deadlockpatchnotes-web-image-proxy/1.0",
};

export function parseAllowedImageURL(raw: string | null) {
  if (!raw) {
    return null;
  }

  try {
    const parsed = new URL(raw);
    if (parsed.protocol !== "https:" || !ALLOWED_IMAGE_HOSTS.has(parsed.hostname)) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

export async function fetchAllowedImage(initialTarget: URL, fetcher: typeof fetch = fetch) {
  let target = initialTarget;

  for (let redirectCount = 0; redirectCount <= MAX_REDIRECTS; redirectCount += 1) {
    const response = await fetcher(target.toString(), {
      headers: IMAGE_PROXY_REQUEST_HEADERS,
      redirect: "manual",
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
