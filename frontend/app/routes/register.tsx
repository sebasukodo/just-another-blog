import { Link } from "react-router";
import type { Route } from "./+types/register";
import { useEffect, useState } from "react";
import { getAPIEndpoint, isErrorResponse, stdErrorMsg } from "~/utils";
import type { UserResponse } from "~/types/users";
import type { GenericError } from "~/types/error";
import { useNavigate } from "react-router";
import { useAuth } from "~/hooks/useAuth";
import { ErrorMessages } from "~/components/errorMessages";

export function meta(_args: Route.MetaArgs) {
  return [
    { title: "Register" },
    { name: "description", content: "user registration page" },
  ];
}

export default function Register() {
  const navigate = useNavigate();

  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<Record<string, string[]> | null>(null);

  const [isLoading, setIsLoading] = useState(false);

  const { user, login, isAuthLoading } = useAuth();

  useEffect(() => {
    if (!isAuthLoading && user) {
      navigate("/");
    }
  }, [isAuthLoading, user, navigate]);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
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

      login(data.user);
      navigate("/");
    } catch {
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

            <ErrorMessages errors={errors} />

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
