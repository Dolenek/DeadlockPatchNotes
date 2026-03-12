import { describe, expect, it } from "vitest";
import {
  buildPatchTimelineHref,
  clampPage,
  extractTimelineSourceBlockID,
  formatCompactDate,
  formatDisplayDate,
  formatForumDate,
  formatUpdateLabel,
  normalizeLookupKey,
  sectionAnchor,
  slugifyLookup,
  timelineBlockAnchor
} from "@/lib/utils";

describe("clampPage", () => {
  it("falls back to page 1 for invalid input", () => {
    expect(clampPage(undefined)).toBe(1);
    expect(clampPage("0")).toBe(1);
    expect(clampPage("-7")).toBe(1);
    expect(clampPage("abc")).toBe(1);
  });

  it("uses positive integer values", () => {
    expect(clampPage("3")).toBe(3);
    expect(clampPage(["7", "2"])).toBe(7);
  });
});

describe("formatDisplayDate", () => {
  it("formats ISO date into long US style", () => {
    expect(formatDisplayDate("2026-03-06T22:36:00Z")).toBe("March 6, 2026");
  });
});

describe("formatCompactDate", () => {
  it("formats ISO date into D.M. YYYY", () => {
    expect(formatCompactDate("2026-03-06T22:36:00Z")).toBe("6.3. 2026");
  });
});

describe("sectionAnchor", () => {
  it("prefixes section id", () => {
    expect(sectionAnchor("general")).toBe("section-general");
  });
});

describe("timelineBlockAnchor", () => {
  it("prefixes timeline block id", () => {
    expect(timelineBlockAnchor("update-2026-03-10")).toBe("timeline-update-2026-03-10");
  });
});

describe("normalizeLookupKey", () => {
  it("normalizes punctuation and spacing", () => {
    expect(normalizeLookupKey("Card Types")).toBe("card types");
  });
});

describe("slugifyLookup", () => {
  it("slugifies spell titles", () => {
    expect(slugifyLookup("Siphon Life")).toBe("siphon-life");
    expect(slugifyLookup("Mo & Krill")).toBe("mo-krill");
  });

  it("falls back to entry for empty slugs", () => {
    expect(slugifyLookup("___")).toBe("entry");
  });
});

describe("extractTimelineSourceBlockID", () => {
  it("strips entity slug suffix from block id", () => {
    expect(extractTimelineSourceBlockID("03-06-2026-update-abrams", "abrams")).toBe("03-06-2026-update");
    expect(extractTimelineSourceBlockID("03-06-2026-update-lady-geist", "lady-geist")).toBe("03-06-2026-update");
  });

  it("returns empty when suffix does not match", () => {
    expect(extractTimelineSourceBlockID("03-06-2026-update", "abrams")).toBe("");
  });
});

describe("buildPatchTimelineHref", () => {
  it("builds anchored patch href when source id is derivable", () => {
    expect(buildPatchTimelineHref("2026-03-06-update", "03-06-2026-update-abrams", "abrams")).toBe(
      "/patches/2026-03-06-update#timeline-03-06-2026-update"
    );
  });

  it("falls back to patch root when source id is unknown", () => {
    expect(buildPatchTimelineHref("2026-03-06-update", "03-06-2026-update", "abrams")).toBe(
      "/patches/2026-03-06-update"
    );
  });
});

describe("formatForumDate", () => {
  it("formats ISO date into MM-DD-YYYY UTC", () => {
    expect(formatForumDate("2026-03-06T22:36:00Z")).toBe("03-06-2026");
  });
});

describe("formatUpdateLabel", () => {
  it("uses update prefix for initial blocks", () => {
    expect(formatUpdateLabel("initial", "2026-03-06T22:36:00Z")).toBe("Update 03-06-2026");
  });

  it("uses patch prefix for hotfix blocks", () => {
    expect(formatUpdateLabel("hotfix", "2026-03-07T12:00:00Z")).toBe("Patch 03-07-2026");
  });
});
