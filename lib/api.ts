//import { API_URL } from "./env";
import { useAuth } from "../store/auth";
import { saveRefreshToken, getRefreshToken, clearRefreshToken } from "./tokens";

const API_URL = process.env.EXPO_PUBLIC_API_URL;

function getAccess() {
  return useAuth.getState().accessToken;
}
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

async function asError(res: Response) {
  let body: any = null;
  try {
    body = await res.json();
  } catch {}
  const msg = body?.error || body?.message || `${res.status} ${res.statusText}`;
  const err = new Error(msg) as any;
  err.status = res.status;
  err.body = body;
  return err;
}

// ---- Auth endpoints ----
export async function apiRegister(
  email: string,
  password: string,
  name: string
) {
  const res = await baseFetch(
    "/v1/auth/register",
    {
      method: "POST",
      body: JSON.stringify({ email, password, name }),
    },
    null // no access token for register
  );

  if (!res.ok) throw await asError(res);

  // Expecting { access_token, refresh_token, user }
  const data = await res.json();

  await saveRefreshToken(data.refresh_token);
  // Put access + user in Zustand
  const { setSession } = useAuth.getState();
  setSession(data.access_token, data.user);

  return data;
}

export async function apiLogin(email: string, password: string) {
  const res = await baseFetch(
    "/v1/auth/login",
    {
      method: "POST",
      body: JSON.stringify({ email, password }),
    },
    null
  );
  if (!res.ok) throw await asError(res);
  const data = await res.json();
  await saveRefreshToken(data.refresh_token);
  setAccess(data.access_token, data.user);
  return data;
}

export async function apiRefresh(refreshToken: string) {
  const res = await baseFetch(
    "/v1/auth/refresh",
    {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    },
    null
  );
  if (!res.ok) throw await asError(res);
  return res.json();
}

export async function apiMe(access?: string) {
  const res = await baseFetch(
    "/v1/users/me",
    { method: "GET" },
    access ?? undefined
  );
  if (!res.ok) throw await asError(res);
  return res.json();
}

export async function apiLogout() {
  // client-side: clear tokens; server has no /logout in memory mode
  await clearRefreshToken();
  setAccess(null, null);
}

// JSON helper with auto refresh on 401
export async function apiJson<T = any>(
  path: string,
  init?: RequestInit
): Promise<T> {
  let res = await baseFetch(path, init);
  if (res.status === 401) {
    const rt = await getRefreshToken();
    if (!rt) throw await asError(res);
    const refreshed = await apiRefresh(rt);
    await saveRefreshToken(refreshed.refresh_token);
    setAccess(refreshed.access_token);
    res = await baseFetch(path, init); // retry
  }
  if (!res.ok) throw await asError(res);
  return res.json();
}

export async function apiHealth() {
  const res = await baseFetch("/healthz", { method: "GET" });
  return res.ok;
}
