import { NextRequest, NextResponse } from "next/server";
import { fetchAllowedImage, parseAllowedImageURL } from "@/lib/image-proxy";

export async function GET(request: NextRequest) {
  const target = parseAllowedImageURL(request.nextUrl.searchParams.get("url"));
  if (!target) {
    return NextResponse.json(
      { error: "invalid image url" },
      { status: 400, headers: { "Cache-Control": "no-store" } },
    );
  }

  try {
    const upstream = await fetchAllowedImage(target);
    if (!upstream.ok || !upstream.body) {
      return NextResponse.json(
        { error: "upstream image unavailable" },
        { status: upstream.status || 502, headers: { "Cache-Control": "no-store" } },
      );
    }

    const responseHeaders = new Headers();
    const contentType = upstream.headers.get("content-type") || "application/octet-stream";
    responseHeaders.set("Content-Type", contentType);
    responseHeaders.set("Cache-Control", "public, max-age=86400, stale-while-revalidate=604800");

    return new NextResponse(upstream.body, {
      status: 200,
      headers: responseHeaders,
    });
  } catch {
    return NextResponse.json(
      { error: "failed to fetch image" },
      { status: 502, headers: { "Cache-Control": "no-store" } },
    );
  }
}
