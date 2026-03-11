import { NextRequest, NextResponse } from "next/server";

const ALLOWED_IMAGE_HOSTS = new Set([
  "assets-bucket.deadlock-api.com",
  "assets.deadlock-api.com",
  "clan.fastly.steamstatic.com",
  "shared.fastly.steamstatic.com",
]);

const REQUEST_HEADERS = {
  "User-Agent": "deadlockpatchnotes-web-image-proxy/1.0",
};

function parseExternalURL(raw: string | null) {
  if (!raw) {
    return null;
  }

  try {
    const parsed = new URL(raw);
    if (parsed.protocol !== "https:") {
      return null;
    }
    if (!ALLOWED_IMAGE_HOSTS.has(parsed.hostname)) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

export async function GET(request: NextRequest) {
  const target = parseExternalURL(request.nextUrl.searchParams.get("url"));
  if (!target) {
    return NextResponse.json(
      { error: "invalid image url" },
      { status: 400, headers: { "Cache-Control": "no-store" } },
    );
  }

  try {
    const upstream = await fetch(target.toString(), { headers: REQUEST_HEADERS, redirect: "follow" });
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
