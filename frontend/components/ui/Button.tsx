"use client";

import { forwardRef } from "react";
import { clsx } from "clsx";

/* Hallmark · component: Button · genre: modern-minimal · theme: Cobalt
 * states: default · hover · focus · active · disabled · loading · error · success
 * contrast: pass (46-50)
 */

type Variant = "primary" | "secondary" | "ghost" | "danger";
type Size = "sm" | "md" | "lg";
type FeedbackState = "default" | "loading" | "error" | "success";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  state?: FeedbackState;
  fullWidth?: boolean;
}

const variantClasses: Record<Variant, string> = {
  primary: [
    "bg-[--brand] text-white",
    "hover:bg-[--brand-hover] hover:shadow-[0_0_28px_var(--brand-glow)]",
    "shadow-[0_0_20px_var(--brand-glow)]",
  ].join(" "),
  secondary: [
    "apple-glass text-[--fg-primary]",
    "hover:bg-white/[0.08] hover:border-white/[0.14]",
  ].join(" "),
  ghost: [
    "bg-transparent text-[--fg-secondary]",
    "hover:text-[--fg-primary] hover:bg-white/[0.06]",
  ].join(" "),
  danger: [
    "bg-[--error]/15 text-[--error] border border-[--error]/25",
    "hover:bg-[--error]/25",
  ].join(" "),
};

const sizeClasses: Record<Size, string> = {
  sm: "h-8 px-3 text-xs gap-1.5 rounded-lg",
  md: "h-10 px-4 text-sm gap-2 rounded-xl",
  lg: "h-12 px-6 text-sm sm:text-base gap-2.5 rounded-xl font-semibold",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = "primary",
      size = "md",
      loading = false,
      state = "default",
      fullWidth = false,
      className,
      children,
      disabled,
      ...props
    },
    ref
  ) => {
    const isWorking = loading || state === "loading";
    const isError = state === "error";
    const isSuccess = state === "success";
    const isDisabled = disabled || isWorking;

    return (
      <button
        ref={ref}
        disabled={isDisabled}
        data-state={isWorking ? "loading" : state}
        className={clsx(
          // Base
          "inline-flex items-center justify-center font-medium select-none touch-manipulation",
          "-webkit-tap-highlight-color-transparent whitespace-nowrap",
          // Apple Pointer-down tactile response
          "transition-all duration-75 ease-out",
          "active:scale-[0.97] active:duration-50",
          // Focus state
          "focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2",
          // Disabled state
          "disabled:opacity-40 disabled:cursor-not-allowed disabled:active:scale-100",
          // Error state override
          isError && "bg-[--error]/20 text-[--error] border border-[--error]/40",
          // Success state override
          isSuccess && "bg-[--success]/20 text-[--success] border border-[--success]/40",
          // Variant & Size
          variantClasses[variant],
          sizeClasses[size],
          fullWidth && "w-full",
          className
        )}
        {...props}
      >
        {isWorking ? (
          <>
            <Spinner size={size === "sm" ? 12 : 14} />
            <span className="opacity-80">{children}</span>
          </>
        ) : isSuccess ? (
          <>
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
              <path d="M2.5 7.5L5.5 10.5L11.5 3.5" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <span>{children}</span>
          </>
        ) : isError ? (
          <>
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
              <path d="M7 3v5M7 11h.01" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round"/>
            </svg>
            <span>{children}</span>
          </>
        ) : (
          children
        )}
      </button>
    );
  }
);

Button.displayName = "Button";

function Spinner({ size }: { size: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 16 16"
      fill="none"
      className="animate-spin"
      aria-hidden="true"
    >
      <circle
        cx="8"
        cy="8"
        r="6"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeDasharray="28"
        strokeDashoffset="14"
        opacity="0.35"
      />
      <path
        d="M8 2a6 6 0 0 1 6 6"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}
