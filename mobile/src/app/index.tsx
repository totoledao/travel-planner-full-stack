import {
  ArrowRight,
  Calendar as CalendarIcon,
  MapPin,
  Settings2,
  UserRoundPlus,
} from "lucide-react-native";
import { Image, Text, View } from "react-native";

import bg from "@/assets/bg.png";
import logo from "@/assets/logo.png";
import { Button } from "@/components/button";
import { Input } from "@/components/input";
import { colors } from "@/styles/colors";
import { useState } from "react";

enum Steps {
  TRIP_DETAILS = 1,
  ADD_EMAIL = 2,
}

export default function index() {
  const [step, setStep] = useState(Steps.TRIP_DETAILS);

  function handleNextStep() {
    if (step == Steps.TRIP_DETAILS) {
      return setStep(Steps.ADD_EMAIL);
    } else {
      return;
    }
  }

  return (
    <View className="flex-1 justify-center items-center px-5">
      <Image source={bg} className="absolute w-full" resizeMode="contain" />
      <Image source={logo} className="h-8" resizeMode="contain" />

      <Text className="text-zinc-400 font-regular text-center text-lg mt-3">
        Plan and organize trips with your friends.{`\n`}Choose destinations,
        finalize itineraries, save useful links and make every journey
        memorable!
      </Text>

      <View className="w-full bg-zinc-900 p-4 rounded-xl my-8 border border-zinc-800">
        {step == Steps.TRIP_DETAILS && (
          <>
            <Input>
              <MapPin color={colors.zinc[400]} size={20} />
              <Input.Field placeholder="Going to" />
            </Input>

            <Input>
              <CalendarIcon color={colors.zinc[400]} size={20} />
              <Input.Field placeholder="When" />
            </Input>
          </>
        )}

        {step == Steps.ADD_EMAIL && (
          <>
            <View className="border-b py-3 border-zinc-800">
              <Button
                variant="secondary"
                onPress={() => setStep(Steps.TRIP_DETAILS)}
              >
                <Button.Title>Change destination / date</Button.Title>
                <Settings2 color={colors.zinc[200]} size={20} />
              </Button>
            </View>

            <Input>
              <UserRoundPlus color={colors.zinc[400]} size={20} />
              <Input.Field placeholder="Who" />
            </Input>
          </>
        )}

        <Button onPress={handleNextStep}>
          <Button.Title>
            {step == Steps.TRIP_DETAILS ? "Continue" : "Confirm trip"}
          </Button.Title>
          <ArrowRight color={colors.lime[950]} size={20} />
        </Button>
      </View>

      <Text className="text-zinc-500 font-regular text-center text-base">
        Using Travel Planner, you automatically agree to our{" "}
        <Text className="text-zinc-300 underline">
          terms of service and privacy policy
        </Text>
        .
      </Text>
    </View>
  );
}
