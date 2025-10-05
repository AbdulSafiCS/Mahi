// lib/session.ts
import { API_URL } from "./env";
import { useAuth } from "@/store/auth";
import { saveRefreshToken, getRefreshToken, clearRefreshToken } from "./tokens";

/**
 * POST helper for small calls without bringing in api.ts
 */
async function postJson<T>(path: string, body: any): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json();
}

/**
 * GET helper with optional Bearer token
 */
async function getJson<T>(path: string, access?: string): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (access) headers.Authorization = `Bearer ${access}`;
  const res = await fetch(`${API_URL}${path}`, { headers });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json();
}

/**
 * apiFetch:
 * - Adds Authorization automatically (if we have an access token)
 * - If the server returns 401, tries to refresh using the stored refresh_token,
 *   updates state, and retries the original request once.
 */
export async function apiFetch(input: string, init: RequestInit = {}) {
  const { accessToken, setSession, clearSession, user } = useAuth.getState();

  const makeHeaders = (token?: string) => {
    const h = new Headers(init.headers || {});
    if (!h.has("Content-Type")) h.set("Content-Type", "application/json");
    if (token) h.set("Authorization", `Bearer ${token}`);
    return h;
  };

  // First attempt with the current access token (if present)
  let res = await fetch(`${API_URL}${input}`, {
    ...init,
    headers: makeHeaders(accessToken || undefined),
  });

  // If unauthorized, try to refresh and retry once
  if (res.status === 401) {
    const rt = await getRefreshToken();
    if (!rt) {
      clearSession();
      throw new Error("Unauthorized (no refresh token)");
    }

    const refreshRes = await fetch(`${API_URL}/v1/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: rt }),
    });

    if (!refreshRes.ok) {
      // Refresh failed → clear everything
      await clearRefreshToken();
      clearSession();
      const txt = await refreshRes.text().catch(() => "");
      throw new Error(`Session refresh failed: ${refreshRes.status} ${txt}`);
    }

    // Update local session with the new tokens
    const data = await refreshRes.json();
    const newAccess: string = data.access_token;
    const newRefresh: string = data.refresh_token;
    await saveRefreshToken(newRefresh);
    setSession(newAccess, user); // keep existing user; callers can refetch /me if they need fresh user

    // Retry original request exactly once with the new access token
    res = await fetch(`${API_URL}${input}`, {
      ...init,
      headers: makeHeaders(newAccess),
    });
  }

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`${res.status} ${text || res.statusText}`);
  }
  // Try to parse JSON; if body is empty, return null
  const text = await res.text();
  return text ? JSON.parse(text) : null;
}

/**
 * Called once from app/_layout.tsx to restore a session at app launch.
 * - Uses stored refresh token to mint a new access token
 * - Then fetches /v1/users/me and hydrates Zustand
 */
export async function bootstrapSession() {
  const rt = await getRefreshToken();
  if (!rt) return;

  try {
    // 1) Refresh access token
    const refreshed = await postJson<{
      access_token: string;
      refresh_token: string;
    }>("/v1/auth/refresh", { refresh_token: rt });

    await saveRefreshToken(refreshed.refresh_token);

    // 2) Fetch current user with new access token
    const me = await getJson<any>("/v1/users/me", refreshed.access_token);

    // 3) Hydrate Zustand
    useAuth.getState().setSession(refreshed.access_token, me);
  } catch {
    // If refresh fails, clear stored token
    await clearRefreshToken();
  }
}

/**
 * Logout:
 * - Best-effort POST to /v1/auth/logout with the stored refresh token
 * - Always clears local tokens and Zustand session
 */
export async function logout() {
  const rt = await getRefreshToken();
  if (rt) {
    try {
      await fetch(`${API_URL}/v1/auth/logout`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: rt }),
      });
    } catch (err) {
      console.warn("Logout request failed:", err);
    }
  }
  await clearRefreshToken();
  useAuth.getState().clearSession();
}
