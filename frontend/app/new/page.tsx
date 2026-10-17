import type { Metadata } from "next";
import { Suspense } from "react";
import { NewRoomClient } from "./NewRoomClient";

export const metadata: Metadata = {
  title: "New Room",
  description: "Create a new collaborative coding room and invite your team.",
};

export default function NewRoomPage() {
  return (
    <Suspense fallback={<NewRoomFallback />}>
      <NewRoomClient />
    </Suspense>
  );
}

function NewRoomFallback() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[--surface-base]">
      <div className="flex flex-col items-center gap-3">
        <div
          className="w-8 h-8 rounded-full border-2 border-t-[--brand] border-white/10 animate-spin"
          aria-hidden="true"
        />
        <span className="text-xs text-[--fg-tertiary]">Loading workspace setup...</span>
      </div>
    </div>
  );
}
