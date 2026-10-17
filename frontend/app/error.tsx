"use client";

import { useEffect } from "react";
import Link from "next/link";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4 text-center">
      <p className="text-label text-[--error] mb-4">Something went wrong</p>
      <h1 className="text-title mb-3">Unexpected error</h1>
      <p className="text-body text-[--fg-secondary] mb-8 max-w-sm">
        {error.message ?? "An unexpected error occurred. Please try again."}
      </p>
      <div className="flex items-center gap-3">
        <button
          onClick={reset}
          className="
            inline-flex items-center gap-2 h-11 px-6 rounded-full
            bg-[--brand] text-white text-sm font-semibold
            hover:bg-[--brand-hover] active:scale-[0.96]
            transition-all duration-100 ease-out
            shadow-[0_0_24px_var(--brand-glow)] hover:shadow-[0_0_36px_var(--brand-glow-strong)]
            focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2
          "
        >
          Try again
        </button>
        <Link
          href="/"
          className="
            inline-flex items-center gap-2 h-11 px-5 rounded-full
            text-sm font-medium text-[--fg-secondary]
            hover:text-[--fg-primary]
            apple-glass
            transition-all duration-150 ease-out active:scale-[0.97]
            focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2
          "
        >
          Go home
        </Link>
      </div>
    </div>
  );
}
