import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  images: {
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
    formats: ["image/webp"],
    deviceSizes: [640, 750, 828, 1080, 1200, 1536, 1920],
    imageSizes: [48, 64, 74, 96, 128, 160],
    qualities: [54, 56, 58, 60, 68, 74, 75],
    minimumCacheTTL: 86400
  }
};

export default nextConfig;
