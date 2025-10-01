import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { View, Text, TextInput, Pressable, Alert } from "react-native";
import { apiLogin } from "../../lib/api";
import { useRouter } from "expo-router";

const LoginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(6, "Min 6 characters"),
});
type LoginValues = z.infer<typeof LoginSchema>;

export default function Login() {
  const router = useRouter();
  const {
    register,
    setValue,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({
    resolver: zodResolver(LoginSchema),
    defaultValues: { email: "", password: "" },
  });

  const onSubmit = async (values: LoginValues) => {
    try {
      await apiLogin(values.email, values.password);
      router.replace("/");
    } catch (e: any) {
      Alert.alert("Login failed", e?.message ?? "Please try again");
    }
  };

  return (
    <View style={{ flex: 1, justifyContent: "center", padding: 20, gap: 12 }}>
      <Text style={{ fontSize: 24, fontWeight: "600" }}>Login</Text>

      <TextInput
        placeholder="Email"
        autoCapitalize="none"
        keyboardType="email-address"
        style={{ borderWidth: 1, borderRadius: 8, padding: 12 }}
        {...register("email")}
        onChangeText={(t) => setValue("email", t, { shouldValidate: true })}
      />
      {!!errors.email && (
        <Text style={{ color: "red" }}>{errors.email.message}</Text>
      )}

      <TextInput
        placeholder="Password"
        secureTextEntry
        style={{ borderWidth: 1, borderRadius: 8, padding: 12 }}
        {...register("password")}
        onChangeText={(t) => setValue("password", t, { shouldValidate: true })}
      />
      {!!errors.password && (
        <Text style={{ color: "red" }}>{errors.password.message}</Text>
      )}

      <Pressable
        onPress={handleSubmit(onSubmit)}
        disabled={isSubmitting}
        style={{
          backgroundColor: "black",
          padding: 14,
          borderRadius: 10,
          opacity: isSubmitting ? 0.6 : 1,
        }}
      >
        <Text
          style={{ color: "white", textAlign: "center", fontWeight: "600" }}
        >
          {isSubmitting ? "Signing in…" : "Sign in"}
        </Text>
      </Pressable>
    </View>
  );
}
