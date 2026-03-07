import Link from "next/link";

export default function NotFoundPage() {
  return (
    <main className="shell not-found-page">
      <h1>Patch Not Found</h1>
      <p>The requested patch entry does not exist in the current dataset.</p>
      <Link href="/patches">Back to Patch List</Link>
    </main>
  );
}
