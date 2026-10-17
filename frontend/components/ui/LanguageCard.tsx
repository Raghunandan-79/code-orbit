"use client";

import { clsx } from "clsx";
import type { SupportedLanguage } from "@/lib/types";
import type { LanguageMeta } from "@/lib/languages";

interface LanguageCardProps {
  language: LanguageMeta;
  selected: boolean;
  onSelect: (id: SupportedLanguage) => void;
}

export function LanguageCard({ language, selected, onSelect }: LanguageCardProps) {
  return (
    <button
      onClick={() => onSelect(language.id)}
      aria-pressed={selected}
      className={clsx(
        "group relative flex flex-col items-center gap-3 p-4 rounded-[14px]",
        "transition-all duration-150 ease-out",
        "select-none touch-manipulation",
        "focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2",
        "active:scale-[0.96] active:duration-[60ms]",
        selected
          ? [
              "bg-[--brand-soft] border border-[--brand]/40",
              "shadow-[0_0_24px_var(--brand-glow)]",
            ]
          : [
              "bg-white/[0.04] border border-white/[0.08]",
              "hover:bg-white/[0.07] hover:border-white/[0.14]",
            ]
      )}
    >
      {/* Language badge */}
      <div
        className={clsx(
          "w-10 h-10 rounded-[10px] flex items-center justify-center",
          "font-mono text-xs font-bold tracking-wider",
          "transition-transform duration-150",
          "group-active:scale-95",
          selected ? "text-white" : "text-[--fg-secondary]"
        )}
        style={{
          background: selected
            ? `${language.color}22`
            : "rgba(255,255,255,0.06)",
          color: selected ? language.color : undefined,
        }}
      >
        {language.icon}
      </div>

      {/* Label */}
      <span
        className={clsx(
          "text-xs font-medium",
          selected ? "text-[--fg-primary]" : "text-[--fg-secondary]"
        )}
      >
        {language.label}
      </span>

      {/* Selected ring */}
      {selected && (
        <span
          className="absolute inset-0 rounded-[14px] pointer-events-none"
          style={{
            boxShadow: `inset 0 0 0 1.5px ${language.color}60`,
          }}
          aria-hidden="true"
        />
      )}
    </button>
  );
}
