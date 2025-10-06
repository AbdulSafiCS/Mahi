// app/(auth)/_layout.tsx
import { Stack, Redirect } from "expo-router";
import { useAuth } from "@/store/auth";

export default function AuthLayout() {
  const { accessToken } = useAuth();
  if (accessToken) return <Redirect href="/(app)/profile" />;
  return (
    <Stack>
      <Stack.Screen name="login" options={{ title: "Sign in" }} />
      <Stack.Screen name="register" options={{ title: "Create account" }} />
    </Stack>
  );
}
