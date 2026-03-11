"use client";

import { useEffect, useMemo, useState } from "react";

type TableOfContentsProps = {
  groups: TableOfContentsGroup[];
};

export type TableOfContentsSectionLink = {
  id: string;
  label: string;
};

export type TableOfContentsGroup = {
  id: string;
  label: string;
  sections: TableOfContentsSectionLink[];
};

const DESKTOP_BREAKPOINT = "(min-width: 1025px)";
const ACTIVE_OFFSET_PX = 120;

export function TableOfContents({ groups }: TableOfContentsProps) {
  const [open, setOpen] = useState(false);
  const [activeId, setActiveId] = useState("");
  const anchorOrder = useMemo(() => groups.flatMap((group) => [group.id, ...group.sections.map((section) => section.id)]), [groups]);

  useEffect(() => {
    if (anchorOrder.length === 0) {
      return;
    }

    setActiveId((current) => (current && anchorOrder.includes(current) ? current : anchorOrder[0]));
  }, [anchorOrder]);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const media = window.matchMedia(DESKTOP_BREAKPOINT);
    const syncOpenState = () => {
      setOpen(media.matches);
    };

    syncOpenState();
    media.addEventListener("change", syncOpenState);
    return () => {
      media.removeEventListener("change", syncOpenState);
    };
  }, []);

  useEffect(() => {
    if (typeof window === "undefined" || anchorOrder.length === 0) {
      return;
    }

    const syncFromHash = () => {
      const targetID = decodeURIComponent(window.location.hash.replace(/^#/, ""));
      if (targetID !== "" && anchorOrder.includes(targetID)) {
        setActiveId(targetID);
      }
    };

    syncFromHash();
    window.addEventListener("hashchange", syncFromHash);
    return () => {
      window.removeEventListener("hashchange", syncFromHash);
    };
  }, [anchorOrder]);

  useEffect(() => {
    if (typeof window === "undefined" || anchorOrder.length === 0) {
      return;
    }

    const visibleEntries = new Map<string, IntersectionObserverEntry>();

    const resolveActiveID = () => {
      const visibleInOrder = anchorOrder.filter((id) => visibleEntries.has(id));
      if (visibleInOrder.length === 0) {
        return "";
      }

      let activeAnchor = visibleInOrder[0];
      for (const id of visibleInOrder) {
        const entry = visibleEntries.get(id);
        if (!entry) {
          continue;
        }

        if (entry.boundingClientRect.top <= ACTIVE_OFFSET_PX) {
          activeAnchor = id;
          continue;
        }

        break;
      }

      return activeAnchor;
    };

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          const id = entry.target.id;
          if (entry.isIntersecting) {
            visibleEntries.set(id, entry);
          } else {
            visibleEntries.delete(id);
          }
        }

        const nextActive = resolveActiveID();
        if (nextActive !== "") {
          setActiveId(nextActive);
        }
      },
      {
        rootMargin: "-88px 0px -55% 0px",
        threshold: [0, 0.25, 0.5, 1]
      }
    );

    const elements = anchorOrder
      .map((id) => document.getElementById(id))
      .filter((element): element is HTMLElement => element !== null);

    for (const element of elements) {
      observer.observe(element);
    }

    return () => {
      observer.disconnect();
    };
  }, [anchorOrder]);

  const handleLinkClick = () => {
    if (typeof window !== "undefined" && !window.matchMedia(DESKTOP_BREAKPOINT).matches) {
      setOpen(false);
    }
  };

  return (
    <aside className="toc" aria-label="Table of contents">
      <button
        aria-expanded={open}
        className="toc-toggle"
        type="button"
        onClick={() => setOpen((value) => !value)}
      >
        On This Patch
        <span className="toc-toggle-symbol" aria-hidden="true">
          {open ? "−" : "+"}
        </span>
      </button>
      <ul className={open ? "toc-list is-open" : "toc-list"}>
        {groups.map((group) => {
          const groupActive = activeId === group.id || group.sections.some((section) => section.id === activeId);

          return (
            <li className="toc-group" key={group.id}>
              <a
                className={groupActive ? "toc-link is-active" : "toc-link"}
                href={`#${group.id}`}
                onClick={handleLinkClick}
              >
                {group.label}
              </a>
              {group.sections.length > 0 ? (
                <ul className="toc-sublist">
                  {group.sections.map((section) => (
                    <li key={section.id}>
                      <a
                        className={activeId === section.id ? "toc-link toc-sublink is-active" : "toc-link toc-sublink"}
                        href={`#${section.id}`}
                        onClick={handleLinkClick}
                      >
                        {section.label}
                      </a>
                    </li>
                  ))}
                </ul>
              ) : null}
            </li>
          );
        })}
      </ul>
    </aside>
  );
}
