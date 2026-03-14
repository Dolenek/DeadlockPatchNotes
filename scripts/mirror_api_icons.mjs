#!/usr/bin/env node
import crypto from "node:crypto";
import fs from "node:fs/promises";
import syncFS from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const REQUEST_HEADERS = {
  "User-Agent": "deadlockpatchnotes-icon-mirror/1.0",
};
const RETRYABLE_STATUS = new Set([408, 425, 429, 500, 502, 503, 504]);
const DEFAULT_CONCURRENCY = 8;
const EXT_BY_CONTENT_TYPE = {
  "image/png": ".png",
  "image/webp": ".webp",
  "image/jpeg": ".jpg",
  "image/jpg": ".jpg",
  "image/svg+xml": ".svg",
  "image/gif": ".gif",
  "image/avif": ".avif",
};

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT = path.resolve(__dirname, "..");
const DEFAULT_WEB_PUBLIC_DIR = path.join(ROOT, "web", "public");
const DEFAULT_MANIFEST_PATH = path.join(DEFAULT_WEB_PUBLIC_DIR, "assets", "mirror", "manifest.json");

function parseArgs(argv) {
  const options = {
    auditPath: "",
    webPublicDir: DEFAULT_WEB_PUBLIC_DIR,
    manifestPath: DEFAULT_MANIFEST_PATH,
    concurrency: DEFAULT_CONCURRENCY,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const token = argv[index];
    const next = argv[index + 1];

    if (token === "--audit" && next) {
      options.auditPath = path.resolve(next);
      index += 1;
      continue;
    }
    if (token === "--web-public-dir" && next) {
      options.webPublicDir = path.resolve(next);
      index += 1;
      continue;
    }
    if (token === "--manifest-out" && next) {
      options.manifestPath = path.resolve(next);
      index += 1;
      continue;
    }
    if (token === "--concurrency" && next) {
      const parsed = Number(next);
      if (Number.isFinite(parsed) && parsed > 0) {
        options.concurrency = Math.min(32, Math.floor(parsed));
      }
      index += 1;
      continue;
    }
    if (token === "-h" || token === "--help") {
      printUsage();
      process.exit(0);
    }

    throw new Error(`Unknown argument: ${token}`);
  }

  if (!options.auditPath) {
    throw new Error("Missing required --audit <path> argument");
  }

  return options;
}

function printUsage() {
  process.stdout.write(
    [
      "Usage: node scripts/mirror_api_icons.mjs --audit <report.json> [options]",
      "",
      "Options:",
      "  --audit <path>          JSON report from audit_api_icons.mjs",
      "  --web-public-dir <dir>  Local web/public directory",
      "  --manifest-out <path>   Output mirror manifest path",
      "  --concurrency <n>       Parallel downloads (default: 8)",
      "",
    ].join("\n"),
  );
}

function sha1(value) {
  return crypto.createHash("sha1").update(value).digest("hex");
}

function shortHash(value, length = 12) {
  return sha1(value).slice(0, length);
}

function toFileSlug(value, maxLength = 72) {
  const normalized = String(value ?? "")
    .toLowerCase()
    .replace(/[%_]+/g, "-")
    .replace(/[^a-z0-9-]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");

  if (!normalized) {
    return "";
  }

  return normalized.slice(0, maxLength).replace(/-+$/g, "");
}

function isHexOnly(value) {
  return /^[a-f0-9]{12,}$/i.test(value);
}

function extensionFromURL(url) {
  try {
    const parsed = new URL(url);
    const ext = path.extname(parsed.pathname).toLowerCase();
    if (/^\.[a-z0-9]{1,8}$/.test(ext)) {
      return ext;
    }
  } catch {
    return "";
  }
  return "";
}

function extensionFromContentType(contentType) {
  if (!contentType) {
    return "";
  }

  const normalized = String(contentType).toLowerCase().split(";")[0].trim();
  return EXT_BY_CONTENT_TYPE[normalized] ?? "";
}

function toRelativeDiskPath(localPath) {
  return localPath.replace(/^\/+/, "");
}

function baseNameFromURL(url) {
  try {
    const parsed = new URL(url);
    const rawName = path.basename(parsed.pathname, path.extname(parsed.pathname));
    return toFileSlug(decodeURIComponent(rawName));
  } catch {
    return "";
  }
}

function firstSourceToken(source) {
  const directMatch = source.match(/\b(?:patches|heroes|items|spells)\.(?:detail|list)\(([^)]+)\)/);
  if (directMatch?.[1]) {
    return toFileSlug(directMatch[1]);
  }

  const fallbackMatch = source.match(/\(([^)]+)\)/);
  if (fallbackMatch?.[1]) {
    return toFileSlug(fallbackMatch[1]);
  }

  return "";
}

