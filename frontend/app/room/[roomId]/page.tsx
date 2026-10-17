import type { Metadata } from "next";
import { Suspense } from "react";
import { EditorPageClient } from "./EditorPageClient";

// This page reads searchParams at runtime: must be dynamic
export const dynamic = "force-dynamic";

interface Props {
  params: Promise<{ roomId: string }>;
  searchParams: Promise<{ nick?: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { roomId } = await params;
  return {
    title: `Room ${roomId.slice(0, 8).toUpperCase()}`,
    description: "Collaborative code editing session on Code Orbit.",
    robots: { index: false, follow: false },
  };
}

export default async function RoomPage({ params, searchParams }: Props) {
  const { roomId } = await params;
  const { nick } = await searchParams;
  return (
    <Suspense fallback={<RoomLoadingSkeleton />}>
      <EditorPageClient roomId={roomId} initialNickname={nick ?? ""} />
    </Suspense>
  );
}

function RoomLoadingSkeleton() {
  return (
    <div
      className="h-screen flex items-center justify-center"
      style={{ background: "var(--surface-base)" }}
    >
      <div className="flex flex-col items-center gap-3">
        <div
          className="w-8 h-8 rounded-full border-2 border-t-[--brand] border-white/10 animate-spin"
          aria-hidden="true"
        />
        <span className="text-xs text-[--fg-tertiary]">Connecting to room…</span>
      </div>
    </div>
  );
}
