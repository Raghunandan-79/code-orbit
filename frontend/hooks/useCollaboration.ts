"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { getRoomWsUrl } from "@/lib/api";
import type {
  AwarenessCursor,
  JoinRoomPayload,
  ParticipantJoinedPayload,
  ParticipantLeftPayload,
  ParticipantSnapshot,
  RoomJoinedPayload,
  SupportedLanguage,
  WsEnvelope,
} from "@/lib/types";

const PROTOCOL_VERSION = 1;
const PING_INTERVAL_MS = 20_000;
const RECONNECT_BASE_MS = 1_000;
const RECONNECT_MAX_MS = 30_000;
const MAX_RECONNECT_ATTEMPTS = 8;
const AWARENESS_TIMEOUT_MS = 30_000;

export type ConnectionState =
  | "idle"
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected"
  | "error";

export interface CollaborationState {
  connectionState: ConnectionState;
  participantId: string | null;
  participants: ParticipantSnapshot[];
  language: SupportedLanguage | null;
  documentTitle: string | null;
  maxParticipants: number;
  cursors: AwarenessCursor[];
  error: string | null;
}

export interface CollaborationActions {
  sendContentUpdate: (data: Uint8Array) => void;
  sendAwareness: (cursor: Omit<AwarenessCursor, "participantId" | "colorIndex" | "lastSeen">) => void;
  connect: () => void;
  disconnect: () => void;
}

const CURSOR_COLORS = [
  "#6d6aff", "#34d399", "#fbbf24", "#f87171",
  "#60a5fa", "#f472b6", "#a78bfa", "#fb923c",
];

let colorCounter = 0;
function assignColorIndex(): number {
  return colorCounter++ % CURSOR_COLORS.length;
}

