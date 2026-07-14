export function resolveFallbackImageSource(currentSrc: string | undefined, fallbackSrc: string | undefined) {
  if (fallbackSrc && currentSrc !== fallbackSrc) {
    return fallbackSrc;
  }

  return currentSrc;
}
