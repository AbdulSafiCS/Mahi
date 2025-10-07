import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { StatusBar } from "expo-status-bar";

export default function HomeScreen() {
  return (
    <SafeAreaView style={styles.safe}>
      <StatusBar style="light" />

      {/* Header */}
      <View style={styles.headerWrap}>
        <Text style={styles.headerTitle}>Your Homepage</Text>
      </View>
      <View style={styles.wrap}>
        <Text style={styles.title}>Home</Text>
        <Text style={styles.subtle}>Welcome to the app 🎉</Text>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#0b1220", padding: 20 },
  wrap: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: 20,
  },
  title: { color: "white", fontSize: 28, fontWeight: "800" },
  subtle: { color: "#9CA3AF", marginTop: 8, fontSize: 14 },
  headerWrap: { marginBottom: 16 },
  headerTitle: {
    color: "white",
    fontSize: 28,
    fontWeight: "800",
    letterSpacing: 0.3,
  },
  headerSubtitle: { color: "#9CA3AF", marginTop: 4, fontSize: 14 },
});
