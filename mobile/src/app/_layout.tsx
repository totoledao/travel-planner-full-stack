import {
  Inter_400Regular,
  Inter_500Medium,
  Inter_600SemiBold,
  useFonts,
} from "@expo-google-fonts/inter";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { Slot } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { StatusBar, View } from "react-native";

import "@/styles/global.css";
import { useCallback } from "react";

// Keep the splash screen visible while fetching resources
SplashScreen.preventAutoHideAsync();

// Add UTC plugin to dayjs
dayjs.extend(utc);

export default function Layout() {
  const [fontsLoaded] = useFonts({
    Inter_400Regular,
    Inter_500Medium,
    Inter_600SemiBold,
  });

  const onLayoutRootView = useCallback(async () => {
    if (fontsLoaded) {
      // Hide the splash screen when done fetching resources
      await SplashScreen.hideAsync();
    }
  }, [fontsLoaded]);

  if (!fontsLoaded) {
    return null;
  }

  return (
    <View className="flex-1 bg-zinc-950" onLayout={onLayoutRootView}>
      <StatusBar
        barStyle={"light-content"}
        backgroundColor="transparent"
        translucent
      />
      <Slot />
    </View>
  );
}
