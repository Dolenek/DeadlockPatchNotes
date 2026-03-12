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
  const patchList = await getPatches(page, 12);

  return (
    <main className="page-like-patches">
      <section className="patch-list-masthead">
        <div className="shell">
          <p className="eyebrow">Deadlock Updates</p>
          <h1>Patch Notes</h1>
        </div>
      </section>

      <section className="shell patch-list-section">
        <div className="patch-grid">
          {patchList.items.map((patch, index) => (
            <PatchCard key={patch.id} patch={patch} index={index} />
          ))}
        </div>

        <Pagination page={patchList.page} totalPages={patchList.totalPages} />
      </section>
    </main>
  );
}
