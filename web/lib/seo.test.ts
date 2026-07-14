import { describe, expect, it } from "vitest";
import { resolveSiteURL } from "@/lib/seo";

describe("resolveSiteURL", () => {
  it("normalizes the configured site to an origin", () => {
    expect(resolveSiteURL("https://www.example.com/some/path?preview=1#top")).toBe("https://www.example.com");
  });

  it("uses the production default for missing and blank values", () => {
    expect(resolveSiteURL(undefined)).toBe("https://www.deadlockpatchnotes.com");
    expect(resolveSiteURL("  ")).toBe("https://www.deadlockpatchnotes.com");
  });

  it("rejects non-HTTP URLs and embedded credentials", () => {
    expect(() => resolveSiteURL("ftp://example.com/path")).toThrow("Invalid SITE_URL");
    expect(() => resolveSiteURL("https://user:secret@example.com")).toThrow("Invalid SITE_URL");
  });
});
