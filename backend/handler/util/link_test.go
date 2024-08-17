package handler

import (
	"oitm/model"
	"oitm/query"
	"strings"
	"testing"
)

func TestGetIDsOfLinksHavingCategories(t *testing.T) {

	// single cat
	test_cats_str := test_single_cat[0]
	link_ids, err := GetIDsOfLinksHavingCategories(test_cats_str)
	if err != nil {
		t.Fatal(err)
	}

	if len(link_ids) == 0 {
		t.Fatalf("no link IDs returned for cat %s", test_cats_str)
	}

	// confirm that links with IDs have cat
	for _, lid := range link_ids {
		var link_cats string
		err = TestClient.QueryRow(`
			SELECT global_cats 
			FROM Links 
			WHERE id = ?`, 
		lid).Scan(&link_cats)
		if err != nil {
			t.Fatal(err)
		} else if !strings.Contains(link_cats, test_cats_str) {
			t.Fatalf(
				"link %s cats (%s) do not contain %s", 
				lid,
				link_cats, 
				test_cats_str,
			)
		}
	}

	// multiple cats
	test_cats_str = strings.Join(test_multiple_cats, ",")
	link_ids, err = GetIDsOfLinksHavingCategories(test_cats_str)
	if err != nil {
		t.Fatal(err)
	}

	if len(link_ids) == 0 {
		t.Fatalf("no link IDs returned for cats %s", test_cats_str)
	}

	// confirm that links with IDs have cat
	for _, lid := range link_ids {
		var link_cats string
		err = TestClient.QueryRow(`
			SELECT global_cats 
			FROM Links 
			WHERE id = ?`, 
		lid).Scan(&link_cats)
		if err != nil {
			t.Fatal(err)
		} else if !strings.Contains(link_cats, test_multiple_cats[0]) || !strings.Contains(link_cats, test_multiple_cats[1]) {
			t.Fatalf(
				"link %s cats (%s) do not contain all test cats: %s", 
				lid,
				link_cats, 
				test_cats_str,
			)
		}
	}
}

