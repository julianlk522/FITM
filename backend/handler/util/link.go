package handler

import (
	"errors"
	"log"
	"oitm/db"
	"oitm/model"
	"oitm/query"
	"regexp"
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

// Get Links
func GetIDsOfLinksHavingCats(cats_str string) (link_ids []string, err error) {
	link_ids_sql := query.NewLinkIDs(cats_str)
	if link_ids_sql.Error != nil {
		err = link_ids_sql.Error
	}

	rows, err := db.Client.Query(link_ids_sql.Text)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var lid string
		if err := rows.Scan(&lid); err != nil {
			log.Fatal(err)
		}

		link_ids = append(link_ids, lid)
	}

	return link_ids, err
}

func RenderZeroLinks(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, &model.PaginatedLinks[model.Link]{NextPage: -1})
	render.Status(r, http.StatusOK)
}

func ScanLinks[T model.LinkSignedIn | model.Link](get_links_sql *query.TopLinks, req_user_id string) (*[]T, error) {
	var links interface{}

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

		links = &signed_in_links

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
	}

	return links.(*[]T), nil
}

func RenderPaginatedLinks[T model.LinkSignedIn | model.Link](links *[]T, page int, w http.ResponseWriter, r *http.Request) {
	if len(*links) == 0 {
		RenderZeroLinks(w, r)
	} else if len(*links) == LINKS_PAGE_LIMIT+1 {
		sliced := (*links)[:LINKS_PAGE_LIMIT]
		render.JSON(w, r, &model.PaginatedLinks[T]{
			Links:    &sliced,
			NextPage: page + 1,
		})
	} else {
		render.JSON(w, r, &model.PaginatedLinks[T]{
			Links:    links,
			NextPage: -1,
		})
	}
}

// Add link
func ResolveAndAssignURL(url string, request *model.NewLinkRequest) (*http.Response, error) {
	has_protocol_regex, err := regexp.Compile(`^(http(s?)\:\/\/)`)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	var ErrRedirect error = errors.New("invalid link destination: redirect detected")

	// Protocol specified: check as-is
	if has_protocol_regex.MatchString(url) {
		resp, err = http.Get(url)
		if err != nil || resp.StatusCode == 404 {
			return nil, InvalidURLError(url)
		} else if IsRedirect(resp.StatusCode) {
			return nil, ErrRedirect
		}

		// Protocol not specified: try https then http
	} else {

		// https
		modified_url := "https://" + url
		resp, err = http.Get(modified_url)
		if err != nil || resp.StatusCode == 404 {

			// http
			modified_url = "http://" + url
			resp, err = http.Get(modified_url)
			if err != nil || resp.StatusCode == 404 {
				return nil, InvalidURLError(modified_url)
			} else if IsRedirect(resp.StatusCode) {
				return nil, ErrRedirect
			}

		} else if IsRedirect(resp.StatusCode) {
			return nil, ErrRedirect
		}
	}

	// save updated URL after any redirects e.g., to wwww.
	// remove trailing slash
	request.URL = strings.TrimSuffix(resp.Request.URL.String(), "/")

	return resp, nil
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
