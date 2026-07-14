import { describe, expect, it } from "vitest";
import { resolveAPIBaseURL } from "@/lib/api-base-url";

describe("resolveAPIBaseURL", () => {
  it("normalizes host-only and /api-suffixed values", () => {
    expect(resolveAPIBaseURL("https://deadlockpatchnotes.com")).toBe("https://deadlockpatchnotes.com");
    expect(resolveAPIBaseURL("https://deadlockpatchnotes.com/api")).toBe("https://deadlockpatchnotes.com");
  });

  it("normalizes the default for missing and blank values", () => {
    expect(resolveAPIBaseURL(undefined)).toBe("https://deadlockpatchnotes.com");
    expect(resolveAPIBaseURL("   ")).toBe("https://deadlockpatchnotes.com");
  });

  it("preserves intentional non-api base paths", () => {
    expect(resolveAPIBaseURL("https://example.com/backend/")).toBe("https://example.com/backend");
  });

  it("rejects invalid non-empty values", () => {
    expect(() => resolveAPIBaseURL("not a URL")).toThrow("Invalid API_BASE_URL");
    expect(() => resolveAPIBaseURL("ftp://example.com")).toThrow("Invalid API_BASE_URL");
    expect(() => resolveAPIBaseURL("https://user:secret@example.com")).toThrow("Invalid API_BASE_URL");
  });
});
