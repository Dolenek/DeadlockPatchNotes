import {
  HeroChangesResponse,
  HeroListResponse,
  ItemChangesResponse,
  ItemListResponse,
  PatchDetail,
  PatchListResponse,
  SpellChangesResponse,
  SpellListResponse
} from "@/lib/types";

const DEFAULT_API_BASE_URL = "https://api.deadlock.jakubdolenek.xyz";

function resolveAPIBaseURL() {
  const candidate = (process.env.API_BASE_URL ?? DEFAULT_API_BASE_URL).trim();
  if (candidate === "") {
    return DEFAULT_API_BASE_URL;
  }

  try {
    const parsed = new URL(candidate);
    const path = parsed.pathname === "/" ? "" : parsed.pathname.replace(/\/+$/, "");
    return `${parsed.origin}${path}`;
  } catch {
    throw new Error(`Invalid API_BASE_URL: ${candidate}`);
  }
}

const API_BASE_URL = resolveAPIBaseURL();

export class APIError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function apiFetch<T>(path: string): Promise<T> {
  const target = `${API_BASE_URL}${path}`;
  let response: Response;
  try {
    response = await fetch(target, {
      next: { revalidate: 30 }
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new APIError(0, `API fetch failed for ${target}: ${message}`);
  }

  if (!response.ok) {
    const detail = await response.text().catch(() => "");
    const suffix = detail ? ` ${detail.slice(0, 200)}` : "";
    throw new APIError(response.status, `API request failed for ${target}: ${response.status}${suffix}`);
  }

  return (await response.json()) as T;
}

export async function getPatches(page: number, limit = 12): Promise<PatchListResponse> {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return apiFetch<PatchListResponse>(`/api/v1/patches?${params.toString()}`);
}

export async function getPatchBySlug(slug: string): Promise<PatchDetail> {
  return apiFetch<PatchDetail>(`/api/v1/patches/${slug}`);
}

type HeroChangesQuery = {
  skill?: string;
  from?: string;
  to?: string;
};

export async function getHeroes(): Promise<HeroListResponse> {
  return apiFetch<HeroListResponse>("/api/v1/heroes");
}

export async function getHeroChanges(slug: string, query: HeroChangesQuery = {}): Promise<HeroChangesResponse> {
  const params = new URLSearchParams();
  if (query.skill) {
    params.set("skill", query.skill);
  }
  if (query.from) {
    params.set("from", query.from);
  }
  if (query.to) {
    params.set("to", query.to);
  }
  const suffix = params.size > 0 ? `?${params.toString()}` : "";
  return apiFetch<HeroChangesResponse>(`/api/v1/heroes/${encodeURIComponent(slug)}/changes${suffix}`);
}

type TimelineDateQuery = {
  from?: string;
  to?: string;
};

function buildDateQuerySuffix(query: TimelineDateQuery = {}) {
  const params = new URLSearchParams();
  if (query.from) {
    params.set("from", query.from);
  }
  if (query.to) {
    params.set("to", query.to);
  }
  return params.size > 0 ? `?${params.toString()}` : "";
}

export async function getItems(): Promise<ItemListResponse> {
  return apiFetch<ItemListResponse>("/api/v1/items");
}

export async function getItemChanges(slug: string, query: TimelineDateQuery = {}): Promise<ItemChangesResponse> {
  const suffix = buildDateQuerySuffix(query);
  return apiFetch<ItemChangesResponse>(`/api/v1/items/${encodeURIComponent(slug)}/changes${suffix}`);
}

export async function getSpells(): Promise<SpellListResponse> {
  return apiFetch<SpellListResponse>("/api/v1/spells");
}

export async function getSpellChanges(slug: string, query: TimelineDateQuery = {}): Promise<SpellChangesResponse> {
  const suffix = buildDateQuerySuffix(query);
  return apiFetch<SpellChangesResponse>(`/api/v1/spells/${encodeURIComponent(slug)}/changes${suffix}`);
}
