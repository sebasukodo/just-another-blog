import { useAuth } from "~/context/auth";
import type { Route } from "./+types/main";
import { useEffect, useState } from "react";
import {
  getAPIEndpoint,
  isErrorResponse,
  stdErrorMsg,
  userProfilePicture,
} from "~/utils";
import { useParams } from "react-router";
import type { Profile, ProfileResponse } from "~/types/profile";
import type { GenericError } from "~/types/error";
import { ErrorMessages } from "~/components/errorMessages";
import { Link } from "react-router";
import type { Article, ArticleResponse, Articles } from "~/types/articles";
import ArticlePreview from "~/components/article/articlePreview";
import { useNavigate } from "react-router";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "User Profile" },
    { name: "description", content: "this is a profile page" },
  ];
}

const searchLimit = 20;
type FeedType = "myArticles" | "favorites";

export default function Profile() {
  const { user, isAuthLoading } = useAuth();
  const { username } = useParams();
  const navigate = useNavigate();

  const [profileUser, setProfileUser] = useState<Profile | null>(null);
  const [errors, setErrors] = useState<Record<string, string[]> | null>(null);
  const [followErrors, setFollowErrors] = useState<Record<
    string,
    string[]
  > | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const [activeTab, setActiveTab] = useState<FeedType>("myArticles");

  const [articles, setArticles] = useState<Articles>({
    articles: [],
    articlesCount: 0,
  });

  const [currentPage, setCurrentPage] = useState(1);
  const totalPages = Math.ceil(articles.articlesCount / searchLimit);

  const isAuthor =
    profileUser != null ? user?.username === profileUser.username : false;

  const headers: HeadersInit = { "Content-Type": "application/json" };
  if (user?.token) {
    headers.Authorization = `Token ${user.token}`;
  }

  useEffect(() => {
    if (isAuthLoading) return;

    async function getProfile() {
      setErrors(null);
      setIsLoading(true);

      try {
        const res = await fetch(getAPIEndpoint(`profiles/${username}`), {
          method: "GET",
          headers: headers,
        });

        const data: ProfileResponse | GenericError = await res.json();

        if (!res.ok || isErrorResponse(data)) {
          setErrors((data as GenericError).errors);
          return;
        }

        setProfileUser(data.profile);
      } catch (err) {
        setErrors({ general: [stdErrorMsg] });
      } finally {
        setIsLoading(false);
      }
    }
    getProfile();
  }, [isAuthLoading, username]);

  useEffect(() => {
    if (profileUser) {
      fetchArticles(1, activeTab);
    }
  }, [profileUser?.username]);

  async function fetchArticles(page: number, tab: FeedType) {
    setIsLoading(true);
    setErrors(null);
    setActiveTab(tab);
    setCurrentPage(page);

    const filter =
      tab == "myArticles"
        ? `&author=${profileUser?.username}`
        : `&favorited=${profileUser?.username}`;
    const apiPath = `articles?offset=${(page - 1) * searchLimit}&limit=${searchLimit}${filter}`;

    try {
      const res = await fetch(getAPIEndpoint(apiPath), {
        method: "GET",
        headers: headers,
      });

      const data: Articles | GenericError = await res.json();
      if (!res.ok || isErrorResponse(data)) {
        setErrors((data as GenericError).errors);
        return;
      }

      setArticles({
        articles: data.articles,
        articlesCount: data.articlesCount,
      });
    } catch (err) {
      setErrors({ general: [stdErrorMsg] });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleFollow() {
    if (!user) {
      navigate("/login");
      return;
    }

    setFollowErrors(null);
    const method = profileUser?.following ? "DELETE" : "POST";
    try {
      const res = await fetch(
        getAPIEndpoint(`profiles/${profileUser?.username}/follow`),
        {
          method: method,
          headers: headers,
        },
      );

      const data: ProfileResponse | GenericError = await res.json();

      if (!res.ok || isErrorResponse(data)) {
        setErrors((data as GenericError).errors);
        return;
      }

      setProfileUser(data.profile);
    } catch (err) {
      setFollowErrors({ general: [stdErrorMsg] });
    }
  }

  async function handleFavorite(article: Article) {
    if (!user) return;

    const method = article.favorited ? "DELETE" : "POST";
    setErrors(null);
    try {
      const res = await fetch(
        getAPIEndpoint(`articles/${article.slug}/favorite`),
        {
          method: method,
          headers: headers,
        },
      );

      const data: ArticleResponse | GenericError = await res.json();
      if (!res.ok || isErrorResponse(data)) {
        setErrors((data as GenericError).errors);
        return;
      }

      setArticles((prev) => ({
        ...prev,
        articles: prev.articles.map((a) =>
          a.slug === data.article.slug ? data.article : a,
        ),
      }));
    } catch (err) {
      setErrors({ general: [stdErrorMsg] });
    }
  }

  if (!profileUser && errors) return <ErrorMessages errors={errors} />;
  else if (!profileUser) return null;

  return (
    <div className="profile-page">
      <div className="user-info">
        <div className="container">
          <div className="row">
            <div className="col-xs-12 col-md-10 offset-md-1">
              <img
                src={userProfilePicture(profileUser.image)}
                className="user-img"
                alt={`${profileUser.username}'s profile picture`}
              />
              <h4>{profileUser.username}</h4>
              <p>{profileUser?.bio}</p>
              {isAuthor ? (
                <Link
                  to="/settings"
                  className="btn btn-sm btn-outline-secondary action-btn"
                >
                  <i className="ion-gear-a"></i>
                  &nbsp; Edit Profile Settings
                </Link>
              ) : (
                <button
                  onClick={() => handleFollow()}
                  className="btn btn-sm btn-outline-secondary action-btn"
                >
                  <i
                    className={`ion-${
                      profileUser.following ? "minus" : "plus"
                    }-round`}
                  ></i>
                  &nbsp;{" "}
                  {profileUser.following
                    ? "Unollow" + ` ${profileUser.username}`
                    : "Follow" + ` ${profileUser.username}`}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="container">
        <div className="row">
          <div className="col-xs-12 col-md-10 offset-md-1">
            <div className="articles-toggle">
              <ul className="nav nav-pills outline-active">
                <li className="nav-item">
                  <button
                    onClick={() => fetchArticles(1, "myArticles")}
                    className={
                      activeTab == "myArticles" ? "nav-link active" : "nav-link"
                    }
                  >
                    My Articles
                  </button>
                </li>
                <li className="nav-item">
                  <button
                    onClick={() => fetchArticles(1, "favorites")}
                    className={
                      activeTab == "favorites" ? "nav-link active" : "nav-link"
                    }
                  >
                    Favorited Articles
                  </button>
                </li>
              </ul>
            </div>

            {isLoading && <p>fetching data from server...</p>}

            <ErrorMessages errors={errors} />

            {articles.articlesCount > 0 ? (
              articles.articles.map((article) => {
                return (
                  <ArticlePreview
                    article={article}
                    onFavorite={handleFavorite}
                    key={`article-${article.slug}`}
                  />
                );
              })
            ) : (
              <p>no articles found</p>
            )}

            {totalPages > 1 && (
              <ul className="pagination">
                {Array.from({ length: totalPages }, (_, index) => {
                  const page = index + 1;
                  return (
                    <li
                      key={`page-${page}`}
                      className={
                        page === currentPage ? "page-item active" : "page-item"
                      }
                    >
                      <button
                        className="page-link"
                        onClick={() => fetchArticles(page, activeTab)}
                      >
                        {page}
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
