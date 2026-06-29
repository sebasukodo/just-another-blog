import { useAuth } from "~/context/auth";
import type { Route } from "./+types/home";
import { useEffect, useState } from "react";
import { getAPIEndpoint, isErrorResponse, stdErrorMsg } from "~/utils";
import type { Articles, Tags } from "~/types/articles";
import type { GenericError } from "~/types/error";
import ArticlePreview from "~/components/article/articlePreview";
import { ErrorMessages } from "~/components/errorMessages";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "conduit" },
    { name: "description", content: "Welcome to conduit!" },
  ];
}

const searchLimit = 20;
type FeedType = "feed" | "global" | "tag";

export default function Home() {
  const { user, isAuthLoading } = useAuth();
  const loggedIn = user != null;

  const [activeTab, setActiveTab] = useState<FeedType>(
    loggedIn ? "feed" : "global",
  );
  const [articleErrors, setArticleErrors] = useState<Record<
    string,
    string[]
  > | null>(null);
  const [tagsErrors, setTagsErrors] = useState<Record<string, string[]> | null>(
    null,
  );
  const [isArticleLoading, setIsArticleLoading] = useState(false);
  const [isTagsLoading, setIsTagsLoading] = useState(false);

  const [tag, setTag] = useState<string | null>(null);
  const [popularTags, setPopularTags] = useState<Tags>({ tags: [] });
  const [articles, setArticles] = useState<Articles>({
    articles: [],
    articlesCount: 0,
  });

  const [currentPage, setCurrentPage] = useState(1);
  const totalPages = Math.ceil(articles.articlesCount / searchLimit);

  useEffect(() => {
    let url = "/articles";
    if (activeTab == "feed") {
      url = "/articles/feed";
    }
    fetchArticles(url);
    fetchTags();
  }, []);

  async function fetchTags() {
    setIsTagsLoading(true);
    setTagsErrors(null);
    try {
      const res = await fetch(getAPIEndpoint("/tags"), {
        method: "GET",
      });

      const data: Tags | GenericError = await res.json();
      if (!res.ok || isErrorResponse(data)) {
        setTagsErrors((data as GenericError).errors);
        return;
      }

      setPopularTags(data);
    } catch (err) {
      setTagsErrors({ general: [stdErrorMsg] });
    } finally {
      setIsTagsLoading(false);
    }
  }

  async function fetchArticles(apiPath: string) {
    setIsArticleLoading(true);
    setArticleErrors(null);

    const token = localStorage.getItem("token");
    const headers: HeadersInit = { "Content-Type": "application/json" };
    if (token) {
      headers.Authorization = `Token ${token}`;
    }

    try {
      const res = await fetch(getAPIEndpoint(apiPath), {
        method: "GET",
        headers: headers,
      });

      const data: Articles | GenericError = await res.json();
      if (!res.ok || isErrorResponse(data)) {
        setArticleErrors((data as GenericError).errors);
        return;
      }

      setArticles({
        articles: data.articles,
        articlesCount: data.articlesCount,
      });
    } catch (err) {
      setArticleErrors({ general: [stdErrorMsg] });
    } finally {
      setIsArticleLoading(false);
    }
  }

  async function handleFeed(page: number) {
    setActiveTab("feed");
    fetchArticles(
      `articles/feed?offset=${(page - 1) * searchLimit}&limit=${searchLimit}`,
    );
    setCurrentPage(page);
  }

  async function handleGlobal(page: number) {
    setActiveTab("global");
    fetchArticles(
      `articles?offset=${(page - 1) * searchLimit}&limit=${searchLimit}`,
    );
    setCurrentPage(page);
  }

  async function handleTag(page: number, selectedTag: string | null) {
    setActiveTab("tag");
    fetchArticles(
      `articles?offset=${(page - 1) * searchLimit}&limit=${searchLimit}&tag=${selectedTag}`,
    );
    setCurrentPage(page);
  }

  return (
    <div className="home-page">
      <div className="banner">
        <div className="container">
          <h1 className="logo-font">conduit</h1>
          <p>A place to share your knowledge.</p>
        </div>
      </div>

      <div className="container page">
        <div className="row">
          <div className="col-md-9">
            <div className="feed-toggle">
              <ul className="nav nav-pills outline-active">
                {loggedIn && (
                  <li className="nav-item">
                    <button
                      className={
                        activeTab == "feed" ? "nav-link active" : "nav-link"
                      }
                      onClick={() => handleFeed(1)}
                    >
                      Your Feed
                    </button>
                  </li>
                )}
                <li className="nav-item">
                  <button
                    className={
                      activeTab == "global" ? "nav-link active" : "nav-link"
                    }
                    onClick={() => handleGlobal(1)}
                  >
                    Global Feed
                  </button>
                </li>
                {tag && (
                  <li className="nav-item">
                    <button
                      className={
                        activeTab == "tag" ? "nav-link active" : "nav-link"
                      }
                      onClick={() => handleTag(1, tag)}
                    >
                      # {tag}
                    </button>
                  </li>
                )}
              </ul>
            </div>

            {isArticleLoading && <p>fetching data from server...</p>}

            <ErrorMessages errors={articleErrors} />

            {articles.articlesCount > 0 ? (
              articles.articles.map((article) => {
                return (
                  <ArticlePreview
                    article={article}
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
                        onClick={() => {
                          if (activeTab === "feed") handleFeed(page);
                          else if (activeTab === "global") handleGlobal(page);
                          else if (activeTab === "tag") handleTag(page, tag);
                        }}
                      >
                        {page}
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>

          <div className="col-md-3">
            <div className="sidebar">
              <p>Popular Tags</p>

              {isTagsLoading && <p>fetching data from server...</p>}

              <ErrorMessages errors={tagsErrors} />

              <div className="tag-list">
                {popularTags.tags.length > 0 ? (
                  popularTags.tags.map((tag, index) => {
                    return (
                      <button
                        onClick={() => {
                          setTag(tag);
                          handleTag(1, tag);
                        }}
                        key={`tag-${index}-${tag}`}
                        className="tag-pill tag-default"
                      >
                        {tag}
                      </button>
                    );
                  })
                ) : (
                  <p>no articles found</p>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
