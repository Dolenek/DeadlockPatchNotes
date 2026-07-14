import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import { WEB_SECURITY_HEADERS } from "@/lib/security-headers";

describe("WEB_SECURITY_HEADERS", () => {
  it("sets browser isolation, transport, and content restrictions", () => {
    const headers = new Map(WEB_SECURITY_HEADERS.map(({ key, value }) => [key, value]));
    expect(headers.get("Strict-Transport-Security")).toContain("max-age=31536000");
    expect(headers.get("X-Content-Type-Options")).toBe("nosniff");
    expect(headers.get("X-Frame-Options")).toBe("DENY");
    expect(headers.get("Content-Security-Policy")).toContain("frame-ancestors 'none'");
    expect(headers.get("Content-Security-Policy")).toContain("object-src 'none'");
  });

  it("keeps runtime font loading same-origin", () => {
    const globalsCSS = readFileSync(new URL("../app/globals.css", import.meta.url), "utf8");
    const layoutSource = readFileSync(new URL("../app/layout.tsx", import.meta.url), "utf8");
    const headers = new Map(WEB_SECURITY_HEADERS.map(({ key, value }) => [key, value]));

    expect(globalsCSS).not.toContain("fonts.googleapis.com");
    expect(layoutSource).toContain('from "next/font/google"');
    expect(headers.get("Content-Security-Policy")).toContain("font-src 'self' data:");
    expect(headers.get("Content-Security-Policy")).not.toContain("fonts.gstatic.com");
  });
});
