import { Image, Text, View } from "react-native";

import logo from "@/assets/logo.png";

export default function index() {
  return (
    <View className="flex-1 justify-center items-center">
      <Image source={logo} className="h-8" resizeMode="contain" />

      <Text className="text-zinc-400 font-regular text-center text-lg mt-3">
        Plan and organize trips with your friends.{`\n`}Choose destinations,
        finalize itineraries, save useful links and make every journey
        memorable!
      </Text>

    </View>
  );
}
