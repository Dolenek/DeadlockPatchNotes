"use client";

import { useEffect, useRef } from "react";

type ItemAbstractPatternProps = {
  blobCount?: number;
};

function randomInRange(min: number, max: number) {
  return min + Math.random() * (max - min);
}

export function ItemAbstractPattern({ blobCount = 9 }: ItemAbstractPatternProps) {
  const blobRefs = useRef<Array<HTMLSpanElement | null>>([]);

  useEffect(() => {
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | null = null;

    const blobs = blobRefs.current.filter((blob): blob is HTMLSpanElement => blob instanceof HTMLSpanElement);
    if (blobs.length === 0) {
      return;
    }

    const applyRandomState = (instant: boolean) => {
      for (const blob of blobs) {
        const x = randomInRange(4, 96);
        const y = randomInRange(6, 94);
        const scale = randomInRange(0.58, 1.75);
        const rotation = randomInRange(0, 360);
        const opacity = randomInRange(0.12, 0.34);
        const blurPx = randomInRange(26, 76);
        const durationMs = randomInRange(3600, 7800);

        blob.style.transitionDuration = instant ? "0ms" : `${Math.round(durationMs)}ms`;
        blob.style.setProperty("--item-blob-x", `${x}%`);
        blob.style.setProperty("--item-blob-y", `${y}%`);
        blob.style.setProperty("--item-blob-scale", String(scale));
        blob.style.setProperty("--item-blob-rot", `${rotation}deg`);
        blob.style.setProperty("--item-blob-opacity", String(opacity));
        blob.style.setProperty("--item-blob-blur", `${Math.round(blurPx)}px`);
      }
    };

    const tick = () => {
      if (cancelled) {
        return;
      }
      applyRandomState(false);
      timer = setTimeout(tick, Math.round(randomInRange(2600, 5400)));
    };

    applyRandomState(true);
    timer = setTimeout(tick, 140);

    return () => {
      cancelled = true;
      if (timer) {
        clearTimeout(timer);
      }
    };
  }, []);

  return (
    <div className="item-abstract-field" aria-hidden="true">
      {Array.from({ length: blobCount }).map((_, index) => (
        <span
          className={`item-abstract-blob item-abstract-blob--tone-${index % 3}`}
          key={`blob-${index}`}
          ref={(node) => {
            blobRefs.current[index] = node;
          }}
        />
      ))}
    </div>
  );
}
