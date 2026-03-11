"use client";

import { useEffect } from "react";

type RGB = {
  r: number;
  g: number;
  b: number;
};

type ItemGradientFromImageProps = {
  targetID: string;
  src?: string;
  fallbackSrc?: string;
};

function colorDistance(left: RGB, right: RGB) {
  const dr = left.r - right.r;
  const dg = left.g - right.g;
  const db = left.b - right.b;
  return Math.sqrt(dr * dr + dg * dg + db * db);
}

function clampByte(value: number) {
  return Math.max(0, Math.min(255, Math.round(value)));
}

function mixColor(from: RGB, to: RGB, weight: number): RGB {
  return {
    r: clampByte(from.r + (to.r - from.r) * weight),
    g: clampByte(from.g + (to.g - from.g) * weight),
    b: clampByte(from.b + (to.b - from.b) * weight),
  };
}

function rgba(color: RGB, alpha: number) {
  return `rgba(${color.r}, ${color.g}, ${color.b}, ${alpha})`;
}

function rgb(color: RGB) {
  return `rgb(${color.r}, ${color.g}, ${color.b})`;
}

function deriveSecondary(primary: RGB): RGB {
  const shifted = {
    r: primary.b,
    g: primary.r,
    b: primary.g,
  };
  return mixColor(shifted, { r: 90, g: 120, b: 170 }, 0.35);
}

function normalizeForBackground(color: RGB): RGB {
  const softened = mixColor(color, { r: 255, g: 255, b: 255 }, 0.06);
  return mixColor(softened, { r: 0, g: 0, b: 0 }, 0.18);
}

async function loadImage(source: string): Promise<HTMLImageElement> {
  const image = new Image();
  image.decoding = "async";

  return new Promise((resolve, reject) => {
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error(`Failed to load ${source}`));
    image.src = source;
  });
}

function resolveSampleSource(source: string) {
  if (source.startsWith("/") || source.startsWith("data:")) {
    return source;
  }

  try {
    const parsed = new URL(source);
    if (parsed.protocol === "http:" || parsed.protocol === "https:") {
      return `/api/image-proxy?url=${encodeURIComponent(parsed.toString())}`;
    }
  } catch {
    return source;
  }

  return source;
}

async function extractPalette(source: string): Promise<{ primary: RGB; secondary: RGB }> {
  const image = await loadImage(source);
  const canvas = document.createElement("canvas");
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context) {
    throw new Error("No 2D context");
  }

  const sampleWidth = 48;
  const sampleHeight = 48;
  canvas.width = sampleWidth;
  canvas.height = sampleHeight;
  context.drawImage(image, 0, 0, sampleWidth, sampleHeight);

  const { data } = context.getImageData(0, 0, sampleWidth, sampleHeight);
  const buckets = new Map<number, { count: number; r: number; g: number; b: number }>();

  for (let i = 0; i < data.length; i += 16) {
    const r = data[i];
    const g = data[i + 1];
    const b = data[i + 2];
    const a = data[i + 3];
    if (a < 32) {
      continue;
    }

    const key = ((r >> 4) << 8) | ((g >> 4) << 4) | (b >> 4);
    const current = buckets.get(key);
    if (current) {
      current.count += 1;
      current.r += r;
      current.g += g;
      current.b += b;
      continue;
    }

    buckets.set(key, { count: 1, r, g, b });
  }

  if (buckets.size === 0) {
    throw new Error("No color samples");
  }

  const ordered = [...buckets.values()]
    .sort((left, right) => right.count - left.count)
    .map((entry) => ({
      r: clampByte(entry.r / entry.count),
      g: clampByte(entry.g / entry.count),
      b: clampByte(entry.b / entry.count),
    }));

  const primary = ordered[0];
  const secondary = ordered.find((candidate) => colorDistance(candidate, primary) > 70) ?? ordered[1] ?? deriveSecondary(primary);

  return { primary: normalizeForBackground(primary), secondary: normalizeForBackground(secondary) };
}

function setThemeVariables(target: HTMLElement, primary: RGB, secondary: RGB) {
  const cardBorder = mixColor(primary, { r: 200, g: 220, b: 245 }, 0.42);
  const link = mixColor(primary, { r: 245, g: 213, b: 158 }, 0.62);
  target.style.setProperty("--item-gradient-accent-1", rgba(primary, 0.27));
  target.style.setProperty("--item-gradient-accent-2", rgba(secondary, 0.23));
  target.style.setProperty("--item-card-border", rgba(cardBorder, 0.44));
  target.style.setProperty("--item-meta-link", rgb(link));
}

function clearThemeVariables(target: HTMLElement) {
  target.style.removeProperty("--item-gradient-accent-1");
  target.style.removeProperty("--item-gradient-accent-2");
  target.style.removeProperty("--item-card-border");
  target.style.removeProperty("--item-meta-link");
}

export function ItemGradientFromImage({ targetID, src, fallbackSrc }: ItemGradientFromImageProps) {
  useEffect(() => {
    const target = document.getElementById(targetID);
    if (!target) {
      return;
    }

    const candidates = [...new Set([src, fallbackSrc].filter(Boolean))] as string[];
    if (candidates.length === 0) {
      clearThemeVariables(target);
      return;
    }

    let cancelled = false;

    const applyFromCandidates = async () => {
      for (const raw of candidates) {
        try {
          const sampled = resolveSampleSource(raw);
          const palette = await extractPalette(sampled);
          if (cancelled) {
            return;
          }
          setThemeVariables(target, palette.primary, palette.secondary);
          return;
        } catch {
          // Try the next candidate image source.
        }
      }

      if (!cancelled) {
        clearThemeVariables(target);
      }
    };

    applyFromCandidates();

    return () => {
      cancelled = true;
      clearThemeVariables(target);
    };
  }, [fallbackSrc, src, targetID]);

  return null;
}
