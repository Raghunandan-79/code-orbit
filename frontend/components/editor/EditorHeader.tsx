"use client";

import Link from "next/link";
import { OrbitIcon } from "@/components/icons/OrbitIcon";
import { ConnectionBadge } from "@/components/ui/ConnectionBadge";
import { ParticipantAvatarStack } from "@/components/ui/ParticipantAvatar";
import type { ConnectionState } from "@/hooks/useCollaboration";
import type { SupportedLanguage } from "@/lib/types";
import type { ParticipantSnapshot } from "@/lib/types";
import { getLanguageMeta } from "@/lib/languages";
import { clsx } from "clsx";

interface EditorHeaderProps {
  roomId: string;
  language: SupportedLanguage | null;
  connectionState: ConnectionState;
  participantCount: number;
  participants: ParticipantSnapshot[];
  currentParticipantId: string | null;
  onShareClick: () => void;
  onSidebarToggle: () => void;
  sidebarOpen: boolean;
}

export function EditorHeader({
  roomId,
  language,
  connectionState,
  participantCount,
  participants,
  currentParticipantId,
  onShareClick,
  onSidebarToggle,
  sidebarOpen,
}: EditorHeaderProps) {
  const langMeta = language ? getLanguageMeta(language) : null;
  const shortId = roomId.slice(0, 8).toUpperCase();

  return (
    <header
      className="shrink-0 h-12 flex items-center gap-2 px-3 sm:px-4 relative z-20"
      style={{
        background: "rgba(9,9,13,0.88)",
        backdropFilter: "blur(24px) saturate(190%)",
        WebkitBackdropFilter: "blur(24px) saturate(190%)",
        borderBottom: "1px solid rgba(255,255,255,0.08)",
        boxShadow: "0 2px 12px rgba(0,0,0,0.4)",
      }}
    >
      {/* Logo */}
      <Link
        href="/"
        className="flex items-center shrink-0 p-1 rounded-md focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2 active:scale-[0.96] transition-transform duration-75"
        aria-label="Code Orbit home"
      >
        <OrbitIcon size={22} />
      </Link>

      <span className="text-[--fg-quaternary] text-xs mx-1" aria-hidden="true">/</span>

      {/* Room ID */}
      <div
        className="hidden sm:inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-white/[0.04] border border-white/[0.06] text-xs font-mono text-[--fg-secondary]"
        title={`Room ID: ${roomId}`}
      >
        <span className="text-[10px] text-[--fg-quaternary] uppercase tracking-wider font-sans">Room</span>
        <span className="font-semibold text-[--fg-primary]">{shortId}</span>
      </div>

      {/* Language badge */}
      {langMeta && (
        <span
          className="hidden sm:inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md text-[10px] font-mono font-semibold"
          style={{
            background: `${langMeta.color}15`,
            color: langMeta.color,
            border: `1px solid ${langMeta.color}30`,
          }}
        >
          {langMeta.icon}
          <span className="hidden md:inline">{langMeta.label}</span>
        </span>
      )}

      {/* Spacer */}
      <div className="flex-1" />

      {/* Connection badge */}
      <ConnectionBadge state={connectionState} className="hidden sm:flex" />

      {/* Participants */}
      {participants.length > 0 && (
        <button
          onClick={onSidebarToggle}
          className={clsx(
            "flex items-center gap-2 px-2.5 py-1 rounded-lg transition-all duration-150 active:scale-[0.96]",
            "focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2",
            sidebarOpen
              ? "bg-[--brand-soft] text-[--brand] border border-[--brand]/30"
              : "bg-white/[0.03] hover:bg-white/[0.07] border border-white/[0.06] text-[--fg-secondary]"
          )}
          aria-label={`${participantCount} participants, toggle panel`}
          aria-expanded={sidebarOpen}
        >
          <ParticipantAvatarStack
            participants={participants}
            currentParticipantId={currentParticipantId}
            maxVisible={3}
          />
          <span className="text-xs font-medium hidden sm:block tabular-data">
            {participantCount}
          </span>
        </button>
      )}

      {/* Share button */}
      <button
        onClick={onShareClick}
        className="
          inline-flex items-center gap-1.5 h-7.5 px-3 rounded-lg text-xs font-medium
          bg-[--brand-soft] text-[--brand]
          hover:bg-[--brand]/20
          border border-[--brand]/25
          shadow-[0_0_14px_var(--brand-glow)]
          transition-all duration-150
          active:scale-[0.96]
          focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2
        "
        aria-label="Share room"
      >
        <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
          <path d="M9 4a1.5 1.5 0 100-3 1.5 1.5 0 000 3zM3 7.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3zM9 11a1.5 1.5 0 100-3 1.5 1.5 0 000 3z" stroke="currentColor" strokeWidth="1.2"/>
          <path d="M4.5 6.75l3 1.5M7.5 3.75l-3 1.5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round"/>
        </svg>
        <span className="hidden sm:inline">Share</span>
      </button>
    </header>
  );
}
