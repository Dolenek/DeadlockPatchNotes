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

const DEFAULT_API_BASE_URL = "https://deadlock.jakubdolenek.xyz/api";

function normalizeBasePath(pathname: string) {
  const trimmed = pathname.replace(/\/+$/, "");
  if (trimmed === "" || trimmed === "/") {
    return "";
  }
  if (trimmed === "/api") {
    return "";
  }
  return trimmed;
}

function resolveAPIBaseURL() {
  const candidate = (process.env.API_BASE_URL ?? DEFAULT_API_BASE_URL).trim();
  if (candidate === "") {
    return DEFAULT_API_BASE_URL;
  }

  try {
    const parsed = new URL(candidate);
    const path = normalizeBasePath(parsed.pathname);
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
    let detail = "";
    try {
      const payload = (await response.json()) as { error?: { message?: string } };
      detail = payload?.error?.message ? String(payload.error.message) : "";
    } catch {
      detail = await response.text().catch(() => "");
    }

    const suffix = detail ? ` ${detail.slice(0, 200)}` : "";
    throw new APIError(response.status, `API request failed for ${target}: ${response.status}${suffix}`);
  }

  return (await response.json()) as T;
}

function normalizeTimelineSummary(raw: any) {
  return {
    id: String(raw?.id ?? ""),
    releaseType: String(raw?.releaseType ?? raw?.kind ?? ""),
    title: String(raw?.title ?? ""),
    releasedAt: String(raw?.releasedAt ?? "")
  };
}

function normalizePatchSource(raw: any) {
  if (raw && typeof raw === "object") {
    return {
      type: String(raw.type ?? ""),
      url: String(raw.url ?? "")
    };
  }
  return { type: "", url: "" };
}

function normalizePatchSummary(raw: any) {
  return {
    id: String(raw?.id ?? ""),
    slug: String(raw?.slug ?? ""),
    title: String(raw?.title ?? ""),
    publishedAt: String(raw?.publishedAt ?? ""),
    category: String(raw?.category ?? ""),
    imageUrl: String(raw?.imageUrl ?? raw?.coverImageUrl ?? ""),
    source: raw?.source ? normalizePatchSource(raw.source) : normalizePatchSource({ type: "forum", url: raw?.sourceUrl ?? "" }),
    releaseTimeline: Array.isArray(raw?.releaseTimeline)
      ? raw.releaseTimeline.map(normalizeTimelineSummary)
      : Array.isArray(raw?.timeline)
        ? raw.timeline.map(normalizeTimelineSummary)
        : []
  };
}

function normalizePatchTimelineBlock(raw: any) {
  return {
    id: String(raw?.id ?? ""),
    releaseType: String(raw?.releaseType ?? raw?.kind ?? ""),
    title: String(raw?.title ?? ""),
    releasedAt: String(raw?.releasedAt ?? ""),
    source: normalizePatchSource(raw?.source),
    changes: Array.isArray(raw?.changes) ? raw.changes : [],
    sections: Array.isArray(raw?.sections) ? raw.sections : undefined
  };
}

function normalizePatchListResponse(raw: any): PatchListResponse {
  const patches = Array.isArray(raw?.patches) ? raw.patches.map(normalizePatchSummary) : Array.isArray(raw?.items) ? raw.items.map(normalizePatchSummary) : [];
  const pagination = raw?.pagination && typeof raw.pagination === "object"
    ? {
        page: Number(raw.pagination.page ?? 1),
        pageSize: Number(raw.pagination.pageSize ?? 12),
        totalItems: Number(raw.pagination.totalItems ?? patches.length),
        totalPages: Number(raw.pagination.totalPages ?? 1)
      }
    : {
        page: Number(raw?.page ?? 1),
        pageSize: Number(raw?.pageSize ?? raw?.limit ?? 12),
        totalItems: Number(raw?.totalItems ?? raw?.total ?? patches.length),
        totalPages: Number(raw?.totalPages ?? 1)
      };

  return { patches, pagination };
}

function normalizePatchDetail(raw: any): PatchDetail {
  return {
    id: String(raw?.id ?? ""),
    slug: String(raw?.slug ?? ""),
    title: String(raw?.title ?? ""),
    publishedAt: String(raw?.publishedAt ?? ""),
    category: String(raw?.category ?? ""),
    source: normalizePatchSource(raw?.source),
    imageUrl: String(raw?.imageUrl ?? raw?.heroImageUrl ?? ""),
    intro: String(raw?.intro ?? ""),
    sections: Array.isArray(raw?.sections) ? raw.sections : [],
    releaseTimeline: Array.isArray(raw?.releaseTimeline)
      ? raw.releaseTimeline.map(normalizePatchTimelineBlock)
      : Array.isArray(raw?.timeline)
        ? raw.timeline.map(normalizePatchTimelineBlock)
        : undefined
  };
}

