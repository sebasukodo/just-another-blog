import { Link } from "react-router";
import type { Route } from "./+types/main";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "404 - Page not found" },
    { name: "description", content: "the requested page does not exist" },
  ];
}

export default function CatchAllPage() {
  return (
    <div className="banner">
      <div className="container page">
        <h1>404 - Page not found</h1>
        <Link className="nav-link" to="/">
          Back to Homepage
        </Link>
      </div>
    </div>
  );
}
