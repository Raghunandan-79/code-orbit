// Matches backend collaboration.SupportedLanguage
export type SupportedLanguage =
  | "c"
  | "cpp"
  | "java"
  | "python"
  | "go"
  | "javascript"
  | "typescript"
  | "csharp"
  | "rust";

export interface CreateRoomRequest {
  language: SupportedLanguage;
}

export interface CreateRoomResponse {
  room_id: string;
  ws_url: string;
  language: SupportedLanguage;
  document_title: string;
  created_at: string;
  expires_at: string;
  max_participants: number;
}

export interface RoomMetadataResponse {
  room_id: string;
  language: SupportedLanguage;
  document_title: string;
  participant_count: number;
  max_participants: number;
  created_at: string;
  last_active_at: string;
  update_count: number;
}

// WebSocket protocol types
export type WsMessageType =
  | "join_room"
  | "room_joined"
  | "sync_request"
  | "sync_update"
  | "awareness_update"
  | "participant_joined"
  | "participant_left"
  | "room_state"
  | "error"
  | "ping"
  | "pong";

export interface WsEnvelope<T = unknown> {
  version: number;
  type: WsMessageType;
  room_id?: string;
  participant_id?: string;
  doc_id?: string;
  timestamp: number;
  payload?: T;
}

export interface JoinRoomPayload {
  nickname?: string;
}

export interface ParticipantSnapshot {
  participant_id: string;
  nickname: string;
  joined_at: string;
}

export interface RoomJoinedPayload {
  room_id: string;
  participant_id: string;
  language: SupportedLanguage;
  document_title: string;
  created_at: string;
  max_participants: number;
  participants: ParticipantSnapshot[];
}

export interface ParticipantJoinedPayload {
  participant_id: string;
  nickname?: string;
  joined_at: string;
}

export interface ParticipantLeftPayload {
  participant_id: string;
  reason: string;
}

export interface RoomStatePayload {
  room_id: string;
  language: SupportedLanguage;
  participant_count: number;
  max_participants: number;
  participants: ParticipantSnapshot[];
}

export interface ErrorPayload {
  code: string;
  message: string;
}

export interface SyncRequestPayload {
  since_seq: number;
}

export interface SyncUpdatePayload {
  seq: number;
  doc_id?: string;
  data: number[];
}

// Awareness cursor data (client-side)
export interface AwarenessCursor {
  participantId: string;
  nickname: string;
  colorIndex: number;
  line: number;
  column: number;
  selection?: {
    startLine: number;
    startColumn: number;
    endLine: number;
    endColumn: number;
  };
  lastSeen: number;
}
