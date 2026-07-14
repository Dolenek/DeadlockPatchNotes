"use client";

import { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { PrefetchKind } from "next/dist/client/components/router-reducer/router-reducer-types";
import { markIntentPrefetch, resolveIntentPrefetchHref } from "@/lib/intent-prefetch";

function findIntentLink(eventTarget: EventTarget | null) {
  return eventTarget instanceof Element
    ? eventTarget.closest<HTMLAnchorElement>('a[data-prefetch="intent"]')
    : null;
}

export function IntentPrefetchManager() {
  const router = useRouter();
  const prefetchedHrefs = useRef(new Set<string>());

  useEffect(() => {
    const prefetchLink = (eventTarget: EventTarget | null) => {
      const anchor = findIntentLink(eventTarget);
      const href = anchor ? resolveIntentPrefetchHref(anchor.href, window.location.href) : null;
      if (!href || !markIntentPrefetch(prefetchedHrefs.current, href)) {
        return;
      }

      router.prefetch(href, {
        kind: PrefetchKind.AUTO,
        onInvalidate: () => {
          prefetchedHrefs.current.delete(href);
        },
      });
    };

    const handlePointerOver = (event: PointerEvent) => {
      if (event.pointerType !== "touch") {
        prefetchLink(event.target);
      }
    };
    const handleFocusIn = (event: FocusEvent) => prefetchLink(event.target);

    document.addEventListener("pointerover", handlePointerOver, true);
    document.addEventListener("focusin", handleFocusIn, true);
    return () => {
      document.removeEventListener("pointerover", handlePointerOver, true);
      document.removeEventListener("focusin", handleFocusIn, true);
    };
  }, [router]);

  return null;
}
