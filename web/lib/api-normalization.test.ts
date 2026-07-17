import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("next/navigation", () => ({ unstable_rethrow: vi.fn() }));

import { getHeroChanges, getHeroes, getPatchBySlug, getPatches } from "@/lib/api";

const fetchMock = vi.fn();
const mirroredIcon = "https://assets-bucket.deadlock-api.com/assets-api-res/images/abilities/archer/archer_charged_shot.png";

describe("frontend API normalization", () => {
  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal("fetch", fetchMock);
  });

  it("normalizes legacy patch list fields and mirrored assets", async () => {
    fetchMock.mockResolvedValue(response({
      items: [{
        id: 7,
        slug: "patch",
        title: "Patch",
        coverImageUrl: mirroredIcon,
        sourceUrl: "https://forum.test/patch",
        timeline: [{ id: "one", kind: "initial", title: "Initial", releasedAt: "2026-07-01" }],
      }],
      page: "2",
      limit: "5",
      total: "6",
      totalPages: "2",
    }));

    const payload = await getPatches(2, 5);

    expect(payload.pagination).toEqual({ page: 2, pageSize: 5, totalItems: 6, totalPages: 2 });
    expect(payload.patches[0]).toMatchObject({
      id: "7",
      imageUrl: "/assets/mirror/icons/archer-charged-shot-cce123bde7ef.png",
      source: { type: "forum", url: "https://forum.test/patch" },
      releaseTimeline: [{ id: "one", releaseType: "initial" }],
    });
  });

  it("normalizes legacy patch detail and malformed collections safely", async () => {
    fetchMock.mockResolvedValue(response({
      id: 9,
      slug: "detail",
      heroImageUrl: "/local.png",
      sections: "invalid",
      timeline: [{
        id: "block",
        kind: "hotfix",
        changes: "invalid",
        sections: [{ id: "general", title: "General", kind: "general", entries: "invalid" }],
      }],
    }));

    const payload = await getPatchBySlug("detail");

    expect(payload.id).toBe("9");
    expect(payload.imageUrl).toBe("/local.png");
    expect(payload.sections).toEqual([]);
    expect(payload.releaseTimeline?.[0]).toMatchObject({ releaseType: "hotfix", changes: [] });
    expect(payload.releaseTimeline?.[0].sections?.[0].entries).toEqual([]);
  });

  it("supports legacy hero list and timeline aliases", async () => {
    fetchMock.mockResolvedValueOnce(response({
      items: [{ slug: "hero", name: "Hero", iconUrl: mirroredIcon, lastChangedAt: "2026-07-01" }],
    }));
    const heroes = await getHeroes();
    expect(heroes.heroes[0].iconUrl).toBe("/assets/mirror/icons/archer-charged-shot-cce123bde7ef.png");

    fetchMock.mockResolvedValueOnce(response({
      hero: { slug: "hero", name: "Hero" },
      items: [{ id: 1, kind: "initial", label: "Update", patch: { slug: "p", title: "P" }, skills: "invalid" }],
    }));
    const changes = await getHeroChanges("hero");
    expect(changes.timeline[0]).toMatchObject({
      id: "1",
      releaseType: "initial",
      displayLabel: "Update",
      patchRef: { slug: "p", title: "P" },
      skills: [],
    });
  });

  it("falls back to empty lists for malformed list payloads", async () => {
    fetchMock.mockResolvedValue(response({ patches: "invalid", pagination: null }));

    const payload = await getPatches(1);

    expect(payload.patches).toEqual([]);
    expect(payload.pagination).toEqual({ page: 1, pageSize: 12, totalItems: 0, totalPages: 1 });
  });
});

function response(payload: unknown) {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
