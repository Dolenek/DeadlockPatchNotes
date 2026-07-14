import Image from "next/image";
import type { Metadata } from "next";
import { IntentLink as Link } from "@/components/IntentLink";

export const metadata: Metadata = {
  title: "Page Not Found",
  robots: {
    index: false,
    follow: false,
  },
};

export default function NotFoundPage() {
  return (
    <main className="not-found-page">
      <section className="not-found-scene">
        <h1>Oops</h1>
        <Image
          src="/lil_troopers.png"
          alt="Lil Troopers"
          className="not-found-image"
          width={1600}
          height={900}
          sizes="(max-width: 720px) 98vw, 768px"
          quality={50}
          loading="lazy"
        />
        <Link href="/patches" className="not-found-link">
          Go back
        </Link>
      </section>
    </main>
  );
}