function buildNameHint(url, auditRow) {
  const fromURL = baseNameFromURL(url);
  if (fromURL && !isHexOnly(fromURL)) {
    return fromURL;
  }

  const sources = Array.isArray(auditRow?.sources) ? auditRow.sources : [];
  for (const source of sources) {
    const token = firstSourceToken(String(source ?? ""));
    if (!token || isHexOnly(token) || token.startsWith("post-")) {
      continue;
    }
    return token;
  }

  return fromURL || "asset";
}

async function fetchBufferWithRetry(url, attempts = 4) {
  let lastError = null;

  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 25_000);

    try {
      const response = await fetch(url, {
        headers: REQUEST_HEADERS,
        signal: controller.signal,
      });

      if (!response.ok) {
        if (!RETRYABLE_STATUS.has(response.status) || attempt >= attempts) {
          throw new Error(`Asset download failed: ${url} (${response.status})`);
        }
      } else {
        const bytes = Buffer.from(await response.arrayBuffer());
        return {
          bytes,
          contentType: response.headers.get("content-type") || "",
        };
      }
    } catch (error) {
      lastError = error;
      if (attempt >= attempts) {
        break;
      }
    } finally {
      clearTimeout(timeout);
    }

    await new Promise((resolve) => setTimeout(resolve, 350 * attempt));
  }

  throw lastError instanceof Error ? lastError : new Error(String(lastError));
}

async function mapWithConcurrency(items, concurrency, worker) {
  let cursor = 0;

  async function run() {
    while (cursor < items.length) {
      const index = cursor;
      cursor += 1;
      await worker(items[index], index);
    }
  }

  const runners = [];
  const count = Math.min(concurrency, Math.max(1, items.length));
  for (let i = 0; i < count; i += 1) {
    runners.push(run());
  }

  await Promise.all(runners);
}

function buildLocalPath(url, contentType, nameHint = "") {
  const ext = extensionFromURL(url) || extensionFromContentType(contentType) || ".bin";
  const slug = toFileSlug(nameHint) || "asset";
  return `/assets/mirror/icons/${slug}-${shortHash(url)}${ext}`;
}

async function collectFilesRecursive(dirPath) {
  const entries = await fs.readdir(dirPath, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await collectFilesRecursive(fullPath)));
      continue;
    }
    if (entry.isFile()) {
      files.push(fullPath);
    }
  }

  return files;
}

async function pruneUnusedMirrorFiles(webPublicDir, usedLocalPaths) {
  const iconsDir = path.join(webPublicDir, "assets", "mirror", "icons");
  if (!syncFS.existsSync(iconsDir)) {
    return 0;
  }

  const usedDiskPaths = new Set(
    [...usedLocalPaths].map((localPath) => path.join(webPublicDir, toRelativeDiskPath(localPath))),
  );
  const allDiskFiles = await collectFilesRecursive(iconsDir);
  let removed = 0;

  for (const diskPath of allDiskFiles) {
    if (usedDiskPaths.has(diskPath)) {
      continue;
    }
    await fs.rm(diskPath, { force: true });
    removed += 1;
  }

  return removed;
}

