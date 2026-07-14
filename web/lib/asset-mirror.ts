import mirrorManifest from "@/public/assets/mirror/manifest.json";

type ManifestAsset = {
  url?: string;
  localPath?: string;
};

type MirrorManifest = {
  assets?: ManifestAsset[];
};

let mirrorLookup: Map<string, string> | null = null;

function loadMirrorLookup(): Map<string, string> {
  const lookup = new Map<string, string>();
  const payload = mirrorManifest as MirrorManifest;
  for (const asset of payload.assets ?? []) {
    const remoteURL = String(asset?.url ?? "").trim();
    const localPath = String(asset?.localPath ?? "").trim();
    if (!remoteURL || !localPath.startsWith("/")) {
      continue;
    }
    lookup.set(remoteURL, localPath);
  }
  return lookup;
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
