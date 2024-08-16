package handler

import (
	"oitm/db"

	"database/sql"
	"net/http"
	e "oitm/error"
	m "oitm/middleware"
	"oitm/model"
	"oitm/query"
	"slices"
	"strings"
)

// Get treasure map
func UserExists(login_name string) (bool, error) {
	var u sql.NullString
	err := db.Client.QueryRow("SELECT id FROM Users WHERE login_name = ?;", login_name).Scan(&u)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func GetTmapForUser[T model.TmapLink | model.TmapLinkSignedIn](login_name string, r *http.Request) (interface{}, error) {
	submitted_sql := query.NewTmapSubmitted(login_name)
	copied_sql := query.NewTmapCopied(login_name)
	tagged_sql := query.NewTmapTagged(login_name)
	
	cats_params := r.URL.Query().Get("cats") 
	has_cat_filter := cats_params != ""

	var cats []string
	var profile *model.Profile
	if has_cat_filter {
		cats = strings.Split(cats_params, ",")
	} else {
		var err error
		profile_sql := query.NewTmapProfile(login_name)
		profile, err = ScanTmapProfile(profile_sql)
		if err != nil {
			return nil, err
		}
	}

	if has_cat_filter {
		submitted_sql = submitted_sql.FromCategories(cats)
		copied_sql = copied_sql.FromCategories(cats)
		tagged_sql = tagged_sql.FromCategories(cats)
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	req_login_name := r.Context().Value(m.LoginNameKey).(string)

	// Requesting user signed in: get IsLiked / IsCopied / IsTagged for each link
	if req_user_id != "" {	
		submitted_sql = submitted_sql.AsSignedInUser(req_user_id, req_login_name)
		copied_sql = copied_sql.AsSignedInUser(req_user_id, req_login_name)
		tagged_sql = tagged_sql.AsSignedInUser(req_user_id, req_login_name)
	}

	submitted, err := ScanTmapLinks[T](submitted_sql.Query)
	if err != nil {
		return nil, err
	}
	copied, err := ScanTmapLinks[T](copied_sql.Query)
	if err != nil {
		return nil, err
	}
	tagged, err := ScanTmapLinks[T](tagged_sql.Query)
	if err != nil {
		return nil, err
	}

	all_links := slices.Concat(*submitted, *copied, *tagged)
	var cat_counts *[]model.CatCount
	if has_cat_filter {
		cat_counts = GetTmapCatCounts(&all_links, cats)
	} else {
		cat_counts = GetTmapCatCounts(&all_links, nil)
	}

	sections := &model.TreasureMapSections[T]{
		Submitted: submitted,
		Copied: copied,
		Tagged: tagged,
		Categories: cat_counts,
	}

	if has_cat_filter {
		    return model.FilteredTreasureMap[T]{
			TreasureMapSections: sections,
		}, nil

	} else {
		return model.TreasureMap[T]{
			Profile: profile, 
			TreasureMapSections: sections,
		}, nil
	}
}

func ScanTmapProfile(sql string) (*model.Profile, error) {
	var u model.Profile
	err := db.Client.
		QueryRow(sql).
		Scan(
			&u.LoginName, 
			&u.About, 
			&u.PFP, 
			&u.Created,
		)
	if err != nil {
		return nil, e.ErrNoUserWithLoginName
	}

	return &u, nil
}

func ScanTmapLinks[T model.TmapLink | model.TmapLinkSignedIn](sql query.Query) (*[]T, error) {
	rows, err := db.Client.Query(sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links interface{}

	switch any(new(T)).(type) {
		case *model.TmapLinkSignedIn:
			var signed_in_links = []model.TmapLinkSignedIn{}

			for rows.Next() {
				l := model.TmapLinkSignedIn{}
				err := rows.Scan(
					&l.ID, 
					&l.URL, 
					&l.SubmittedBy, 
					&l.SubmitDate, 
					&l.Categories, 
					&l.CategoriesFromUser, 
					&l.Summary, 
					&l.SummaryCount, 
					&l.LikeCount, 
					&l.TagCount,
					&l.ImgURL,
					
					// Add IsLiked / IsCopied / IsTagged 
					&l.IsLiked, 
					&l.IsCopied,
					&l.IsTagged) 
				if err != nil {
					return nil, err
				}
				signed_in_links = append(signed_in_links, l)
			}

			links = &signed_in_links

		case *model.TmapLink:
			var signed_out_links = []model.TmapLink{}

			for rows.Next() {
				l := model.TmapLink{}
				err := rows.Scan(
					&l.ID, 
					&l.URL, 
					&l.SubmittedBy, 
					&l.SubmitDate, 
					&l.Categories, 
					&l.CategoriesFromUser, 
					&l.Summary, 
					&l.SummaryCount, 
					&l.LikeCount, 
					&l.TagCount,
					&l.ImgURL)
				if err != nil {
					return nil, err
				}
				signed_out_links = append(signed_out_links, l)
			}

			links = &signed_out_links
	}

	return links.(*[]T), nil	
}

// Get counts of each category found in links
// Omit any categories passed via omitted_cats
// (omit used to retrieve subcategories by passing directly searched categories)
// TODO: refactor to make this clearer
func GetTmapCatCounts[T model.TmapLink | model.TmapLinkSignedIn] (links *[]T, omitted_cats []string) *[]model.CatCount {
	counts := []model.CatCount{}
	found_cats := []string{}
	var found bool

	for _, link := range *links {
		var categories string
		switch l := any(link).(type) {
			case model.TmapLinkSignedIn:
				categories = l.Categories
			case model.TmapLink:
				categories = l.Categories
		}

		for _, cat := range strings.Split(categories, ",") {
			if omitted_cats != nil && slices.Contains(omitted_cats, cat) {
				continue
			}

			found = false
			for _, found_cat := range found_cats {
				if cat == found_cat {
					found = true

					for i, count := range counts {
						if count.Category == cat {
							counts[i].Count++
							break
						}
					}
				}
			}

			if !found {
				counts = append(counts, model.CatCount{Category: cat, Count: 1})

				// add to found categories
				found_cats = append(found_cats, cat)
			}
		}
	}

	SortAndLimitCatCounts(&counts)

	return &counts
}