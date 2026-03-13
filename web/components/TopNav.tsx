import Link from "next/link";

export function TopNav() {
  return (
    <header className="top-nav">
      <div className="top-nav-inner shell">
        <Link href="/" className="brand-wordmark">
          <img src="/deadlock_logo_white.svg" alt="Deadlock" className="brand-logo" />
        </Link>
        <nav className="main-links" aria-label="Main navigation">
          <Link href="/patches">Patch Notes</Link>
          <Link href="/heroes">Heroes</Link>
          <Link href="/items">Items</Link>
          <Link href="/spells">Spells</Link>
          <a href="https://forums.playdeadlock.com/forums/changelog.10/" target="_blank" rel="noreferrer" className="main-links-right-anchor">
            Changelog Forum
          </a>
          <a href="/api/scalar" target="_blank" rel="noreferrer">
            PatchNotes API
          </a>
          <a href="https://assets.deadlock-api.com/scalar" target="_blank" rel="noreferrer">
            Assets API
          </a>
        </nav>
        <a className="play-button" href="https://store.steampowered.com/app/1422450/Deadlock/" target="_blank" rel="noreferrer">
          Open on Steam
        </a>
      </div>
    </header>
  );
}
