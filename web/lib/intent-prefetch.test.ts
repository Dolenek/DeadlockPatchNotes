import { describe, expect, it } from "vitest";
import { markIntentPrefetch, resolveIntentPrefetchHref } from "@/lib/intent-prefetch";

const CURRENT_URL = "https://www.deadlockpatchnotes.com/heroes?page=2";

describe("resolveIntentPrefetchHref", () => {
  it("returns an internal path with its query and without its fragment", () => {
    expect(resolveIntentPrefetchHref("/patches?page=3#update", CURRENT_URL)).toBe("/patches?page=3");
  });

  it("filters external and non-http links", () => {
    expect(resolveIntentPrefetchHref("https://example.com/patches", CURRENT_URL)).toBeNull();
    expect(resolveIntentPrefetchHref("mailto:test@example.com", CURRENT_URL)).toBeNull();
  });

  it("filters the current route and same-page fragments", () => {
    expect(resolveIntentPrefetchHref("/heroes?page=2", CURRENT_URL)).toBeNull();
    expect(resolveIntentPrefetchHref("#hero-card", CURRENT_URL)).toBeNull();
  });

  it("rejects malformed input", () => {
    expect(resolveIntentPrefetchHref("/patches", "not-a-url")).toBeNull();
  });

  it("deduplicates a URL until its prefetch is invalidated", () => {
    const prefetchedHrefs = new Set<string>();
    expect(markIntentPrefetch(prefetchedHrefs, "/patches")).toBe(true);
    expect(markIntentPrefetch(prefetchedHrefs, "/patches")).toBe(false);

    prefetchedHrefs.delete("/patches");
    expect(markIntentPrefetch(prefetchedHrefs, "/patches")).toBe(true);
  });
});
