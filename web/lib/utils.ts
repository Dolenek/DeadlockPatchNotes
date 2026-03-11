export function clampPage(value: string | string[] | undefined): number {
  const raw = Array.isArray(value) ? value[0] : value;
  if (!raw) {
    return 1;
  }

  const parsed = Number.parseInt(raw, 10);
  if (Number.isNaN(parsed) || parsed < 1) {
    return 1;
  }

  return parsed;
}

export function formatDisplayDate(isoDate: string): string {
  const date = new Date(isoDate);
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric"
  });
}

export function sectionAnchor(sectionID: string): string {
  return `section-${sectionID}`;
}

export function timelineBlockAnchor(blockID: string): string {
  return `timeline-${blockID}`;
}

export function formatForumDate(isoDate: string): string {
  const date = new Date(isoDate);
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  const day = String(date.getUTCDate()).padStart(2, "0");
  const year = date.getUTCFullYear();
  return `${month}-${day}-${year}`;
}

export function formatUpdateLabel(kind: string, isoDate: string): string {
  const prefix = kind === "initial" ? "Update" : "Patch";
  return `${prefix} ${formatForumDate(isoDate)}`;
}
