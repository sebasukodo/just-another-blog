import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { useParams } from "react-router";
import { ErrorMessages } from "~/components/errorMessages";
import { useAuth } from "~/hooks/useAuth";
import type {
  Article,
  ArticleFormData,
  ArticleResponse,
} from "~/types/articles";
import type { GenericError } from "~/types/error";
import { getAPIEndpoint, isErrorResponse, stdErrorMsg } from "~/utils";
import type { Route } from "./+types/editor";

export function meta(_args: Route.MetaArgs) {
  return [
    { title: "Edit/Create Articles" },
    {
      name: "description",
      content: "this is a page to edit or create articles",
    },
  ];
}

export default function Editor() {
  const navigate = useNavigate();
  const { user, isAuthLoading } = useAuth();
  const { slug } = useParams();
  const editMode = slug ? true : false;

  const [article, setArticle] = useState<Article | null>(null);
  const [formData, setFormData] = useState<ArticleFormData>({
    title: "",
    description: "",
    body: "",
  });

  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState("");

  const [errors, setErrors] = useState<Record<string, string[]> | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const headers = useMemo(() => {
    const h: HeadersInit = { "Content-Type": "application/json" };
    if (user?.token) {
      h.Authorization = `Token ${user.token}`;
    }
    return h;
  }, [user]);

  useEffect(() => {
    if (isAuthLoading) return;
    if (article !== null && user?.username !== article.author.username) {
      navigate("/");
    }
  }, [article, user, isAuthLoading, navigate]);

  useEffect(() => {
    if (!editMode) return;

    if (!isAuthLoading && !user) {
      navigate("/login");
    }

    async function getArticle() {
      setIsLoading(true);
      setErrors(null);

      try {
        const res = await fetch(getAPIEndpoint(`articles/${slug}`), {
          method: "GET",
        });

        const data: ArticleResponse | GenericError = await res.json();

        if (!res.ok || isErrorResponse(data)) {
          setErrors((data as GenericError).errors);
          return;
        }

        setArticle(data.article);
        setFormData({
          title: data.article.title,
          description: data.article.description,
          body: data.article.body ?? "",
        });
        setTags(data.article.tagList);
      } catch {
        setErrors({ general: [stdErrorMsg] });
      } finally {
        setIsLoading(false);
      }
    }
    getArticle();
  }, [editMode, isAuthLoading, slug, user, navigate]);

  async function handleSubmit() {
    setIsLoading(true);
    setErrors(null);

    let method = "POST";
    let url = "articles";
    if (editMode) {
      method = "PUT";
      url = `articles/${slug}`;
    }

    const requestBody = {
      article: {
        title: formData.title,
        description: formData.description,
        body: formData.body,
        tagList: tags,
      },
    };

    try {
      const res = await fetch(getAPIEndpoint(url), {
        method: method,
        headers: headers,
        body: JSON.stringify(requestBody),
      });

      const data: ArticleResponse | GenericError = await res.json();

      if (!res.ok || isErrorResponse(data)) {
        setErrors((data as GenericError).errors);
        return;
      }

      setArticle(data.article);
      navigate(`/article/${data.article.slug}`);
    } catch {
      setErrors({ general: [stdErrorMsg] });
    } finally {
      setIsLoading(false);
    }
  }

  function handleFormChange(
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) {
    const { name, value } = event.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  }

  function handleAddTags(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Enter") {
      event.preventDefault();

      const trimmed = tagInput.trim();
      if (trimmed !== "" && !tags.includes(trimmed)) {
        setTags((prev) => [...prev, trimmed]);
      }

      setTagInput("");
    }
  }

  function handleRemoveTags(tagToRemove: string) {
    setTags((prev) => prev.filter((tag) => tag !== tagToRemove));
  }

  if (isAuthLoading) return null;
  if (!user) return null;
  if (article !== null && user.username !== article.author.username)
    return null;

  return (
    <div className="editor-page">
      <div className="container page">
        <div className="row">
          <div className="col-md-10 offset-md-1 col-xs-12">
            {isLoading && <p>fetching data from server...</p>}
            <ErrorMessages errors={errors} />

            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleSubmit();
              }}
            >
              <fieldset>
                <fieldset className="form-group">
                  <input
                    type="text"
                    className="form-control form-control-lg"
                    placeholder="Article Title"
                    value={formData.title}
                    onChange={(e) => handleFormChange(e)}
                    name="title"
                  />
                </fieldset>
                <fieldset className="form-group">
                  <input
                    type="text"
                    className="form-control"
                    placeholder="What's this article about?"
                    value={formData.description}
                    onChange={(e) => handleFormChange(e)}
                    name="description"
                  />
                </fieldset>
                <fieldset className="form-group">
                  <textarea
                    className="form-control"
                    rows={8}
                    placeholder="Write your article (in markdown)"
                    value={formData.body}
                    onChange={(e) => handleFormChange(e)}
                    name="body"
                  ></textarea>
                </fieldset>
                <fieldset className="form-group">
                  <input
                    type="text"
                    className="form-control"
                    placeholder="Enter tags"
                    onKeyDown={handleAddTags}
                    value={tagInput}
                    onChange={(e) => setTagInput(e.target.value)}
                  />
                  {tags.length > 0 && (
                    <div className="tag-list">
                      {tags.map((tag, index) => {
                        return (
                          <span
                            key={`article-tag-${tag}-${index}`}
                            className="tag-default tag-pill"
                          >
                            {" "}
                            <i
                              className="ion-close-round"
                              onClick={() => handleRemoveTags(tag)}
                            ></i>{" "}
                            {tag}
                          </span>
                        );
                      })}
                    </div>
                  )}
                </fieldset>
                <button
                  className="btn btn-lg pull-xs-right btn-primary"
                  type="submit"
                >
                  Publish Article
                </button>
              </fieldset>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
