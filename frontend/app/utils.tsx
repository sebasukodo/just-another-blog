import type { GenericError } from "./types/error";
import type { UserResponse } from "./types/users";

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
