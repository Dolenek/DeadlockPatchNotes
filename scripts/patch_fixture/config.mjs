import path from "node:path";
import { fileURLToPath } from "node:url";

export const PATCH_SLUG = "2026-03-06-update";
export const PATCH_STEAM_TITLE = "Gameplay Update - 03-06-2026";
export const PATCH_STEAM_GID = "1826362059925616";
export const SOURCE_URL = "https://store.steampowered.com/news/app/1422450/view/519740319207522795";
export const HERO_IMAGE_URL =
  "https://clan.akamai.steamstatic.com/images/45164767/1a200778c94a048c5b2580a1e1a36071679ff19e.png";
export const STEAM_NEWS_URL =
  "https://api.steampowered.com/ISteamNews/GetNewsForApp/v2/?appid=1422450&count=120&maxlength=0&format=json";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export const ROOT = path.resolve(__dirname, "../..");
export const WEB_PUBLIC_DIR = path.join(ROOT, "web", "public");
export const ASSET_PREFIX = `/assets/patches/${PATCH_SLUG}`;
export const PATCH_ASSET_DIR = path.join(WEB_PUBLIC_DIR, ASSET_PREFIX.replace(/^\//, ""));
export const FIXTURE_PATH = path.join(ROOT, "api", "internal", "patches", "data", `${PATCH_SLUG}.json`);
export const MANIFEST_PATH = path.join(PATCH_ASSET_DIR, "manifest.json");
