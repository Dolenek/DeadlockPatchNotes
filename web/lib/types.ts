export type PatchSummary = {
  id: string;
  slug: string;
  title: string;
  publishedAt: string;
  category: string;
  excerpt: string;
  coverImageUrl: string;
  sourceUrl: string;
};

export type PatchChange = {
  id: string;
  text: string;
};

export type PatchEntry = {
  id: string;
  entityName: string;
  entityIconUrl?: string;
  summary?: string;
  changes: PatchChange[];
};

export type PatchSection = {
  id: string;
  title: string;
  kind: "general" | "items" | "heroes";
  entries: PatchEntry[];
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
  heroImageUrl: string;
  intro: string;
  sections: PatchSection[];
};

export type PatchListResponse = {
  items: PatchSummary[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};
