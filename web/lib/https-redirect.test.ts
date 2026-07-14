import { describe, expect, it } from "vitest";
import { buildHTTPSRedirectURL, firstForwardedProtocol, shouldRedirectToHTTPS } from "@/lib/https-redirect";

describe("firstForwardedProtocol", () => {
  it("normalizes the first proxy value", () => {
    expect(firstForwardedProtocol(" HTTP, https ")).toBe("http");
    expect(firstForwardedProtocol(null)).toBe("");
  });
});

describe("shouldRedirectToHTTPS", () => {
  it("redirects forwarded HTTP only in production", () => {
    expect(shouldRedirectToHTTPS("production", "http")).toBe(true);
    expect(shouldRedirectToHTTPS("development", "http")).toBe(false);
    expect(shouldRedirectToHTTPS("production", "https")).toBe(false);
  });
});

describe("buildHTTPSRedirectURL", () => {
  it("uses the configured canonical HTTPS origin and preserves path and query", () => {
    expect(buildHTTPSRedirectURL("https://www.deadlockpatchnotes.com", "/patches", "?page=2").toString()).toBe(
      "https://www.deadlockpatchnotes.com/patches?page=2",
    );
  });

  it("does not use an untrusted request host or an insecure SITE_URL", () => {
    expect(buildHTTPSRedirectURL("http://attacker.example", "/api/healthz", "").hostname).toBe(
      "www.deadlockpatchnotes.com",
    );
  });
});
