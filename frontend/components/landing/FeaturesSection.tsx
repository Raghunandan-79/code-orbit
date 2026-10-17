"use client";

import { LANGUAGES } from "@/lib/languages";

interface Capability {
  title: string;
  category: string;
  summary: string;
  spec: string;
}

const CAPABILITIES: Capability[] = [
  {
    title: "Real-time Delta Broadcast",
    category: "Synchronization",
    summary: "Keystrokes are encoded as compact binary updates and dispatched across WebSocket connections in under 50 milliseconds.",
    spec: "Sequential versioning with automated reconnection and exponential backoff.",
  },
  {
    title: "Awareness & Cursor Projection",
    category: "Presence",
    summary: "See where team members are typing and selecting in real time. Floating nametags clarify active ownership across lines.",
    spec: "Color-coded participant indicators with line and column tracking.",
  },
  {
    title: "Monaco Syntax Engine",
    category: "Editor",
    summary: "Built on the same core editor that powers VS Code. Native keybindings, auto-closing brackets, indentation guides, and line numbers.",
    spec: "Optimized for desktop pair programming and code reviews.",
  },
  {
    title: "Ephemeral Room Lifecycle",
    category: "Privacy",
    summary: "Zero database footprint. Coding sessions live exclusively in memory during active collaboration and cleanly dissolve after expiry.",
    spec: "Instant room creation via URL sharing with no account required.",
  },
];

export function FeaturesSection() {
  return (
    <section
      id="capabilities"
      className="w-full py-20 sm:py-28 relative"
      aria-labelledby="capabilities-heading"
    >
      <div className="max-w-5xl mx-auto px-4 sm:px-6">
        {/* Section Title */}
        <div className="max-w-xl mb-14">
          <p className="text-label text-[--brand] mb-2.5">
            Architecture
          </p>
          <h2 id="capabilities-heading" className="text-title mb-4">
            Engineered for low-latency pair programming.
          </h2>
          <p className="text-body-lg">
            A focused collaboration layer built on standard web protocols. No heavy accounts, no setup scripts, no telemetry.
          </p>
        </div>

        {/* Capability Grid — card hover: lift + glow rim */}
        <div className="grid md:grid-cols-2 gap-4 mb-12">
          {CAPABILITIES.map((cap) => (
            <div
              key={cap.title}
              className="
                group relative p-6 sm:p-7 rounded-2xl apple-card
                flex flex-col justify-between min-h-[220px]
                transition-all duration-200 ease-out
                hover:-translate-y-1
                hover:border-white/[0.14]
                hover:shadow-[0_12px_36px_rgba(0,0,0,0.5),0_0_0_1px_rgba(99,102,241,0.18),0_0_48px_rgba(99,102,241,0.06)]
              "
            >
              {/* Top row: category + dot */}
              <div>
                <div className="flex items-center justify-between mb-3 text-xs font-mono text-[--fg-tertiary]">
                  <span
                    className="px-2.5 py-0.5 rounded-full text-[10px] font-semibold tracking-wide uppercase"
                    style={{
                      background: "rgba(99,102,241,0.09)",
                      color: "rgba(163,161,255,0.85)",
                      border: "1px solid rgba(99,102,241,0.18)",
                    }}
                  >
                    {cap.category}
                  </span>
                  {/* Dot brightens to brand on hover */}
                  <span className="w-1.5 h-1.5 rounded-full bg-white/[0.15] group-hover:bg-[--brand] group-hover:shadow-[0_0_8px_var(--brand)] transition-all duration-200" />
                </div>

                <h3 className="text-base font-semibold text-[--fg-primary] mb-2 tracking-tight">
                  {cap.title}
                </h3>

                <p className="text-sm text-[--fg-secondary] leading-relaxed mb-4">
                  {cap.summary}
                </p>
              </div>

              {/* Spec line */}
              <div className="pt-4 border-t border-white/[0.06] text-xs font-mono text-[--fg-tertiary] group-hover:text-[--fg-secondary] transition-colors duration-200">
                {cap.spec}
              </div>

              {/* Subtle inner glow that appears on hover */}
              <div
                className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity duration-300"
                style={{
                  background: "radial-gradient(ellipse 70% 50% at 50% 0%, rgba(99,102,241,0.05) 0%, transparent 70%)",
                }}
                aria-hidden="true"
              />
            </div>
          ))}
        </div>

        {/* 9 Languages Strip */}
        <div className="p-6 sm:p-7 rounded-2xl bg-white/[0.02] border border-white/[0.07] shadow-[inset_0_1px_0_rgba(255,255,255,0.04)]">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-5">
            <div>
              <h4 className="text-sm font-semibold text-[--fg-primary]">
                Nine native language environments
              </h4>
              <p className="text-xs text-[--fg-tertiary] mt-0.5">
                Full syntax tokenization and starter buffers configured out of the box.
              </p>
            </div>
            <span className="text-xs font-mono text-[--fg-tertiary] px-2.5 py-1 rounded-md bg-white/[0.03] border border-white/[0.06] w-fit">
              Monaco Language Registry
            </span>
          </div>

          <div className="grid grid-cols-3 sm:grid-cols-9 gap-2.5">
            {LANGUAGES.map((lang) => (
              <div
                key={lang.id}
                className="
                  group flex flex-col items-center gap-1.5 py-3.5 px-2 rounded-xl
                  bg-white/[0.02] border border-white/[0.05]
                  hover:bg-white/[0.06] hover:border-white/[0.12]
                  hover:shadow-[0_4px_16px_rgba(0,0,0,0.3)]
                  transition-all duration-150 ease-out
                  hover:-translate-y-0.5
                  cursor-default select-none
                "
              >
                <span
                  className="
                    w-7.5 h-7.5 rounded-lg flex items-center justify-center
                    font-mono text-[10px] font-bold
                    transition-all duration-150
                    group-hover:scale-105
                  "
                  style={{
                    background: `${lang.color}18`,
                    color: lang.color,
                    border: `1px solid ${lang.color}35`,
                  }}
                >
                  {lang.icon}
                </span>
                <span className="text-[11px] font-medium text-[--fg-secondary] truncate group-hover:text-[--fg-primary] transition-colors duration-150">
                  {lang.label}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
