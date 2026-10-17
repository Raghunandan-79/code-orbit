"use client";

import { motion, AnimatePresence } from "motion/react";
import type { ParticipantSnapshot, AwarenessCursor } from "@/lib/types";
import { ParticipantAvatar } from "@/components/ui/ParticipantAvatar";

interface ParticipantsPanelProps {
  open: boolean;
  participants: ParticipantSnapshot[];
  currentParticipantId: string | null;
  cursors: AwarenessCursor[];
  maxParticipants: number;
  onClose: () => void;
}

const CURSOR_COLORS = [
  "#6d6aff", "#34d399", "#fbbf24", "#f87171",
  "#60a5fa", "#f472b6", "#a78bfa", "#fb923c",
];

export function ParticipantsPanel({
  open,
  participants,
  currentParticipantId,
  cursors,
  maxParticipants,
  onClose,
}: ParticipantsPanelProps) {
  const getCursor = (id: string) => cursors.find((c) => c.participantId === id);

  return (
    <>
      {/* Mobile overlay */}
      <AnimatePresence>
        {open && (
          <motion.div
            key="overlay"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-0 bg-black/60 backdrop-blur-sm z-30 sm:hidden"
            onClick={onClose}
            aria-hidden="true"
          />
        )}
      </AnimatePresence>

      {/* Panel */}
      <AnimatePresence>
        {open && (
          <motion.aside
            key="panel"
            initial={{ x: "100%", opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            exit={{ x: "100%", opacity: 0 }}
            transition={{ type: "spring", bounce: 0, duration: 0.35 }}
            className="
              fixed right-0 top-12 bottom-0 z-40
              sm:relative sm:top-auto sm:bottom-auto sm:z-auto
              w-64 shrink-0 flex flex-col overflow-hidden
            "
            style={{
              background: "rgba(14,14,19,0.95)",
              backdropFilter: "blur(24px) saturate(190%)",
              WebkitBackdropFilter: "blur(24px) saturate(190%)",
              borderLeft: "1px solid rgba(255,255,255,0.08)",
              boxShadow: "-8px 0 32px rgba(0,0,0,0.5)",
            }}
            aria-label="Participants"
          >
            {/* Header */}
            <div
              className="shrink-0 flex items-center justify-between px-4 h-12 border-b"
              style={{ borderColor: "rgba(255,255,255,0.07)" }}
            >
              <span className="text-xs font-semibold text-[--fg-secondary] uppercase tracking-wider">
                Participants
              </span>
              <div className="flex items-center gap-2">
                <span className="text-xs text-[--fg-tertiary] tabular-data">
                  {participants.length} / {maxParticipants}
                </span>
                <button
                  onClick={onClose}
                  className="
                    w-7 h-7 rounded-lg flex items-center justify-center
                    text-[--fg-tertiary] hover:text-[--fg-secondary]
                    hover:bg-white/[0.06]
                    transition-all duration-150 active:scale-[0.93]
                    focus-visible:outline-2 focus-visible:outline-[--brand]
                  "
                  aria-label="Close participants panel"
                >
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
                    <path d="M1 1l10 10M11 1L1 11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                  </svg>
                </button>
              </div>
            </div>

            {/* List */}
            <div className="flex-1 overflow-y-auto py-3 px-3 space-y-1">
              {participants.length === 0 && (
                <p className="text-xs text-[--fg-quaternary] text-center mt-8">
                  No participants yet
                </p>
              )}
              {participants.map((p, i) => {
                const isYou = p.participant_id === currentParticipantId;
                const cursor = getCursor(p.participant_id);
                const colorIdx = i % CURSOR_COLORS.length;
                const color = CURSOR_COLORS[colorIdx];

                return (
                  <motion.div
                    key={p.participant_id}
                    initial={{ opacity: 0, x: 12 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: 12 }}
                    transition={{ duration: 0.25, ease: [0.25, 0.46, 0.45, 0.94] }}
                    className="flex items-start gap-2.5 p-2 rounded-[10px] hover:bg-white/[0.04] transition-colors duration-150"
                  >
                    <ParticipantAvatar
                      nickname={p.nickname || "Anonymous"}
                      colorIndex={colorIdx}
                      isYou={isYou}
                      size="md"
                    />
                    <div className="flex-1 min-w-0">
                      <p className="text-xs font-medium text-[--fg-primary] truncate">
                        {p.nickname || "Anonymous"}
                        {isYou && (
                          <span className="ml-1 text-[--fg-quaternary] font-normal">(you)</span>
                        )}
                      </p>
                      {cursor && (
                        <p
                          className="text-[10px] mt-0.5 truncate"
                          style={{ color: `${color}80` }}
                        >
                          Ln {cursor.line}, Col {cursor.column}
                        </p>
                      )}
                      {!cursor && (
                        <p className="text-[10px] text-[--fg-quaternary] mt-0.5">
                          {isYou ? "Editing" : "Viewing"}
                        </p>
                      )}
                    </div>

                    {/* Live indicator */}
                    <div
                      className="w-1.5 h-1.5 rounded-full mt-1.5 shrink-0"
                      style={{ background: color, opacity: cursor ? 1 : 0.3 }}
                      aria-hidden="true"
                    />
                  </motion.div>
                );
              })}
            </div>

            {/* Footer */}
            <div
              className="shrink-0 px-4 py-3 border-t"
              style={{ borderColor: "rgba(255,255,255,0.06)" }}
            >
              <p className="text-[10px] text-[--fg-quaternary] leading-relaxed">
                Up to {maxParticipants} people can collaborate in this room simultaneously.
              </p>
            </div>
          </motion.aside>
        )}
      </AnimatePresence>
    </>
  );
}
