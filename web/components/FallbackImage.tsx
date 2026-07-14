"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { resolveFallbackImageSource } from "@/lib/fallback-image";

type FallbackImageProps = {
  src?: string;
  fallbackSrc?: string;
  alt: string;
  className?: string;
  loading?: "lazy" | "eager";
  decoding?: "async" | "auto" | "sync";
  fetchPriority?: "high" | "low" | "auto";
  width: number;
  height: number;
  sizes?: string;
  quality?: number;
};

export function FallbackImage({
  src,
  fallbackSrc,
  alt,
  className,
  loading = "lazy",
  decoding = "async",
  fetchPriority = "auto",
  width, height, sizes, quality = 50,
}: FallbackImageProps) {
  const [currentSrc, setCurrentSrc] = useState<string | undefined>(src);

  useEffect(() => {
    setCurrentSrc(src);
  }, [src]);

  const resolvedSrc = currentSrc ?? fallbackSrc;
  if (!resolvedSrc) {
    return null;
  }

  return (
    <Image
      src={resolvedSrc}
      alt={alt}
      className={className}
      loading={loading}
      decoding={decoding}
      fetchPriority={fetchPriority}
      width={width}
      height={height}
      sizes={sizes}
      quality={quality}
      onError={() => {
        setCurrentSrc((failedSrc) => resolveFallbackImageSource(failedSrc, fallbackSrc));
      }}
    />
  );
}
