export type PatchSummary = {
  id: string;
  slug: string;
  title: string;
  publishedAt: string;
  category: string;
  imageUrl: string;
  source: {
    type: string;
    url: string;
  };
  releaseTimeline?: PatchTimelineSummary[];
};

export type PatchTimelineSummary = {
  id: string;
  releaseType: string;
  title: string;
  releasedAt: string;
};

export type PatchChange = {
  id: string;
  text: string;
};

export type PatchEntryGroup = {
  id: string;
  title: string;
  iconUrl?: string;
  iconFallbackUrl?: string;
  changes: PatchChange[];
};

export type PatchEntry = {
  id: string;
  entityName: string;
  entityIconUrl?: string;
  entityIconFallbackUrl?: string;
  summary?: string;
  changes: PatchChange[];
  groups?: PatchEntryGroup[];
};

export type PatchSection = {
  id: string;
  title: string;
  kind: "general" | "items" | "heroes";
  entries: PatchEntry[];
};

export type PatchTimelineBlock = {
  id: string;
  releaseType: string;
  title: string;
  releasedAt: string;
  source: {
    type: string;
    url: string;
  };
  changes: PatchChange[];
  sections?: PatchSection[];
};

export type PatchDetail = {
  id: string;
  slug: string;
  title: string;
  publishedAt: string;
  category: string;
  source: {
    type: string;
    url: string;
  };
  imageUrl: string;
  intro: string;
  sections: PatchSection[];
  releaseTimeline?: PatchTimelineBlock[];
};

export type PatchListResponse = {
  patches: PatchSummary[];
  pagination: {
    page: number;
    pageSize: number;
    totalItems: number;
    totalPages: number;
  };
};

export type HeroSummary = {
  slug: string;
  name: string;
  iconUrl?: string;
  iconFallbackUrl?: string;
  lastChangedAt: string;
};

export type HeroListResponse = {
  heroes: HeroSummary[];
};

export type HeroTimelineSkill = {
  id: string;
  title: string;
  iconUrl?: string;
  iconFallbackUrl?: string;
  changes: PatchChange[];
};

export type HeroTimelineBlock = {
  id: string;
  releaseType: string;
  displayLabel: string;
  releasedAt: string;
  patchRef: {
    slug: string;
    title: string;
  };
  source: {
    type: string;
    url: string;
  };
  generalChanges?: PatchChange[];
  skills: HeroTimelineSkill[];
};

export type HeroChangesResponse = {
  hero: HeroSummary;
  timeline: HeroTimelineBlock[];
};

export type ItemSummary = {
  slug: string;
  name: string;
  iconUrl?: string;
  iconFallbackUrl?: string;
  lastChangedAt: string;
};

export type ItemListResponse = {
  items: ItemSummary[];
};

export type ItemTimelineBlock = {
  id: string;
  releaseType: string;
  displayLabel: string;
  releasedAt: string;
  patchRef: {
    slug: string;
    title: string;
  };
  source: {
    type: string;
    url: string;
  };
  changes: PatchChange[];
};

export type ItemChangesResponse = {
  item: ItemSummary;
  timeline: ItemTimelineBlock[];
};

export type SpellSummary = {
  slug: string;
  name: string;
  iconUrl?: string;
  iconFallbackUrl?: string;
  lastChangedAt: string;
};

export type SpellListResponse = {
  spells: SpellSummary[];
};

export type SpellTimelineEntry = {
  id: string;
  heroSlug?: string;
  heroName?: string;
  heroIconUrl?: string;
  heroIconFallbackUrl?: string;
  changes: PatchChange[];
};

export type SpellTimelineBlock = {
  id: string;
  releaseType: string;
  displayLabel: string;
  releasedAt: string;
  patchRef: {
    slug: string;
    title: string;
  };
  source: {
    type: string;
    url: string;
  };
  entries: SpellTimelineEntry[];
};

export type SpellChangesResponse = {
  spell: SpellSummary;
  timeline: SpellTimelineBlock[];
};
