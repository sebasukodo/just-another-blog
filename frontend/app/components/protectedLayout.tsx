import { Outlet } from "react-router";
import { Navigate, useLocation } from "react-router";
import { useAuth } from "~/context/auth";

export default function ProtectedLayout() {
  const { user, isAuthLoading } = useAuth();
  const location = useLocation();

  if (isAuthLoading) return null;

  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <Outlet />;
}
