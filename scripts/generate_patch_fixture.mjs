#!/usr/bin/env node
import path from "node:path";
import {
  createAssetRegistry,
  fetchJson,
  replaceAssetDirectory,
  writeTextFileAtomically,
} from "./patch_fixture/assets.mjs";
import {
  FIXTURE_PATH,
  MANIFEST_PATH,
  PATCH_ASSET_DIR,
  PATCH_SLUG,
  ROOT,
  STEAM_NEWS_URL,
  WEB_PUBLIC_DIR,
} from "./patch_fixture/config.mjs";
import { buildPatchDetail, findSteamPatchItem } from "./patch_fixture/build_patch_detail.mjs";
import { heroLookupFromAssets } from "./patch_fixture/lookups.mjs";

async function loadSourceData() {
  const steamPayload = await fetchJson(STEAM_NEWS_URL);
  const steamItem = findSteamPatchItem(steamPayload);
  if (!steamItem) {
    throw new Error("Could not locate configured Steam patch item");
  }

  const [allHeroes, allItems] = await Promise.all([
    fetchJson("https://assets.deadlock-api.com/v2/heroes"),
    fetchJson("https://assets.deadlock-api.com/v2/items"),
  ]);

  return { steamItem, allHeroes, allItems };
}

function buildManifest(assetsRegistry) {
  return {
    generatedAt: new Date().toISOString(),
    patchSlug: PATCH_SLUG,
    assetCount: assetsRegistry.entries().length,
    assets: assetsRegistry.entries().map((entry) => ({
      url: entry.url,
      localPath: entry.publicPath,
    })),
  };
}

async function writeOutputs(detail, assetsRegistry) {
  await replaceAssetDirectory(assetsRegistry, WEB_PUBLIC_DIR, PATCH_ASSET_DIR);

  await writeTextFileAtomically(FIXTURE_PATH, `${JSON.stringify(detail, null, 2)}\n`);

  const manifest = buildManifest(assetsRegistry);
  await writeTextFileAtomically(MANIFEST_PATH, `${JSON.stringify(manifest, null, 2)}\n`);
}

function printSummary(stats) {
  process.stdout.write(`wrote fixture ${path.relative(ROOT, FIXTURE_PATH)}\n`);
  process.stdout.write(`wrote manifest ${path.relative(ROOT, MANIFEST_PATH)}\n`);
  process.stdout.write(`general lines: ${stats.generalLines}\n`);
  process.stdout.write(`item changes: ${stats.itemChanges}\n`);
  process.stdout.write(`hero changes: ${stats.heroChanges}\n`);
}

async function main() {
  const { steamItem, allHeroes, allItems } = await loadSourceData();
  const assetsRegistry = createAssetRegistry();

  const { detail, stats } = await buildPatchDetail({
    steamItem,
    allItems,
    assetsRegistry,
    fetchJson,
    heroesLookup: heroLookupFromAssets(allHeroes),
  });

  await writeOutputs(detail, assetsRegistry);
  printSummary(stats);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
