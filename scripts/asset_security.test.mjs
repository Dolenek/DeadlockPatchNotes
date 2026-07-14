import assert from "node:assert/strict";
import path from "node:path";
import test from "node:test";
import {
  fetchAllowedURL,
  parseAllowedDownloadURL,
  readLimitedResponse,
  resolveContainedPath,
  validateImageBytes,
} from "./asset_security.mjs";

test("download URL validation rejects credentials, ports, HTTP, and other hosts", () => {
  assert.equal(parseAllowedDownloadURL("https://assets.deadlock-api.com/icon.png")?.hostname, "assets.deadlock-api.com");
  assert.equal(parseAllowedDownloadURL("http://assets.deadlock-api.com/icon.png"), null);
  assert.equal(parseAllowedDownloadURL("https://user:pass@assets.deadlock-api.com/icon.png"), null);
  assert.equal(parseAllowedDownloadURL("https://assets.deadlock-api.com:8443/icon.png"), null);
  assert.equal(parseAllowedDownloadURL("https://127.0.0.1/icon.png"), null);
});

test("redirects are revalidated against the allowlist", async () => {
  const fetcher = async () => new Response(null, { status: 302, headers: { location: "https://127.0.0.1/private" } });
  await assert.rejects(fetchAllowedURL("https://assets.deadlock-api.com/icon.png", { fetcher }), /not allowed/);
});

test("contained paths cannot escape their configured root", () => {
  const root = path.resolve("web", "public");
  assert.equal(resolveContainedPath(root, "/assets/icon.png"), path.join(root, "assets", "icon.png"));
  assert.throws(() => resolveContainedPath(root, "/../secret.txt"), /must stay inside/);
  assert.throws(() => resolveContainedPath(root, "assets/../../secret.txt"), /must stay inside/);
});

test("response reads stop at the byte limit", async () => {
  await assert.rejects(readLimitedResponse(new Response("12345"), 4), /exceeds/);
});

test("image validation checks raster signatures and blocks active SVG", () => {
  const png = Buffer.from("89504e470d0a1a0a", "hex");
  assert.equal(validateImageBytes(png, "image/png"), "image/png");
  assert.throws(() => validateImageBytes(Buffer.from("not png"), "image/png"), /does not match/);
  assert.throws(() => validateImageBytes(Buffer.from("<svg><script>alert(1)</script></svg>"), "image/svg+xml"), /active/);
  assert.equal(validateImageBytes(Buffer.from("<svg xmlns='http://www.w3.org/2000/svg'><path d='M0 0'/></svg>"), "image/svg+xml"), "image/svg+xml");
});
