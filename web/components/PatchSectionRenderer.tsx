import { FallbackImage } from "@/components/FallbackImage";
import { PatchEntry, PatchEntryGroup, PatchSection } from "@/lib/types";
import { sectionAnchor } from "@/lib/utils";

type PatchSectionRendererProps = {
  section: PatchSection;
};

type EntryHeaderProps = {
  entry: PatchEntry;
  portraitLayout: boolean;
};

function getEntryClassName(kind: PatchSection["kind"]) {
  const classes = ["patch-entry"];
  if (kind === "heroes") {
    classes.push("patch-entry--hero");
  }
  if (kind === "items") {
    classes.push("patch-entry--item");
  }
  return classes.join(" ");
}

function EntryHeader({ entry, portraitLayout }: EntryHeaderProps) {
  return (
    <header className="patch-entry-header">
      <FallbackImage
        src={entry.entityIconUrl}
        fallbackSrc={entry.entityIconFallbackUrl}
        alt={entry.entityName}
        className={portraitLayout ? "entry-portrait" : "entry-icon"}
      />
      <div className="entry-heading-copy">
        <h3>{entry.entityName}</h3>
      </div>
    </header>
  );
}

function ChangeList({ changes }: { changes: PatchEntry["changes"] }) {
  return (
    <ul className="entry-change-list">
      {changes.map((change) => (
        <li key={change.id}>{change.text}</li>
      ))}
    </ul>
  );
}

function EntryGroups({ groups }: { groups: PatchEntryGroup[] }) {
  return (
    <div className="entry-groups">
      {groups.map((group) => (
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
          <ChangeList changes={group.changes} />
        </section>
      ))}
    </div>
  );
}

function PatchEntryArticle({ entry, kind }: { entry: PatchEntry; kind: PatchSection["kind"] }) {
  const portraitLayout = kind === "heroes" || kind === "items";

  return (
    <article className={getEntryClassName(kind)}>
      <EntryHeader entry={entry} portraitLayout={portraitLayout} />

      {entry.summary ? <blockquote className="entry-quote">{entry.summary}</blockquote> : null}

      {entry.changes.length ? (
        <section className="entry-general-block">
          <ChangeList changes={entry.changes} />
        </section>
      ) : null}

      {entry.groups?.length ? <EntryGroups groups={entry.groups} /> : null}
    </article>
  );
}

export function PatchSectionRenderer({ section }: PatchSectionRendererProps) {
  return (
    <section id={sectionAnchor(section.id)} className={`patch-section patch-section--${section.kind}`}>
      <header className="patch-section-header">
        <h2>{section.title}</h2>
      </header>

      <div className="patch-entry-list">
        {section.entries.map((entry) => (
          <PatchEntryArticle key={entry.id} entry={entry} kind={section.kind} />
        ))}
      </div>
    </section>
  );
}