func TestScanLinks(t *testing.T) {
	links_sql := query.NewTopLinks()
	// NewTopLinks().Error tested in query/link_test.go
	
	// signed out
	links_signed_out, err := ScanLinks[model.Link](links_sql, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(*links_signed_out) == 0 { 
		t.Fatal("no links")
	}

	// signed in
	links_signed_in, err := ScanLinks[model.LinkSignedIn](links_sql, test_req_user_id)
	if err != nil {
		t.Fatal(err)
	}

	if len(*links_signed_in) == 0 { 
		t.Fatal("no links")
	}
}

func TestResolveAndAssignURL(t *testing.T) {
	var test_urls = []struct{
		URL string
		Valid bool
	}{
		{"abc.com", true},
		{"www.abc.com", true},
		{"https://www.abc.com", true},
		{"about.google.com", true},
		{"julianlk.com/notreal", false},
		{"gobblety gook", false},
	}
	test_request := &model.NewLinkRequest{
		NewLink: &model.NewLink{
			URL: "",
			Categories: "",
			Summary: "",
		},
	}

	for _, u := range test_urls {
		_, err := ResolveAndAssignURL(u.URL, test_request)
		if u.Valid && err != nil {
			t.Fatal(err)
		} else if !u.Valid && err == nil {
			t.Fatalf("expected error for url %s", u.URL)
		}
	}
}

func TestURLAlreadyAdded(t *testing.T) {
	var test_urls = []struct{
		URL string
		Added bool
	}{
		{"https://stackoverflow.co/", true},
		{"https://www.ronjarzombek.com", true},
		{"https://somethingnotonfitm", false},
		{"jimminy jillickers", false},
	}

	for _, u := range test_urls {
		added := URLAlreadyAdded(u.URL)
		if u.Added && !added {
			t.Fatalf("expected url %s to be added", u.URL)
		} else if !u.Added && added {
			t.Fatalf("%s NOT added, expected error", u.URL)
		}
	}
}

func TestAssignMetadata(t *testing.T) {
	mock_metas := []HTMLMeta{
		// Auto Summary should be og:description,
		// og:image should be set
		{
			Title:         "title",
			Description:   "description",
			OGTitle:       "og:title",
			OGDescription: "og:description",
			OGImage:       "https://i.ytimg.com/vi/L4gaqVH0QHU/maxresdefault.jpg",
			OGAuthor:      "",
			OGPublisher:   "",
			OGSiteName:    "og:site_name",
		},
		// Auto Summary should be description
		{
			Title:         "",
			Description:   "description",
			OGTitle:       "",
			OGDescription: "",
			OGImage:       "",
			OGAuthor:      "",
			OGSiteName:    "",
			OGPublisher:   "",
		},
		// Auto Summary should be og:title
		{
			Title:         "title",
			Description:   "",
			OGTitle:       "og:title",
			OGDescription: "",
			OGImage:       "",
			OGAuthor:      "",
			OGSiteName:    "",
			OGPublisher:   "",
		},
		// Auto Summary should be title
		{
			Title:         "title",
			Description:   "",
			OGTitle:       "",
			OGDescription: "",
			OGImage:       "",
			OGAuthor:      "",
			OGSiteName:    "goopis",
			OGPublisher:   "",
		},
		// Auto Summary should be goopis
		// og:image should be set
		{
			Title:         "",
			Description:   "",
			OGTitle:       "",
			OGDescription: "",
			OGImage:       "https://i.ytimg.com/vi/XdfoXdzGmr0/maxresdefault.jpg",
			OGAuthor:      "",
			OGSiteName:    "goopis",
			OGPublisher:   "",
		},
	}

	for i, meta := range mock_metas {
		mock_request := &model.NewLinkRequest{
			NewLink: &model.NewLink{
				URL: "",
				Categories: "",
				Summary: "",
			},
		}

		AssignMetadata(meta, mock_request)

		switch i {
			case 0:
				if mock_request.AutoSummary != "og:description" {
					t.Fatalf("og:description provided but auto summary set to: %s", mock_request.AutoSummary)
				} else if mock_request.ImgURL != "https://i.ytimg.com/vi/L4gaqVH0QHU/maxresdefault.jpg" {
					t.Fatal("expected og:image to be set")
				}
			case 1:
				if mock_request.AutoSummary != "description" {
					t.Fatalf("description provided but auto summary set to: %s", mock_request.AutoSummary)
				}
			case 2:
				if mock_request.AutoSummary != "og:title" {
					t.Fatalf("og:title provided but auto summary set to: %s", mock_request.AutoSummary)
				}
			case 3:
				if mock_request.AutoSummary != "title" {
					t.Fatalf("title provided but auto summary set to: %s", mock_request.AutoSummary)
				}
			case 4:
				if mock_request.AutoSummary != "goopis" {
					t.Fatalf("goopis provided but auto summary set to: %s", mock_request.AutoSummary)
				} else if mock_request.ImgURL != "https://i.ytimg.com/vi/XdfoXdzGmr0/maxresdefault.jpg" {
					t.Fatal("expected og:image to be set")
				}
			default:
				t.Fatal("unhandled case, you f'ed up dawg")
		}
	}
}

// IsRedirect / AssignSortedCategories are pretty simple
// don't really need tests

// Like / unlike link
func TestUserSubmittedLink(t *testing.T) {
	var test_links = []struct{
		ID string
		SubmittedByTestUser bool
	}{
		// user goolian submitted links with ID 7, 13, 23
		// (not 0, 1, or 86)
		{"7", true},
		{"13", true},
		{"23", true},
		{"0", false},
		{"1", false},
		{"86", false},
	}

	for _, l := range test_links {
		return_true := UserSubmittedLink(test_login_name, l.ID)
		if l.SubmittedByTestUser && !return_true {
			t.Fatalf("expected link %s to be submitted by user", l.ID)
		} else if !l.SubmittedByTestUser && return_true {
			t.Fatalf("%s NOT submitted by user, expected error", l.ID)
		}
	}
}

func TestUserHasLikedLink(t *testing.T) {
	var test_links = []struct{
		ID string
		LikedByTestUser bool
	}{
		// user goolian liked links with ID 21, 24, 32
		// (not 9, 11, or 15)
		{"21", true},
		{"24", true},
		{"32", true},
		{"9", false},
		{"11", false},
		{"15", false},
	}

	for _, l := range test_links {
		return_true := UserHasLikedLink(test_user_id, l.ID)
		if l.LikedByTestUser && !return_true {
			t.Fatalf("expected link %s to be liked by user", l.ID)
		} else if !l.LikedByTestUser && return_true {
			t.Fatalf("%s NOT liked by user, expected error", l.ID)
		}
	}
}

// Copy link
func TestUserHasCopiedLink(t *testing.T) {
	var test_links = []struct{
		ID string
		CopiedByTestUser bool
	}{
		// user goolian copied links with ID 19, 31, 32
		// (not 0, 1, or 99)
		{"19", true},
		{"31", true},
		{"32", true},
		{"0", false},
		{"1", false},
		{"104", false},
	}

	for _, l := range test_links {
		return_true := UserHasCopiedLink(test_user_id, l.ID)
		if l.CopiedByTestUser && !return_true {
			t.Fatalf("expected link %s to be copied by user", l.ID)
		} else if !l.CopiedByTestUser && return_true {
			t.Fatalf("%s NOT copied by user, expected error", l.ID)
		}
	}
}