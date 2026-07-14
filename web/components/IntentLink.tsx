import Link, { type LinkProps } from "next/link";
import type { AnchorHTMLAttributes } from "react";

type IntentLinkProps = LinkProps &
  Omit<AnchorHTMLAttributes<HTMLAnchorElement>, keyof LinkProps | "href">;

export function IntentLink({ prefetch: _prefetch, ...props }: IntentLinkProps) {
  return <Link {...props} prefetch={false} data-prefetch="intent" />;
}
