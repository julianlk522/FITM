package handler

import (
	"oitm/db"
	e "oitm/error"
	"oitm/model"
	"oitm/query"
	"slices"
	"strings"

	"database/sql"
	"fmt"

	"net/http"

	"github.com/go-chi/render"
)

const (
	LINKS_PAGE_LIMIT    int = 20
	CATEGORY_PAGE_LIMIT int = 15
)

func RenderZeroLinks(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, &model.PaginatedLinks[model.Link]{NextPage: -1})
	render.Status(r, http.StatusOK)
}

func ScanLinks[T model.LinkSignedIn | model.Link](get_links_sql *query.TopLinks, req_user_id string) (interface{}, error) {
	rows, err := db.Client.Query(get_links_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	switch any(new(T)).(type) {
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
				&i.IsTagged,
			); err != nil {
				return nil, err
			}

			signed_in_links = append(signed_in_links, i)
		}

		return &signed_in_links, nil

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

		return &signed_out_links, nil
	}

	return nil, e.ErrCouldNotScanLinks
}

func RenderPaginatedLinks(li interface{}, page int, w http.ResponseWriter, r *http.Request) () {

	// Note: this could be generic, and the duplicate code could be removed,
	// but that requires duplicating code in the GetLinks handler to handle
	// the generic
    switch l := li.(type) {
    case *[]model.LinkSignedIn:
		if len(*l) == 0 {
			RenderZeroLinks(w, r)
		}

		if len(*l) == LINKS_PAGE_LIMIT+1 {
			sliced := (*l)[0:LINKS_PAGE_LIMIT]
			paginated_links := &model.PaginatedLinks[model.LinkSignedIn]{
				NextPage: page + 1,
				Links: &sliced,
			}
			render.JSON(w, r, paginated_links)
		} else {
			render.JSON(w, r, &model.PaginatedLinks[model.LinkSignedIn]{
				NextPage: -1,
				Links:    l,
			})
		}
    case *[]model.Link:
		if len(*l) == 0 {
			RenderZeroLinks(w, r)
		}

		if len(*l) == LINKS_PAGE_LIMIT+1 {
			sliced := (*l)[0:LINKS_PAGE_LIMIT]
			paginated_links := &model.PaginatedLinks[model.Link]{
				NextPage: page + 1,
				Links: &sliced,
			}
			render.JSON(w, r, paginated_links)
		} else {
			render.JSON(w, r, &model.PaginatedLinks[model.Link]{
				NextPage: -1,
				Links:    l,
			})
		}
    default:
        render.Render(w, r, e.ErrInvalidRequest(e.ErrCouldNotPaginateLinks))
        return
    }
}

// Add link
func ResolveURL(url string) (*http.Response, error) {
	protocols := []string{"", "https://", "http://"}

	for _, prefix := range protocols {
		fullURL := prefix + url
		resp, err := http.Get(fullURL)

		if err != nil || resp.StatusCode == http.StatusNotFound {
			continue
		} else if IsRedirect(resp.StatusCode) {
			return nil, e.ErrRedirect
		}

		// URL is valid:  and return
		return resp, nil
	}

	return nil, InvalidURLError(url)
}

func InvalidURLError(url string) error {
	return fmt.Errorf("invalid URL: %s", url)
}

func URLAlreadyAdded(url string) bool {
	var u sql.NullString

	err := db.Client.QueryRow("SELECT url FROM Links WHERE url = ?", url).Scan(&u)
	return err == nil && u.Valid
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

func AssignSortedCats(unsorted_cats string, link *model.NewLinkRequest) {
	split_cats := strings.Split(unsorted_cats, ",")
	slices.Sort(split_cats)

	link.Cats = strings.Join(split_cats, ",")
}

// Add link

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
