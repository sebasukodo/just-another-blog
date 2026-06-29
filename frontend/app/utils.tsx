import type { GenericError } from "./types/error";
import type { UserResponse } from "./types/users";
import defaultAvatar from "~/assets/default-avatar.svg";

export function getAPIEndpoint(endpoint: string): string {
  const host = "http://localhost:7337/api";
  return endpoint.startsWith("/")
    ? `${host}${endpoint}`
    : `${host}/${endpoint}`;
}

export function isErrorResponse(
  data: UserResponse | GenericError,
): data is GenericError {
  return "errors" in data;
}

export const stdErrorMsg = "network error, please try again.";

export function formatDate(date: string): string {
  return new Intl.DateTimeFormat("de-DE", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(new Date(date));
}

export function userProfilePicture(path: string | null): string {
  return path != null ? path : defaultAvatar;
}
