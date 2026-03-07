import { describe, expect, it } from "vitest";
import { clampPage, formatDisplayDate, sectionAnchor } from "@/lib/utils";

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

describe("sectionAnchor", () => {
  it("prefixes section id", () => {
    expect(sectionAnchor("general")).toBe("section-general");
  });
});
