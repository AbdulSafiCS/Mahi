import { useEffect, useState } from "react";
import { Text, View, Pressable, Alert } from "react-native";
import { Link } from "expo-router";
import { apiHealth, apiJson } from "../lib/api";
import { useAuth } from "../store/auth";
import { API_URL } from "../lib/env";

export default function Home() {
  const [healthy, setHealthy] = useState<string>("");
  const { user, accessToken, clearSession } = useAuth();

  useEffect(() => {
    if (!API_URL) return;
    apiHealth()
      .then((ok) => setHealthy(ok ? "ok" : "error"))
      .catch(() => setHealthy("offline"));
  }, []);

  async function callMe() {
    try {
      const me = await apiJson("/v1/users/me", { method: "GET" });
      Alert.alert("Me", JSON.stringify(me, null, 2));
    } catch (e: any) {
      Alert.alert("Error", e?.message ?? "failed");
    }
  }

  function signOut() {
    clearSession();
    Alert.alert("Signed out", "Local tokens cleared.");
  }

  return (
    <View style={{ flex: 1, padding: 20, gap: 12, justifyContent: "center" }}>
      <Text style={{ fontSize: 22, fontWeight: "600" }}>Home</Text>
      <Text>API_URL: {API_URL || "(not set)"}</Text>
      <Text>API health: {healthy || "checking..."}</Text>

      {user ? (
        <>
          <Text>Signed in as: {user.email}</Text>
          <Text numberOfLines={1}>
            Access token: {accessToken ? "(present)" : "(none)"}
          </Text>

          <Pressable
            onPress={callMe}
            style={{ padding: 12, borderWidth: 1, borderRadius: 8 }}
          >
            <Text>Call /v1/users/me</Text>
          </Pressable>

          <Pressable
            onPress={signOut}
            style={{ padding: 12, borderWidth: 1, borderRadius: 8 }}
          >
            <Text>Sign out (client-side)</Text>
          </Pressable>
        </>
      ) : (
        <>
          <Link href="/(auth)/login" asChild>
            <Pressable style={{ padding: 12, borderWidth: 1, borderRadius: 8 }}>
              <Text>Go to Login</Text>
            </Pressable>
          </Link>
          <Link href="/(auth)/register" asChild>
            <Pressable style={{ padding: 12, borderWidth: 1, borderRadius: 8 }}>
              <Text>Create an account</Text>
            </Pressable>
          </Link>
        </>
      )}
    </View>
  );
}