export function useCollaboration(
  roomId: string | null,
  nickname: string,
  onContentUpdate?: (data: Uint8Array) => void
): CollaborationState & CollaborationActions {
  const wsRef = useRef<WebSocket | null>(null);
  const pingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const colorMapRef = useRef<Map<string, number>>(new Map());
  const mountedRef = useRef(true);
  const connectRef = useRef<() => void>(() => {});

  const [state, setState] = useState<CollaborationState>({
    connectionState: "idle",
    participantId: null,
    participants: [],
    language: null,
    documentTitle: null,
    maxParticipants: 25,
    cursors: [],
    error: null,
  });

  const setPartialState = useCallback((patch: Partial<CollaborationState>) => {
    if (mountedRef.current) {
      setState((prev) => ({ ...prev, ...patch }));
    }
  }, []);

  const getColorIndex = useCallback((participantId: string): number => {
    if (!colorMapRef.current.has(participantId)) {
      colorMapRef.current.set(participantId, assignColorIndex());
    }
    return colorMapRef.current.get(participantId)!;
  }, []);

  const clearTimers = useCallback(() => {
    if (pingTimerRef.current) {
      clearInterval(pingTimerRef.current);
      pingTimerRef.current = null;
    }
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
  }, []);

  const sendEnvelope = useCallback(
    (type: string, payload?: unknown) => {
      const ws = wsRef.current;
      if (!ws || ws.readyState !== WebSocket.OPEN) return;
      const env: WsEnvelope = {
        version: PROTOCOL_VERSION,
        type: type as WsEnvelope["type"],
        timestamp: Date.now(),
        payload,
      };
      ws.send(JSON.stringify(env));
    },
    []
  );

  const handleMessage = useCallback(
    (event: MessageEvent) => {
      // Binary frames = CRDT updates
      if (event.data instanceof ArrayBuffer || event.data instanceof Blob) {
        const toBytes = async (data: Blob | ArrayBuffer) => {
          const buf = data instanceof Blob ? await data.arrayBuffer() : data;
          onContentUpdate?.(new Uint8Array(buf));
        };
        toBytes(event.data);
        return;
      }

      let env: WsEnvelope<unknown>;
      try {
        env = JSON.parse(event.data as string);
      } catch {
        return;
      }

      switch (env.type) {
        case "room_joined": {
          const p = env.payload as RoomJoinedPayload;
          reconnectAttemptsRef.current = 0;
          setPartialState({
            connectionState: "connected",
            participantId: p.participant_id,
            participants: p.participants ?? [],
            language: p.language,
            documentTitle: p.document_title,
            maxParticipants: p.max_participants,
            error: null,
          });
          break;
        }

        case "participant_joined": {
          const p = env.payload as ParticipantJoinedPayload;
          setState((prev) => {
            const already = prev.participants.some(
              (x) => x.participant_id === p.participant_id
            );
            if (already) return prev;
            return {
              ...prev,
              participants: [
                ...prev.participants,
                {
                  participant_id: p.participant_id,
                  nickname: p.nickname ?? "Anonymous",
                  joined_at: p.joined_at,
                },
              ],
            };
          });
          break;
        }

        case "participant_left": {
          const p = env.payload as ParticipantLeftPayload;
          setState((prev) => ({
            ...prev,
            participants: prev.participants.filter(
              (x) => x.participant_id !== p.participant_id
            ),
            cursors: prev.cursors.filter(
              (c) => c.participantId !== p.participant_id
            ),
          }));
          colorMapRef.current.delete(p.participant_id);
          break;
        }

        case "awareness_update": {
          const p = env.payload as {
            participant_id: string;
            nickname?: string;
            line?: number;
            column?: number;
            selection?: AwarenessCursor["selection"];
          };
          if (!p.participant_id) break;
          const colorIndex = getColorIndex(p.participant_id);
          setState((prev) => {
            const existing = prev.cursors.findIndex(
              (c) => c.participantId === p.participant_id
            );
            const updated: AwarenessCursor = {
              participantId: p.participant_id,
              nickname: p.nickname ?? "Anonymous",
              colorIndex,
              line: p.line ?? 1,
              column: p.column ?? 1,
              selection: p.selection,
              lastSeen: Date.now(),
            };
            const cursors =
              existing >= 0
                ? prev.cursors.map((c, i) => (i === existing ? updated : c))
                : [...prev.cursors, updated];
            return { ...prev, cursors };
          });
          break;
        }

        case "sync_update": {
          const p = env.payload as { data?: number[] };
          if (p.data) {
            onContentUpdate?.(new Uint8Array(p.data));
          }
          break;
        }

        case "error": {
          const p = env.payload as { message?: string };
          setPartialState({ error: p.message ?? "Unknown error" });
          break;
        }

        case "pong":
          // heartbeat acknowledged
          break;
      }
    },
    [onContentUpdate, setPartialState, getColorIndex]
  );

  const connect = useCallback(() => {
    if (!roomId) return;
    clearTimers();

    // Close any previous socket cleanly before creating a new one
    if (wsRef.current) {
      try {
        wsRef.current.onclose = null;
        wsRef.current.onerror = null;
        wsRef.current.onmessage = null;
        wsRef.current.close(1000, "Reconnecting");
      } catch {
        // Ignore
      }
      wsRef.current = null;
    }

    const url = getRoomWsUrl(roomId);
    setPartialState({
      connectionState:
        reconnectAttemptsRef.current > 0 ? "reconnecting" : "connecting",
      error: null,
    });

    const ws = new WebSocket(url);
    ws.binaryType = "arraybuffer";
    wsRef.current = ws;

    ws.onopen = () => {
      // Send join_room immediately
      const joinPayload: JoinRoomPayload = { nickname };
      const env: WsEnvelope<JoinRoomPayload> = {
        version: PROTOCOL_VERSION,
        type: "join_room",
        timestamp: Date.now(),
        payload: joinPayload,
      };
      ws.send(JSON.stringify(env));

      // Start ping heartbeat
      pingTimerRef.current = setInterval(() => {
        sendEnvelope("ping");
      }, PING_INTERVAL_MS);
    };

    ws.onmessage = handleMessage;

    ws.onerror = () => {
      setPartialState({ error: "Connection error" });
    };

    ws.onclose = (event) => {
      clearTimers();
      wsRef.current = null;

      if (!mountedRef.current) return;

      // Clean disconnect (code 1000/1001): don't reconnect
      if (event.code === 1000 || event.code === 1001) {
        setPartialState({ connectionState: "disconnected" });
        return;
      }

      // Exponential backoff reconnection
      if (reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
        const delay = Math.min(
          RECONNECT_BASE_MS * Math.pow(2, reconnectAttemptsRef.current),
          RECONNECT_MAX_MS
        );
        reconnectAttemptsRef.current++;
        setPartialState({ connectionState: "reconnecting" });
        reconnectTimerRef.current = setTimeout(() => {
          connectRef.current();
        }, delay);
      } else {
        setPartialState({
          connectionState: "error",
          error: "Failed to reconnect after multiple attempts.",
        });
      }
    };
  }, [roomId, nickname, handleMessage, sendEnvelope, clearTimers, setPartialState]);

  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  const disconnect = useCallback(() => {
    clearTimers();
    if (wsRef.current) {
      try {
        wsRef.current.onclose = null;
        wsRef.current.onerror = null;
        wsRef.current.onmessage = null;
        wsRef.current.close(1000, "User left");
      } catch {
        // Ignore if already closing/closed
      }
      wsRef.current = null;
    }
    setPartialState({ connectionState: "disconnected" });
  }, [clearTimers, setPartialState]);

  const sendContentUpdate = useCallback((data: Uint8Array) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    // Send as binary frame
    ws.send(data.buffer);
  }, []);

  const sendAwareness = useCallback(
    (cursor: Omit<AwarenessCursor, "participantId" | "colorIndex" | "lastSeen">) => {
      sendEnvelope("awareness_update", cursor);
    },
    [sendEnvelope]
  );

  // Prune stale cursors (no update for > 30s)
  useEffect(() => {
    const timer = setInterval(() => {
      setState((prev) => ({
        ...prev,
        cursors: prev.cursors.filter(
          (c) => Date.now() - c.lastSeen < AWARENESS_TIMEOUT_MS
        ),
      }));
    }, 10_000);
    return () => clearInterval(timer);
  }, []);

  // Connect when roomId is provided
  useEffect(() => {
    mountedRef.current = true;
    if (roomId) connect();
    return () => {
      mountedRef.current = false;
      clearTimers();
      if (wsRef.current) {
        try {
          wsRef.current.onclose = null;
          wsRef.current.onerror = null;
          wsRef.current.onmessage = null;
          wsRef.current.close(1000, "Component unmounted");
        } catch {
          // Ignore if already closing/closed
        }
        wsRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roomId]);

  return {
    ...state,
    connect,
    disconnect,
    sendContentUpdate,
    sendAwareness,
  };
}
