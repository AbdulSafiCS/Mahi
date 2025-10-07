// lib/api.ts
import { useAuth } from "../store/auth";
import { saveRefreshToken, getRefreshToken, clearRefreshToken } from "./tokens";
import { ApiError } from "./errors"; // <-- use the class above

const API_URL = process.env.EXPO_PUBLIC_API_URL;

function setAccess(token: string | null, user?: any) {
  useAuth.getState().setSession(token, user ?? useAuth.getState().user);
}

export async function baseFetch(
  path: string,
  init: RequestInit = {},
  access?: string | null
) {
  const headers = new Headers(init.headers || {});
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (access) headers.set("Authorization", `Bearer ${access}`);
  const url = `${API_URL}${path}`;
  return fetch(url, { ...init, headers });
}

async function toApiError(res: Response): Promise<ApiError> {
  let body: any = null;
  try {
    body = await res.json();
  } catch {
    // keep body null if not JSON
  }
  // server uses writeErr(w, status, code, details) -> { error: code, details?: any, message?: string }
  const code = body?.error;
  const message =
    body?.message || body?.error || `${res.status} ${res.statusText}`;
  return new ApiError({
    status: res.status,
    message,
    code,
    details: body?.details ?? body,
  });
}

// ---- Auth endpoints ----
export async function apiRegister(
  email: string,
  password: string,
  name: string
) {
  const res = await baseFetch(
    "/v1/auth/register",
    { method: "POST", body: JSON.stringify({ email, password, name }) },
    null
  );
  if (!res.ok) throw await toApiError(res);

  const data = await res.json(); // { access_token, refresh_token, user }
  await saveRefreshToken(data.refresh_token);
  setAccess(data.access_token, data.user);
  return data;
}

export async function apiLogin(email: string, password: string) {
  const res = await baseFetch(
    "/v1/auth/login",
    { method: "POST", body: JSON.stringify({ email, password }) },
    null
  );
  if (!res.ok) throw await toApiError(res);
  const data = await res.json();
  await saveRefreshToken(data.refresh_token);
  setAccess(data.access_token, data.user);
  return data;
}

export async function apiRefresh(refreshToken: string) {
  const res = await baseFetch(
    "/v1/auth/refresh",
    { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) },
    null
  );
  if (!res.ok) throw await toApiError(res);
  return res.json();
}

export async function apiMe(access?: string) {
  const res = await baseFetch(
    "/v1/users/me",
    { method: "GET" },
    access ?? undefined
  );
  if (!res.ok) throw await toApiError(res);
  return res.json();
}

export async function apiLogout() {
  await clearRefreshToken();
  setAccess(null, null);
}
