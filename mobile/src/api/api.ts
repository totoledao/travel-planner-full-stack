import axios from "axios";

const apiUrl = process.env.EXPO_PUBLIC_API_URL;

export const API = axios.create({
  baseURL: `http://${apiUrl}:8080/`,
});
