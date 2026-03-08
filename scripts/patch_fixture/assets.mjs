import fs from "node:fs/promises";
import path from "node:path";

const REQUEST_HEADERS = {
  "User-Agent": "deadlockpatchnotes-fixture-generator/1.0",
};

export async function fetchJson(url) {
  const response = await fetch(url, { headers: REQUEST_HEADERS });
  if (!response.ok) {
    throw new Error(`Request failed: ${url} (${response.status})`);
  }
  return response.json();
}

async function fetchBuffer(url) {
  const response = await fetch(url, { headers: REQUEST_HEADERS });
  if (!response.ok) {
    throw new Error(`Asset download failed: ${url} (${response.status})`);
  }
  const bytes = await response.arrayBuffer();
  return Buffer.from(bytes);
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

export async function downloadAssets(registry, webPublicDir) {
  for (const asset of registry.entries()) {
    const outputPath = path.join(webPublicDir, asset.relPath);
    await fs.mkdir(path.dirname(outputPath), { recursive: true });

    try {
      const bytes = await fetchBuffer(asset.url);
      await fs.writeFile(outputPath, bytes);
      process.stdout.write(`downloaded ${asset.relPath}\n`);
    } catch (error) {
      process.stdout.write(`warn: ${asset.url} -> ${String(error.message)}\n`);
    }
  }
}
