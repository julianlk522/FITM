package query

import (
	"database/sql"
	"strconv"
	"strings"
	"testing"

	"oitm/model"
)

// Tags Page Link
func TestNewTagPageLink(t *testing.T) {
	test_link_id, test_user_id := "1", "1"

	tag_sql := NewTagPageLink(test_link_id, test_user_id)
	if tag_sql.Error != nil {
		t.Fatal(tag_sql.Error)
	}

	var l model.LinkSignedIn
	if err := TestClient.QueryRow(tag_sql.Text).Scan(
		&l.ID,
		&l.URL,
		&l.SubmittedBy,
		&l.SubmitDate,
		&l.Categories,
		&l.Summary,
		&l.SummaryCount,
		&l.LikeCount,
		&l.ImgURL,
		&l.IsLiked,
		&l.IsCopied,
	); err != nil {
		t.Fatal(err)
	}

	link_id_str := strconv.FormatInt(l.ID, 10)
	if link_id_str != test_link_id {
		t.Fatalf("got %d, want %s", l.ID, test_link_id)
	}
}


// Top Overlap Scores (Internal)
func TestNewTopOverlapScores(t *testing.T) {
	test_link_id := "1"

	tags_sql := NewTopOverlapScores(test_link_id)
	if tags_sql.Error != nil {
		t.Fatal(tags_sql.Error)
	}

	rows, err := TestClient.Query(tags_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	// test first row only, rest are same
	if rows.Next() {
		var tr model.TagRanking
		if err := rows.Scan(
			&tr.LifeSpanOverlap,
			&tr.Categories,
		); err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatalf("no overlap scores for test link %s", test_link_id)
	}

	// reset and modify fields to check for correct link_id (test _FromLink())
	tags_sql = NewTopOverlapScores(test_link_id)

	tags_sql.Text = strings.Replace(tags_sql.Text, 
		`SELECT 
	(julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) AS lifespan_overlap, 
	categories`, 
		`SELECT link_id`, 
	1)
	tags_sql.Text = strings.Replace(tags_sql.Text,
		"ORDER BY lifespan_overlap DESC",
		"",
	1)

	rows, err = TestClient.Query(tags_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		var link_id string
		if err := rows.Scan(&link_id); err != nil {
			t.Fatal(err)
		}

		if link_id != test_link_id {
			t.Fatalf("got %s, want %s", link_id, test_link_id)
		}
	} else {
		t.Fatalf("failed link_id check with modified query: no overlap scores for test link %s", test_link_id)
	}
}



// Tag Rankings (Public Overlap Scores)
func TestNewTagRankingsForLink(t *testing.T) {
	test_link_id := "1"

	tags_sql := NewTagRankingsForLink(test_link_id)
	if tags_sql.Error != nil {
		t.Fatal(tags_sql.Error)
	}

	rows, err := TestClient.Query(tags_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	// only necessary to test first row cols since they will all be the same
	if rows.Next() {
		var tr model.TagRankingPublic

		if err := rows.Scan(
			&tr.LifeSpanOverlap,
			&tr.Categories,
			&tr.SubmittedBy,
			&tr.LastUpdated,
		); err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatalf("no tag rankings for test link %s", test_link_id)
	}

	// Verify link_id (test _FromLink())
	// reset and modify fields 
	tags_sql = NewTagRankingsForLink(test_link_id)
	tags_sql.Text = strings.Replace(tags_sql.Text, 
		TAG_RANKINGS_BASE,
		`SELECT link_id 
		FROM Tags 
		INNER JOIN Links 
		ON Links.id = Tags.link_id`, 
	1)
	tags_sql.Text = strings.Replace(tags_sql.Text,
		"ORDER BY lifespan_overlap DESC",
		"",
	1)

	rows, err = TestClient.Query(tags_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		var row_link_id string

		if err := rows.Scan(&row_link_id); err != nil {
			t.Fatal(err)
		}
		if row_link_id != test_link_id {
			t.Fatalf("got %s, want %s", row_link_id, test_link_id)
		}
	} else {
		t.Fatalf("failed link_id check with modified query: no tag rankings for test link %s", test_link_id)
	}
}



// All Global Cats
func TestNewTopGlobalCatCounts(t *testing.T) {
	counts_sql := NewTopGlobalCatCounts()
	// no opportunity for counts_sql.Error to have been set so no need to check

	_, err := TestClient.Query(counts_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewTopGlobalCatCountsDuringPeriod(t *testing.T) {
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

	for _, tp := range test_periods {
		tags_sql := NewTopGlobalCatCounts().DuringPeriod(tp.Period)
		if tp.Valid && tags_sql.Error != nil {
			t.Fatalf("unexpected error for period %s", tp.Period)
		} else if !tp.Valid && tags_sql.Error == nil {
			t.Fatalf("expected error for period %s", tp.Period)
		}

		_, err := TestClient.Query(tags_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		} 
	}
}