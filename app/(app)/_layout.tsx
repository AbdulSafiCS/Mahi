import { Stack, Redirect } from "expo-router";
import { useAuth } from "@/store/auth";

export default function AppLayout() {
  const { accessToken } = useAuth();
  if (!accessToken) return <Redirect href="/(auth)/login" />;
  return (
    <Stack>
      <Stack.Screen name="profile" options={{ title: "Profile" }} />
    </Stack>
  );
}
