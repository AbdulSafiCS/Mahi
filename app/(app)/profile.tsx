import { useEffect, useState } from "react";
import { View, Text, ActivityIndicator, Button } from "react-native";
import { apiFetch } from "../../lib/session";
import { useAuth } from "../../store/auth";
import * as SecureStore from "expo-secure-store";

export default function ProfileScreen() {
  const { user, setSession, clearSession } = useAuth();
  const [loading, setLoading] = useState(true);
  const [me, setMe] = useState<any>(user);

  useEffect(() => {
    (async () => {
      try {
        const data = await apiFetch("/v1/users/me");
        setMe(data);
        // keep store user synced
        setSession(useAuth.getState().accessToken!, data);
      } catch (e) {
        console.warn("me failed", e);
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const logout = async () => {
    try {
      const rt = await SecureStore.getItemAsync("refresh_token");
      if (rt) {
        await fetch(`${process.env.EXPO_PUBLIC_API_URL}/v1/auth/logout`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: rt }),
        });
      }
    } finally {
      await SecureStore.deleteItemAsync("refresh_token");
      clearSession();
    }
  };

  if (loading) return <ActivityIndicator />;
  if (!me) return <Text>Not logged in</Text>;

  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: "600" }}>
        Hello, {me.name || me.email}
      </Text>
      <Text style={{ marginTop: 8 }}>User ID: {me.id}</Text>
      <Button title="Log out" onPress={logout} />
    </View>
  );
}
