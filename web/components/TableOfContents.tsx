"use client";

import { useState } from "react";
import { PatchSection } from "@/lib/types";
import { sectionAnchor } from "@/lib/utils";

type TableOfContentsProps = {
  sections: PatchSection[];
};

export function TableOfContents({ sections }: TableOfContentsProps) {
  const [open, setOpen] = useState(false);

  return (
    <aside className="toc" aria-label="Table of contents">
      <button className="toc-toggle" type="button" onClick={() => setOpen((value) => !value)}>
        Sections {open ? "−" : "+"}
      </button>
      <ul className={open ? "toc-list is-open" : "toc-list"}>
        {sections.map((section) => (
          <li key={section.id}>
            <a href={`#${sectionAnchor(section.id)}`}>{section.title}</a>
          </li>
        ))}
      </ul>
    </aside>
  );
}
