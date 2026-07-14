import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { cp, mkdir, mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { setTimeout as delay } from "node:timers/promises";

const webRoot = fileURLToPath(new URL("..", import.meta.url));
const port = Number(process.env.RUNTIME_TEST_PORT ?? 39000 + (process.pid % 1000));
const origin = `http://127.0.0.1:${port}`;
const runtimeParent = await mkdtemp(join(tmpdir(), "deadlock-web-runtime-"));
const runtimeRoot = join(runtimeParent, "app");
let serverOutput = "";

await cp(join(webRoot, ".next", "standalone"), runtimeRoot, { recursive: true });
await mkdir(join(runtimeRoot, ".next"), { recursive: true });
await cp(join(webRoot, ".next", "static"), join(runtimeRoot, ".next", "static"), { recursive: true });
await cp(join(webRoot, "public"), join(runtimeRoot, "public"), { recursive: true });

const server = spawn(process.execPath, [join(runtimeRoot, "server.js")], {
  cwd: runtimeRoot,
  env: {
    ...process.env,
    NODE_ENV: "production",
    HOSTNAME: "127.0.0.1",
    PORT: String(port),
    SITE_URL: "https://www.deadlockpatchnotes.com",
  },
  stdio: ["ignore", "pipe", "pipe"],
});

for (const stream of [server.stdout, server.stderr]) {
  stream.on("data", (chunk) => {
    serverOutput += chunk.toString();
  });
}

async function request(pathname, forwardedProtocol = "https") {
  return fetch(`${origin}${pathname}`, {
    headers: { "x-forwarded-proto": forwardedProtocol },
    redirect: "manual",
    signal: AbortSignal.timeout(10_000),
  });
}

async function waitUntilReady() {
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    if (server.exitCode !== null) {
      throw new Error(`Next.js exited before becoming ready.\n${serverOutput}`);
    }
    try {
      const response = await request("/healthz");
      if (response.ok) {
        return;
      }
    } catch {
      // The server is still starting.
    }
    await delay(200);
  }
  throw new Error(`Next.js did not become ready.\n${serverOutput}`);
}

async function verifyRedirectAndImages() {
  const documentRedirect = await request("/patches", "http");
  assert.equal(documentRedirect.status, 308);
  assert.equal(documentRedirect.headers.get("location"), "https://www.deadlockpatchnotes.com/patches");

  const staticImage = await request("/header_heroes.png", "http");
  assert.equal(staticImage.status, 200);
  assert.equal(staticImage.headers.get("content-type"), "image/png");

  const optimizedImage = await request("/_next/image?url=%2Fheader_heroes.png&w=640&q=74");
  assert.equal(optimizedImage.status, 200);
  assert.match(optimizedImage.headers.get("content-type") ?? "", /^image\/(?:avif|png|webp)$/);
  assert.ok((await optimizedImage.arrayBuffer()).byteLength > 0);
}

async function loadGeneratedStylesheets(html) {
  const stylesheetPaths = [...html.matchAll(/href="([^"]+\.css[^"]*)"/g)].map((match) => match[1]);
  assert.ok(stylesheetPaths.length > 0, "Expected at least one generated stylesheet");

  return Promise.all(
    stylesheetPaths.map(async (pathname) => {
      const response = await request(pathname);
      assert.equal(response.status, 200);
      return { pathname, css: await response.text() };
    }),
  );
}

async function verifySelfHostedFonts(html) {
  const stylesheets = await loadGeneratedStylesheets(html);
  const combinedCSS = stylesheets.map(({ css }) => css).join("\n");
  assert.doesNotMatch(`${html}\n${combinedCSS}`, /fonts\.(?:googleapis|gstatic)\.com/);
  assert.match(combinedCSS, /--font-(?:barlow|cinzel|jetbrains-mono)/);

  const fontStylesheet = stylesheets.find(({ css }) => /url\([^)]+\.woff2\)/.test(css));
  const fontReference = fontStylesheet?.css.match(/url\(([^)]+\.woff2)\)/)?.[1]?.replaceAll(/["']/g, "");
  assert.ok(fontReference && fontStylesheet, "Expected a self-hosted WOFF2 font reference");
  const fontURL = new URL(fontReference, new URL(fontStylesheet.pathname, origin));
  const fontResponse = await request(`${fontURL.pathname}${fontURL.search}`);
  assert.equal(fontResponse.status, 200);
  assert.equal(fontResponse.headers.get("content-type"), "font/woff2");
}

async function verifyFontsAndDynamicRendering() {
  const home = await request("/");
  assert.equal(home.status, 200);
  assert.match(home.headers.get("cache-control") ?? "", /no-store/);

  const sitemap = await request("/sitemap.xml");
  assert.equal(sitemap.status, 200);
  assert.match(sitemap.headers.get("cache-control") ?? "", /max-age=0/);
  assert.doesNotMatch(sitemap.headers.get("cache-control") ?? "", /s-maxage/);

  const html = await home.text();
  await verifySelfHostedFonts(html);
}

try {
  await waitUntilReady();
  await verifyRedirectAndImages();
  await verifyFontsAndDynamicRendering();
  process.stdout.write("Production runtime checks passed.\n");
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.stack : error}\n${serverOutput.slice(-4000)}`);
  process.exitCode = 1;
} finally {
  server.kill("SIGTERM");
  await Promise.race([
    new Promise((resolve) => server.once("exit", resolve)),
    delay(5_000),
  ]);
  await rm(runtimeParent, { recursive: true, force: true });
}
