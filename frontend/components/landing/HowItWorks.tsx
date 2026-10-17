"use client";


const STEPS = [
  {
    step: "01",
    title: "Initialize session",
    desc: "Select a language and generate an ephemeral workspace. The room boots in milliseconds.",
  },
  {
    step: "02",
    title: "Dispatch invite link",
    desc: "Copy the unique room URL and share it directly with teammates. No login required.",
  },
  {
    step: "03",
    title: "Code in unison",
    desc: "Edit together in Monaco with live spatial cursor presence and sub-50ms delta sync.",
  },
];

export function HowItWorks() {
  return (
    <section
      id="workflow"
      className="w-full py-16 sm:py-24 border-y border-white/[0.06] relative overflow-hidden"
      aria-labelledby="workflow-heading"
    >
      {/* Subtle ambient glow at center */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          background: "radial-gradient(ellipse 50% 60% at 50% 50%, rgba(99,102,241,0.04) 0%, transparent 70%)",
        }}
        aria-hidden="true"
      />

      <div className="relative max-w-5xl mx-auto px-4 sm:px-6">
        <div className="grid lg:grid-cols-[220px_1fr] gap-10 lg:gap-16 items-start">
          {/* Header Column */}
          <div>
            <p className="text-label text-[--brand] mb-2">Workflow</p>
            <h2 id="workflow-heading" className="text-title">
              Three steps to pair.
            </h2>
          </div>

          {/* Steps Grid */}
          <div className="relative grid sm:grid-cols-3 gap-8">
            {/* ── Connector line (desktop only) ── */}
            <div
              className="absolute hidden sm:block top-[13px] left-0 right-0 h-px"
              style={{ background: "rgba(255,255,255,0.06)" }}
              aria-hidden="true"
            />
            <div
              className="absolute hidden sm:block top-[13px] left-0 h-px origin-left"
              style={{
                background: "linear-gradient(90deg, var(--brand) 0%, rgba(99,102,241,0.2) 100%)",
                width: "66.66%", /* reaches the end of step 02 midpoint */
              }}
              aria-hidden="true"
            />

            {STEPS.map((s, i) => (
              <div
                key={s.step}
                className="relative flex flex-col"
              >
                {/* Number */}
                <div
                  className="
                    w-7 h-7 rounded-full flex items-center justify-center mb-4
                    text-[11px] font-mono font-semibold
                    relative z-10 transition-transform duration-150 hover:scale-105
                  "
                  style={{
                    background: i === 0
                      ? "var(--brand)"
                      : "rgba(99,102,241,0.12)",
                    color: i === 0
                      ? "#fff"
                      : "rgba(163,161,255,0.7)",
                    border: `1px solid ${i === 0 ? "transparent" : "rgba(99,102,241,0.25)"}`,
                    boxShadow: i === 0 ? "0 0 16px var(--brand-glow)" : "none",
                  }}
                >
                  {s.step}
                </div>

                <h3 className="text-sm font-semibold text-[--fg-primary] mb-2 tracking-tight">
                  {s.title}
                </h3>
                <p className="text-xs text-[--fg-secondary] leading-relaxed">
                  {s.desc}
                </p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
