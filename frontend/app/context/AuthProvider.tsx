import { useEffect, useState, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { AuthContext } from "~/context/AuthContext";
import type { GenericError } from "~/types/error";
import type { User, UserResponse, UserUpdate } from "~/types/users";
import { getAPIEndpoint, isErrorResponse } from "~/utils";

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
      } catch {
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
