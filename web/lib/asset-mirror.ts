import fs from "node:fs";
import path from "node:path";

type ManifestAsset = {
  url?: string;
  localPath?: string;
};

type MirrorManifest = {
  assets?: ManifestAsset[];
};

const MANIFEST_REL_PATH = path.join("assets", "mirror", "manifest.json");

let mirrorLookup: Map<string, string> | null = null;

function candidateManifestPaths(): string[] {
  const cwd = process.cwd();
  return [
    path.join(cwd, "public", MANIFEST_REL_PATH),
    path.join(cwd, "web", "public", MANIFEST_REL_PATH),
  ];
}

function loadMirrorLookup(): Map<string, string> {
  for (const candidatePath of candidateManifestPaths()) {
    try {
      if (!fs.existsSync(candidatePath)) {
        continue;
      }

      const raw = fs.readFileSync(candidatePath, "utf8");
      const payload = JSON.parse(raw) as MirrorManifest;
      const lookup = new Map<string, string>();

      for (const asset of payload.assets ?? []) {
        const remoteURL = String(asset?.url ?? "").trim();
        const localPath = String(asset?.localPath ?? "").trim();
        if (!remoteURL || !localPath.startsWith("/")) {
          continue;
        }
        lookup.set(remoteURL, localPath);
      }

      return lookup;
    } catch {
      return new Map();
    }
  }

  return new Map();
}

function isRemoteURL(value: string): boolean {
  return value.startsWith("https://") || value.startsWith("http://");
}

export function resolveMirroredAssetURL(rawURL: string | undefined): string | undefined {
  const value = String(rawURL ?? "").trim();
  if (value === "") {
    return undefined;
  }
  if (!isRemoteURL(value)) {
    return undefined;
  }

  if (mirrorLookup === null) {
    mirrorLookup = loadMirrorLookup();
  }

  return mirrorLookup.get(value);
}
