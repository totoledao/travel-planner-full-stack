import { API } from "./api";

export type TripDetails = {
  id: string;
  destination: string;
  starts_at: string;
  ends_at: string;
  is_confirmed: boolean;
};

type TripCreate = Omit<TripDetails, "id" | "is_confirmed"> & {
  emails_to_invite: string[];
  owner_name: string;
  owner_email: string;
};

async function get(id: string) {
  try {
    const { data } = await API.get<{ trip: TripDetails }>(`trips/${id}`);
    return data.trip;
  } catch (err) {
    throw err;
  }
}

async function create(trip: TripCreate) {
  try {
    const { data } = await API.post<{ tripId: string }>(`trips`, trip);
    return data;
  } catch (err) {
    throw err;
  }
}

export const tripsApi = { get, create };
