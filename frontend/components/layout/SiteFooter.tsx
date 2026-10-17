import Link from "next/link";
import { OrbitIcon } from "@/components/icons/OrbitIcon";

export function SiteFooter() {
  return (
    <footer
      className="w-full border-t border-white/[0.06] py-12 mt-auto"
      aria-label="Footer"
    >
      <div className="max-w-5xl mx-auto px-4 sm:px-6">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6 pb-8 border-b border-white/[0.04]">
          {/* Brand */}
          <div className="flex items-center gap-2.5">
            <OrbitIcon size={22} />
            <span className="text-sm font-semibold text-[--fg-primary] tracking-tight">
              Code Orbit
            </span>
            <span className="text-xs text-[--fg-quaternary]">/</span>
            <span className="text-xs text-[--fg-tertiary]">
              Zero configuration collaborative coding
            </span>
          </div>

          {/* Navigation */}
          <nav className="flex items-center gap-6" aria-label="Footer navigation">
            <Link
              href="/"
              className="text-xs text-[--fg-secondary] hover:text-[--fg-primary] transition-colors py-1 focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2 rounded-sm"
            >
              Home
            </Link>
            <Link
              href="/new"
              className="text-xs text-[--fg-secondary] hover:text-[--fg-primary] transition-colors py-1 focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2 rounded-sm"
            >
              New room
            </Link>
            <a
              href="#workbench"
              className="text-xs text-[--fg-secondary] hover:text-[--fg-primary] transition-colors py-1 focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2 rounded-sm"
            >
              Workbench
            </a>
            <a
              href="#capabilities"
              className="text-xs text-[--fg-secondary] hover:text-[--fg-primary] transition-colors py-1 focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2 rounded-sm"
            >
              Capabilities
            </a>
          </nav>
        </div>

        {/* Ft2 Single Line Colophon */}
        <div className="pt-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs text-[--fg-quaternary]">
          <p>
            Architected with Next.js, Monaco Editor, and Go WebSocket engine.
          </p>
          <p>
            Code Orbit. Open multiplayer runtime.
          </p>
        </div>
      </div>
    </footer>
  );
}
