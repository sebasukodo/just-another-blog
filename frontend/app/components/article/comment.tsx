import { Link } from "react-router";
import { useAuth } from "~/context/auth";
import type { Comment } from "~/types/articles";
import { formatDate, userProfilePicture } from "~/utils";

export default function CommentComponent({
  comment,
  isCommentAuthor,
  onDeleteComment,
}: {
  comment: Comment;
  isCommentAuthor: boolean;
  onDeleteComment: (comment: Comment) => void;
}) {
  return (
    <div className="card">
      <div className="card-block">
        <p className="card-text">{comment.body}</p>
      </div>
      <div className="card-footer">
        <Link
          to={`/profile/${comment.author.username}`}
          className="comment-author"
        >
          <img
            src={userProfilePicture(comment.author.image)}
            className="comment-author-img"
          />
        </Link>
        &nbsp;
        <Link
          to={`/profile/${comment.author.username}`}
          className="comment-author"
        >
          {comment.author.username}
        </Link>
        <span className="date-posted">{formatDate(comment.createdAt)}</span>
        {isCommentAuthor && (
          <span
            onClick={() => onDeleteComment(comment)}
            className="mod-options"
          >
            <i className="ion-trash-a"></i>
          </span>
        )}
      </div>
    </div>
  );
}
