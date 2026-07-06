// API 客户端 — 与 Go 后端通信

const BASE = ""; // 同源，无需前缀

interface RoomStatus {
  connected: boolean;
  room_code?: string;
  connect_code?: string;
  connection_type: string;
  nat_type?: string;
  players: string[];
  mc_version: string;
  server_running?: boolean;
  step?: string;
  mode?: string;
  need_open_lan?: boolean;
  punch_code?: string;
  punch_required?: boolean;
  instruction?: string;
}

interface ServerStatus {
  running: boolean;
  version: string;
  tps: number;
  memory_usage: number;
  uptime_secs: number;
  players_online: number;
  max_players: number;
}

// 通用 fetch 封装
async function post<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(BASE + path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error((err as { error: string }).error || "请求失败");
  }
  return res.json();
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(BASE + path);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error((err as { error: string }).error || "请求失败");
  }
  return res.json();
}

// ============ 房间 ============

export async function createRoom(params: {
  game_mode?: string;
  room_name?: string;
  mc_version?: string;
  mode?: string;
}): Promise<RoomStatus> {
  return post<RoomStatus>("/api/room/create", params);
}

export async function joinRoom(code: string): Promise<RoomStatus> {
  return post<RoomStatus>("/api/room/join", { code });
}

export async function leaveRoom(): Promise<void> {
  await post("/api/room/leave");
}

export async function getRoomStatus(): Promise<RoomStatus> {
  return get<RoomStatus>("/api/room/status");
}

// ============ 连接码 ============

export async function generateConnectCode(params: {
  mc_version?: string;
  mod_hash?: string;
}): Promise<{ code: string; nat_type: string; public_ip: string }> {
  return post("/api/connect/generate", params);
}

export async function parseConnectCode(code: string): Promise<{
  version: string;
  endpoints: { ip: string; port: number }[];
  mod_hash: string;
}> {
  return post("/api/connect/parse", { code });
}

// ============ 服务器 ============

export async function startServer(params: {
  mc_version?: string;
  game_mode?: string;
}): Promise<ServerStatus> {
  return post<ServerStatus>("/api/server/start", params);
}

export async function stopServer(): Promise<void> {
  await post("/api/server/stop");
}

export async function getServerStatus(): Promise<ServerStatus> {
  return get<ServerStatus>("/api/server/status");
}

// ============ 隧道 ============

export async function getTunnelStatus(): Promise<{
  connected: boolean;
  connection_type: string;
  latency_ms: number;
}> {
  return get("/api/tunnel/status");
}

// ============ 版本检测 ============

interface MCVersion {
  id: string;
  type: string;
}

export async function detectVersions(): Promise<{
  versions: MCVersion[];
  latest: string;
  running: string;
  from_process: boolean;
  minecraft_dir: string;
}> {
  return get("/api/versions");
}

// ============ 配置 ============

export async function getConfig(): Promise<Record<string, unknown>> {
  return get("/api/config");
}

export async function saveConfig(config: Record<string, unknown>): Promise<void> {
  await post("/api/config", config);
}
