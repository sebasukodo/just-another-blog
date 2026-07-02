import type { User, UserUpdate } from "./users";

export interface AuthContextType {
  user: User | null;
  login: (user: User) => void;
  logout: () => void;
  updateUser: (user: UserUpdate) => void;
  isAuthLoading: boolean;
}
