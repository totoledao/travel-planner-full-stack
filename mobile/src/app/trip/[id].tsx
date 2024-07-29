import { useLocalSearchParams } from "expo-router";
import { Text, View } from "react-native";

export default function Trip() {
  const { id } = useLocalSearchParams();

  return (
    <View>
      <Text className="text-zinc-200">{id}</Text>
      <Text className="text-zinc-200">{id}</Text>
      <Text className="text-zinc-200">{id}</Text>
      <Text className="text-zinc-200">{id}</Text>
      <Text className="text-zinc-200">{id}</Text>
    </View>
  );
}
