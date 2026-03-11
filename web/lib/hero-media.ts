import heroMediaManifest from "@/lib/hero-media-manifest.json";

type HeroMediaEntry = {
  name: string;
  sourceName: string;
  backgroundImageUrl?: string;
  nameImageUrl?: string;
};

type HeroMediaManifest = {
  generatedAt: string;
  heroes: Record<string, HeroMediaEntry>;
};

const manifest = heroMediaManifest as HeroMediaManifest;

export function getHeroMediaBySlug(slug: string): HeroMediaEntry | null {
  return manifest.heroes[slug] ?? null;
}
