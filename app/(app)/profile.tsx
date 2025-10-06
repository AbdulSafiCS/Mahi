// app/(app)/profile.tsx
import React, { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  Alert,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  View,
  Pressable,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import * as SecureStore from "expo-secure-store";
import { apiFetch } from "@/lib/session";
import { useAuth } from "@/store/auth";

/**
 * ProfileScreen
 * - Loads /v1/users/me using the current access token
 * - Shows a nice profile card with name/email & ID
 * - Pull-to-refresh to re-fetch "me"
 * - Log out clears refresh token on server (best-effort) + local state
 */
export default function ProfileScreen() {
  const { user, setSession, clearSession } = useAuth();

  const [me, setMe] = useState<any>(user);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  // Fetch the current user from the API
  const loadMe = useCallback(async () => {
    try {
      const data = await apiFetch("/v1/users/me");
      setMe(data);
      // keep our Zustand store in sync with the latest user payload
      setSession(useAuth.getState().accessToken!, data);
    } catch (e) {
      console.warn("me failed", e);
      Alert.alert("Couldn’t load profile", "Please try again.");
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [setSession]);

  useEffect(() => {
    loadMe();
  }, [loadMe]);

  // Pull-to-refresh handler
  const onRefresh = useCallback(() => {
    setRefreshing(true);
    loadMe();
  }, [loadMe]);

  // Log out:
  // 1) Try to inform the server (best-effort)
  // 2) Always clear local secure storage + Zustand session
  const logout = useCallback(async () => {
    try {
      const rt = await SecureStore.getItemAsync("refresh_token");
      if (rt) {
        await fetch(`${process.env.EXPO_PUBLIC_API_URL}/v1/auth/logout`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: rt }),
        });
      }
    } catch (err) {
      console.warn("logout warn:", err);
    } finally {
      await SecureStore.deleteItemAsync("refresh_token");
      clearSession();
    }
  }, [clearSession]);

  if (loading) {
    return (
      <SafeAreaView style={styles.center}>
        <ActivityIndicator size="large" />
        <Text style={styles.loadingText}>Loading your profile…</Text>
      </SafeAreaView>
    );
  }

  if (!me) {
    return (
      <SafeAreaView style={styles.center}>
        <Text style={styles.title}>Not signed in</Text>
        <Text style={styles.subtle}>
          Please log in again to view your profile.
        </Text>
      </SafeAreaView>
    );
  }

  const initials = (
    me.name
      ?.trim()
      ?.split(/\s+/)
      .map((s: string) => s[0])
      .join("") ||
    me.email?.[0] ||
    "U"
  )?.toUpperCase();

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView
        contentContainerStyle={styles.scroll}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
      >
        {/* Header */}
        <View style={styles.headerWrap}>
          <Text style={styles.headerTitle}>Your Profile</Text>
          <Text style={styles.headerSubtitle}>Welcome back</Text>
        </View>

        {/* Card */}
        <View style={styles.card}>
          {/* Avatar circle with initials */}
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{initials}</Text>
          </View>

          <Text style={styles.nameText}>{me.name || "Unnamed User"}</Text>
          <Text style={styles.emailText}>{me.email}</Text>

          <View style={styles.divider} />

          <Row label="User ID" value={me.id} />
          <Row label="Full name" value={me.name || "-"} />
          <Row label="Email" value={me.email} />
          <Row label="Plan" value="Free" />
          <Row label="Member since" value={me.createdAt} />
          <View style={styles.actions}>
            <Pressable
              onPress={logout}
              style={({ pressed }) => [
                styles.button,
                styles.buttonDanger,
                pressed && styles.buttonPressed,
              ]}
              accessibilityRole="button"
              accessibilityLabel="Log out"
            >
              <Text style={styles.buttonText}>Log out</Text>
            </Pressable>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

/** Small reusable key/value row */
function Row({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.row}>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={styles.rowValue} numberOfLines={2}>
        {value}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#0b1220" },
  scroll: { padding: 20, paddingBottom: 32 },
  center: {
    flex: 1,
    backgroundColor: "#0b1220",
    alignItems: "center",
    justifyContent: "center",
    gap: 12,
  },
  loadingText: { color: "#c7d2fe", fontSize: 14 },

  title: {
    fontSize: 22,
    fontWeight: "700",
    color: "white",
    marginBottom: 4,
    textAlign: "center",
  },
  subtle: {
    fontSize: 14,
    color: "#9CA3AF",
    textAlign: "center",
  },
  // ------------------------------------

  headerWrap: { marginBottom: 16 },
  headerTitle: {
    color: "white",
    fontSize: 28,
    fontWeight: "800",
    letterSpacing: 0.3,
  },
  headerSubtitle: { color: "#9CA3AF", marginTop: 4, fontSize: 14 },

  card: {
    backgroundColor: "#0f172a",
    borderRadius: 20,
    padding: 20,
    borderWidth: 1,
    borderColor: "#1f2a44",
    shadowColor: "#000",
    shadowOpacity: 0.2,
    shadowRadius: 12,
    shadowOffset: { width: 0, height: 8 },
    elevation: 6,
  },

  avatar: {
    width: 80,
    height: 80,
    borderRadius: 40,
    alignSelf: "center",
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "#1d4ed8",
    marginBottom: 14,
  },
  avatarText: { color: "white", fontWeight: "800", fontSize: 28 },

  nameText: {
    color: "white",
    fontSize: 20,
    fontWeight: "700",
    textAlign: "center",
  },
  emailText: {
    color: "#c7d2fe",
    fontSize: 14,
    textAlign: "center",
    marginTop: 4,
  },

  divider: {
    height: 1,
    backgroundColor: "#1f2a44",
    marginVertical: 16,
  },

  row: {
    paddingVertical: 10,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: "#1f2a44",
  },
  rowLabel: { color: "#9CA3AF", fontSize: 13, marginBottom: 4 },
  rowValue: { color: "white", fontSize: 15, fontWeight: "600" },

  actions: { marginTop: 18, gap: 12 },
  button: {
    height: 48,
    borderRadius: 12,
    alignItems: "center",
    justifyContent: "center",
  },
  buttonDanger: {
    backgroundColor: "#ef4444",
  },
  buttonPressed: {
    opacity: 0.85,
    transform: [{ scale: 0.99 }],
  },
  buttonText: {
    color: "white",
    fontSize: 16,
    fontWeight: "700",
    letterSpacing: 0.3,
  },
});