function normalizeHeroListResponse(raw: any): HeroListResponse {
  return {
    heroes: Array.isArray(raw?.heroes) ? raw.heroes : Array.isArray(raw?.items) ? raw.items : []
  };
}

function normalizeItemListResponse(raw: any): ItemListResponse {
  return {
    items: Array.isArray(raw?.items) ? raw.items : []
  };
}

function normalizeSpellListResponse(raw: any): SpellListResponse {
  return {
    spells: Array.isArray(raw?.spells) ? raw.spells : Array.isArray(raw?.items) ? raw.items : []
  };
}

function normalizePatchRef(raw: any) {
  return {
    slug: String(raw?.slug ?? ""),
    title: String(raw?.title ?? "")
  };
}

function normalizeHeroChangesResponse(raw: any): HeroChangesResponse {
  const blocks = Array.isArray(raw?.timeline) ? raw.timeline : Array.isArray(raw?.items) ? raw.items : [];
  return {
    hero: raw?.hero,
    timeline: blocks.map((block: any) => ({
      ...block,
      releaseType: String(block?.releaseType ?? block?.kind ?? ""),
      displayLabel: String(block?.displayLabel ?? block?.label ?? ""),
      patchRef: normalizePatchRef(block?.patchRef ?? block?.patch)
    }))
  };
}

function normalizeItemChangesResponse(raw: any): ItemChangesResponse {
  const blocks = Array.isArray(raw?.timeline) ? raw.timeline : Array.isArray(raw?.items) ? raw.items : [];
  return {
    item: raw?.item,
    timeline: blocks.map((block: any) => ({
      ...block,
      releaseType: String(block?.releaseType ?? block?.kind ?? ""),
      displayLabel: String(block?.displayLabel ?? block?.label ?? ""),
      patchRef: normalizePatchRef(block?.patchRef ?? block?.patch)
    }))
  };
}

function normalizeSpellChangesResponse(raw: any): SpellChangesResponse {
  const blocks = Array.isArray(raw?.timeline) ? raw.timeline : Array.isArray(raw?.items) ? raw.items : [];
  return {
    spell: raw?.spell,
    timeline: blocks.map((block: any) => ({
      ...block,
      releaseType: String(block?.releaseType ?? block?.kind ?? ""),
      displayLabel: String(block?.displayLabel ?? block?.label ?? ""),
      patchRef: normalizePatchRef(block?.patchRef ?? block?.patch)
    }))
  };
}

export async function getPatches(page: number, limit = 12): Promise<PatchListResponse> {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  const payload = await apiFetch<any>(`/api/v1/patches?${params.toString()}`);
  return normalizePatchListResponse(payload);
}

export async function getPatchBySlug(slug: string): Promise<PatchDetail> {
  const payload = await apiFetch<any>(`/api/v1/patches/${slug}`);
  return normalizePatchDetail(payload);
}

type HeroChangesQuery = {
  skill?: string;
  from?: string;
  to?: string;
};

export async function getHeroes(): Promise<HeroListResponse> {
  const payload = await apiFetch<any>("/api/v1/heroes");
  return normalizeHeroListResponse(payload);
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
  const payload = await apiFetch<any>(`/api/v1/heroes/${encodeURIComponent(slug)}/changes${suffix}`);
  return normalizeHeroChangesResponse(payload);
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
  const payload = await apiFetch<any>("/api/v1/items");
  return normalizeItemListResponse(payload);
}

export async function getItemChanges(slug: string, query: TimelineDateQuery = {}): Promise<ItemChangesResponse> {
  const suffix = buildDateQuerySuffix(query);
  const payload = await apiFetch<any>(`/api/v1/items/${encodeURIComponent(slug)}/changes${suffix}`);
  return normalizeItemChangesResponse(payload);
}

export async function getSpells(): Promise<SpellListResponse> {
  const payload = await apiFetch<any>("/api/v1/spells");
  return normalizeSpellListResponse(payload);
}

export async function getSpellChanges(slug: string, query: TimelineDateQuery = {}): Promise<SpellChangesResponse> {
  const suffix = buildDateQuerySuffix(query);
  const payload = await apiFetch<any>(`/api/v1/spells/${encodeURIComponent(slug)}/changes${suffix}`);
  return normalizeSpellChangesResponse(payload);
}
