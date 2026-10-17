"use client";

import { clsx } from "clsx";
import type { ConnectionState } from "@/hooks/useCollaboration";

const CONFIG: Record<
  ConnectionState,
  { label: string; dotColor: string; pulse: boolean }
> = {
  idle:          { label: "Not connected",  dotColor: "bg-[--fg-quaternary]", pulse: false },
  connecting:    { label: "Connecting…",    dotColor: "bg-[--warning]",       pulse: true  },
  connected:     { label: "Live",           dotColor: "bg-[--success]",       pulse: false },
  reconnecting:  { label: "Reconnecting…", dotColor: "bg-[--warning]",       pulse: true  },
  disconnected:  { label: "Disconnected",  dotColor: "bg-[--fg-quaternary]", pulse: false },
  error:         { label: "Error",         dotColor: "bg-[--error]",         pulse: false },
};

interface ConnectionBadgeProps {
  state: ConnectionState;
  className?: string;
}

export function ConnectionBadge({ state, className }: ConnectionBadgeProps) {
  const { label, dotColor, pulse } = CONFIG[state];

  return (
    <div
      className={clsx(
        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full select-none",
        "bg-white/[0.04] border border-white/[0.07] shadow-[inset_0_1px_0_rgba(255,255,255,0.04)]",
        "text-xs font-medium text-[--fg-secondary]",
        className
      )}
      role="status"
      aria-live="polite"
      aria-label={`Connection status: ${label}`}
    >
      <span
        className={clsx(
          "w-1.5 h-1.5 rounded-full shrink-0",
          dotColor,
          pulse && "animate-pulse",
          state === "connected" && "shadow-[0_0_8px_var(--success)]"
        )}
        aria-hidden="true"
      />
      {label}
    </div>
  );
}
