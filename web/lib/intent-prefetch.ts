export function resolveIntentPrefetchHref(rawHref: string, currentURL: string): string | null {
  let targetURL: URL;
  let activeURL: URL;

  try {
    activeURL = new URL(currentURL);
    targetURL = new URL(rawHref, activeURL);
  } catch {
    return null;
  }

  if (
    targetURL.origin !== activeURL.origin ||
    (targetURL.protocol !== "http:" && targetURL.protocol !== "https:") ||
    (targetURL.pathname === activeURL.pathname && targetURL.search === activeURL.search)
  ) {
    return null;
  }

  return `${targetURL.pathname}${targetURL.search}`;
}

export function markIntentPrefetch(prefetchedHrefs: Set<string>, href: string) {
  if (prefetchedHrefs.has(href)) {
    return false;
  }

  prefetchedHrefs.add(href);
  return true;
}
