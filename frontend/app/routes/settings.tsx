import { useState } from "react";
import { ErrorMessages } from "~/components/errorMessages";
import { useAuth } from "~/context/auth";
import type { GenericError } from "~/types/error";
import type { UserResponse } from "~/types/users";
import { getAPIEndpoint, isErrorResponse, stdErrorMsg } from "~/utils";
import type { Route } from "./+types/settings";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "User Settings" },
    { name: "description", content: "user settings page" },
  ];
}

export default function Settings() {
  const { user, logout, updateUser } = useAuth();
  if (!user) return;

  const [errors, setErrors] = useState<Record<string, string[]> | null>(null);
  const [username, setUsername] = useState<string>(user.username);
  const [email, setEmail] = useState<string>(user.email);
  const [password, setPassword] = useState<string | null>(null);
  const [bio, setBio] = useState<string | null>(user.bio);
  const [image, setImage] = useState<string | null>(user.image);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErrors(null);
    const token = localStorage.getItem("token");

    try {
      const res = await fetch(getAPIEndpoint("/user"), {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Token ${token}`,
        },
        body: JSON.stringify({
          user: { username, email, password, bio, image },
        }),
      });

      const data: UserResponse | GenericError = await res.json();

      if (!res.ok || isErrorResponse(data)) {
        setErrors((data as GenericError).errors);
        return;
      }

      updateUser(data.user);
    } catch (err) {
      setErrors({ general: [stdErrorMsg] });
    }
  }

  return (
    <div className="settings-page">
      <div className="container page">
        <div className="row">
          <div className="col-md-6 offset-md-3 col-xs-12">
            <h1 className="text-xs-center">Your Settings</h1>

            <ErrorMessages errors={errors} />

            <form onSubmit={handleSubmit}>
              <fieldset>
                <fieldset className="form-group">
                  <input
                    className="form-control"
                    type="text"
                    placeholder="URL of profile picture"
                    name="image"
                    onChange={(e) => setImage(e.target.value)}
                    value={!image ? "" : image}
                  />
                </fieldset>
                <fieldset className="form-group">
                  <input
                    className="form-control form-control-lg"
                    type="text"
                    placeholder="Your Name"
                    name="username"
                    onChange={(e) => setUsername(e.target.value)}
                    value={username}
                  />
                </fieldset>
                <fieldset className="form-group">
                  <textarea
                    className="form-control form-control-lg"
                    rows={8}
                    placeholder="Short bio about you"
                    name="bio"
                    onChange={(e) => setBio(e.target.value)}
                    value={!bio ? "" : bio}
                  ></textarea>
                </fieldset>
                <fieldset className="form-group">
                  <input
                    className="form-control form-control-lg"
                    type="text"
                    placeholder="Email"
                    name="email"
                    onChange={(e) => setEmail(e.target.value)}
                    value={!email ? "" : email}
                  />
                </fieldset>
                <fieldset className="form-group">
                  <input
                    className="form-control form-control-lg"
                    type="password"
                    placeholder="New Password"
                    name="password"
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </fieldset>
                <button
                  type="submit"
                  className="btn btn-lg btn-primary pull-xs-right"
                >
                  Update Settings
                </button>
              </fieldset>
            </form>
            <hr />
            <button onClick={logout} className="btn btn-outline-danger">
              Or click here to logout.
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
