import { Link } from "react-router";
import { useAuth } from "~/context/auth";
import defaultAvatar from "~/assets/default-avatar.svg";

export default function Header() {
  const { user } = useAuth();

  const header =
    user != null ? (
      <nav className="navbar navbar-light">
        <div className="container">
          <Link className="navbar-brand" to="/">
            conduit
          </Link>
          <ul className="nav navbar-nav pull-xs-right">
            <li className="nav-item">
              <Link className="nav-link active" to="/">
                Home
              </Link>
            </li>
            <li className="nav-item">
              <Link className="nav-link" to="/editor">
                {" "}
                <i className="ion-compose"></i>&nbsp;New Article{" "}
              </Link>
            </li>
            <li className="nav-item">
              <Link className="nav-link" to="/settings">
                {" "}
                <i className="ion-gear-a"></i>&nbsp;Settings{" "}
              </Link>
            </li>
            <li className="nav-item">
              <Link className="nav-link" to={`/profile/${user.username}`}>
                <img
                  src={user.image != null ? user.image : defaultAvatar}
                  className="user-pic"
                />
                {user.username}
              </Link>
            </li>
          </ul>
        </div>
      </nav>
    ) : (
      <nav className="navbar navbar-light">
        <div className="container">
          <Link className="navbar-brand" to="/">
            conduit
          </Link>
          <ul className="nav navbar-nav pull-xs-right">
            <li className="nav-item">
              <Link className="nav-link active" to="/">
                Home
              </Link>
            </li>
            <li className="nav-item">
              <Link className="nav-link" to="/login">
                Sign in
              </Link>
            </li>
            <li className="nav-item">
              <Link className="nav-link" to="/register">
                Sign up
              </Link>
            </li>
          </ul>
        </div>
      </nav>
    );

  return header;
}
