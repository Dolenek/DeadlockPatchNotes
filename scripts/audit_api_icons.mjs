#!/usr/bin/env node
import fs from "node:fs/promises";
import syncFS from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const DEFAULT_API_BASE_URL = "https://deadlockpatchnotes.com/api";
const DEFAULT_CONCURRENCY = 8;
const REQUEST_HEADERS = {
  "User-Agent": "deadlockpatchnotes-icon-audit/1.0",
};
const RETRYABLE_STATUS = new Set([408, 425, 429, 500, 502, 503, 504]);
const ALLOWED_REMOTE_HOSTS = new Set([
  "assets-bucket.deadlock-api.com",
  "assets.deadlock-api.com",
]);

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT = path.resolve(__dirname, "..");
const DEFAULT_WEB_PUBLIC_DIR = path.join(ROOT, "web", "public");

function normalizeBasePath(pathname) {
  const trimmed = pathname.replace(/\/+$/, "");
  if (trimmed === "" || trimmed === "/") {
    return "";
  }
  if (trimmed === "/api") {
    return "";
  }
  return trimmed;
}

function normalizeAPIBaseURL(raw) {
  const candidate = String(raw || DEFAULT_API_BASE_URL).trim();
  if (!candidate) {
    throw new Error("API base URL cannot be empty");
  }

  const parsed = new URL(candidate);
  const basePath = normalizeBasePath(parsed.pathname);
  return `${parsed.origin}${basePath}`;
}

function parseArgs(argv) {
  const options = {
    apiBaseURL: DEFAULT_API_BASE_URL,
    webPublicDir: DEFAULT_WEB_PUBLIC_DIR,
    jsonOut: "",
    csvOut: "",
    concurrency: DEFAULT_CONCURRENCY,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const token = argv[index];
    const next = argv[index + 1];

    if (token === "--api-base-url" && next) {
      options.apiBaseURL = next;
      index += 1;
      continue;
    }
    if (token === "--web-public-dir" && next) {
      options.webPublicDir = path.resolve(next);
      index += 1;
      continue;
    }
    if (token === "--json-out" && next) {
      options.jsonOut = path.resolve(next);
      index += 1;
      continue;
    }
    if (token === "--csv-out" && next) {
      options.csvOut = path.resolve(next);
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

  const stamp = new Date().toISOString().replace(/[:.]/g, "-");
  if (!options.jsonOut) {
    options.jsonOut = path.join("/tmp", `deadlock-icon-audit-${stamp}.json`);
  }
  if (!options.csvOut) {
    options.csvOut = path.join("/tmp", `deadlock-icon-audit-${stamp}.csv`);
  }

  options.apiBaseURL = normalizeAPIBaseURL(options.apiBaseURL);
  return options;
}

function printUsage() {
  process.stdout.write(
    [
      "Usage: node scripts/audit_api_icons.mjs [options]",
      "",
      "Options:",
      "  --api-base-url <url>    API host (default: https://deadlockpatchnotes.com/api)",
      "  --web-public-dir <dir>  Local web/public directory",
      "  --json-out <path>       Output JSON path (default: /tmp)",
      "  --csv-out <path>        Output CSV path (default: /tmp)",
      "  --concurrency <n>       Parallel fetches for detail endpoints (default: 8)",
      "",
    ].join("\n"),
  );
}

async function fetchJSONWithRetry(url, attempts = 4) {
  let lastError = null;

  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 20_000);

    try {
      const response = await fetch(url, {
        headers: REQUEST_HEADERS,
        signal: controller.signal,
      });
      if (!response.ok) {
        if (!RETRYABLE_STATUS.has(response.status) || attempt >= attempts) {
          throw new Error(`Request failed: ${url} (${response.status})`);
        }
      } else {
        return await response.json();
      }
    } catch (error) {
      lastError = error;
      if (attempt >= attempts) {
        break;
      }
    } finally {
      clearTimeout(timeout);
    }

    const waitMs = 300 * attempt;
    await new Promise((resolve) => setTimeout(resolve, waitMs));
  }

  throw lastError instanceof Error ? lastError : new Error(String(lastError));
}

