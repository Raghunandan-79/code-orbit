"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { OrbitIcon } from "@/components/icons/OrbitIcon";

export function SiteHeader() {
  const [scrolled, setScrolled] = useState(false);

  // Scroll-aware: tighten glass material once page scrolls past ~24px
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header
      className="sticky top-0 z-50 w-full px-4 sm:px-6 pt-3 pb-2 pointer-events-none transition-opacity duration-200"
    >
      <div className="max-w-5xl mx-auto flex items-center justify-between pointer-events-auto">
        {/* Floating Apple Glass Navigation Bar */}
        <div
          className={`
            w-full flex items-center justify-between h-13 px-4 sm:px-5 rounded-full
            apple-glass transition-all duration-300 ease-out
            ${scrolled
              ? "shadow-[0_12px_36px_rgba(0,0,0,0.65),0_1px_0_rgba(255,255,255,0.10)] border-white/[0.12]"
              : "shadow-[0_4px_20px_rgba(0,0,0,0.3)] border-white/[0.08]"
            }
          `}
          style={{
            backdropFilter: scrolled
              ? "blur(28px) saturate(200%)"
              : "blur(20px) saturate(180%)",
            WebkitBackdropFilter: scrolled
              ? "blur(28px) saturate(200%)"
              : "blur(20px) saturate(180%)",
          }}
        >
          {/* Brand */}
          <Link
            href="/"
            className="flex items-center gap-2.5 rounded-full py-1 px-1 focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2 transition-transform duration-75 active:scale-[0.97]"
            aria-label="Code Orbit home"
          >
            <div className="relative shrink-0">
              <OrbitIcon size={24} />
            </div>
            <span className="font-semibold text-sm tracking-tight text-[--fg-primary]">
              Code Orbit
            </span>
          </Link>

          {/* Links */}
          <nav className="hidden md:flex items-center gap-7" aria-label="Primary navigation">
            {[
              { href: "#workbench", label: "Workbench" },
              { href: "#capabilities", label: "Capabilities" },
              { href: "#workflow", label: "Workflow" },
            ].map((link) => (
              <a
                key={link.href}
                href={link.href}
                className="
                  text-xs font-medium text-[--fg-secondary] hover:text-[--fg-primary]
                  transition-colors duration-150 whitespace-nowrap relative py-1
                  focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2 rounded-sm
                  after:absolute after:bottom-[-2px] after:left-0 after:w-0 after:h-[1.5px]
                  after:bg-[--brand] after:rounded-full after:transition-all after:duration-200
                  hover:after:w-full
                "
              >
                {link.label}
              </a>
            ))}
          </nav>

          {/* Quick Action — glow-pulse CTA */}
          <div className="flex items-center gap-2">
            <Link
              href="/new"
              className="
                inline-flex items-center justify-center gap-1.5 h-8.5 px-4 rounded-full
                text-xs font-semibold text-white
                bg-[--brand] hover:bg-[--brand-hover]
                transition-all duration-100 ease-out
                active:scale-[0.96]
                shadow-[0_0_20px_var(--brand-glow)]
                hover:shadow-[0_0_28px_var(--brand-glow-strong)]
                whitespace-nowrap
                focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2
              "
            >
              <span>Create room</span>
              <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
                <path d="M2.5 6h7M6.5 3l3 3-3 3" stroke="currentColor" strokeWidth="1.25" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </Link>
          </div>
        </div>
      </div>
    </header>
  );
}
