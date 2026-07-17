import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("next/navigation", () => ({ unstable_rethrow: vi.fn() }));

import { unstable_rethrow } from "next/navigation";
import {
  APIError,
  getHeroChanges,
  getItemChanges,
  getPatches,
  getSpellChanges,
} from "@/lib/api";

const fetchMock = vi.fn();

describe("frontend API requests", () => {
  beforeEach(() => {
    fetchMock.mockReset();
    vi.mocked(unstable_rethrow).mockReset();
    vi.stubGlobal("fetch", fetchMock);
  });

  it("builds pagination and encoded entity query URLs", async () => {
    fetchMock.mockResolvedValue(jsonResponse({ patches: [] }));
    await getPatches(3, 25);
    expect(requestPath(0)).toBe("/api/v1/patches?page=3&limit=25");

    fetchMock.mockResolvedValue(jsonResponse({ hero: {}, timeline: [] }));
    await getHeroChanges("hero/name", { skill: "A&B", from: "2026-07-01", to: "2026-07-02" });
    expect(requestPath(1)).toBe("/api/v1/heroes/hero%2Fname/changes?skill=A%26B&from=2026-07-01&to=2026-07-02");

    fetchMock.mockResolvedValue(jsonResponse({ item: {}, timeline: [] }));
    await getItemChanges("item/name", { from: "2026-07-01" });
    expect(requestPath(2)).toBe("/api/v1/items/item%2Fname/changes?from=2026-07-01");

    fetchMock.mockResolvedValue(jsonResponse({ spell: {}, timeline: [] }));
    await getSpellChanges("spell/name", { to: "2026-07-02" });
    expect(requestPath(3)).toBe("/api/v1/spells/spell%2Fname/changes?to=2026-07-02");
  });

  it("wraps a network failure in an APIError", async () => {
    fetchMock.mockRejectedValue(new Error("offline"));

    const error = await getPatches(1).catch((caught) => caught);

    expect(error).toBeInstanceOf(APIError);
    expect(error).toMatchObject({ status: 0 });
    expect(error.message).toContain("offline");
    expect(unstable_rethrow).toHaveBeenCalledOnce();
  });

  it("includes a structured API error without leaking more than the limit", async () => {
    fetchMock.mockResolvedValue(jsonResponse({ error: { message: "x".repeat(250) } }, 422));

    const error = await getPatches(1).catch((caught) => caught);

    expect(error).toBeInstanceOf(APIError);
    expect(error.status).toBe(422);
    expect(error.message).toContain("x".repeat(200));
    expect(error.message).not.toContain("x".repeat(201));
  });

  it("uses text details when error JSON decoding fails", async () => {
    fetchMock.mockResolvedValue({
      ok: false,
      status: 502,
      json: vi.fn().mockRejectedValue(new SyntaxError("invalid JSON")),
      text: vi.fn().mockResolvedValue("gateway unavailable"),
    } as unknown as Response);

    await expect(getPatches(1)).rejects.toMatchObject({
      status: 502,
      message: expect.stringContaining("gateway unavailable"),
    });
  });

  it("preserves Next.js control-flow exceptions", async () => {
    const controlFlowError = new Error("NEXT_REDIRECT");
    fetchMock.mockRejectedValue(controlFlowError);
    vi.mocked(unstable_rethrow).mockImplementationOnce((error) => {
      throw error;
    });

    await expect(getPatches(1)).rejects.toBe(controlFlowError);
  });
});

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function requestPath(callIndex: number) {
  const target = String(fetchMock.mock.calls[callIndex][0]);
  return new URL(target).pathname + new URL(target).search;
}
