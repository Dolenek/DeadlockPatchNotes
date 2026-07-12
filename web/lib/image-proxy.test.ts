import { describe, expect, it, vi } from "vitest";
import { fetchAllowedImage, parseAllowedImageURL } from "@/lib/image-proxy";

describe("parseAllowedImageURL", () => {
  it("accepts configured HTTPS image hosts", () => {
    expect(parseAllowedImageURL("https://clan.akamai.steamstatic.com/image.png")?.hostname).toBe(
      "clan.akamai.steamstatic.com"
    );
  });

  it("rejects other hosts and protocols", () => {
    expect(parseAllowedImageURL("https://example.com/image.png")).toBeNull();
    expect(parseAllowedImageURL("http://assets.deadlock-api.com/image.png")).toBeNull();
  });
});

describe("fetchAllowedImage", () => {
  it("follows an allowed relative redirect manually", async () => {
    const fetcher = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(new Response(null, { status: 302, headers: { location: "/final.png" } }))
      .mockResolvedValueOnce(new Response("image", { status: 200 }));
    const target = parseAllowedImageURL("https://assets.deadlock-api.com/start.png");

    const response = await fetchAllowedImage(target!, fetcher);

    expect(response.status).toBe(200);
    expect(fetcher).toHaveBeenNthCalledWith(
      2,
      "https://assets.deadlock-api.com/final.png",
      expect.objectContaining({ redirect: "manual" })
    );
  });

  it("rejects redirects outside the allowlist", async () => {
    const fetcher = vi
      .fn<typeof fetch>()
      .mockResolvedValue(new Response(null, { status: 302, headers: { location: "https://example.com/private" } }));
    const target = parseAllowedImageURL("https://assets.deadlock-api.com/start.png");

    await expect(fetchAllowedImage(target!, fetcher)).rejects.toThrow("not allowed");
    expect(fetcher).toHaveBeenCalledTimes(1);
  });
});
