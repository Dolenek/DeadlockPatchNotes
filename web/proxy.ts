import { NextRequest, NextResponse } from "next/server";
import {
  buildHTTPSRedirectURL,
  shouldBypassHTTPSRedirect,
  shouldRedirectToHTTPS,
} from "@/lib/https-redirect";

export function proxy(request: NextRequest) {
  if (
    shouldBypassHTTPSRedirect(request.nextUrl.pathname) ||
    !shouldRedirectToHTTPS(process.env.NODE_ENV, request.headers.get("x-forwarded-proto"))
  ) {
    return NextResponse.next();
  }

  const redirectURL = buildHTTPSRedirectURL(
    process.env.SITE_URL,
    request.nextUrl.pathname,
    request.nextUrl.search,
  );
  return NextResponse.redirect(redirectURL, 308);
}

export const config = {
  matcher: ["/:path*"],
};
