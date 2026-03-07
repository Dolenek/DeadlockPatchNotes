import { Pagination } from "@/components/Pagination";
import { PatchCard } from "@/components/PatchCard";
import { getPatches } from "@/lib/api";
import { clampPage } from "@/lib/utils";

type PatchesPageProps = {
  searchParams: Promise<{ page?: string | string[] }>;
};

export default async function PatchesPage({ searchParams }: PatchesPageProps) {
  const resolvedParams = await searchParams;
  const page = clampPage(resolvedParams.page);
  const data = await getPatches(page, 12);

  return (
    <main>
      <section className="patch-list-masthead">
        <div className="shell">
          <p className="eyebrow">Deadlock Updates</p>
          <h1>Patch Notes</h1>
          <p>
            A focused archive of gameplay updates with a visual style inspired by modern competitive game patch portals.
          </p>
        </div>
      </section>

      <section className="shell patch-list-section">
        <div className="patch-grid">
          {data.items.map((patch) => (
            <PatchCard key={patch.id} patch={patch} />
          ))}
        </div>

        <Pagination page={data.page} totalPages={data.totalPages} />
      </section>
    </main>
  );
}
