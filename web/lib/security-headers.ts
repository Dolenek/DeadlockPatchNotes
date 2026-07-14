const contentSecurityPolicy = [
  "default-src 'self'",
  "base-uri 'self'",
  "form-action 'self'",
  "frame-ancestors 'none'",
  "object-src 'none'",
  "script-src 'self' 'unsafe-inline'",
  "style-src 'self' 'unsafe-inline'",
  "font-src 'self' data:",
  [
    "img-src 'self' data: blob:",
    "https://assets-bucket.deadlock-api.com",
    "https://assets.deadlock-api.com",
    "https://clan.akamai.steamstatic.com",
    "https://clan.fastly.steamstatic.com",
    "https://shared.fastly.steamstatic.com",
  ].join(" "),
  "connect-src 'self'",
].join("; ");

export const WEB_SECURITY_HEADERS = [
  { key: "Content-Security-Policy", value: contentSecurityPolicy },
  { key: "Cross-Origin-Opener-Policy", value: "same-origin" },
  { key: "Permissions-Policy", value: "camera=(), geolocation=(), microphone=(), payment=(), usb=()" },
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  { key: "Strict-Transport-Security", value: "max-age=31536000; includeSubDomains" },
  { key: "X-Content-Type-Options", value: "nosniff" },
  { key: "X-Frame-Options", value: "DENY" },
] as const;
