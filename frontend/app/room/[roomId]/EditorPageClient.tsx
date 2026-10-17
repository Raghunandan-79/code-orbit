"use client";

import { useCallback, useRef, useState } from "react";
import { useCollaboration } from "@/hooks/useCollaboration";
import { EditorHeader } from "@/components/editor/EditorHeader";
import { EditorPane } from "@/components/editor/EditorPane";
import { ParticipantsPanel } from "@/components/editor/ParticipantsPanel";
import { NicknameModal } from "@/components/editor/NicknameModal";
import { ShareModal } from "@/components/editor/ShareModal";

interface Props {
  roomId: string;
  initialNickname: string;
}

export function EditorPageClient({ roomId, initialNickname }: Props) {
  const [showNicknameModal, setShowNicknameModal] = useState(!initialNickname);
  const [showShareModal, setShowShareModal] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [resolvedNickname, setResolvedNickname] = useState(() =>
    initialNickname || `Guest ${Math.floor(Math.random() * 9000) + 1000}`
  );

  // Content state: shared between hook and editor
  const pendingUpdateRef = useRef<Uint8Array | null>(null);

  const handleContentUpdate = useCallback((data: Uint8Array) => {
    pendingUpdateRef.current = data;
  }, []);

  const collab = useCollaboration(roomId, resolvedNickname, handleContentUpdate);

  const handleNicknameSet = useCallback((name: string) => {
    const resolved = name.trim() || `Guest ${Math.floor(Math.random() * 9000) + 1000}`;
    setResolvedNickname(resolved);
    setShowNicknameModal(false);
  }, []);

  return (
    <div className="h-screen flex flex-col overflow-hidden" style={{ background: "var(--surface-base)" }}>
      {/* Header */}
      <EditorHeader
        roomId={roomId}
        language={collab.language}
        connectionState={collab.connectionState}
        participantCount={collab.participants.length}
        participants={collab.participants}
        currentParticipantId={collab.participantId}
        onShareClick={() => setShowShareModal(true)}
        onSidebarToggle={() => setSidebarOpen((o) => !o)}
        sidebarOpen={sidebarOpen}
      />

      {/* Body */}
      <div className="flex flex-1 overflow-hidden">
        {/* Editor */}
        <main className="flex-1 overflow-hidden">
          <EditorPane
            language={collab.language}
            collab={collab}
            pendingUpdateRef={pendingUpdateRef}
          />
        </main>

        {/* Participants sidebar */}
        <ParticipantsPanel
          open={sidebarOpen}
          participants={collab.participants}
          currentParticipantId={collab.participantId}
          cursors={collab.cursors}
          maxParticipants={collab.maxParticipants}
          onClose={() => setSidebarOpen(false)}
        />
      </div>

      {/* Modals */}
      {showNicknameModal && (
        <NicknameModal onConfirm={handleNicknameSet} />
      )}
      {showShareModal && (
        <ShareModal
          roomId={roomId}
          onClose={() => setShowShareModal(false)}
        />
      )}
    </div>
  );
}
