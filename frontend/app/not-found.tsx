import Link from "next/link";

export default function NotFound() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4 text-center">
      {/* Glow */}
      <div
        className="absolute top-1/4 left-1/2 -translate-x-1/2 w-96 h-96 rounded-full pointer-events-none opacity-10"
        style={{
          background: "radial-gradient(circle, #6d6aff 0%, transparent 70%)",
          filter: "blur(60px)",
        }}
        aria-hidden="true"
      />

      <p className="text-label text-[--brand] mb-4">404</p>
      <h1 className="text-title mb-3">Page not found</h1>
      <p className="text-body text-[--fg-secondary] mb-8 max-w-sm">
        This room may have expired, or the link is invalid.
      </p>
      <Link
        href="/"
        className="
          inline-flex items-center gap-2 h-11 px-6 rounded-full
          bg-[--brand] text-white text-sm font-semibold
          hover:bg-[--brand-hover] active:scale-[0.96]
          transition-all duration-100 ease-out
          shadow-[0_0_24px_var(--brand-glow)] hover:shadow-[0_0_36px_var(--brand-glow-strong)]
          focus-visible:outline-2 focus-visible:outline-[--brand] focus-visible:outline-offset-2
        "
      >
        Go home
      </Link>
    </div>
  );
}
