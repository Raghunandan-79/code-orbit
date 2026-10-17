"use client";

import { useState } from "react";
import Link from "next/link";
import { motion, AnimatePresence } from "motion/react";
import { LANGUAGES } from "@/lib/languages";

// Apple spring presets
const SPRING       = { type: "spring" as const, bounce: 0,    duration: 0.45 };

interface SampleSnippet {
  filename: string;
  langLabel: string;
  color: string;
  lines: Array<{
    num: number;
    tokens: Array<{ text: string; color: string }>;
    cursor?: { name: string; color: string; label?: string };
    highlight?: boolean;
  }>;
}

const SNIPPETS: Record<string, SampleSnippet> = {
  typescript: {
    filename: "session.ts",
    langLabel: "TypeScript",
    color: "#3178c6",
    lines: [
      {
        num: 1,
        tokens: [
          { text: "import", color: "#c792ea" },
          { text: " { ", color: "#89ddff" },
          { text: "useCollaboration", color: "#82aaff" },
          { text: " } ", color: "#89ddff" },
          { text: "from", color: "#c792ea" },
          { text: " ", color: "#ffffff" },
          { text: "'@/hooks/collab'", color: "#c3e88d" },
          { text: ";", color: "#89ddff" },
        ],
      },
      { num: 2, tokens: [] },
      {
        num: 3,
        tokens: [
          { text: "export function", color: "#c792ea" },
          { text: " ", color: "#ffffff" },
          { text: "PairSession", color: "#ffcb6b" },
          { text: "({ ", color: "#89ddff" },
          { text: "roomId", color: "#f78c6c" },
          { text: " }: ", color: "#89ddff" },
          { text: "SessionProps", color: "#ffcb6b" },
          { text: ") {", color: "#89ddff" },
        ],
      },
      {
        num: 4,
        tokens: [
          { text: "  const", color: "#c792ea" },
          { text: " { editor, cursors, peers } = ", color: "#89ddff" },
          { text: "useCollaboration", color: "#82aaff" },
          { text: "(roomId);", color: "#89ddff" },
        ],
        cursor: { name: "You", color: "#6366f1" },
      },
      {
        num: 5,
        tokens: [
          { text: "  ", color: "#ffffff" },
          { text: "// Sub-50ms synchronized delta broadcast across active peers", color: "rgba(255,255,255,0.3)" },
        ],
      },
      {
        num: 6,
        tokens: [
          { text: "  return", color: "#c792ea" },
          { text: " (", color: "#89ddff" },
        ],
      },
      {
        num: 7,
        tokens: [
          { text: "    <", color: "#89ddff" },
          { text: "MonacoWorkspace", color: "#ffcb6b" },
          { text: " sync=", color: "#89ddff" },
          { text: "{editor.state}", color: "#f78c6c" },
          { text: " awareness=", color: "#89ddff" },
          { text: "{cursors}", color: "#f78c6c" },
          { text: " />", color: "#89ddff" },
        ],
        cursor: { name: "Elena", color: "#34d399", label: "editing" },
        highlight: true,
      },
      {
        num: 8,
        tokens: [
          { text: "  );", color: "#89ddff" },
        ],
      },
      {
        num: 9,
        tokens: [
          { text: "}", color: "#89ddff" },
        ],
      },
    ],
  },
  python: {
    filename: "orchestrator.py",
    langLabel: "Python",
    color: "#3572a5",
    lines: [
      {
        num: 1,
        tokens: [
          { text: "from", color: "#c792ea" },
          { text: " dataclasses ", color: "#82aaff" },
          { text: "import", color: "#c792ea" },
          { text: " dataclass", color: "#ffcb6b" },
        ],
      },
      {
        num: 2,
        tokens: [
          { text: "from", color: "#c792ea" },
          { text: " typing ", color: "#82aaff" },
          { text: "import", color: "#c792ea" },
          { text: " List, Optional", color: "#ffcb6b" },
        ],
      },
      { num: 3, tokens: [] },
      {
        num: 4,
        tokens: [
          { text: "@dataclass", color: "#ffcb6b" },
        ],
      },
      {
        num: 5,
        tokens: [
          { text: "class", color: "#c792ea" },
          { text: " RoomSession", color: "#ffcb6b" },
          { text: ":", color: "#89ddff" },
        ],
        cursor: { name: "Marcus", color: "#fbbf24", label: "reviewing" },
      },
      {
        num: 6,
        tokens: [
          { text: "    room_id: ", color: "#89ddff" },
          { text: "str", color: "#f78c6c" },
        ],
      },
      {
        num: 7,
        tokens: [
          { text: "    participants: ", color: "#89ddff" },
          { text: "List[str]", color: "#f78c6c" },
        ],
      },
      {
        num: 8,
        tokens: [
          { text: "    def broadcast(self, delta: bytes) -> None:", color: "#82aaff" },
        ],
        cursor: { name: "You", color: "#6366f1" },
      },
      {
        num: 9,
        tokens: [
          { text: "        # In-memory distribution to connected websockets", color: "rgba(255,255,255,0.3)" },
        ],
      },
    ],
  },
  go: {
    filename: "hub.go",
    langLabel: "Go",
    color: "#00add8",
    lines: [
      {
        num: 1,
        tokens: [
          { text: "package", color: "#c792ea" },
          { text: " collaboration", color: "#ffcb6b" },
        ],
      },
      { num: 2, tokens: [] },
      {
        num: 3,
        tokens: [
          { text: "type", color: "#c792ea" },
          { text: " Hub ", color: "#ffcb6b" },
          { text: "struct", color: "#c792ea" },
          { text: " {", color: "#89ddff" },
        ],
      },
      {
        num: 4,
        tokens: [
          { text: "    rooms      ", color: "#ffffff" },
          { text: "map[string]*Room", color: "#f78c6c" },
        ],
      },
      {
        num: 5,
        tokens: [
          { text: "    broadcast  ", color: "#ffffff" },
          { text: "chan []byte", color: "#f78c6c" },
        ],
        cursor: { name: "Elena", color: "#34d399" },
      },
      {
        num: 6,
        tokens: [
          { text: "    register   ", color: "#ffffff" },
          { text: "chan *Client", color: "#f78c6c" },
        ],
      },
      {
        num: 7,
        tokens: [
          { text: "}", color: "#89ddff" },
        ],
      },
      {
        num: 8,
        tokens: [
          { text: "func (h *Hub) Run() {", color: "#82aaff" },
        ],
        cursor: { name: "You", color: "#6366f1" },
      },
      {
        num: 9,
        tokens: [
          { text: "    // Concurrent non-blocking lockless multiplexing", color: "rgba(255,255,255,0.3)" },
        ],
      },
    ],
  },
  rust: {
    filename: "engine.rs",
    langLabel: "Rust",
    color: "#ce412b",
    lines: [
      {
        num: 1,
        tokens: [
          { text: "pub struct", color: "#c792ea" },
          { text: " PeerState", color: "#ffcb6b" },
          { text: " {", color: "#89ddff" },
        ],
      },
      {
        num: 2,
        tokens: [
          { text: "    pub id: ", color: "#89ddff" },
          { text: "String", color: "#ffcb6b" },
          { text: ",", color: "#89ddff" },
        ],
      },
      {
        num: 3,
        tokens: [
          { text: "    pub cursor: ", color: "#89ddff" },
          { text: "Option<CursorPosition>", color: "#ffcb6b" },
          { text: ",", color: "#89ddff" },
        ],
        cursor: { name: "Marcus", color: "#fbbf24" },
      },
      {
        num: 4,
        tokens: [
          { text: "}", color: "#89ddff" },
        ],
      },
      { num: 5, tokens: [] },
      {
        num: 6,
        tokens: [
          { text: "impl", color: "#c792ea" },
          { text: " PeerState {", color: "#89ddff" },
        ],
      },
      {
        num: 7,
        tokens: [
          { text: "    pub fn new(id: impl Into<String>) -> Self {", color: "#82aaff" },
        ],
        cursor: { name: "You", color: "#6366f1" },
      },
      {
        num: 8,
        tokens: [
          { text: "        Self { id: id.into(), cursor: None }", color: "#ffffff" },
        ],
      },
      {
        num: 9,
        tokens: [
          { text: "    }", color: "#89ddff" },
        ],
      },
    ],
  },
};

