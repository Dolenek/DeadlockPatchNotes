import Image from "next/image";
import { resolveDecorativeImageLoading } from "@/lib/image-loading";

type DecorativeImageLayer = {
  src: string;
  className?: string;
  quality?: number;
  preload?: boolean;
  loading?: "eager" | "lazy";
  fetchPriority?: "high" | "low" | "auto";
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
        <ImageLayer key={`${layer.src}-${index}`} layer={layer} />
      ))}
    </div>
  );
}

function ImageLayer({ layer }: { layer: DecorativeImageLayer }) {
  const loadingOptions = resolveDecorativeImageLoading(layer);

  return (
    <Image
      src={layer.src}
      alt=""
      fill
      quality={layer.quality ?? 60}
      {...loadingOptions}
      sizes={layer.sizes ?? "100vw"}
      className={layer.className ? `decorative-image-layer ${layer.className}` : "decorative-image-layer"}
    />
  );
}