async function loadJSON(filePath) {
  const raw = await fs.readFile(filePath, "utf8");
  return JSON.parse(raw);
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const auditReport = await loadJSON(options.auditPath);
  const auditRows = Array.isArray(auditReport?.urls) ? auditReport.urls : [];

  const targetRows = auditRows.filter((row) => row?.kind === "remote" && row?.allowedHost === true);
  const targetURLs = [...new Set(targetRows.map((row) => String(row?.url ?? "").trim()).filter(Boolean))];
  const targetRowByURL = new Map(
    targetRows.map((row) => [String(row?.url ?? "").trim(), row]),
  );

  const existingByURL = new Map();
  if (syncFS.existsSync(options.manifestPath)) {
    const existing = await loadJSON(options.manifestPath);
    for (const asset of existing?.assets ?? []) {
      const url = String(asset?.url ?? "").trim();
      const localPath = String(asset?.localPath ?? "").trim();
      if (!url || !localPath.startsWith("/")) {
        continue;
      }
      existingByURL.set(url, localPath);
    }
  }

  const resolvedByURL = new Map();
  const failures = [];
  let downloadedCount = 0;
  let skippedExisting = 0;
  let renamedExisting = 0;

  await mapWithConcurrency(targetURLs, options.concurrency, async (url) => {
    const row = targetRowByURL.get(url);
    const nameHint = buildNameHint(url, row);
    const existingLocalPath = existingByURL.get(url);
    if (existingLocalPath) {
      const existingDiskPath = path.join(options.webPublicDir, toRelativeDiskPath(existingLocalPath));
      if (syncFS.existsSync(existingDiskPath)) {
        const desiredLocalPath = buildLocalPath(url, extensionFromURL(existingLocalPath), nameHint);
        if (desiredLocalPath !== existingLocalPath) {
          const desiredDiskPath = path.join(options.webPublicDir, toRelativeDiskPath(desiredLocalPath));
          await fs.mkdir(path.dirname(desiredDiskPath), { recursive: true });
          await fs.copyFile(existingDiskPath, desiredDiskPath);
          resolvedByURL.set(url, desiredLocalPath);
          renamedExisting += 1;
          process.stdout.write(`renamed ${desiredLocalPath}\n`);
          return;
        }
        resolvedByURL.set(url, existingLocalPath);
        skippedExisting += 1;
        return;
      }
    }

    try {
      const { bytes, contentType } = await fetchBufferWithRetry(url);
      const localPath = buildLocalPath(url, contentType, nameHint);
      const outputPath = path.join(options.webPublicDir, toRelativeDiskPath(localPath));
      await fs.mkdir(path.dirname(outputPath), { recursive: true });
      await fs.writeFile(outputPath, bytes);
      resolvedByURL.set(url, localPath);
      downloadedCount += 1;
      process.stdout.write(`downloaded ${localPath}\n`);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      failures.push({ url, error: message });
      process.stdout.write(`warn: ${url} -> ${message}\n`);
    }
  });

  for (const [url, localPath] of existingByURL.entries()) {
    if (resolvedByURL.has(url)) {
      continue;
    }
    const diskPath = path.join(options.webPublicDir, toRelativeDiskPath(localPath));
    if (syncFS.existsSync(diskPath)) {
      resolvedByURL.set(url, localPath);
    }
  }

  const assets = [...resolvedByURL.entries()]
    .map(([url, localPath]) => ({ url, localPath }))
    .sort((left, right) => left.url.localeCompare(right.url));

  const removedStaleFiles = await pruneUnusedMirrorFiles(
    options.webPublicDir,
    new Set(assets.map((entry) => entry.localPath)),
  );

  const manifest = {
    generatedAt: new Date().toISOString(),
    sourceApiBase: String(auditReport?.apiBaseURL ?? ""),
    assetCount: assets.length,
    assets,
    failed: failures,
  };

  await fs.mkdir(path.dirname(options.manifestPath), { recursive: true });
  await fs.writeFile(options.manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);

  process.stdout.write(`wrote manifest ${options.manifestPath}\n`);
  process.stdout.write(`target remote urls: ${targetURLs.length}\n`);
  process.stdout.write(`downloaded: ${downloadedCount}\n`);
  process.stdout.write(`reused existing: ${skippedExisting}\n`);
  process.stdout.write(`renamed existing: ${renamedExisting}\n`);
  process.stdout.write(`removed stale files: ${removedStaleFiles}\n`);
  process.stdout.write(`failures: ${failures.length}\n`);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
