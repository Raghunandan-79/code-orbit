"use client";

import { clsx } from "clsx";

const CURSOR_COLORS = [
  "#6d6aff", "#34d399", "#fbbf24", "#f87171",
  "#60a5fa", "#f472b6", "#a78bfa", "#fb923c",
];

interface ParticipantAvatarProps {
  nickname: string;
  colorIndex: number;
  isYou?: boolean;
  size?: "sm" | "md";
  showName?: boolean;
}

function initials(name?: string | null): string {
  if (!name || typeof name !== "string") return "?";
  const clean = name.trim();
  if (!clean) return "?";
  return (
    clean
      .split(/\s+/)
      .slice(0, 2)
      .map((w) => w[0]?.toUpperCase() ?? "")
      .join("") || "?"
  );
}

export function ParticipantAvatar({
  nickname,
  colorIndex = 0,
  isYou = false,
  size = "md",
  showName = false,
}: ParticipantAvatarProps) {
  const safeName = (nickname && typeof nickname === "string" && nickname.trim()) || "Guest";
  const color = CURSOR_COLORS[Math.abs(colorIndex || 0) % CURSOR_COLORS.length];
  const abbr = initials(safeName);

  const avatarSize = size === "sm" ? "w-7 h-7 text-[10px]" : "w-8 h-8 text-xs";

  return (
    <div
      className={clsx(
        "flex items-center gap-2",
        showName && "min-w-0"
      )}
      title={isYou ? `${safeName} (you)` : safeName}
    >
      <div
        className={clsx(
          "relative rounded-full flex items-center justify-center",
          "font-mono font-semibold shrink-0 select-none",
          "ring-2 ring-[--surface-base]",
          avatarSize
        )}
        style={{
          background: `${color}20`,
          color,
          border: `1px solid ${color}40`,
        }}
      >
        {abbr}
        {isYou && (
          <span
            className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full bg-[--success] ring-2 ring-[--surface-base] shadow-[0_0_6px_var(--success)]"
            aria-label="You"
          />
        )}
      </div>
      {showName && (
        <span className="text-xs text-[--fg-secondary] truncate">
          {isYou ? `${safeName} (you)` : safeName}
        </span>
      )}
    </div>
  );
}

/** Stacked avatar group: shows first N + overflow count */
export function ParticipantAvatarStack({
  participants,
  currentParticipantId,
  maxVisible = 4,
}: {
  participants: Array<{ participant_id: string; nickname?: string }>;
  currentParticipantId?: string | null;
  maxVisible?: number;
}) {
  const visible = participants.slice(0, maxVisible);
  const overflow = participants.length - maxVisible;

  return (
    <div className="flex items-center -space-x-2" aria-label={`${participants.length} participants`}>
      {visible.map((p, i) => (
        <div key={p.participant_id || i} style={{ zIndex: visible.length - i }}>
          <ParticipantAvatar
            nickname={p.nickname || "Guest"}
            colorIndex={i}
            isYou={p.participant_id === currentParticipantId}
            size="sm"
          />
        </div>
      ))}
      {overflow > 0 && (
        <div
          className="relative w-7 h-7 rounded-full bg-white/[0.08] border border-white/10 flex items-center justify-center text-[10px] font-medium text-[--fg-secondary]"
          style={{ zIndex: 0 }}
        >
          +{overflow}
        </div>
      )}
    </div>
  );
}
