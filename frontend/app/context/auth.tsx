import { useContext, useEffect, useState, type ReactNode } from "react";
import { createContext } from "react";
import { useNavigate } from "react-router";
import { Navigate } from "react-router";
import type { AuthContextType } from "~/types/context";
import type { GenericError } from "~/types/error";
import type { User, UserResponse, UserUpdate } from "~/types/users";
import { getAPIEndpoint, isErrorResponse } from "~/utils";

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate();

  const [user, setUser] = useState<User | null>(null);
  const [isAuthLoading, setIsAuthLoading] = useState(true);

  useEffect(() => {
    async function loadUser() {
      const token = localStorage.getItem("token");

      if (!token) {
        setIsAuthLoading(false);
        return;
      }

      try {
        const res = await fetch(getAPIEndpoint("/user"), {
          headers: {
            Authorization: `Token ${token}`,
          },
        });
        const data: UserResponse | GenericError = await res.json();

        if (!res.ok || isErrorResponse(data)) {
          localStorage.removeItem("token");
          setUser(null);
          return;
        }

        setUser(data.user);
      } catch (err) {
        setUser(null);
      } finally {
        setIsAuthLoading(false);
      }
    }
    loadUser();
  }, []);

  function login(user: User) {
    localStorage.setItem("token", user.token);
    setUser(user);
  }

  function logout() {
    localStorage.removeItem("token");
    setUser(null);
    navigate("/");
  }

  function updateUser(updatedFields: UserUpdate) {
    setUser((currentUser) => {
      if (!updatedFields) return currentUser;
      return { ...currentUser, ...updatedFields } as User;
    });
  }

  return (
    <AuthContext.Provider
      value={{ user, login, logout, updateUser, isAuthLoading }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth could not find context");
  }
  return context;
}
