import { describe, expect, it } from "vitest";
import { resolveDecorativeImageLoading } from "@/lib/image-loading";

describe("resolveDecorativeImageLoading", () => {
  it("loads decorative layers lazily by default", () => {
    expect(resolveDecorativeImageLoading({})).toEqual({
      preload: false,
      loading: "lazy",
      fetchPriority: undefined,
    });
  });

  it("removes conflicting loading hints from a preloaded image", () => {
    expect(resolveDecorativeImageLoading({ preload: true, loading: "lazy", fetchPriority: "low" })).toEqual({
      preload: true,
      loading: undefined,
      fetchPriority: undefined,
    });
  });

  it("preserves explicit eager loading without preloading", () => {
    expect(resolveDecorativeImageLoading({ loading: "eager", fetchPriority: "auto" })).toEqual({
      preload: false,
      loading: "eager",
      fetchPriority: "auto",
    });
  });
});
