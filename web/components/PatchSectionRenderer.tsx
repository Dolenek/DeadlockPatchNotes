import { PatchSection } from "@/lib/types";
import { sectionAnchor } from "@/lib/utils";

type PatchSectionRendererProps = {
  section: PatchSection;
};

export function PatchSectionRenderer({ section }: PatchSectionRendererProps) {
  return (
    <section id={sectionAnchor(section.id)} className="patch-section">
      <h2>{section.title}</h2>
      {section.entries.map((entry) => (
        <article key={entry.id} className="patch-entry">
          <header className="patch-entry-header">
            {entry.entityIconUrl ? <img src={entry.entityIconUrl} alt="" className="entry-icon" /> : null}
            <h3>{entry.entityName}</h3>
          </header>
          {entry.summary ? <p className="entry-summary">{entry.summary}</p> : null}
          <ul>
            {entry.changes.map((change) => (
              <li key={change.id}>{change.text}</li>
            ))}
          </ul>
        </article>
      ))}
    </section>
  );
}
