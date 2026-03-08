"use client";

import { useEffect, useState } from "react";

type FallbackImageProps = {
  src?: string;
  fallbackSrc?: string;
  alt: string;
  className?: string;
};

export function FallbackImage({ src, fallbackSrc, alt, className }: FallbackImageProps) {
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
      loading="lazy"
      onError={() => {
        if (fallbackSrc && currentSrc !== fallbackSrc) {
          setCurrentSrc(fallbackSrc);
        }
      }}
    />
  );
}
