import { FallbackImage } from "@/components/FallbackImage";
import { PatchSection } from "@/lib/types";
import { sectionAnchor } from "@/lib/utils";

type PatchSectionRendererProps = {
  section: PatchSection;
};

export function PatchSectionRenderer({ section }: PatchSectionRendererProps) {
  const sectionClass = `patch-section patch-section--${section.kind}`;
  const isHeroSection = section.kind === "heroes";
  const isItemSection = section.kind === "items";

  return (
    <section id={sectionAnchor(section.id)} className={sectionClass}>
      <header className="patch-section-header">
        <h2>{section.title}</h2>
      </header>

      <div className="patch-entry-list">
        {section.entries.map((entry) => {
          const entryClasses = [
            "patch-entry",
            isHeroSection ? "patch-entry--hero" : "",
            isItemSection ? "patch-entry--item" : "",
          ]
            .filter(Boolean)
            .join(" ");

          return (
            <article key={entry.id} className={entryClasses}>
              <header className="patch-entry-header">
                <FallbackImage
                  src={entry.entityIconUrl}
                  fallbackSrc={entry.entityIconFallbackUrl}
                  alt={entry.entityName}
                  className={isHeroSection || isItemSection ? "entry-portrait" : "entry-icon"}
                />
                <div className="entry-heading-copy">
                  <h3>{entry.entityName}</h3>
                </div>
              </header>

              {entry.summary ? <blockquote className="entry-quote">{entry.summary}</blockquote> : null}

              {isHeroSection && entry.changes.length ? (
                <section className="entry-general-block">
                  <ul className="entry-change-list">
                    {entry.changes.map((change) => (
                      <li key={change.id}>{change.text}</li>
                    ))}
                  </ul>
                </section>
              ) : null}

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
