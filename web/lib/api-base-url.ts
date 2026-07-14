const DEFAULT_API_BASE_URL = "https://deadlockpatchnotes.com/api";

function normalizeBasePath(pathname: string) {
  const trimmed = pathname.replace(/\/+$/, "");
  if (trimmed === "" || trimmed === "/" || trimmed === "/api") {
    return "";
  }
  return trimmed;
}

export function resolveAPIBaseURL(rawValue: string | undefined) {
  const configuredValue = String(rawValue ?? "").trim();
  const candidate = configuredValue === "" ? DEFAULT_API_BASE_URL : configuredValue;

  try {
    const parsed = new URL(candidate);
    const usesHTTP = parsed.protocol === "http:" || parsed.protocol === "https:";
    if (!usesHTTP || parsed.hostname === "" || parsed.username !== "" || parsed.password !== "") {
      throw new Error("unsupported API URL");
    }
    const path = normalizeBasePath(parsed.pathname);
    return `${parsed.origin}${path}`;
  } catch {
    throw new Error(`Invalid API_BASE_URL: ${candidate}`);
  }
}
