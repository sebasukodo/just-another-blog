import { Link } from "react-router";
import type { Article } from "~/types/articles";
import { formatDate, userProfilePicture } from "~/utils";

export default function ArticlePreview({ article }: { article: Article }) {
  return (
    <div className="article-preview">
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
        <button className="btn btn-outline-primary btn-sm pull-xs-right">
          <i className="ion-heart"></i> {article.favoritesCount}
        </button>
      </div>
      <Link to={`/article/${article.slug}`} className="preview-link">
        <h1>{article.title}</h1>
        <p>{article.description}</p>
        <span>Read more...</span>
        <ul className="tag-list">
          {article.tagList.map((tag, index) => {
            return (
              <li
                key={`${article.slug}-tag-${index}`}
                className="tag-default tag-pill tag-outline"
              >
                {tag}
              </li>
            );
          })}
        </ul>
      </Link>
    </div>
  );
}
