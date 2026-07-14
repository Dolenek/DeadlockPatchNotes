import { describe, expect, it } from "vitest";
import { resolveFallbackImageSource } from "@/lib/fallback-image";

describe("resolveFallbackImageSource", () => {
  it("switches a failed primary source to the fallback", () => {
    expect(resolveFallbackImageSource("/primary.png", "/fallback.png")).toBe("/fallback.png");
  });

  it("does not loop after the fallback source fails", () => {
    expect(resolveFallbackImageSource("/fallback.png", "/fallback.png")).toBe("/fallback.png");
  });

  it("preserves the primary source when no fallback is available", () => {
    expect(resolveFallbackImageSource("/primary.png", undefined)).toBe("/primary.png");
  });
});
