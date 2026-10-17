"use client";

import { useState, useRef, useEffect } from "react";
import { motion } from "motion/react";
import { Button } from "@/components/ui/Button";

interface NicknameModalProps {
  onConfirm: (nickname: string) => void;
}

// Apple materialize: blur radius + scale together
const MATERIALIZE = {
  type: "spring" as const,
  bounce: 0,
  duration: 0.38,
};

export function NicknameModal({ onConfirm }: NicknameModalProps) {
  const [value,   setValue]   = useState("");
  const [focused, setFocused] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    // Slight delay so the materialize animation starts before focus fires
    const t = setTimeout(() => inputRef.current?.focus(), 120);
    return () => clearTimeout(t);
  }, []);

  const handleSubmit = () => {
    onConfirm(value.trim());
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="nickname-modal-title"
    >
      {/* ── Backdrop scrim gradient ── */}
      <motion.div
        className="absolute inset-0"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.2 }}
        style={{
          background: "radial-gradient(ellipse 80% 60% at 50% 55%, rgba(9,9,11,0.72) 0%, rgba(0,0,0,0.60) 100%)",
          backdropFilter: "blur(6px) saturate(120%)",
          WebkitBackdropFilter: "blur(6px) saturate(120%)",
        }}
        aria-hidden="true"
      />

      {/* ── Materialize panel ── */}
      <motion.div
        initial={{ opacity: 0, scale: 0.94, y: 16, filter: "blur(8px)" }}
        animate={{ opacity: 1, scale: 1,    y: 0,  filter: "blur(0px)" }}
        transition={MATERIALIZE}
        className="relative w-full max-w-sm rounded-[20px] p-6 sm:p-8"
        style={{
          background: "rgba(18, 18, 22, 0.96)",
          backdropFilter: "blur(40px) saturate(200%)",
          WebkitBackdropFilter: "blur(40px) saturate(200%)",
          border: "1px solid rgba(255,255,255,0.10)",
          boxShadow: "0 24px 80px rgba(0,0,0,0.75), inset 0 1px 0 rgba(255,255,255,0.08)",
        }}
      >
        {/* Top ambient glow */}
        <div
          className="absolute top-0 left-0 right-0 h-20 rounded-t-[20px] pointer-events-none"
          style={{
            background: "radial-gradient(ellipse 80% 100% at 50% 0%, rgba(99,102,241,0.06) 0%, transparent 100%)",
          }}
          aria-hidden="true"
        />

        {/* Header */}
        <div className="relative mb-6">
          {/* Icon badge */}
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center mb-4"
            style={{
              background: "rgba(99,102,241,0.12)",
              border: "1px solid rgba(99,102,241,0.25)",
            }}
          >
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none" aria-hidden="true">
              <circle cx="9" cy="6" r="3" stroke="#6366f1" strokeWidth="1.5"/>
              <path d="M3.5 15.5c0-3.038 2.462-5.5 5.5-5.5s5.5 2.462 5.5 5.5" stroke="#6366f1" strokeWidth="1.5" strokeLinecap="round"/>
            </svg>
          </div>

          <h2
            id="nickname-modal-title"
            className="text-base font-semibold text-[--fg-primary] mb-1"
          >
            What should we call you?
          </h2>
          <p className="text-sm text-[--fg-secondary]">
            Your name is visible to other participants in this room.
          </p>
        </div>

        {/* Input with focus glow */}
        <div className="relative mb-4">
          <input
            ref={inputRef}
            type="text"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
            onFocus={() => setFocused(true)}
            onBlur={() => setFocused(false)}
            placeholder="Your name…"
            maxLength={30}
            autoComplete="off"
            spellCheck={false}
            className="
              w-full h-11 px-3.5 rounded-xl
              text-sm text-[--fg-primary] placeholder:text-[--fg-quaternary]
              outline-none transition-all duration-150
            "
            style={{
              background: "rgba(255,255,255,0.06)",
              borderWidth: "1px",
              borderStyle: "solid",
              borderColor: focused ? "rgba(99,102,241,0.55)" : "rgba(255,255,255,0.10)",
              boxShadow: focused
                ? "0 0 0 3px rgba(99,102,241,0.14)"
                : "none",
            }}
          />
        </div>

        {/* Actions */}
        <div className="flex gap-2.5">
          <Button
            variant="ghost"
            size="md"
            onClick={() => onConfirm("")}
            className="flex-1"
          >
            Skip
          </Button>
          <Button
            variant="primary"
            size="md"
            onClick={handleSubmit}
            className="flex-1 shadow-[0_0_20px_var(--brand-glow)]"
          >
            Join room
          </Button>
        </div>
      </motion.div>
    </div>
  );
}
