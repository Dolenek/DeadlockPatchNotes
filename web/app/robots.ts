import type { MetadataRoute } from "next";
import { SEO_BASE_URL, buildAbsoluteURL } from "@/lib/seo";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
        disallow: ["/api", "/api/", "/image-proxy"],
      },
    ],
    sitemap: buildAbsoluteURL("/sitemap.xml"),
    host: SEO_BASE_URL,
  };
}