function apiURL(apiBaseURL, endpointPath) {
  return `${apiBaseURL}${endpointPath}`;
}

async function mapWithConcurrency(items, concurrency, worker) {
  const outputs = new Array(items.length);
  let cursor = 0;

  async function run() {
    while (cursor < items.length) {
      const index = cursor;
      cursor += 1;
      outputs[index] = await worker(items[index], index);
    }
  }

  const runners = [];
  const count = Math.min(concurrency, Math.max(1, items.length));
  for (let i = 0; i < count; i += 1) {
    runners.push(run());
  }

  await Promise.all(runners);
  return outputs;
}

function addURL(records, rawURL, source) {
  const url = String(rawURL ?? "").trim();
  if (!url) {
    return;
  }
  if (!records.has(url)) {
    records.set(url, new Set());
  }
  records.get(url).add(source);
}

function extractSectionURLs(records, sections, sourcePrefix) {
  if (!Array.isArray(sections)) {
    return;
  }

  for (const section of sections) {
    const sectionID = String(section?.id ?? "section");
    const entries = Array.isArray(section?.entries) ? section.entries : [];

    for (const entry of entries) {
      const entryID = String(entry?.id ?? "entry");
      const entryPrefix = `${sourcePrefix}.section(${sectionID}).entry(${entryID})`;
      addURL(records, entry?.entityIconUrl, `${entryPrefix}.entityIconUrl`);
      addURL(records, entry?.entityIconFallbackUrl, `${entryPrefix}.entityIconFallbackUrl`);

      const groups = Array.isArray(entry?.groups) ? entry.groups : [];
      for (const group of groups) {
        const groupID = String(group?.id ?? "group");
        const groupPrefix = `${entryPrefix}.group(${groupID})`;
        addURL(records, group?.iconUrl, `${groupPrefix}.iconUrl`);
        addURL(records, group?.iconFallbackUrl, `${groupPrefix}.iconFallbackUrl`);
      }
    }
  }
}

async function collectPatchSlugs(apiBaseURL) {
  const slugs = [];
  const pageSize = 50;
  let page = 1;
  let totalPages = 1;

  while (page <= totalPages) {
    const payload = await fetchJSONWithRetry(apiURL(apiBaseURL, `/api/v1/patches?page=${page}&limit=${pageSize}`));
    const patches = Array.isArray(payload?.patches) ? payload.patches : [];
    for (const patch of patches) {
      const slug = String(patch?.slug ?? "").trim();
      if (!slug) {
        continue;
      }
      slugs.push(slug);
    }

    totalPages = Number(payload?.pagination?.totalPages ?? 1) || 1;
    page += 1;
  }

  return [...new Set(slugs)];
}

function classifyURL(url, webPublicDir) {
  if (url.startsWith("/")) {
    const relativePath = url.replace(/^\/+/, "");
    const diskPath = path.join(webPublicDir, relativePath);
    const localExists = syncFS.existsSync(diskPath);
    return {
      kind: "local",
      localPath: `/${relativePath}`,
      localExists,
    };
  }

  if (url.startsWith("https://") || url.startsWith("http://")) {
    try {
      const host = new URL(url).hostname;
      return {
        kind: "remote",
        host,
        allowedHost: ALLOWED_REMOTE_HOSTS.has(host),
      };
    } catch {
      return { kind: "other" };
    }
  }

  return { kind: "other" };
}

function formatCSV(rows) {
  const header = [
    "kind",
    "url",
    "host",
    "allowed_host",
    "local_path",
    "local_exists",
    "source_count",
    "sources",
  ];

  const csvRows = [header.join(",")];
  for (const row of rows) {
    const values = [
      row.kind,
      row.url,
      row.host ?? "",
      row.allowedHost === undefined ? "" : String(row.allowedHost),
      row.localPath ?? "",
      row.localExists === undefined ? "" : String(row.localExists),
      String(row.sourceCount),
      row.sources.join(" | "),
    ];

    csvRows.push(values.map(csvEscape).join(","));
  }

  return `${csvRows.join("\n")}\n`;
}

