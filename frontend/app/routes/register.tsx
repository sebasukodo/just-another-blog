import { Link } from "react-router";
import type { Route } from "./+types/main";
import { useState } from "react";
import { getAPIEndpoint, isErrorResponse, stdErrorMsg } from "~/utils";
import type { UserResponse } from "~/types/users";
import type { GenericError } from "~/types/error";
import { useNavigate } from "react-router";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Register" },
    { name: "description", content: "user registration page" },
  ];
}

export default function register() {
  const navigate = useNavigate();

  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<Record<string, string[]> | null>(null);

  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormEvent>) {
    e.preventDefault();
    setErrors(null);
    setIsLoading(false);

    try {
      const res = await fetch(getAPIEndpoint("users"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user: { username, email, password } }),
      });

      const data: UserResponse | GenericError = await res.json();

      if (!res.ok || isErrorResponse(data)) {
        setErrors((data as GenericError).errors);
        return;
      }

      localStorage.setItem("token", data.user.token);
      navigate("/");
    } catch (err) {
      setErrors({ general: [stdErrorMsg] });
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="auth-page">
      <div className="container page">
        <div className="row">
          <div className="col-md-6 offset-md-3 col-xs-12">
            <h1 className="text-xs-center">Sign up</h1>
            <p className="text-xs-center">
              <Link to="/login">Have an account?</Link>
            </p>

            {isLoading && (
              <p className="text-xs-center">trying to register...</p>
            )}

            {errors && (
              <ul className="error-messages">
                {Object.entries(errors).map(([field, messages]) =>
                  messages.map((msg) => (
                    <li key={`${field}-${msg}`}>
                      {field} {msg}
                    </li>
                  )),
                )}
              </ul>
            )}

            <form onSubmit={handleSubmit}>
              <fieldset className="form-group">
                <input
                  className="form-control form-control-lg"
                  type="text"
                  placeholder="Username"
                  name="username"
                  onChange={(e) => setUsername(e.target.value)}
                />
              </fieldset>
              <fieldset className="form-group">
                <input
                  className="form-control form-control-lg"
                  type="text"
                  placeholder="Email"
                  name="email"
                  onChange={(e) => setEmail(e.target.value)}
                />
              </fieldset>
              <fieldset className="form-group">
                <input
                  className="form-control form-control-lg"
                  type="password"
                  placeholder="Password"
                  name="password"
                  onChange={(e) => setPassword(e.target.value)}
                />
              </fieldset>
              <button
                type="submit"
                className="btn btn-lg btn-primary pull-xs-right"
              >
                Sign up
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
