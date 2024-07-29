import AsyncStorage from "@react-native-async-storage/async-storage";

const TRIP_STORAGE_KEY = "@travel-planner:tripId";

async function save(id: string) {
  try {
    await AsyncStorage.setItem(TRIP_STORAGE_KEY, id);
  } catch (err) {
    throw err;
  }
}

async function get() {
  try {
    return await AsyncStorage.getItem(TRIP_STORAGE_KEY);
  } catch (err) {
    throw err;
  }
}

async function remove() {
  try {
    await AsyncStorage.removeItem(TRIP_STORAGE_KEY);
  } catch (err) {
    throw err;
  }
}

export const tripStorage = { get, remove, save };
