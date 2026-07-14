import type { NextConfig } from "next";
import { WEB_SECURITY_HEADERS } from "./lib/security-headers";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  output: "standalone",
  outputFileTracingIncludes: {
    "/*": ["./node_modules/sharp/**/*", "./node_modules/@img/**/*"],
  },
  poweredByHeader: false,
  async headers() {
    return [
      {
        source: "/assets/mirror/:path(.*-[0-9a-f]{12}\\.[^/]+)",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
      {
        source: "/(.*)",
        headers: [...WEB_SECURITY_HEADERS],
      },
    ];
  },
  images: {
    localPatterns: [{ pathname: "/**", search: "" }],
    remotePatterns: [
      {
        protocol: "https",
        hostname: "clan.akamai.steamstatic.com"
      },
      {
        protocol: "https",
        hostname: "clan.fastly.steamstatic.com"
      },
      {
        protocol: "https",
        hostname: "shared.fastly.steamstatic.com"
      },
      {
        protocol: "https",
        hostname: "assets-bucket.deadlock-api.com"
      },
      {
        protocol: "https",
        hostname: "assets.deadlock-api.com"
      }
    ],
    formats: ["image/avif", "image/webp"],
    deviceSizes: [640, 750, 828, 1080, 1200, 1536, 1920, 2560],
    imageSizes: [32, 37, 48, 53, 64, 74, 96, 128, 160, 192, 256, 320],
    qualities: [45, 50, 54, 56, 58, 60, 68, 74, 75],
    minimumCacheTTL: 86400,
    maximumResponseBody: 8 * 1024 * 1024,
    maximumRedirects: 2,
    dangerouslyAllowLocalIP: false,
  }
};

export default nextConfig;
