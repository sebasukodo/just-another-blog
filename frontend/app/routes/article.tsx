import { useEffect, useState } from "react";
import { useParams } from "react-router";
import { useNavigate } from "react-router";
import { useAuth } from "~/context/auth";
import type {
  Article,
  ArticleResponse,
  Comment,
  CommentResponse,
  CommentsResponse,
} from "~/types/articles";
import type { GenericError } from "~/types/error";
import type { ProfileResponse } from "~/types/profile";
import ReactMarkdown from "react-markdown";
import {
  formatDate,
  getAPIEndpoint,
  isErrorResponse,
  stdErrorMsg,
  userProfilePicture,
} from "~/utils";
import { Link } from "react-router";
import { ErrorMessages } from "~/components/errorMessages";
import CommentComponent from "~/components/article/comment";

export default function Article() {
  const navigate = useNavigate();
  const { user, isAuthLoading } = useAuth();
  const { slug } = useParams();

  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState<Record<string, string[]> | null>(null);
  const [article, setArticle] = useState<Article | null>(null);
  const isAuthor =
    article != null ? user?.username === article.author.username : false;

  const [comments, setComments] = useState<Comment[]>([]);
  const [commentErrors, setCommentErrors] = useState<Record<
    string,
    string[]
  > | null>(null);
  const [isCommentLoading, setIsCommentLoading] = useState(false);

  const [commentInput, setCommentInput] = useState("");

  const headers: HeadersInit = { "Content-Type": "application/json" };
  if (user?.token) {
    headers.Authorization = `Token ${user.token}`;
  }

  useEffect(() => {
    async function getArticle() {
      setIsLoading(true);
      setErrors(null);

      try {
        const res = await fetch(getAPIEndpoint(`articles/${slug}`), {
          method: "GET",
          headers: headers,
        });

        const data: ArticleResponse | GenericError = await res.json();

        if (!res.ok || isErrorResponse(data)) {
          setErrors((data as GenericError).errors);
          return;
        }

        setArticle(data.article);
      } catch (err) {
        setErrors({ general: [stdErrorMsg] });
      } finally {
        setIsLoading(false);
      }
    }

    async function getComments() {
      setIsCommentLoading(true);
      setCommentErrors(null);

      try {
        const res = await fetch(getAPIEndpoint(`articles/${slug}/comments`), {
          method: "GET",
          headers: headers,
        });

        const data: CommentsResponse | GenericError = await res.json();

        if (!res.ok || isErrorResponse(data)) {
          setCommentErrors((data as GenericError).errors);
          return;
        }

        setComments(data.comments);
      } catch (err) {
        setCommentErrors({ general: [stdErrorMsg] });
      } finally {
        setIsCommentLoading(false);
      }
    }

    getArticle();
    getComments();
  }, [isAuthLoading, slug]);

  async function handleDelete() {
    if (!isAuthor) return;
    try {
      const res = await fetch(getAPIEndpoint(`articles/${article?.slug}`), {
        method: "DELETE",
        headers: headers,
      });

      if (!res.ok) {
        const data: GenericError = await res.json();
        setErrors(data.errors);
        return;
      }

      navigate("/");
    } catch (err) {
      setErrors({ general: [stdErrorMsg] });
    }
  }

  async function handleFollow() {
    if (!user) {
      navigate("/login");
      return;
    }

    setErrors(null);
    const method = article?.author.following ? "DELETE" : "POST";
    try {
      const res = await fetch(
        getAPIEndpoint(`profiles/${article?.author.username}/follow`),
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

      setArticle((prev) =>
        prev
          ? {
              ...prev,
              author: { ...prev.author, following: data.profile.following },
            }
          : prev,
      );
    } catch (err) {
      setErrors({ general: [stdErrorMsg] });
    }
  }

  async function handleFavorite() {
    if (!user) return;
    if (article === null) return;

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

      setArticle(data.article);
    } catch (err) {
      setErrors({ general: [stdErrorMsg] });
    }
  }

  async function handleAddComment(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!user) {
      navigate("/login");
      return;
    }

    setCommentErrors(null);
    try {
      const res = await fetch(
        getAPIEndpoint(`articles/${article?.slug}/comments`),
        {
          method: "POST",
          headers: headers,
          body: JSON.stringify({ comment: { body: commentInput } }),
        },
      );

      const data: CommentResponse | GenericError = await res.json();

      if (!res.ok || isErrorResponse(data)) {
        setCommentErrors((data as GenericError).errors);
        return;
      }

      setComments((prev) => [...prev, data.comment]);
      setCommentInput("");
    } catch (err) {
      setCommentErrors({ general: [stdErrorMsg] });
    }
  }

  async function onDeleteComment(comment: Comment) {
    if (user?.username !== comment.author.username) return;

    setCommentErrors(null);
    try {
      const res = await fetch(
        getAPIEndpoint(`articles/${article?.slug}/comments/${comment.id}`),
        {
          method: "DELETE",
          headers: headers,
        },
      );

      if (!res.ok) {
        const data: GenericError = await res.json();
        setCommentErrors(data.errors);
        return;
      }

      setComments((prev) => {
        return prev.filter((old) => old.id !== comment.id);
      });
    } catch (err) {
      setCommentErrors({ general: [stdErrorMsg] });
    }
  }

  if (article === null) return null;

  const articleMeta = (
    <div className="article-meta">
      <Link to={`/profile/${article.author.username}`}>
        <img src={userProfilePicture(article.author.image)} />
      </Link>
      <div className="info">
        <Link to={`/profile/${article.author.username}`} className="author">
          {article.author.username}
        </Link>
        <span className="date">{formatDate(article.createdAt)}</span>
      </div>

      {isAuthor ? (
        <>
          <Link
            to={`/editor/${article.slug}`}
            className="btn btn-sm btn-outline-secondary"
          >
            <i className="ion-edit"></i> Edit Article
          </Link>
          <button
            onClick={() => handleDelete()}
            className="btn btn-sm btn-outline-danger"
          >
            <i className="ion-trash-a"></i> Delete Article
          </button>{" "}
        </>
      ) : (
        <>
          <button
            onClick={() => handleFollow()}
            className="btn btn-sm btn-outline-secondary action-btn"
          >
            <i
              className={`ion-${
                article.author.following ? "minus" : "plus"
              }-round`}
            ></i>
            &nbsp;{" "}
            {article.author.following
              ? "Unfollow" + ` ${article.author.username}`
              : "Follow" + ` ${article.author.username}`}
          </button>
          &nbsp;&nbsp;
          <button
            onClick={() => handleFavorite()}
            className="btn btn-sm btn-outline-primary"
          >
            <i
              className={article.favorited ? "ion-heart-broken" : "ion-heart"}
            ></i>{" "}
            &nbsp; {article.favorited ? "Unfavorite Post" : "Favorite Post"}{" "}
            <span className="counter">({article.favoritesCount})</span>
          </button>
        </>
      )}
    </div>
  );

  return (
    <div className="article-page">
      <div className="banner">
        <div className="container">
          <h1>{article.title}</h1>
          {isLoading && <p>fetching data from server...</p>}

          <ErrorMessages errors={errors} />
          {articleMeta}
        </div>
      </div>

      <div className="container page">
        <div className="row article-content">
          <div className="col-md-12">
            <ReactMarkdown>{article.body}</ReactMarkdown>
            <ul className="tag-list">
              {article.tagList.map((tag) => {
                return (
                  <li
                    key={`article-tag-${tag}`}
                    className="tag-default tag-pill tag-outline"
                  >
                    {tag}
                  </li>
                );
              })}
            </ul>
          </div>
        </div>

        <hr />

        <div className="article-actions">{articleMeta}</div>

        <div className="row">
          <div className="col-xs-12 col-md-8 offset-md-2">
            <form
              onSubmit={(event) => handleAddComment(event)}
              className="card comment-form"
            >
              <div className="card-block">
                <textarea
                  className="form-control"
                  placeholder="Write a comment..."
                  rows={3}
                  name="body"
                  value={commentInput}
                  onChange={(e) => setCommentInput(e.target.value)}
                ></textarea>
              </div>
              <div className="card-footer">
                <img
                  src={userProfilePicture(user !== null ? user.image : null)}
                  className="comment-author-img"
                />
                <button type="submit" className="btn btn-sm btn-primary">
                  Post Comment
                </button>
              </div>
            </form>

            {isCommentLoading && (
              <p>fetching comments for article from server...</p>
            )}

            <ErrorMessages errors={commentErrors} />
            {comments.length > 0 &&
              comments.map((comment) => {
                return (
                  <CommentComponent
                    key={`article-comment-${comment.id}`}
                    comment={comment}
                    isCommentAuthor={comment.author.username === user?.username}
                    onDeleteComment={onDeleteComment}
                  />
                );
              })}
          </div>
        </div>
      </div>
    </div>
  );
}