export function WorkbenchHero() {
  const [selectedLang, setSelectedLang] = useState<string>("typescript");
  const snippet = SNIPPETS[selectedLang] ?? SNIPPETS.typescript;

  return (
    <section
      id="workbench"
      className="relative w-full pt-16 sm:pt-24 pb-20 overflow-hidden"
      aria-label="Workbench"
    >
      {/* ── Animated ambient glow blob ── */}
      <div
        className="absolute top-0 left-1/2 -translate-x-1/2 w-[720px] h-[420px] pointer-events-none animate-ambient-pulse"
        style={{
          background: "radial-gradient(ellipse 60% 60% at 50% 0%, rgba(99, 102, 241, 0.18) 0%, rgba(99,102,241,0.04) 55%, transparent 80%)",
          filter: "blur(1px)",
        }}
        aria-hidden="true"
      />
      {/* Secondary accent blob */}
      <div
        className="absolute top-24 left-[30%] w-[340px] h-[220px] pointer-events-none opacity-25 animate-ambient-pulse"
        style={{
          background: "radial-gradient(ellipse, rgba(52, 211, 153, 0.12) 0%, transparent 70%)",
          animationDelay: "1.8s",
        }}
        aria-hidden="true"
      />

      <div className="relative max-w-5xl mx-auto px-4 sm:px-6">
        {/* ── Staggered text entrance ── */}
        <div className="max-w-2xl mb-12">
          {/* Badge */}
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-[11px] font-medium mb-5 bg-white/[0.04] border border-white/[0.08] text-[--fg-secondary]">
            <span className="w-1.5 h-1.5 rounded-full bg-[--success] animate-dot-pulse" />
            <span>Zero configuration pair programming</span>
          </div>

          {/* Headline — Solid ink (no tacky AI purple text gradient per Hallmark gate 46) */}
          <h1 className="text-display mb-4 text-[--fg-primary]">
            Multiplayer code editor.{" "}
            <span className="text-white">
              In your browser.
            </span>
          </h1>

          {/* Sub-copy */}
          <p className="text-body-lg mb-8 max-w-xl text-[--fg-secondary]">
            Type simultaneously with your team in Monaco. Zero setup, zero accounts, sub-50 millisecond synchronization across nine languages.
          </p>

          {/* Action Row */}
          <div className="flex flex-wrap items-center gap-3.5">
            <Link
              href="/new"
              className="
                inline-flex items-center gap-2 h-11 px-6 rounded-full
                bg-[--brand] text-white text-sm font-semibold
                hover:bg-[--brand-hover] transition-all duration-100 ease-out
                active:scale-[0.96] whitespace-nowrap
                shadow-[0_0_24px_var(--brand-glow)] hover:shadow-[0_0_36px_var(--brand-glow-strong)]
                focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2
              "
            >
              <span>Start coding now</span>
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
                <path d="M2.5 7h9M8 3.5l3.5 3.5L8 10.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </Link>
            <a
              href="#capabilities"
              className="
                inline-flex items-center gap-1.5 h-11 px-5 rounded-full
                text-sm font-medium text-[--fg-secondary] hover:text-[--fg-primary]
                bg-white/[0.03] hover:bg-white/[0.07] border border-white/[0.08]
                hover:border-white/[0.16]
                transition-all duration-150 ease-out active:scale-[0.97]
                whitespace-nowrap
                focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2
              "
            >
              <span>Explore capabilities</span>
            </a>
          </div>
        </div>

        {/* ── Workbench surface ── */}
        <div
          className="relative rounded-2xl overflow-hidden"
          style={{
            background: "var(--surface-card)",
            border: "1px solid rgba(255,255,255,0.09)",
            boxShadow: "0 32px 100px rgba(0,0,0,0.85), inset 0 1px 0 rgba(255,255,255,0.08)",
          }}
        >
          {/* Top Workbench Navigation & Control Bar */}
          <div className="flex items-center justify-between px-3 sm:px-4 h-12 bg-black/40 border-b border-white/[0.08]">
            {/* Language Tabs */}
            <div className="flex items-center gap-1 overflow-x-auto no-scrollbar py-1" role="tablist">
              {(["typescript", "python", "go", "rust"] as const).map((langKey) => {
                const isActive = selectedLang === langKey;
                const langMeta = LANGUAGES.find((l) => l.id === langKey);
                return (
                  <button
                    key={langKey}
                    role="tab"
                    aria-selected={isActive}
                    onClick={() => setSelectedLang(langKey)}
                    className={`
                      relative flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-mono
                      transition-all duration-100 ease-out active:scale-[0.96] whitespace-nowrap
                      ${isActive ? "text-[--fg-primary] bg-white/[0.08]" : "text-[--fg-tertiary] hover:text-[--fg-secondary] hover:bg-white/[0.03]"}
                    `}
                  >
                    <span
                      className="w-2 h-2 rounded-full"
                      style={{ background: langMeta?.color ?? "#6366f1" }}
                    />
                    <span>{snippet.filename.split(".")[0]}.{langMeta?.extension.replace(".", "")}</span>
                    {isActive && (
                      <motion.div
                        layoutId="activeTabPill"
                        className="absolute inset-0 rounded-lg border border-white/[0.12] pointer-events-none"
                        transition={SPRING}
                      />
                    )}
                  </button>
                );
              })}
            </div>

            {/* Room Status */}
            <div className="hidden sm:flex items-center gap-3 text-xs font-mono text-[--fg-tertiary]">
              <span className="inline-flex items-center gap-1.5 text-[--success]">
                <span className="w-1.5 h-1.5 rounded-full bg-[--success] animate-dot-pulse" />
                Live Sync
              </span>
              <span>3 peers active</span>
              <span className="px-2 py-0.5 rounded bg-white/[0.05] text-[--fg-secondary] tabular-data">24ms</span>
            </div>
          </div>

          {/* Code Surface */}
          <div className="relative flex min-h-[290px] sm:min-h-[320px] bg-[#0c0c0f] font-mono text-xs sm:text-[13px] leading-relaxed select-none">
            {/* Gutter */}
            <div className="w-10 sm:w-12 py-4 select-none text-right pr-3 sm:pr-4 border-r border-white/[0.05] text-[--fg-quaternary]">
              {snippet.lines.map((line) => (
                <div key={line.num} className="leading-6 sm:leading-7 tabular-data">
                  {line.num}
                </div>
              ))}
            </div>

            {/* Code Body */}
            <div className="flex-1 py-4 px-4 sm:px-6 overflow-x-auto">
              <AnimatePresence mode="wait">
                <motion.div
                  key={selectedLang}
                  initial={false}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -4 }}
                  transition={{ duration: 0.18 }}
                >
                  {snippet.lines.map((line) => (
                    <div
                      key={line.num}
                      className={`relative flex items-center h-6 sm:h-7 whitespace-pre ${
                        line.highlight ? "bg-white/[0.035] -mx-4 sm:-mx-6 px-4 sm:px-6 rounded" : ""
                      }`}
                    >
                      {line.tokens.length === 0 ? (
                        <span>&nbsp;</span>
                      ) : (
                        line.tokens.map((tok, j) => (
                          <span key={j} style={{ color: tok.color }}>{tok.text}</span>
                        ))
                      )}

                      {/* Live Cursor Chip with pulse caret */}
                      {line.cursor && (
                        <span className="relative inline-flex items-center ml-1">
                          <span
                            className="inline-block w-[2px] h-4 align-middle cursor-blink"
                            style={{ background: line.cursor.color }}
                          />
                          <span
                            className="absolute -top-5 left-0 px-1.5 py-0.5 rounded text-[9px] font-sans font-semibold text-black tracking-tight pointer-events-none shadow-sm whitespace-nowrap"
                            style={{ background: line.cursor.color }}
                          >
                            {line.cursor.name}
                            {line.cursor.label ? ` · ${line.cursor.label}` : ""}
                          </span>
                        </span>
                      )}
                    </div>
                  ))}
                </motion.div>
              </AnimatePresence>
            </div>
          </div>

          {/* Workbench Footer Bar */}
          <div className="flex flex-col sm:flex-row items-center justify-between px-4 py-3 bg-[#111115] border-t border-white/[0.06] gap-3 text-xs">
            <div className="flex items-center gap-2 text-[--fg-tertiary]">
              <span className="font-medium text-[--fg-secondary]">{snippet.langLabel}</span>
              <span>·</span>
              <span>Monaco Kernel</span>
              <span>·</span>
              <span>Shared buffer</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-[--fg-tertiary] hidden md:inline">Ready to collaborate?</span>
              <Link
                href={`/new?lang=${selectedLang}`}
                className="
                  inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg
                  bg-white/[0.06] hover:bg-white/[0.12] text-[--fg-primary] font-medium
                  border border-white/[0.10] hover:border-white/[0.20]
                  transition-all duration-150 ease-out
                  active:scale-[0.96] whitespace-nowrap
                "
              >
                <span>Launch this {snippet.langLabel} room</span>
                <span className="text-[--fg-tertiary]">→</span>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
