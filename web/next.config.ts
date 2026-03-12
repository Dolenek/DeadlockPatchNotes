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
    formats: ["image/avif", "image/webp"],
    minimumCacheTTL: 86400
  }
};

export default nextConfig;
