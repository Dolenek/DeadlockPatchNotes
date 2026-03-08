import { PatchDetail, PatchListResponse } from "@/lib/types";

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
  const response = await fetch(`${API_BASE_URL}${path}`, {
    next: { revalidate: 30 }
  });

  if (!response.ok) {
    throw new APIError(response.status, `API request failed: ${response.status}`);
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
