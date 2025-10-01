import { API_URL } from "./env";
import { useAuth } from "@/store/auth";
import { saveRefreshToken, getRefreshToken, clearRefreshToken } from "./tokens";

// Minimal helpers to call the API without importing api.ts
async function postJson<T>(path: string, body: any): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json();
}

async function getJson<T>(path: string, access?: string): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (access) headers.Authorization = `Bearer ${access}`;
  const res = await fetch(`${API_URL}${path}`, { headers });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json();
}

// Called once from app/_layout.tsx to restore a session
export async function bootstrapSession() {
  const rt = await getRefreshToken();
  if (!rt) return;

  try {
    // Refresh access token
    const refreshed = await postJson<{
      access_token: string;
      refresh_token: string;
    }>("/v1/auth/refresh", { refresh_token: rt });

    await saveRefreshToken(refreshed.refresh_token);

    // Fetch current user with new access token
    const me = await getJson<any>("/v1/users/me", refreshed.access_token);

    // Put in Zustand
    useAuth.getState().setSession(refreshed.access_token, me);
  } catch {
    // If refresh fails, clear stored token
    await clearRefreshToken();
  }
}
