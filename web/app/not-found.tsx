import Link from "next/link";

export default function NotFoundPage() {
  return (
    <main className="not-found-page">
      <section className="not-found-scene">
        <h1>Oops</h1>
        <img src="/lil_troopers.png" alt="Lil Troopers" className="not-found-image" />
        <Link href="/patches" className="not-found-link">
          Go back
        </Link>
      </section>
    </main>
  );
}
