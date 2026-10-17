"use client";

import { useState, useCallback, useRef, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { motion, AnimatePresence } from "motion/react";
import { api, ApiError } from "@/lib/api";
import { LANGUAGES } from "@/lib/languages";
import type { SupportedLanguage } from "@/lib/types";
import { Button } from "@/components/ui/Button";
import { OrbitIcon } from "@/components/icons/OrbitIcon";
import { clsx } from "clsx";

const SPRING      = { type: "spring" as const, bounce: 0,   duration: 0.45 };
const SPRING_CARD = { type: "spring" as const, bounce: 0.1, duration: 0.35 };

export function NewRoomClient() {
  const router       = useRouter();
  const searchParams = useSearchParams();
  const paramLang    = searchParams.get("lang") as SupportedLanguage | null;
  const initialLang  = paramLang && LANGUAGES.some((l) => l.id === paramLang) ? paramLang : "javascript";

  const [selectedLang, setSelectedLang] = useState<SupportedLanguage>(initialLang);
  const [nickname,     setNickname]     = useState("");
  const [loading,      setLoading]      = useState(false);
  const [error,        setError]        = useState<string | null>(null);
  const [inputFocused, setInputFocused] = useState(false);

  const nicknameRef = useRef<HTMLInputElement>(null);

  // Auto-focus nickname on mount
  useEffect(() => {
    nicknameRef.current?.focus();
  }, []);

  const handleCreate = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const room = await api.createRoom({ language: selectedLang });
      const nickParam = nickname.trim()
        ? `?nick=${encodeURIComponent(nickname.trim())}`
        : "";
      router.push(`/room/${room.room_id}${nickParam}`);
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? err.message
          : "Something went wrong. Please try again.";
      setError(msg);
      setLoading(false);
    }
  }, [selectedLang, nickname, router]);

  const selectedMeta = LANGUAGES.find((l) => l.id === selectedLang)!;

  return (
    <div className="min-h-screen flex flex-col bg-[--surface-base] relative overflow-hidden">
      {/* ── Page-level ambient glow ── */}
      <div
        className="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[400px] pointer-events-none animate-ambient-pulse"
        style={{
          background: "radial-gradient(ellipse 60% 60% at 50% 0%, rgba(99,102,241,0.12) 0%, transparent 70%)",
        }}
        aria-hidden="true"
      />

      {/* ── Minimal Header ── */}
      <header className="relative z-10 px-4 sm:px-6 h-14 flex items-center border-b border-white/[0.06] bg-black/20 backdrop-blur-md">
        <Link
          href="/"
          className="flex items-center gap-2 rounded-full py-1 px-1 focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2 transition-transform duration-75 active:scale-[0.97]"
          aria-label="Code Orbit home"
        >
          <OrbitIcon size={24} />
          <span className="text-sm font-medium text-[--fg-secondary]">Code Orbit</span>
          <span className="text-xs text-[--fg-quaternary]">/</span>
          <span className="text-xs text-[--fg-tertiary]">New Room</span>
        </Link>
      </header>

      {/* ── Main Centered Workbench ── */}
      <main className="relative z-10 flex-1 flex items-center justify-center px-4 sm:px-6 py-12">
        <div className="w-full max-w-md">
          {/* ── Card ── */}
          <div
            className="p-6 sm:p-8 rounded-3xl relative overflow-hidden"
            style={{
              background: "var(--surface-card)",
              border: "1px solid rgba(255,255,255,0.09)",
              boxShadow: "0 32px 80px rgba(0,0,0,0.7), inset 0 1px 0 rgba(255,255,255,0.09)",
            }}
          >
            {/* Inner ambient — top */}
            <div
              className="absolute top-0 left-0 right-0 h-32 pointer-events-none opacity-50"
              style={{
                background: "radial-gradient(ellipse 80% 100% at 50% 0%, rgba(99,102,241,0.08) 0%, transparent 70%)",
              }}
              aria-hidden="true"
            />

            {/* Header copy */}
            <div className="mb-6 relative z-10">
              <p className="text-label text-[--brand] mb-1.5">Setup</p>
              <h1 className="text-headline">Configure your workspace</h1>
              <p className="text-xs text-[--fg-tertiary] mt-1">
                Select an environment and enter a nickname to begin.
              </p>
            </div>

            {/* ── Name input ── */}
            <section
              className="mb-6 relative z-10"
              aria-labelledby="name-label"
            >
              <label
                id="name-label"
                htmlFor="nickname"
                className="block text-xs font-medium mb-2 text-[--fg-secondary]"
              >
                Your nickname
              </label>
              <div className="relative">
                <input
                  ref={nicknameRef}
                  id="nickname"
                  type="text"
                  value={nickname}
                  onChange={(e) => setNickname(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && !loading && handleCreate()}
                  onFocus={() => setInputFocused(true)}
                  onBlur={() => setInputFocused(false)}
                  placeholder="Anonymous coder"
                  maxLength={30}
                  autoComplete="off"
                  spellCheck={false}
                  className="
                    w-full h-11 pl-10 pr-3 rounded-xl
                    text-sm text-[--fg-primary] placeholder:text-[--fg-quaternary]
                    outline-none transition-all duration-150 ease-out
                    bg-white/[0.04] border
                  "
                  style={{
                    borderColor: inputFocused
                      ? "rgba(99,102,241,0.55)"
                      : "rgba(255,255,255,0.08)",
                    boxShadow: inputFocused
                      ? "0 0 0 3px rgba(99,102,241,0.12), inset 0 1px 0 rgba(255,255,255,0.04)"
                      : "none",
                  }}
                />
                <svg
                  className="absolute left-3.5 top-1/2 -translate-y-1/2 pointer-events-none transition-colors duration-150"
                  style={{ color: inputFocused ? "rgba(99,102,241,0.6)" : "rgba(255,255,255,0.25)" }}
                  width="16" height="16" viewBox="0 0 16 16" fill="none"
                  aria-hidden="true"
                >
                  <circle cx="8" cy="5" r="2.5" stroke="currentColor" strokeWidth="1.25"/>
                  <path d="M3 13c0-2.761 2.239-5 5-5s5 2.239 5 5" stroke="currentColor" strokeWidth="1.25" strokeLinecap="round"/>
                </svg>
              </div>
            </section>

            {/* ── Language picker ── */}
            <section
              className="mb-6 relative z-10"
              aria-labelledby="lang-label"
            >
              <div className="flex items-center justify-between mb-2.5">
                <label id="lang-label" className="text-xs font-medium text-[--fg-secondary]">
                  Environment
                </label>
                <AnimatePresence mode="wait">
                  <motion.span
                    key={selectedLang}
                    initial={{ opacity: 0, scale: 0.88, y: -2 }}
                    animate={{ opacity: 1, scale: 1,    y: 0 }}
                    exit={{ opacity: 0, scale: 0.88, y: 2 }}
                    transition={{ duration: 0.14 }}
                    className="text-[10px] font-mono font-semibold px-2 py-0.5 rounded-full"
                    style={{
                      background: `${selectedMeta.color}18`,
                      color: selectedMeta.color,
                      border: `1px solid ${selectedMeta.color}30`,
                    }}
                  >
                    {selectedMeta.label}
                  </motion.span>
                </AnimatePresence>
              </div>

              <div
                className="grid grid-cols-3 gap-2"
                role="listbox"
                aria-labelledby="lang-label"
              >
                {LANGUAGES.map((lang) => {
                  const isSelected = selectedLang === lang.id;
                  return (
                    <button
                      key={lang.id}
                      onClick={() => setSelectedLang(lang.id)}
                      role="option"
                      aria-selected={isSelected}
                      className={clsx(
                        "relative flex flex-col items-center gap-1.5 py-3 px-2 rounded-xl",
                        "transition-all duration-100 ease-out overflow-hidden",
                        "select-none touch-manipulation active:scale-[0.94]",
                        "focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2",
                        isSelected
                          ? "bg-white/[0.08] text-[--fg-primary]"
                          : "bg-white/[0.02] text-[--fg-tertiary] hover:bg-white/[0.05] hover:text-[--fg-secondary] hover:-translate-y-0.5"
                      )}
                      style={{
                        border: isSelected
                          ? `1px solid ${lang.color}60`
                          : "1px solid rgba(255,255,255,0.06)",
                        boxShadow: isSelected
                          ? `0 0 16px ${lang.color}20`
                          : "none",
                      }}
                    >
                      <span
                        className="w-7 h-7 rounded-lg flex items-center justify-center text-[10px] font-mono font-bold transition-transform duration-100 group-active:scale-90"
                        style={{
                          background: isSelected ? `${lang.color}24` : "rgba(255,255,255,0.04)",
                          color: isSelected ? lang.color : "rgba(255,255,255,0.40)",
                        }}
                      >
                        {lang.icon}
                      </span>
                      <span className="text-[11px] font-medium leading-none">
                        {lang.label}
                      </span>

                      {/* Layout-animated selection ring */}
                      {isSelected && (
                        <motion.span
                          layoutId="lang-indicator"
                          className="absolute inset-0 rounded-xl pointer-events-none"
                          style={{ boxShadow: `0 0 16px ${lang.color}25` }}
                          transition={SPRING_CARD}
                          aria-hidden="true"
                        />
                      )}

                      {/* Hover shimmer sweep */}
                      {!isSelected && (
                        <span
                          className="absolute inset-0 rounded-xl pointer-events-none opacity-0 hover:opacity-100 transition-opacity duration-200 animate-shimmer"
                          aria-hidden="true"
                        />
                      )}
                    </button>
                  );
                })}
              </div>
            </section>

            {/* ── Error message ── */}
            <AnimatePresence>
              {error && (
                <motion.div
                  initial={{ opacity: 0, height: 0, marginBottom: 0 }}
                  animate={{ opacity: 1, height: "auto", marginBottom: 16 }}
                  exit={{ opacity: 0, height: 0, marginBottom: 0 }}
                  transition={SPRING}
                  className="px-3.5 py-2.5 rounded-xl text-xs bg-[--error]/10 border border-[--error]/25 text-[--error] relative z-10"
                  role="alert"
                >
                  {error}
                </motion.div>
              )}
            </AnimatePresence>

            {/* ── Create Button ── */}
            <div className="relative z-10">
              <Button
                variant="primary"
                size="lg"
                fullWidth
                loading={loading}
                onClick={handleCreate}
                className="animate-glow-pulse"
              >
                {loading ? "Initializing room…" : "Create room"}
              </Button>

              <p className="text-center text-[11px] mt-4 text-[--fg-quaternary]">
                A shareable link is generated automatically
              </p>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
