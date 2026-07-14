import { NextRequest, NextResponse } from "next/server";
import {
  fetchAllowedImage,
  ImageProxyValidationError,
  parseAllowedImageURL,
  readAllowedImageResponse,
} from "@/lib/image-proxy";

const IMAGE_FETCH_TIMEOUT_MS = 10_000;

function createImageFetchSignal(request: NextRequest) {
  const controller = new AbortController();
  const abortFromRequest = () => controller.abort(request.signal.reason);
  if (request.signal.aborted) {
    abortFromRequest();
  } else {
    request.signal.addEventListener("abort", abortFromRequest, { once: true });
  }
  const timeout = setTimeout(() => controller.abort(), IMAGE_FETCH_TIMEOUT_MS);
  return {
    controller,
    cleanup: () => {
      clearTimeout(timeout);
      request.signal.removeEventListener("abort", abortFromRequest);
    },
  };
}

function proxyError(message: string, status: number) {
  return NextResponse.json(
    { error: message },
    { status, headers: { "Cache-Control": "no-store" } },
  );
}

function caughtProxyError(error: unknown, request: NextRequest, controller: AbortController) {
  if (error instanceof ImageProxyValidationError) {
    return proxyError(error.message, error.status);
  }
  if (controller.signal.aborted && !request.signal.aborted) {
    return proxyError("upstream image timed out", 504);
  }
  return proxyError("failed to fetch image", 502);
}

export async function GET(request: NextRequest) {
  const target = parseAllowedImageURL(request.nextUrl.searchParams.get("url"));
  if (!target) {
    return proxyError("invalid image url", 400);
  }

  const imageFetch = createImageFetchSignal(request);
  const { controller } = imageFetch;

  try {
    const upstream = await fetchAllowedImage(target, fetch, controller.signal);
    if (!upstream.ok || !upstream.body) {
      await upstream.body?.cancel().catch(() => undefined);
      return proxyError("upstream image unavailable", upstream.status || 502);
    }

    const { bytes, contentType } = await readAllowedImageResponse(upstream);
    const responseHeaders = new Headers();
    responseHeaders.set("Content-Type", contentType);
    responseHeaders.set("Content-Length", String(bytes.byteLength));
    responseHeaders.set("Cache-Control", "public, max-age=86400, stale-while-revalidate=604800");
    responseHeaders.set("Content-Security-Policy", "default-src 'none'; sandbox");
    responseHeaders.set("Cross-Origin-Resource-Policy", "same-origin");
    responseHeaders.set("X-Content-Type-Options", "nosniff");

    return new NextResponse(bytes, {
      status: 200,
      headers: responseHeaders,
    });
  } catch (error) {
    return caughtProxyError(error, request, controller);
  } finally {
    imageFetch.cleanup();
  }
}
