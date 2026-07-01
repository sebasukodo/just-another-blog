import {
  type RouteConfig,
  index,
  layout,
  route,
} from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("/login", "routes/login.tsx"),
  route("/register", "routes/register.tsx"),
  layout("components/protectedLayout.tsx", [
    route("/settings", "routes/settings.tsx"),
  ]),
  route("/profile/:username/:tab?", "routes/profile.tsx"),
  route("*", "routes/catchAll.tsx"),
] satisfies RouteConfig;
