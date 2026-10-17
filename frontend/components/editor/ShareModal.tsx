"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "motion/react";
import { Button } from "@/components/ui/Button";

interface ShareModalProps {
  roomId: string;
  onClose: () => void;
}

// Apple materialize: blur radius + scale together so the surface reads as a real material arriving
const MATERIALIZE = {
  type: "spring" as const,
  bounce: 0,
  duration: 0.38,
};

export function ShareModal({ roomId, onClose }: ShareModalProps) {
  const [copied, setCopied] = useState(false);
  const url = typeof window !== "undefined" ? `${window.location.origin}/room/${roomId}` : "";

  const handleCopy = async () => {
    if (!url) return;
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      const input = document.getElementById("share-url-input") as HTMLInputElement | null;
      input?.select();
    }
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) onClose();
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="share-modal-title"
      onClick={handleBackdropClick}
    >
      {/* ── Backdrop scrim with gradient ── */}
      <motion.div
        className="absolute inset-0"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.2 }}
        style={{
          background: "radial-gradient(ellipse 80% 60% at 50% 60%, rgba(9,9,11,0.75) 0%, rgba(0,0,0,0.65) 100%)",
          backdropFilter: "blur(6px) saturate(120%)",
          WebkitBackdropFilter: "blur(6px) saturate(120%)",
        }}
        aria-hidden="true"
      />

      {/* ── Materialize: scale + blur together on enter ── */}
      <motion.div
        initial={{ opacity: 0, scale: 0.94, y: 16, filter: "blur(8px)" }}
        animate={{ opacity: 1, scale: 1,    y: 0,  filter: "blur(0px)" }}
        exit={  { opacity: 0, scale: 0.96, y: 8,   filter: "blur(4px)" }}
        transition={MATERIALIZE}
        className="relative w-full max-w-sm rounded-[20px] p-6 sm:p-8"
        style={{
          background: "rgba(18, 18, 22, 0.96)",
          backdropFilter: "blur(40px) saturate(200%)",
          WebkitBackdropFilter: "blur(40px) saturate(200%)",
          border: "1px solid rgba(255,255,255,0.10)",
          boxShadow: "0 24px 80px rgba(0,0,0,0.75), inset 0 1px 0 rgba(255,255,255,0.08)",
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Top ambient glow on the panel */}
        <div
          className="absolute top-0 left-0 right-0 h-20 rounded-t-[20px] pointer-events-none"
          style={{
            background: "radial-gradient(ellipse 80% 100% at 50% 0%, rgba(99,102,241,0.06) 0%, transparent 100%)",
          }}
          aria-hidden="true"
        />

        {/* Header */}
        <div className="relative flex items-start justify-between mb-5">
          <div>
            <h2
              id="share-modal-title"
              className="text-base font-semibold text-[--fg-primary] mb-0.5"
            >
              Share this room
            </h2>
            <p className="text-sm text-[--fg-secondary]">
              Anyone with this link can join and collaborate.
            </p>
          </div>
          <button
            onClick={onClose}
            className="
              w-7 h-7 rounded-lg flex items-center justify-center ml-3
              text-[--fg-tertiary] hover:text-[--fg-secondary]
              hover:bg-white/[0.06]
              transition-colors duration-150
              focus-visible:outline-2 focus-visible:outline-[--brand]
              active:scale-[0.93]
            "
            aria-label="Close"
          >
            <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
              <path d="M1 1l10 10M11 1L1 11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
            </svg>
          </button>
        </div>

        {/* URL input + copy */}
        <div className="relative flex gap-2 mb-5">
          <input
            id="share-url-input"
            type="text"
            readOnly
            value={url}
            className="
              flex-1 h-10 px-3 rounded-xl
              text-xs font-mono text-[--fg-secondary]
              outline-none cursor-text select-all
              transition-all duration-150
            "
            style={{
              background: "rgba(255,255,255,0.05)",
              border: "1px solid rgba(255,255,255,0.08)",
            }}
            onFocus={(e) => {
              (e.target as HTMLInputElement).select();
              e.currentTarget.style.borderColor = "rgba(99,102,241,0.5)";
              e.currentTarget.style.boxShadow = "0 0 0 3px rgba(99,102,241,0.14)";
            }}
            onBlur={(e) => {
              e.currentTarget.style.borderColor = "rgba(255,255,255,0.08)";
              e.currentTarget.style.boxShadow = "none";
            }}
            aria-label="Room URL"
          />
          <Button
            variant={copied ? "ghost" : "primary"}
            size="md"
            onClick={handleCopy}
            className="shrink-0"
            aria-live="polite"
            aria-label={copied ? "Copied!" : "Copy link"}
          >
            <AnimatePresence mode="wait">
              {copied ? (
                <motion.span
                  key="copied"
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.15 }}
                  className="text-[--success]"
                >
                  Copied!
                </motion.span>
              ) : (
                <motion.span
                  key="copy"
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.15 }}
                >
                  Copy
                </motion.span>
              )}
            </AnimatePresence>
          </Button>
        </div>

        {/* Room ID display */}
        <div
          className="rounded-[10px] px-3 py-2.5 flex items-center justify-between"
          style={{
            background: "rgba(255,255,255,0.03)",
            border: "1px solid rgba(255,255,255,0.06)",
          }}
        >
          <span className="text-xs text-[--fg-tertiary]">Room ID</span>
          <span className="text-xs font-mono text-[--fg-secondary]">{roomId}</span>
        </div>
      </motion.div>
    </div>
  );
}
