import { describe, expect, it, vi } from "vitest";
import { fetchAllowedImage, parseAllowedImageURL, readAllowedImageResponse } from "@/lib/image-proxy";

describe("parseAllowedImageURL", () => {
  it("accepts configured HTTPS image hosts", () => {
    expect(parseAllowedImageURL("https://clan.akamai.steamstatic.com/image.png")?.hostname).toBe(
      "clan.akamai.steamstatic.com"
    );
  });

  it("rejects other hosts and protocols", () => {
    expect(parseAllowedImageURL("https://example.com/image.png")).toBeNull();
    expect(parseAllowedImageURL("http://assets.deadlock-api.com/image.png")).toBeNull();
    expect(parseAllowedImageURL("https://assets.deadlock-api.com:8443/image.png")).toBeNull();
    expect(parseAllowedImageURL("https://user:secret@assets.deadlock-api.com/image.png")).toBeNull();
  });

  it("passes cancellation to every upstream request", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("image", { status: 200, headers: { "content-type": "image/png" } })
    );
    const controller = new AbortController();
    const target = parseAllowedImageURL("https://assets.deadlock-api.com/start.png");

    await fetchAllowedImage(target!, fetcher, controller.signal);

    expect(fetcher).toHaveBeenCalledWith(
      target!.toString(),
      expect.objectContaining({ signal: controller.signal })
    );
  });
});

describe("readAllowedImageResponse", () => {
  const pngBytes = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);

  it("rejects non-image content types", async () => {
    const response = new Response("<html>not an image</html>", {
      headers: { "content-type": "text/html" },
    });
    await expect(readAllowedImageResponse(response)).rejects.toThrow("not an allowed raster image");
  });

  it("rejects bodies that exceed the byte limit", async () => {
    const response = new Response("12345", { headers: { "content-type": "image/png" } });
    await expect(readAllowedImageResponse(response, 4)).rejects.toThrow("too large");
  });

  it("buffers an image within the byte limit", async () => {
    const response = new Response(pngBytes, { headers: { "content-type": "image/png; charset=binary" } });
    const result = await readAllowedImageResponse(response, pngBytes.byteLength);
    expect(result.contentType).toBe("image/png");
    expect(result.bytes.byteLength).toBe(pngBytes.byteLength);
  });

  it("rejects SVG and mismatched raster signatures", async () => {
    await expect(
      readAllowedImageResponse(new Response("<svg/>", { headers: { "content-type": "image/svg+xml" } })),
    ).rejects.toMatchObject({ status: 415 });
    await expect(
      readAllowedImageResponse(new Response("not a png", { headers: { "content-type": "image/png" } })),
    ).rejects.toThrow("signature");
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
