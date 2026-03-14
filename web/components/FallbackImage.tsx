"use client";

import { useEffect, useState } from "react";

type FallbackImageProps = {
  src?: string;
  fallbackSrc?: string;
  alt: string;
  className?: string;
  loading?: "lazy" | "eager";
  decoding?: "async" | "auto" | "sync";
  fetchPriority?: "high" | "low" | "auto";
  width?: number;
  height?: number;
};

export function FallbackImage({
  src,
  fallbackSrc,
  alt,
  className,
  loading = "lazy",
  decoding = "async",
  fetchPriority = "auto",
  width,
  height,
}: FallbackImageProps) {
  const [currentSrc, setCurrentSrc] = useState<string | undefined>(src);

  useEffect(() => {
    setCurrentSrc(src);
  }, [src]);

  if (!currentSrc && !fallbackSrc) {
    return null;
  }

  return (
    <img
      src={currentSrc ?? fallbackSrc}
      alt={alt}
      className={className}
      loading={loading}
      decoding={decoding}
      fetchPriority={fetchPriority}
      width={width}
      height={height}
      onError={() => {
        if (fallbackSrc && currentSrc !== fallbackSrc) {
          setCurrentSrc(fallbackSrc);
        }
      }}
    />
  );
}
