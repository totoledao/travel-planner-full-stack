import dayjs from "dayjs";
import { router } from "expo-router";
import {
  ArrowRight,
  AtSign,
  Calendar as CalendarIcon,
  MapPin,
  Plus,
  Settings2,
  UserRoundPlus,
} from "lucide-react-native";
import { useState } from "react";
import { Alert, Image, Keyboard, Text, View } from "react-native";
import { DateData } from "react-native-calendars";

import { tripsApi } from "@/api/trips";
import { tripStorage } from "@/storage/trip";
import { colors } from "@/styles/colors";
import { calendarUtils, DatesSelected } from "@/utils/calendarUtils";
import { validateInput } from "@/utils/validateInput";

import bg from "@/assets/bg.png";
import logo from "@/assets/logo.png";
import { Button } from "@/components/button";
import { Calendar } from "@/components/calendar";
import { GuestEmail } from "@/components/email";
import { Input } from "@/components/input";
import { Modal } from "@/components/modal";

enum Steps {
  TRIP_DETAILS = 1,
  ADD_EMAIL = 2,
}

enum Modals {
  None = 0,
  CALENDAR = 1,
  GUESTS = 2,
}

export default function Index() {
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState(Steps.TRIP_DETAILS);
  const [showModal, setShowModal] = useState(Modals.None);
  const [destination, setDestination] = useState("");
  const [selectedDates, setSelectedDates] = useState({} as DatesSelected);
  const [organizerName, setOrganizerName] = useState("");
  const [organizerEmail, setOrganizerEmail] = useState("");
  const [newGuest, setNewGuest] = useState("");
  const [guestList, setGuestList] = useState<string[]>([]);

  async function createTrip() {
    async function saveTripOnDevice(tripId: string) {
      try {
        await tripStorage.save(tripId);
        router.navigate(`trip/${tripId}`);
      } catch (err) {
        console.log(err);
        Alert.alert(
          "Error",
          "It was not possible to save the trip details on the device",
          undefined,
          {
            cancelable: true,
          }
        );
      }
    }

    setLoading(true);
    try {
      const { tripId } = await tripsApi.create({
        destination: destination,
        starts_at: dayjs(selectedDates.startsAt!.dateString).utc().format(),
        ends_at: dayjs(selectedDates.endsAt!.dateString).utc().format(),
        emails_to_invite: guestList,
        owner_name: organizerName,
        owner_email: organizerEmail,
      });
      await saveTripOnDevice(tripId);
    } catch (err) {
      console.log(err);
      Alert.alert(
        "Error",
        "It was not possible to create the trip. Try again.",
        undefined,
        {
          cancelable: true,
        }
      );
    } finally {
      setLoading(false);
    }
  }

  function handleNextStep() {
    if (!destination || !selectedDates.endsAt) {
      return Alert.alert(
        "Invalid fields",
        "Please fill in all the fields to continue",
        undefined,
        {
          cancelable: true,
        }
      );
    }

    if (destination.length < 4) {
      return Alert.alert(
        "Invalid fields",
        'The "Going to" field should be at least 4 characters long',
        undefined,
        {
          cancelable: true,
        }
      );
    }

    if (step === Steps.TRIP_DETAILS) {
      return setStep(Steps.ADD_EMAIL);
    }

    if (!isValidOrganizerName()) return;

    if (!isValidOrganizerEmail()) return;

    return Alert.alert(
      "Create trip",
      "Do you want to create the trip? The guests will receive an email invitation to confirm their participation.",
      [
        { text: "Go back", style: "cancel" },
        { text: "Create", onPress: createTrip },
      ],
      {
        cancelable: true,
      }
    );
  }

  function handleSelectDate(selectedDay: DateData) {
    const dates = calendarUtils.orderStartsAtAndEndsAt({
      startsAt: selectedDates.startsAt,
      endsAt: selectedDates.endsAt,
      selectedDay,
    });

    setSelectedDates(dates);
  }

  function handleCloseModal() {
    setShowModal(Modals.None);
  }

  function isValidOrganizerName() {
    if (!organizerName) {
      Alert.alert(
        "Invalid fields",
        "The organizer's name cannot be empty",
        undefined,
        {
          cancelable: true,
        }
      );

      return false;
    }

    return true;
  }

  function isValidOrganizerEmail() {
    if (!validateInput.email(organizerEmail)) {
      setOrganizerEmail("");

      Alert.alert(
        "Invalid fields",
        "The organizer email is not valid",
        undefined,
        {
          cancelable: true,
        }
      );

      return false;
    }

    return true;
  }

  function handleAddGuest() {
    if (!validateInput.email(newGuest)) {
      return Alert.alert(
        "Invalid fields",
        "The guest email is not valid",
        undefined,
        {
          cancelable: true,
        }
      );
    }

    if (guestList.includes(newGuest)) {
      setNewGuest("");
      return Alert.alert(
        "Invalid fields",
        "The guest email is already on the list",
        undefined,
        {
          cancelable: true,
        }
      );
    }

    setGuestList((prev) => [...prev, newGuest]);
    setNewGuest("");
  }

  function handleRemoveGuest(email: string) {
    setGuestList((prev) => prev.filter((str) => str !== email));
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
        <Input>
          <MapPin color={colors.zinc[400]} size={20} />
          <Input.Field
            placeholder="Going to"
            value={destination}
            onChangeText={setDestination}
            editable={step === Steps.TRIP_DETAILS}
          />
        </Input>

        <Input>
          <CalendarIcon color={colors.zinc[400]} size={20} />
          <Input.Field
            placeholder="When"
            value={selectedDates.formatDatesInText}
            onFocus={() => Keyboard.dismiss()}
            showSoftInputOnFocus={false}
            onPressIn={() => setShowModal(Modals.CALENDAR)}
            editable={step === Steps.TRIP_DETAILS}
          />
        </Input>

        {step === Steps.ADD_EMAIL && (
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
              <Input.Field
                placeholder="Who is organizing"
                value={organizerName}
                onChangeText={setOrganizerName}
                onBlur={isValidOrganizerName}
              />
            </Input>

            <Input>
              <AtSign color={colors.zinc[400]} size={20} />
              <Input.Field
                placeholder="Organizer's email"
                value={organizerEmail}
                onChangeText={setOrganizerEmail}
                autoCapitalize="none"
                autoCorrect={false}
                keyboardType="email-address"
                onBlur={isValidOrganizerEmail}
              />
            </Input>

            <View className="border-b border-zinc-800" />

            <Input>
              <UserRoundPlus color={colors.zinc[400]} size={20} />
              <Input.Field
                placeholder="Who is going"
                value={
                  guestList.length > 0
                    ? `${guestList.length} ${
                        guestList.length === 1 ? "guest" : "guests"
                      } on the list`
                    : ""
                }
                onFocus={() => Keyboard.dismiss()}
                showSoftInputOnFocus={false}
                onPressIn={() => setShowModal(Modals.GUESTS)}
              />
            </Input>
          </>
        )}

        <Button onPress={handleNextStep} isLoading={loading}>
          <Button.Title>
            {step === Steps.TRIP_DETAILS ? "Continue" : "Confirm trip"}
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

      <Modal
        title="Select the dates"
        subtitle="Select the start and end dates of your trip"
        visible={showModal === Modals.CALENDAR}
        onClose={handleCloseModal}
      >
        <View className="mt-4 gap-4">
          <Calendar
            onDayPress={handleSelectDate}
            markedDates={selectedDates.dates}
            minDate={dayjs().toISOString()}
          />
          <Button
            onPress={() => {
              setShowModal(Modals.None);
              handleNextStep();
            }}
          >
            <Button.Title>Confirm</Button.Title>
            <ArrowRight color={colors.lime[950]} size={20} />
          </Button>
        </View>
      </Modal>

      <Modal
        title="Who is going?"
        subtitle="Your guests will receive an email invitation to join the trip"
        visible={showModal === Modals.GUESTS}
        onClose={handleCloseModal}
      >
        <View className="my-2 flex-wrap gap-2 border-b border-zinc-800 py-5 items-start">
          {guestList.length > 0 ? (
            guestList.map((str) => (
              <GuestEmail
                key={str}
                email={str}
                onRemove={() => handleRemoveGuest(str)}
              />
            ))
          ) : (
            <Text className="text-zinc-500 font-regular text-base">
              The guest list is empty
            </Text>
          )}
        </View>
        <View className="gap-4 mt-4">
          <Input variant="secondary">
            <AtSign color={colors.zinc[400]} size={20} />
            <Input.Field
              placeholder="guest@email.com"
              value={newGuest}
              onChangeText={setNewGuest}
              autoCapitalize="none"
              autoCorrect={false}
              keyboardType="email-address"
              onSubmitEditing={handleAddGuest}
              blurOnSubmit={false}
            />
          </Input>
          <Button onPress={handleAddGuest}>
            <Button.Title>Add</Button.Title>
            <Plus color={colors.lime[950]} size={20} />
          </Button>
        </View>
      </Modal>
    </View>
  );
}