function csvEscape(value) {
  const raw = String(value ?? "");
  if (!/[",\n]/.test(raw)) {
    return raw;
  }
  return `"${raw.replace(/"/g, '""')}"`;
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const records = new Map();

  const patchSlugs = await collectPatchSlugs(options.apiBaseURL);
  await mapWithConcurrency(patchSlugs, options.concurrency, async (slug) => {
    const payload = await fetchJSONWithRetry(apiURL(options.apiBaseURL, `/api/v1/patches/${encodeURIComponent(slug)}`));
    addURL(records, payload?.imageUrl, `patches.detail(${slug}).imageUrl`);
    extractSectionURLs(records, payload?.sections, `patches.detail(${slug}).sections`);

    const timeline = Array.isArray(payload?.releaseTimeline) ? payload.releaseTimeline : [];
    for (const block of timeline) {
      const blockID = String(block?.id ?? "block");
      extractSectionURLs(records, block?.sections, `patches.detail(${slug}).releaseTimeline(${blockID})`);
    }
  });

  const heroListPayload = await fetchJSONWithRetry(apiURL(options.apiBaseURL, "/api/v1/heroes"));
  const heroes = Array.isArray(heroListPayload?.heroes) ? heroListPayload.heroes : [];
  for (const hero of heroes) {
    addURL(records, hero?.iconUrl, `heroes.list(${hero?.slug ?? ""}).iconUrl`);
    addURL(records, hero?.iconFallbackUrl, `heroes.list(${hero?.slug ?? ""}).iconFallbackUrl`);
  }

  await mapWithConcurrency(heroes, options.concurrency, async (hero) => {
    const slug = String(hero?.slug ?? "").trim();
    if (!slug) {
      return;
    }

    const payload = await fetchJSONWithRetry(
      apiURL(options.apiBaseURL, `/api/v1/heroes/${encodeURIComponent(slug)}/changes`),
    );

    addURL(records, payload?.hero?.iconUrl, `heroes.detail(${slug}).hero.iconUrl`);
    addURL(records, payload?.hero?.iconFallbackUrl, `heroes.detail(${slug}).hero.iconFallbackUrl`);

    const timeline = Array.isArray(payload?.timeline) ? payload.timeline : [];
    for (const block of timeline) {
      const blockID = String(block?.id ?? "block");
      const skills = Array.isArray(block?.skills) ? block.skills : [];
      for (const skill of skills) {
        const skillID = String(skill?.id ?? "skill");
        addURL(records, skill?.iconUrl, `heroes.detail(${slug}).timeline(${blockID}).skill(${skillID}).iconUrl`);
        addURL(
          records,
          skill?.iconFallbackUrl,
          `heroes.detail(${slug}).timeline(${blockID}).skill(${skillID}).iconFallbackUrl`,
        );
      }
    }
  });

  const itemListPayload = await fetchJSONWithRetry(apiURL(options.apiBaseURL, "/api/v1/items"));
  const items = Array.isArray(itemListPayload?.items) ? itemListPayload.items : [];
  for (const item of items) {
    addURL(records, item?.iconUrl, `items.list(${item?.slug ?? ""}).iconUrl`);
    addURL(records, item?.iconFallbackUrl, `items.list(${item?.slug ?? ""}).iconFallbackUrl`);
  }

  await mapWithConcurrency(items, options.concurrency, async (item) => {
    const slug = String(item?.slug ?? "").trim();
    if (!slug) {
      return;
    }

    const payload = await fetchJSONWithRetry(apiURL(options.apiBaseURL, `/api/v1/items/${encodeURIComponent(slug)}/changes`));
    addURL(records, payload?.item?.iconUrl, `items.detail(${slug}).item.iconUrl`);
    addURL(records, payload?.item?.iconFallbackUrl, `items.detail(${slug}).item.iconFallbackUrl`);
  });

  const spellListPayload = await fetchJSONWithRetry(apiURL(options.apiBaseURL, "/api/v1/spells"));
  const spells = Array.isArray(spellListPayload?.spells) ? spellListPayload.spells : [];
  for (const spell of spells) {
    addURL(records, spell?.iconUrl, `spells.list(${spell?.slug ?? ""}).iconUrl`);
    addURL(records, spell?.iconFallbackUrl, `spells.list(${spell?.slug ?? ""}).iconFallbackUrl`);
  }

  await mapWithConcurrency(spells, options.concurrency, async (spell) => {
    const slug = String(spell?.slug ?? "").trim();
    if (!slug) {
      return;
    }

    const payload = await fetchJSONWithRetry(apiURL(options.apiBaseURL, `/api/v1/spells/${encodeURIComponent(slug)}/changes`));
    addURL(records, payload?.spell?.iconUrl, `spells.detail(${slug}).spell.iconUrl`);
    addURL(records, payload?.spell?.iconFallbackUrl, `spells.detail(${slug}).spell.iconFallbackUrl`);

    const timeline = Array.isArray(payload?.timeline) ? payload.timeline : [];
    for (const block of timeline) {
      const blockID = String(block?.id ?? "block");
      const entries = Array.isArray(block?.entries) ? block.entries : [];
      for (const entry of entries) {
        const entryID = String(entry?.id ?? "entry");
        const prefix = `spells.detail(${slug}).timeline(${blockID}).entry(${entryID})`;
        addURL(records, entry?.heroIconUrl, `${prefix}.heroIconUrl`);
        addURL(records, entry?.heroIconFallbackUrl, `${prefix}.heroIconFallbackUrl`);
      }
    }
  });

  const rows = [];
  const summary = {
    uniqueURLs: 0,
    localExisting: 0,
    localMissing: 0,
    remoteAllowed: 0,
    remoteDisallowed: 0,
    other: 0,
  };

  for (const [url, sources] of records.entries()) {
    const classification = classifyURL(url, options.webPublicDir);
    const row = {
      url,
      sourceCount: sources.size,
      sources: [...sources].sort((left, right) => left.localeCompare(right)),
      ...classification,
    };

    rows.push(row);
    summary.uniqueURLs += 1;

    if (row.kind === "local") {
      if (row.localExists) {
        summary.localExisting += 1;
      } else {
        summary.localMissing += 1;
      }
      continue;
    }

    if (row.kind === "remote") {
      if (row.allowedHost) {
        summary.remoteAllowed += 1;
      } else {
        summary.remoteDisallowed += 1;
      }
      continue;
    }

    summary.other += 1;
  }

  rows.sort((left, right) => {
    if (left.kind !== right.kind) {
      return left.kind.localeCompare(right.kind);
    }
    return left.url.localeCompare(right.url);
  });

  const report = {
    generatedAt: new Date().toISOString(),
    apiBaseURL: options.apiBaseURL,
    webPublicDir: options.webPublicDir,
    allowedRemoteHosts: [...ALLOWED_REMOTE_HOSTS].sort((left, right) => left.localeCompare(right)),
    discovery: {
      patches: patchSlugs.length,
      heroes: heroes.length,
      items: items.length,
      spells: spells.length,
    },
    summary,
    urls: rows,
  };

  await fs.mkdir(path.dirname(options.jsonOut), { recursive: true });
  await fs.writeFile(options.jsonOut, `${JSON.stringify(report, null, 2)}\n`);
  await fs.mkdir(path.dirname(options.csvOut), { recursive: true });
  await fs.writeFile(options.csvOut, formatCSV(rows));

  process.stdout.write(`wrote JSON report ${options.jsonOut}\n`);
  process.stdout.write(`wrote CSV report ${options.csvOut}\n`);
  process.stdout.write(`unique urls: ${summary.uniqueURLs}\n`);
  process.stdout.write(`local existing: ${summary.localExisting}\n`);
  process.stdout.write(`local missing: ${summary.localMissing}\n`);
  process.stdout.write(`remote allowed-host: ${summary.remoteAllowed}\n`);
  process.stdout.write(`remote disallowed-host: ${summary.remoteDisallowed}\n`);
  process.stdout.write(`other: ${summary.other}\n`);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
