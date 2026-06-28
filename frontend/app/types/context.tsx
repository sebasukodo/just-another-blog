import type { User } from "./users";

export interface AuthContextType {
  user: User | null;
  login: (user: User) => void;
  logout: () => void;
  isAuthLoading: boolean;
}
