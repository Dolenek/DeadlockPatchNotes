import Link from "next/link";
import { FallbackImage } from "@/components/FallbackImage";
import { PatchEntry, PatchEntryGroup, PatchSection } from "@/lib/types";
import { entryAnchor, normalizeLookupKey, sectionAnchor, slugifyLookup } from "@/lib/utils";

type PatchSectionRendererProps = {
  section: PatchSection;
};

type EntryHeaderProps = {
  entry: PatchEntry;
  kind: PatchSection["kind"];
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

function entryTimelineHref(kind: PatchSection["kind"], entryName: string): string | null {
  const entrySlug = slugifyLookup(entryName);
  if (entrySlug === "entry") {
    return null;
  }
  if (kind === "heroes") {
    return `/heroes/${entrySlug}`;
  }
  if (kind === "items") {
    return `/items/${entrySlug}`;
  }
  return null;
}

function spellTimelineHref(groupTitle: string): string | null {
  const normalizedGroupTitle = normalizeLookupKey(groupTitle);
  if (normalizedGroupTitle === "talents" || normalizedGroupTitle === "card types") {
    return null;
  }
  const groupSlug = slugifyLookup(groupTitle);
  if (groupSlug === "entry") {
    return null;
  }
  return `/spells/${groupSlug}`;
}

function EntryHeader({ entry, kind, portraitLayout }: EntryHeaderProps) {
  const href = entryTimelineHref(kind, entry.entityName);

  return (
    <header className="patch-entry-header">
      <FallbackImage
        src={entry.entityIconUrl}
        fallbackSrc={entry.entityIconFallbackUrl}
        alt={entry.entityName}
        className={portraitLayout ? "entry-portrait" : "entry-icon"}
      />
      <div className="entry-heading-copy">
        <h3>
          {href ? (
            <Link href={href} className="patch-entry-title-link">
              {entry.entityName}
            </Link>
          ) : (
            entry.entityName
          )}
        </h3>
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

function EntryGroups({ groups, kind }: { groups: PatchEntryGroup[]; kind: PatchSection["kind"] }) {
  return (
    <div className="entry-groups">
      {groups.map((group) => {
        const href = kind === "heroes" ? spellTimelineHref(group.title) : null;

        return (
          <section key={group.id} className="entry-group">
            <header className="entry-group-header">
              <FallbackImage
                src={group.iconUrl}
                fallbackSrc={group.iconFallbackUrl}
                alt={group.title}
                className="group-icon"
              />
              <h4>
                {href ? (
                  <Link href={href} className="entry-group-title-link">
                    {group.title}
                  </Link>
                ) : (
                  group.title
                )}
              </h4>
            </header>
            <ChangeList changes={group.changes} />
          </section>
        );
      })}
    </div>
  );
}

function PatchEntryArticle({
  entry,
  kind,
  entryAnchorID
}: {
  entry: PatchEntry;
  kind: PatchSection["kind"];
  entryAnchorID: string;
}) {
  const portraitLayout = kind === "heroes" || kind === "items";
  const changes = Array.isArray(entry.changes) ? entry.changes : [];
  const groups = Array.isArray(entry.groups) ? entry.groups : [];

  return (
    <article id={entryAnchor(entryAnchorID)} className={getEntryClassName(kind)}>
      <EntryHeader entry={entry} kind={kind} portraitLayout={portraitLayout} />

      {entry.summary ? <blockquote className="entry-quote">{entry.summary}</blockquote> : null}

      {changes.length ? (
        <section className="entry-general-block">
          <ChangeList changes={changes} />
        </section>
      ) : null}

      {groups.length ? <EntryGroups groups={groups} kind={kind} /> : null}
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
        {section.entries.map((entry) => {
          const entryID = `${section.id}-${entry.id}`;
          return <PatchEntryArticle key={entryID} entry={entry} kind={section.kind} entryAnchorID={entryID} />;
        })}
      </div>
    </section>
  );
}
