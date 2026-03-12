"use client";

import { useEffect, useMemo, useState } from "react";
import { FallbackImage } from "@/components/FallbackImage";

export type PatchHeroesRailHeroLink = {
  id: string;
  label: string;
  targetId: string;
  iconUrl?: string;
  iconFallbackUrl?: string;
};

export type PatchHeroesRailBlock = {
  id: string;
  label: string;
  heroes: PatchHeroesRailHeroLink[];
};

type PatchHeroesRailProps = {
  blocks: PatchHeroesRailBlock[];
};

const ACTIVE_OFFSET_PX = 120;

export function PatchHeroesRail({ blocks }: PatchHeroesRailProps) {
  const [activeBlockID, setActiveBlockID] = useState("");
  const blockOrder = useMemo(() => blocks.map((block) => block.id), [blocks]);
  const blockByID = useMemo(() => new Map(blocks.map((block) => [block.id, block])), [blocks]);
  const targetToBlock = useMemo(() => {
    const lookup = new Map<string, string>();
    for (const block of blocks) {
      for (const hero of block.heroes) {
        lookup.set(hero.targetId, block.id);
      }
    }
    return lookup;
  }, [blocks]);

  useEffect(() => {
    if (blockOrder.length === 0) {
      return;
    }

    setActiveBlockID((current) => (current !== "" && blockOrder.includes(current) ? current : blockOrder[0]));
  }, [blockOrder]);

  useEffect(() => {
    if (typeof window === "undefined" || blockOrder.length === 0) {
      return;
    }

    const syncFromHash = () => {
      const targetID = decodeURIComponent(window.location.hash.replace(/^#/, ""));
      if (targetID === "") {
        return;
      }
      if (blockOrder.includes(targetID)) {
        setActiveBlockID(targetID);
        return;
      }

      const mappedBlock = targetToBlock.get(targetID);
      if (mappedBlock) {
        setActiveBlockID(mappedBlock);
      }
    };

    syncFromHash();
    window.addEventListener("hashchange", syncFromHash);
    return () => {
      window.removeEventListener("hashchange", syncFromHash);
    };
  }, [blockOrder, targetToBlock]);

  useEffect(() => {
    if (typeof window === "undefined" || blockOrder.length === 0) {
      return;
    }

    const visibleBlocks = new Map<string, IntersectionObserverEntry>();

    const resolveActiveID = () => {
      const visibleInOrder = blockOrder.filter((id) => visibleBlocks.has(id));
      if (visibleInOrder.length === 0) {
        return "";
      }

      let activeBlock = visibleInOrder[0];
      for (const id of visibleInOrder) {
        const entry = visibleBlocks.get(id);
        if (!entry) {
          continue;
        }
        if (entry.boundingClientRect.top <= ACTIVE_OFFSET_PX) {
          activeBlock = id;
          continue;
        }
        break;
      }
      return activeBlock;
    };

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          const id = entry.target.id;
          if (entry.isIntersecting) {
            visibleBlocks.set(id, entry);
          } else {
            visibleBlocks.delete(id);
          }
        }

        const nextActive = resolveActiveID();
        if (nextActive !== "") {
          setActiveBlockID(nextActive);
        }
      },
      {
        rootMargin: "-88px 0px -55% 0px",
        threshold: [0, 0.25, 0.5, 1]
      }
    );

    const elements = blockOrder
      .map((id) => document.getElementById(id))
      .filter((element): element is HTMLElement => element !== null);

    for (const element of elements) {
      observer.observe(element);
    }

    return () => {
      observer.disconnect();
    };
  }, [blockOrder]);

  if (blocks.length === 0) {
    return null;
  }

  const activeBlock = blockByID.get(activeBlockID) ?? blocks[0];
  const heroes = activeBlock?.heroes ?? [];
  const hasHeroes = Boolean(activeBlock) && heroes.length > 0;

  return (
    <aside
      className={hasHeroes ? "patch-heroes-rail" : "patch-heroes-rail patch-heroes-rail--empty"}
      aria-label="Heroes in this update"
    >
      {hasHeroes ? (
        <div className="heroes-rail-panel">
          <p className="heroes-rail-block-label">{activeBlock.label}</p>
          <ul className="heroes-rail-list">
            {heroes.map((hero) => (
              <li key={hero.id}>
                <a
                  className="heroes-rail-link"
                  href={`#${hero.targetId}`}
                  aria-label={`Jump to ${hero.label}`}
                  title={hero.label}
                >
                  <FallbackImage
                    src={hero.iconUrl}
                    fallbackSrc={hero.iconFallbackUrl}
                    alt={hero.label}
                    className="heroes-rail-avatar"
                  />
                </a>
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </aside>
  );
}
