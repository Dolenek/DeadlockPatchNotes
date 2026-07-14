import Image from "next/image";

type HeroBackdropProps = {
  src?: string;
};

export function HeroBackdrop({ src }: HeroBackdropProps) {
  if (!src) {
    return null;
  }

  return (
    <div className="hero-detail-backdrop" aria-hidden>
      <Image
        src={src}
        alt=""
        fill
        sizes="100vw"
        quality={45}
        preload
        className="hero-detail-backdrop__image"
      />
      <div className="hero-detail-backdrop__veil" />
    </div>
  );
}
