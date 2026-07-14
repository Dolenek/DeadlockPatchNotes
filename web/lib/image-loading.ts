type ImageLoadingInput = {
  preload?: boolean;
  loading?: "eager" | "lazy";
  fetchPriority?: "high" | "low" | "auto";
};

export function resolveDecorativeImageLoading({ preload, loading, fetchPriority }: ImageLoadingInput) {
  if (preload) {
    return {
      preload: true,
      loading: undefined,
      fetchPriority: undefined,
    } as const;
  }

  return {
    preload: false,
    loading: loading ?? "lazy",
    fetchPriority,
  } as const;
}
