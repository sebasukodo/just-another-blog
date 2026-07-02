import { Link } from "react-router";
import type { Route } from "./+types/login";
import { useEffect, useState } from "react";
import { getAPIEndpoint, isErrorResponse, stdErrorMsg } from "~/utils";
import type { UserResponse } from "~/types/users";
import type { GenericError } from "~/types/error";
import { useNavigate } from "react-router";
import { useAuth } from "~/context/auth";
import { useLocation } from "react-router";
import { ErrorMessages } from "~/components/errorMessages";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Login" },
    { name: "description", content: "login page to access personal content" },
  ];
}

export default function Login() {
  const navigate = useNavigate();
  const location = useLocation();
  const from = location.state?.from?.pathname || "/";

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
    setIsLoading(true);
    setErrors(null);

    try {
      const res = await fetch(getAPIEndpoint("/users/login"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user: { email, password } }),
      });
      const data: UserResponse | GenericError = await res.json();

      if (!res.ok || isErrorResponse(data)) {
        setErrors((data as GenericError).errors);
        return;
      }

      login(data.user);
      navigate(from, { replace: true });
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
            <h1 className="text-xs-center">Sign in</h1>
            <p className="text-xs-center">
              <Link to="/register">Need an account?</Link>
            </p>

            {isLoading && <p className="text-xs-center">trying to login...</p>}

            <ErrorMessages errors={errors} />

            <form onSubmit={handleSubmit}>
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
                Sign in
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
