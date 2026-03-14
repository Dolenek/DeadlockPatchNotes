import Image from "next/image";

type DecorativeImageLayer = {
  src: string;
  className?: string;
  quality?: number;
  priority?: boolean;
  sizes?: string;
};

type DecorativeImageLayersProps = {
  className?: string;
  layers: DecorativeImageLayer[];
};

export function DecorativeImageLayers({ className, layers }: DecorativeImageLayersProps) {
  if (layers.length === 0) {
    return null;
  }

  const containerClassName = className ? `decorative-image-layers ${className}` : "decorative-image-layers";

  return (
    <div className={containerClassName} aria-hidden>
      {layers.map((layer, index) => (
        <Image
          key={`${layer.src}-${index}`}
          src={layer.src}
          alt=""
          fill
          quality={layer.quality ?? 60}
          priority={layer.priority ?? false}
          sizes={layer.sizes ?? "100vw"}
          className={layer.className ? `decorative-image-layer ${layer.className}` : "decorative-image-layer"}
        />
      ))}
    </div>
  );
}
