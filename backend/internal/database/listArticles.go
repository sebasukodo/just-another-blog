package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ListArticle struct {
	ID               int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AuthorID         uuid.UUID
	Slug             string
	Title            string
	Description      string
	AuthorUsername   string
	AuthorBio        sql.NullString
	AuthorImage      sql.NullString
	Tags             []string
	AuthorIsFollowed bool
	IsFavorited      bool
	FavoritesCount   int64
}

func ScanListArticles(rows *sql.Rows) ([]ListArticle, error) {
	defer rows.Close()

	var articles []ListArticle
	for rows.Next() {
		var la ListArticle
		err := rows.Scan(
			&la.ID,
			&la.CreatedAt,
			&la.UpdatedAt,
			&la.AuthorID,
			&la.Slug,
			&la.Title,
			&la.Description,
			&la.AuthorUsername,
			&la.AuthorBio,
			&la.AuthorImage,
			pq.Array(&la.Tags),
			&la.AuthorIsFollowed,
			&la.IsFavorited,
			&la.FavoritesCount,
		)
		if err != nil {
			return nil, err
		}
		articles = append(articles, la)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}
