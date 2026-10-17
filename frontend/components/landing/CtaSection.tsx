"use client";

import Link from "next/link";

const SPECIFICATIONS = [
  { label: "Room Initialization", value: "Instant" },
  { label: "Max Room Concurrency", value: "25 peers" },
  { label: "Supported Syntaxes", value: "9 languages" },
  { label: "Authentication Required", value: "None" },
];

export function CtaSection() {
  return (
    <section
      className="w-full py-20 sm:py-28 relative overflow-hidden"
      aria-labelledby="cta-heading"
    >
      <div className="max-w-5xl mx-auto px-4 sm:px-6">
        {/* ── Specifications Strip ── */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-6 sm:p-7 rounded-2xl bg-white/[0.02] border border-white/[0.07] shadow-[inset_0_1px_0_rgba(255,255,255,0.04)] mb-16">
          {SPECIFICATIONS.map((spec) => (
            <div
              key={spec.label}
              className="flex flex-col gap-1"
            >
              <span className="font-semibold text-lg sm:text-xl text-[--fg-primary] tabular-data tracking-tight">
                {spec.value}
              </span>
              <span className="text-xs text-[--fg-tertiary]">
                {spec.label}
              </span>
            </div>
          ))}
        </div>

        {/* ── Closing Statement Card ── */}
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-8 p-8 sm:p-12 rounded-3xl apple-card relative overflow-hidden border border-white/[0.09]">
          {/* ── Animated ambient glow ring ── */}
          <div
            className="absolute -right-24 -bottom-24 w-96 h-96 rounded-full pointer-events-none animate-ambient-pulse"
            style={{
              background: "radial-gradient(circle, rgba(99, 102, 241, 0.22) 0%, rgba(99,102,241,0.06) 50%, transparent 70%)",
            }}
            aria-hidden="true"
          />
          {/* Secondary glow — top-left corner */}
          <div
            className="absolute -left-16 -top-16 w-64 h-64 rounded-full pointer-events-none opacity-30 animate-ambient-pulse"
            style={{
              background: "radial-gradient(circle, rgba(52,211,153,0.12) 0%, transparent 70%)",
              animationDelay: "2s",
            }}
            aria-hidden="true"
          />

          {/* Copy */}
          <div className="max-w-xl relative z-10">
            <h2
              id="cta-heading"
              className="text-title mb-4"
            >
              Start coding with your team today.
            </h2>
            <p className="text-body-lg">
              Create a workspace in seconds. Share the URL. Everyone collaborates in the same file immediately with no barrier to entry.
            </p>
          </div>

          {/* CTA — glow button */}
          <div className="relative z-10 flex flex-col sm:flex-row md:flex-col gap-3 shrink-0">
            <Link
              href="/new"
              className="
                inline-flex items-center justify-center gap-2 h-12 px-7 rounded-full
                bg-[--brand] text-white text-sm font-semibold
                hover:bg-[--brand-hover] transition-all duration-100 ease-out
                active:scale-[0.96] whitespace-nowrap
                shadow-[0_0_24px_var(--brand-glow)] hover:shadow-[0_0_36px_var(--brand-glow-strong)]
                focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2
              "
            >
              <span>Create a room</span>
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
                <path d="M2.5 7h9M8 3.5l3.5 3.5L8 10.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </Link>
            <span className="text-[11px] text-[--fg-tertiary] text-center md:text-right">
              Free open collaboration · Zero login
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}
