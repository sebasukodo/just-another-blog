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
    route("/editor/:slug?", "routes/editor.tsx"),
  ]),
  route("/profile/:username/:tab?", "routes/profile.tsx"),
  route("/article/:slug", "routes/article.tsx"),
  route("*", "routes/catchAll.tsx"),
] satisfies RouteConfig;
