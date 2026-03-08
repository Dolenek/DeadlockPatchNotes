import { FallbackImage } from "@/components/FallbackImage";
import { PatchSection } from "@/lib/types";
import { sectionAnchor } from "@/lib/utils";

type PatchSectionRendererProps = {
  section: PatchSection;
};

export function PatchSectionRenderer({ section }: PatchSectionRendererProps) {
  const sectionClass = `patch-section patch-section--${section.kind}`;
  const isHeroSection = section.kind === "heroes";

  return (
    <section id={sectionAnchor(section.id)} className={sectionClass}>
      <header className="patch-section-header">
        <h2>{section.title}</h2>
      </header>

      <div className="patch-entry-list">
        {section.entries.map((entry) => {
          const heroEntryClass = isHeroSection ? "patch-entry patch-entry--hero" : "patch-entry";

          return (
            <article key={entry.id} className={heroEntryClass}>
              <header className="patch-entry-header">
                <FallbackImage
                  src={entry.entityIconUrl}
                  fallbackSrc={entry.entityIconFallbackUrl}
                  alt={entry.entityName}
                  className={isHeroSection ? "entry-portrait" : "entry-icon"}
                />
                <div className="entry-heading-copy">
                  <h3>{entry.entityName}</h3>
                  {entry.summary ? <p className="entry-summary">{entry.summary}</p> : null}

                  {isHeroSection && entry.changes.length ? (
                    <ul className="entry-inline-change-list">
                      {entry.changes.map((change) => (
                        <li key={change.id}>{change.text}</li>
                      ))}
                    </ul>
                  ) : null}
                </div>
              </header>

              {!isHeroSection && entry.changes.length ? (
                <section className="entry-general-block">
                  <ul className="entry-change-list">
                    {entry.changes.map((change) => (
                      <li key={change.id}>{change.text}</li>
                    ))}
                  </ul>
                </section>
              ) : null}

              {entry.groups?.length ? (
                <div className="entry-groups">
                  {entry.groups.map((group) => (
                    <section key={group.id} className="entry-group">
                      <header className="entry-group-header">
                        <FallbackImage
                          src={group.iconUrl}
                          fallbackSrc={group.iconFallbackUrl}
                          alt={group.title}
                          className="group-icon"
                        />
                        <h4>{group.title}</h4>
                      </header>
                      <ul className="entry-change-list">
                        {group.changes.map((change) => (
                          <li key={change.id}>{change.text}</li>
                        ))}
                      </ul>
                    </section>
                  ))}
                </div>
              ) : null}
            </article>
          );
        })}
      </div>
    </section>
  );
}
