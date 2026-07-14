import fs from "node:fs/promises";
import path from "node:path";
import {
  MAX_ASSET_BYTES,
  MAX_JSON_BYTES,
  fetchAllowedURL,
  readLimitedResponse,
  resolveContainedPath,
  validateImageBytes,
} from "../asset_security.mjs";

const REQUEST_HEADERS = {
  "User-Agent": "deadlockpatchnotes-fixture-generator/1.0",
};

export async function fetchJson(url) {
  const response = await fetchAllowedURL(url, { headers: REQUEST_HEADERS });
  if (!response.ok) {
    throw new Error(`Request failed: ${url} (${response.status})`);
  }
  const bytes = await readLimitedResponse(response, MAX_JSON_BYTES);
  return JSON.parse(bytes.toString("utf8"));
}

async function fetchBuffer(url, fetcher = fetch) {
  const response = await fetchAllowedURL(url, { fetcher, headers: REQUEST_HEADERS });
  if (!response.ok) {
    throw new Error(`Asset download failed: ${url} (${response.status})`);
  }
  const bytes = await readLimitedResponse(response, MAX_ASSET_BYTES);
  validateImageBytes(bytes, response.headers.get("content-type"));
  return bytes;
}

export function createAssetRegistry() {
  const byUrl = new Map();

  return {
    register(url, relPath) {
      if (!url) {
        return null;
      }

      const cleanRelPath = relPath.replace(/^\//, "");
      if (!byUrl.has(url)) {
        byUrl.set(url, {
          url,
          relPath: cleanRelPath,
          publicPath: `/${cleanRelPath}`,
        });
      }
      return byUrl.get(url);
    },
    entries() {
      return [...byUrl.values()];
    },
  };
}

async function downloadAsset(asset, outputPath, fetcher) {
  await fs.mkdir(path.dirname(outputPath), { recursive: true });
  const bytes = await fetchBuffer(asset.url, fetcher);
  await fs.writeFile(outputPath, bytes);
  process.stdout.write(`downloaded ${asset.relPath}\n`);
}

export async function downloadAssets(registry, webPublicDir, fetcher = fetch) {
  for (const asset of registry.entries()) {
    const outputPath = resolveContainedPath(webPublicDir, asset.relPath, "Asset output path");
    await downloadAsset(asset, outputPath, fetcher);
  }
}

function relativeDescendantPath(parent, candidate, label) {
  const relativePath = path.relative(parent, candidate);
  if (
    relativePath === "" ||
    relativePath === ".." ||
    relativePath.startsWith(`..${path.sep}`) ||
    path.isAbsolute(relativePath)
  ) {
    throw new Error(`${label} must be inside ${parent}: ${candidate}`);
  }
  return relativePath;
}

async function installStagedDirectory(stagingDir, targetDir, backupDir) {
  let movedExistingTarget = false;
  try {
    await fs.rename(targetDir, backupDir);
    movedExistingTarget = true;
  } catch (error) {
    if (error?.code !== "ENOENT") {
      throw error;
    }
  }

  try {
    await fs.rename(stagingDir, targetDir);
  } catch (error) {
    if (movedExistingTarget) {
      try {
        await fs.rename(backupDir, targetDir);
      } catch (restoreError) {
        throw new AggregateError([error, restoreError], "Failed to install assets and restore the previous directory");
      }
    }
    throw error;
  }

  if (movedExistingTarget) {
    await fs.rm(backupDir, { recursive: true, force: true });
  }
}

export async function replaceAssetDirectory(registry, webPublicDir, targetDir, fetcher = fetch) {
  const publicRoot = path.resolve(webPublicDir);
  const resolvedTarget = path.resolve(targetDir);
  relativeDescendantPath(publicRoot, resolvedTarget, "Asset target");

  const targetParent = path.dirname(resolvedTarget);
  await fs.mkdir(targetParent, { recursive: true });
  const stagingDir = await fs.mkdtemp(path.join(targetParent, `.${path.basename(resolvedTarget)}-staging-`));
  const backupDir = path.join(targetParent, `.${path.basename(resolvedTarget)}-backup-${process.pid}-${Date.now()}`);

  try {
    for (const asset of registry.entries()) {
      const assetPath = path.resolve(publicRoot, asset.relPath);
      const relativeToTarget = relativeDescendantPath(resolvedTarget, assetPath, "Asset path");
      await downloadAsset(asset, path.join(stagingDir, relativeToTarget), fetcher);
    }
    await installStagedDirectory(stagingDir, resolvedTarget, backupDir);
  } finally {
    await fs.rm(stagingDir, { recursive: true, force: true });
  }
}

export async function writeTextFileAtomically(filePath, contents) {
  await fs.mkdir(path.dirname(filePath), { recursive: true });
  const temporaryPath = `${filePath}.tmp-${process.pid}-${Date.now()}`;
  try {
    await fs.writeFile(temporaryPath, contents);
    await fs.rename(temporaryPath, filePath);
  } finally {
    await fs.rm(temporaryPath, { force: true });
  }
}
