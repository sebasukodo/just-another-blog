package handler

import (
	"net/http"
	"strings"

	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type RespondTags struct {
	Tags []string `json:"tags"`
}

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {

	tags, err := h.DbQueries.GetTags(r.Context())
	if err != nil {
		h.RespondWithDatabaseError(w, "tags", err)
		return
	}

	respBody := RespondTags{
		Tags: tags,
	}

	h.RespondWithJSON(w, 200, respBody)

}

func (h *Handler) saveTagsToDatabase(r *http.Request, articleID int64, tags []string) error {
	noDuplication := make(map[string]struct{})

	for _, tag := range tags {
		if _, ok := noDuplication[tag]; ok {
			continue
		}
		noDuplication[tag] = struct{}{}

		newTag, err := h.DbQueries.CreateTags(r.Context(), strings.ToLower(tag))
		if err != nil {
			return err
		}
		_, err = h.DbQueries.CreateArticleTags(r.Context(), database.CreateArticleTagsParams{
			ArticleID: articleID,
			TagID:     newTag.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
