const DEFAULT_HTTPS_SITE_URL = "https://www.deadlockpatchnotes.com";
const STATIC_ASSET_PATH = /\.(?:avif|gif|ico|jpe?g|png|svg|ttf|webp|woff2?)$/i;

export function shouldBypassHTTPSRedirect(pathname: string) {
  return pathname === "/healthz" || pathname.startsWith("/_next/") || STATIC_ASSET_PATH.test(pathname);
}

export function firstForwardedProtocol(rawValue: string | null) {
  return rawValue?.split(",", 1)[0].trim().toLowerCase() ?? "";
}

export function shouldRedirectToHTTPS(nodeEnvironment: string | undefined, forwardedProtocol: string | null) {
  return nodeEnvironment === "production" && firstForwardedProtocol(forwardedProtocol) === "http";
}

export function buildHTTPSRedirectURL(rawSiteURL: string | undefined, pathname: string, search: string) {
  let canonicalOrigin: URL;
  try {
    canonicalOrigin = new URL(rawSiteURL?.trim() || DEFAULT_HTTPS_SITE_URL);
  } catch {
    canonicalOrigin = new URL(DEFAULT_HTTPS_SITE_URL);
  }

  if (
    canonicalOrigin.protocol !== "https:" ||
    canonicalOrigin.username !== "" ||
    canonicalOrigin.password !== ""
  ) {
    canonicalOrigin = new URL(DEFAULT_HTTPS_SITE_URL);
  }

  canonicalOrigin.pathname = pathname.startsWith("/") ? pathname : `/${pathname}`;
  canonicalOrigin.search = search;
  canonicalOrigin.hash = "";
  return canonicalOrigin;
}
