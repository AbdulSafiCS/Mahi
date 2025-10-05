// app/index.tsx
import { Redirect } from "expo-router";
import { useAuth } from "@/store/auth";

export default function Index() {
  const { accessToken } = useAuth();
  return accessToken ? (
    <Redirect href="/(app)/profile" />
  ) : (
    <Redirect href="/(auth)/login" />
  );
}
