import Link from "next/link";

type PaginationProps = {
  page: number;
  totalPages: number;
};

export function Pagination({ page, totalPages }: PaginationProps) {
  if (totalPages <= 1) {
    return null;
  }

  return (
    <nav className="pagination" aria-label="Patch list pagination">
      <Link
        href={page <= 1 ? "/patches" : `/patches?page=${page - 1}`}
        aria-disabled={page <= 1}
        className={page <= 1 ? "is-disabled" : ""}
      >
        Previous
      </Link>
      <span>
        Page {page} of {totalPages}
      </span>
      <Link
        href={page >= totalPages ? `/patches?page=${totalPages}` : `/patches?page=${page + 1}`}
        aria-disabled={page >= totalPages}
        className={page >= totalPages ? "is-disabled" : ""}
      >
        Next
      </Link>
    </nav>
  );
}
