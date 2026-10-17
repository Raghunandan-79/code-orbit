import type {
  CreateRoomRequest,
  CreateRoomResponse,
  RoomMetadataResponse,
} from "@/lib/types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(
  path: string,
  init?: RequestInit
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init,
  });

  if (!res.ok) {
    let message = `HTTP ${res.status}`;
    let code = String(res.status);
    try {
      const body = await res.json();
      message = body.error ?? message;
      code = body.code ?? code;
    } catch {
      // ignore parse errors
    }
    throw new ApiError(res.status, code, message);
  }

  return res.json() as Promise<T>;
}

export const api = {
  /**
   * POST /api/v1/rooms
   * Creates a new collaborative room for the given language.
   */
  createRoom(data: CreateRoomRequest): Promise<CreateRoomResponse> {
    return request<CreateRoomResponse>("/api/v1/rooms", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  /**
   * GET /api/v1/rooms/:roomID
   * Fetches metadata for an existing room.
   */
  getRoom(roomId: string): Promise<RoomMetadataResponse> {
    return request<RoomMetadataResponse>(`/api/v1/rooms/${encodeURIComponent(roomId)}`);
  },
};

export { ApiError };

/**
 * Builds the WebSocket URL for a given room.
 * Converts http(s):// to ws(s):// automatically.
 */
export function getRoomWsUrl(roomId: string): string {
  const base = (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080")
    .replace(/^http/, "ws");
  return `${base}/api/v1/rooms/${encodeURIComponent(roomId)}/ws`;
}
