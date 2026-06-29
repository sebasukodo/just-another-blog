import { Link } from "react-router";
import { useAuth } from "~/context/auth";
import defaultAvatar from "~/assets/default-avatar.svg";
import { NavLink } from "react-router";
import { userProfilePicture } from "~/utils";

export default function Header() {
  const { user } = useAuth();

  const navLinkClass = "nav-link";
  const activeClass = navLinkClass + " active";

  const header =
    user != null ? (
      <nav className="navbar navbar-light">
        <div className="container">
          <Link className="navbar-brand" to="/">
            conduit
          </Link>
          <ul className="nav navbar-nav pull-xs-right">
            <li className="nav-item">
              <NavLink
                className={({ isActive }) =>
                  isActive ? activeClass : navLinkClass
                }
                to="/"
              >
                Home
              </NavLink>
            </li>
            <li className="nav-item">
              <NavLink
                className={({ isActive }) =>
                  isActive ? activeClass : navLinkClass
                }
                to="/editor"
              >
                {" "}
                <i className="ion-compose"></i>&nbsp;New Article{" "}
              </NavLink>
            </li>
            <li className="nav-item">
              <NavLink
                className={({ isActive }) =>
                  isActive ? activeClass : navLinkClass
                }
                to="/settings"
              >
                {" "}
                <i className="ion-gear-a"></i>&nbsp;Settings{" "}
              </NavLink>
            </li>
            <li className="nav-item">
              <NavLink
                className={({ isActive }) =>
                  isActive ? activeClass : navLinkClass
                }
                to={`/profile/${user.username}`}
              >
                <img
                  src={userProfilePicture(user.image)}
                  className="user-pic"
                />
                {user.username}
              </NavLink>
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
              <NavLink
                className={({ isActive }) =>
                  isActive ? activeClass : navLinkClass
                }
                to="/"
              >
                Home
              </NavLink>
            </li>
            <li className="nav-item">
              <NavLink
                className={({ isActive }) =>
                  isActive ? activeClass : navLinkClass
                }
                to="/login"
              >
                Sign in
              </NavLink>
            </li>
            <li className="nav-item">
              <NavLink
                className={({ isActive }) =>
                  isActive ? activeClass : navLinkClass
                }
                to="/register"
              >
                Sign up
              </NavLink>
            </li>
          </ul>
        </div>
      </nav>
    );

  return header;
}
