package handler

import (
	"strings"

	"github.com/julianlk522/fitm/db"
	e "github.com/julianlk522/fitm/error"
	"github.com/julianlk522/fitm/model"
	"github.com/julianlk522/fitm/query"

	"database/sql"
	"fmt"

	"net/http"

	"github.com/go-chi/render"
)

func RenderZeroLinks(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, &model.PaginatedLinks[model.Link]{NextPage: -1})
	render.Status(r, http.StatusOK)
}

func ScanLinks[T model.Link | model.LinkSignedIn](get_links_sql *query.TopLinks) (*[]T, error) {
	rows, err := db.Client.Query(get_links_sql.Text)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var links interface{}

	switch any(new(T)).(type) {
		case *model.Link:
			var signed_out_links = []model.Link{}

			for rows.Next() {
				i := model.Link{}
				err := rows.Scan(
					&i.ID,
					&i.URL,
					&i.SubmittedBy,
					&i.SubmitDate,
					&i.Cats,
					&i.Summary,
					&i.SummaryCount,
					&i.TagCount,
					&i.LikeCount,
					&i.ImgURL,
				)
				if err != nil {
					return nil, err
				}
				signed_out_links = append(signed_out_links, i)
			}

			links = &signed_out_links

		case *model.LinkSignedIn:
			var signed_in_links = []model.LinkSignedIn{}

			for rows.Next() {
				i := model.LinkSignedIn{}
				if err := rows.Scan(
					&i.ID,
					&i.URL,
					&i.SubmittedBy,
					&i.SubmitDate,
					&i.Cats,
					&i.Summary,
					&i.SummaryCount,
					&i.TagCount,
					&i.LikeCount,
					&i.ImgURL,
					&i.IsLiked,
					&i.IsCopied,
				); err != nil {
					return nil, err
				}

				signed_in_links = append(signed_in_links, i)
			}

			links = &signed_in_links
	}
	

	return links.(*[]T), nil
}

func PaginateLinks[T model.LinkSignedIn | model.Link](links *[]T, page int) (interface{}) {
	if links == nil || len(*links) == 0 {
		return &model.PaginatedLinks[model.Link]{NextPage: -1}
	}

	if len(*links) == query.LINKS_PAGE_LIMIT+1 {
		sliced := (*links)[0:query.LINKS_PAGE_LIMIT]
		return &model.PaginatedLinks[T]{
			NextPage: page + 1,
			Links:    &sliced,
		}
	} else {
		return &model.PaginatedLinks[T]{
			NextPage: -1,
			Links:    links,
		}
	}
}

// Add link (non-YT) 
func ObtainURLMetaData(request *model.NewLinkRequest) error {
	resp, err := GetResolvedURLResponse(request.NewLink.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// save updated URL (after any redirects e.g., to wwww.)
	request.URL = resp.Request.URL.String()

	// remove trailing slash
	request.URL = strings.TrimSuffix(request.URL, "/")

	meta := ExtractMetaFromHTMLTokens(resp.Body)
	AssignMetadata(meta, request)

	return nil
}
func GetResolvedURLResponse(url string) (*http.Response, error) {
	protocols := []string{"", "https://", "http://"}

	for _, prefix := range protocols {
		fullURL := prefix + url
		resp, err := http.Get(fullURL)

		if err != nil || resp.StatusCode == http.StatusNotFound {
			continue
		} else if IsRedirect(resp.StatusCode) {
			return nil, e.ErrRedirect
		}

		// URL is valid: return
		return resp, nil
	}

	return nil, InvalidURLError(url)
}

func InvalidURLError(url string) error {
	return fmt.Errorf("invalid URL: %s", url)
}

func AssignMetadata(meta HTMLMeta, link_data *model.NewLinkRequest) {
	switch {
	case meta.OGDescription != "":
		link_data.AutoSummary = meta.OGDescription
	case meta.Description != "":
		link_data.AutoSummary = meta.Description
	case meta.OGTitle != "":
		link_data.AutoSummary = meta.OGTitle
	case meta.Title != "":
		link_data.AutoSummary = meta.Title
	case meta.OGSiteName != "":
		link_data.AutoSummary = meta.OGSiteName
	}

	if meta.OGImage != "" {
		resp, err := http.Get(meta.OGImage)
		if err == nil && resp.StatusCode != 404 && !IsRedirect(resp.StatusCode) {
			link_data.ImgURL = meta.OGImage
		}
	}
}

func IsRedirect(status_code int) bool {
	return status_code > 299 && status_code < 400
}

func LinkAlreadyAdded(url string) (bool, string) {
	var id sql.NullString

	err := db.Client.QueryRow("SELECT id FROM Links WHERE url = ?", url).Scan(&id)

	if err == nil && id.Valid {
		return true, id.String
	} else {
		return false, ""
	}
}

func IncrementSpellfixRanksForCats(cats []string) error {
	for _, cat := range cats {
		_, err := db.Client.Exec(
			"UPDATE global_cats_spellfix SET rank = rank + 1 WHERE word = ?;",
			cat,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Like / unlike link
func UserSubmittedLink(login_name string, link_id string) bool {
	var sb sql.NullString
	err := db.Client.QueryRow("SELECT submitted_by FROM Links WHERE id = ?;", link_id).Scan(&sb)

	if err != nil {
		return false
	}

	return sb.String == login_name
}

func UserHasLikedLink(user_id string, link_id string) bool {
	var l sql.NullString
	err := db.Client.QueryRow("SELECT id FROM 'Link Likes' WHERE user_id = ? AND link_id = ?;", user_id, link_id).Scan(&l)

	return err == nil && l.Valid
}

// Copy link
func UserHasCopiedLink(user_id string, link_id string) bool {
	var l sql.NullString
	err := db.Client.QueryRow("SELECT id FROM 'Link Copies' WHERE user_id = ? AND link_id = ?;", user_id, link_id).Scan(&l)

	return err == nil && l.Valid
}
