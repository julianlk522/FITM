package handler

import (
	"oitm/query"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestScanTagPageLink(t *testing.T) {
	tag_page_link_sql := query.NewTagPageLink(test_link_id, test_req_user_id)
	// NewTagPageLink().Error already tested in query/tag_test.go

	link, err := ScanTagPageLink(tag_page_link_sql)
	if err != nil {
		t.Fatal(err)
	}

	// Verify link ID
	var link_id_str = strconv.FormatInt(link.ID, 10)
	if link_id_str != test_link_id {
		t.Fatalf(
			"got link ID %s, want %s", 
			link_id_str, 
			test_link_id,
		)
	}

	// Verify isLiked / isCopied
	liked := UserHasLikedLink(test_req_user_id, test_link_id)
	if liked && !link.IsLiked {
		t.Fatalf("expected link with ID %s to be liked by user", test_link_id)
	} else if !liked && link.IsLiked {
		t.Fatalf("link with ID %s NOT liked by user, expected error", test_link_id)
	}

	copied := UserHasCopiedLink(test_req_user_id, test_link_id)
	if copied && !link.IsCopied {
		t.Fatalf("expected link with ID %s to be copied by user", test_link_id)
	} else if !copied && link.IsCopied {
		t.Fatalf("link with ID %s NOT copied by user, expected error", test_link_id)
	}
}

func TestGetUserTagForLink(t *testing.T) {
	var test_tag = struct{
		LoginName string
		LinkID string
		Cats string
	}{
		LoginName: test_login_name,
		LinkID: "22",
		Cats: "barbie,magic,wow",
	}

	tag, err := GetUserTagForLink(test_tag.LoginName, test_tag.LinkID)
	if err != nil {
		t.Fatal(err)
	} else if tag == nil {
		t.Fatalf(
			"no tag found for user %s and link %s, expected cats %s", 
			test_tag.LoginName, 
			test_tag.LinkID,
			test_tag.Cats,
		)
	}

	// Verify id and cats
	var id, cats string

	err = TestClient.QueryRow(`
		SELECT id, categories 
		FROM Tags 
		WHERE submitted_by = ?
		AND link_id = ?;`, 
	test_tag.LoginName,
	test_tag.LinkID).Scan(
		&id, 
		&cats,
	)
	if err != nil {
		t.Fatal(err)
	} else if tag.ID != id {
		t.Fatalf(
			"got tag ID %s for user %s and link %s, want %s", 
			tag.ID, 
			test_tag.LoginName,
			test_tag.LinkID,
			id,
		)
	} else if tag.Categories != cats {
		t.Fatalf(
			"got cats %s for user %s and link %s, want %s", 
			tag.Categories, 
			test_tag.LoginName,
			test_tag.LinkID,
			cats,
		)
	}
}

func TestScanTagRankings(t *testing.T) {
	var test_rankings = []struct{
		Cats string
		SubmittedBy string
	}{
		{
			Cats: "flowers",
			SubmittedBy: "xyz",
		},
		{
			Cats: "jungle,idk,something",
			SubmittedBy: "nelson",
		},
		{
			Cats: "star,wars",
			SubmittedBy: "boolian",
		},
		{
			Cats: "i,hate,sql",
			SubmittedBy: "Julian",
		},
		{
			Cats: "monkeys,something",
			SubmittedBy: "goolian",
		},
		{
			Cats: "jungle,knights,monkeys,talladega",
			SubmittedBy: "monkey",
		},
	}
	tag_rankings_sql := query.NewTagRankingsForLink(test_link_id)
	// NewTagRankingsForLink().Error already tested in query/tag_test.go

	rankings, err := ScanTagRankings(tag_rankings_sql)
	if err != nil {
		t.Fatal(err)
	}

	// Verify result length
	if len(*rankings) != len(test_rankings) {
		t.Fatalf(
			"got %d tag rankings, want %d", 
			len(*rankings), 
			len(test_rankings),
		)
	}

	// Verify result order
	for i, ranking := range *rankings {
		if ranking.SubmittedBy != test_rankings[i].SubmittedBy {
			t.Fatalf(
				"expected ranking %d to be submitted by %s, got %s",
				i + 1,
				test_rankings[i].SubmittedBy,
				ranking.SubmittedBy,
			)
		} else if ranking.Categories != test_rankings[i].Cats {
			t.Fatalf(
				"expected ranking %d to have cats %s, got %s",
				i + 1,
				test_rankings[i].Cats,
				ranking.Categories,
			)
		}
	}
}

// Get top global cats
func TestScanAndSplitGlobalCats(t *testing.T) {
	global_cats_sql := query.NewAllGlobalCats()
	// NewAllGlobalCats already tested in query/tag_test.go

	split_global_cats, err := ScanAndSplitGlobalCats(global_cats_sql)
	if err != nil {
		t.Fatal(err)
	}

	var found_cats []string
	
	for _, cat := range *split_global_cats {

		// Verify split (no commas)
		if strings.Contains(cat, ",") {
			t.Fatalf(
				"global cat %s contains comma, was not properly split", 
				cat,
			)

		// Verify no duplicates
		} else if slices.Contains(found_cats, cat) {
			t.Fatalf(
				"global cat %s was found more than once", 
				cat,
			)
		} else {
			found_cats = append(found_cats, cat)
		}
	}
}

func TestGetCatCounts(t *testing.T) {
	counts, err := GetCatCounts(&test_multiple_cats)
	if err != nil {
		t.Fatal(err)
	}

	// Verify result length
	if len(*counts) == 0 {
		t.Fatalf("no counts returned")
	}

	// Verify correct properties set
	for i, c := range *counts {
		if c.Count == 0 {
			t.Fatalf("cat %s returned count 0", c.Category)
		} else if c.Category == "" {
			t.Fatalf("%s returned empty category", test_multiple_cats[i])
		}
	}

	// SortAndLimitCatCounts already tested in handler/util/subcats_test.go
}

func TestGetCatCountsDuringPeriod(t *testing.T) {
	var test_periods = []struct{
		Period string
		Valid bool
	}{
		{"day", true},
		{"week", true},
		{"month", true},
		{"year", true},
		{"all", false},
		{"invalid_period", false},
	}

	var result_counts []struct{
		Count int32
		Cat string
		Period string
	}

	for _, tp := range test_periods {
		counts, err := GetCatCountsDuringPeriod(&test_multiple_cats, tp.Period)
		if tp.Valid && err != nil {
			t.Fatalf("unexpected error for period %s", tp.Period)
		} else if !tp.Valid && err == nil {
			t.Fatalf("expected error for period %s", tp.Period)
		}

		if tp.Valid {
			for _, c := range *counts {
				result := struct{
					Count int32
					Cat string
					Period string
				}{
					Count: c.Count,
					Cat: c.Category,
					Period: tp.Period,
				}
				result_counts = append(result_counts, result)
			}
		}
	}
	
	// Verify that wider periods do not have lower counts
	// (can be same or more, but never less)
	for _, cat := range test_multiple_cats {
		var prev_highest_cat_count int32

		for _, r := range result_counts {
			if r.Cat == cat && r.Count < prev_highest_cat_count {
				t.Fatalf(
					"cat %s had lower count (%d) in period %s than previous (%d)", 
					cat,
					r.Count, 
					r.Period,
					prev_highest_cat_count,
				)
			} else if r.Cat == cat && r.Count > prev_highest_cat_count {
				prev_highest_cat_count = r.Count
			}
		}
	}
}

func TestUserHasTaggedLink(t *testing.T) {
	var test_links = []struct{
		ID string
		TaggedByTestUser bool
	}{
		{"1", true},
		{"13", true},
		{"22", true},
		{"0", false},
		{"10", false},
		{"15", false},
	}

	for _, l := range test_links {
		return_true, err := UserHasTaggedLink(test_login_name, l.ID)
		if err != nil  {
			t.Fatalf("failed with error: %s", err)
		} else if l.TaggedByTestUser && !return_true {
			t.Fatalf("expected tag with ID %s to be tagged by user", l.ID)
		} else if !l.TaggedByTestUser && return_true {
			t.Fatalf("tag with ID %s NOT submitted by user, expected error", l.ID)
		}
	}
}

// AssignNewTagIDToRequest() is simple, no point in testing 

// Edit tag
func TestUserSubmittedTagWithID(t *testing.T) {
	var test_tags = []struct{
		ID string
		SubmittedByTestUser bool
	}{
		{"32", true},
		{"34", true},
		{"114", true},
		{"5", false},
		{"6", false},
		{"11", false},
	}

	for _, tag := range test_tags {
		return_true, err := UserSubmittedTagWithID(test_login_name, tag.ID)
		if err != nil  {
			t.Fatalf("failed with error: %s", err)
		} else if tag.SubmittedByTestUser && !return_true {
			t.Fatalf("expected tag with ID %s to be submitted by user", tag.ID)
		} else if !tag.SubmittedByTestUser && return_true {
			t.Fatalf("tag with ID %s NOT submitted by user, expected error", tag.ID)
		}
	}
}

// AlphabetizeCats() is simple usage of strings.Split / string.Join / slices.Sort
// no point in testing 

func TestGetLinkIDFromTagID(t *testing.T) {
	var test_tags = []struct{
		ID string
		LinkID string
	}{
		{"32", "1"},
		{"34", "13"},
		{"114", "22"},
		{"5", "0"},
		{"6", "8"},
		{"11", "10"},
	}

	for _, tag := range test_tags {
		return_link_id, err := GetLinkIDFromTagID(tag.ID)
		if err != nil  {
			t.Fatalf("failed with error: %s", err)
		} else if tag.LinkID != return_link_id {
			t.Fatalf(
				"expected tag with ID %s to have link ID %s", 
				tag.ID, 
				tag.LinkID,
			)
		}
	}
}

func TestCalculateAndSetGlobalCats(t *testing.T) {
	
	// TODO: refactor test after refactoring CalculateGlobalCatsForLink()

	var test_link_ids = []struct{
		ID string
		GlobalCats string
	}{
		{"0","flowers"},
		{"7","7,baby,lucky"},
		{"11","test"},
	}

	for _, l := range test_link_ids {
		err := CalculateAndSetGlobalCats(l.ID)
		if err != nil  {
			t.Fatalf("failed with error: %s", err)
		}

		// confirm global cats match expected
		var gc string
		err = TestClient.QueryRow(`
			SELECT global_cats
			FROM Links 
			WHERE id = ?`,
			l.ID,
		).Scan(&gc)

		if err != nil {
			t.Fatalf(
				"failed with error: %s for link with ID %s", 
				err,
				l.ID,
			)
		} else if gc != l.GlobalCats {
			t.Fatalf(
				"got global cats %s for link with ID %s, want %s", 
				gc, 
				l.ID,
				l.GlobalCats,
			)
		}
	}
}

// AlphabetizeOverlapScoreCats() is simple usage of slices.Sort()
// no point in testing

// TODO: retest once in-memory DB has suitable test data
func TestSetGlobalCats(t *testing.T) {
	var test_link_id = "11"

	err := SetGlobalCats(test_link_id, "foo,bar")
	if err != nil  {
		t.Fatalf("failed with error: %s", err)
	}

	// confirm global cats match expected
	var gc string
	err = TestClient.QueryRow(`
		SELECT global_cats
		FROM Links 
		WHERE id = ?`,
		test_link_id,
	).Scan(&gc)

	if err != nil {
		t.Fatalf("failed with error: %s", err)
	} else if gc != "foo,bar" {
		t.Fatalf("got global cats %s, want %s", gc, "foo,bar")
	}
}