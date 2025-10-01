import { create } from "zustand";

type User = { id: string; email: string; name?: string };

type AuthState = {
  accessToken: string | null;
  user: User | null;
  setSession: (access: string | null, user: User | null) => void;
  clearSession: () => void;
};

export const useAuth = create<AuthState>(
  (set: (arg0: { accessToken: string | null; user: User | null }) => any) => ({
    accessToken: null,
    user: null,

    // Explicitly type parameters
    setSession: (access: string | null, user: User | null) =>
      set({ accessToken: access, user }),

    clearSession: () => set({ accessToken: null, user: null }),
  })
);
