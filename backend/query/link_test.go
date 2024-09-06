package query

import (
	"database/sql"
	"testing"

	"fmt"
	"strings"
)

// Links
func Test_PaginateLimitClause(t *testing.T) {
	var test_cases = []struct {
		Page int
		Want string
	}{
		{1, fmt.Sprintf(" LIMIT %d;", LINKS_PAGE_LIMIT+1)},
		{2, fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT+1, LINKS_PAGE_LIMIT)},
		{3, fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT+1, 2*LINKS_PAGE_LIMIT)},
		{4, fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT+1, 3*LINKS_PAGE_LIMIT)},
	}

	for _, tc := range test_cases {
		got := _PaginateLimitClause(tc.Page)
		if got != tc.Want {
			t.Fatalf("got %s, want %s", got, tc.Want)
		}
	}
}

func TestNewTopLinks(t *testing.T) {
	links_sql := NewTopLinks()

	if links_sql.Error != nil {
		t.Fatal(links_sql.Error)
	}

	rows, err := TestClient.Query(links_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	cols, err := rows.ColumnTypes()
	if err != nil {
		t.Fatal(err)
	}

	if len(cols) == 0 {
		t.Fatal("no columns")
	} else if len(cols) != 10 {
		t.Fatal("too few columns")
	}

	var test_cols = []struct {
		Want string
	}{
		{"id"},
		{"url"},
		{"sb"},
		{"sd"},
		{"cats"},
		{"summary"},
		{"summary_count"},
		{"tag_count"},
		{"like_count"},
		{"img_url"},
	}

	for i, col := range cols {
		if col.Name() != test_cols[i].Want {
			t.Fatalf("column %d: got %s, want %s", i, col.Name(), test_cols[i].Want)
		}
	}
}

func TestFromCats(t *testing.T) {
	var test_cats = []struct {
		Cats  []string
		Valid bool
	}{
		{[]string{}, false},
		{[]string{""}, false},
		{[]string{"umvc3"}, true},
		{[]string{"umvc3", "flowers"}, true},
	}

	for _, tc := range test_cats {

		// cats only
		links_sql := NewTopLinks().FromCats(tc.Cats)
		if tc.Valid && links_sql.Error != nil {
			t.Fatal(links_sql.Error)
		} else if !tc.Valid && links_sql.Error == nil {
			t.Fatalf("expected error for cats %s", tc.Cats)
		}

		rows, err := TestClient.Query(links_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		defer rows.Close()

		// with period
		links_sql = links_sql.DuringPeriod("month")
		if tc.Valid && links_sql.Error != nil {
			t.Fatal(links_sql.Error)
		} else if !tc.Valid && links_sql.Error == nil {
			t.Fatalf("expected error for cats %s", tc.Cats)
		}

		rows, err = TestClient.Query(links_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		defer rows.Close()
	}
}

func TestLinksDuringPeriod(t *testing.T) {
	var test_periods = []struct {
		Period string
		Valid  bool
	}{
		{"day", true},
		{"week", true},
		{"month", true},
		{"year", true},
		{"all", false},
		{"gobblety gook", false},
	}

	for _, tp := range test_periods {

		// period only
		links_sql := NewTopLinks().DuringPeriod(tp.Period)
		if tp.Valid && links_sql.Error != nil {
			t.Fatal(links_sql.Error)
		} else if !tp.Valid && links_sql.Error == nil {
			t.Fatalf("expected error for period %s", tp.Period)
		}

		rows, err := TestClient.Query(links_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		defer rows.Close()

		// with cats
		// NOT a repeat of TestFromCats; testing order of method calls
		links_sql = links_sql.FromCats([]string{"umvc3"})
		if tp.Valid && links_sql.Error != nil {
			t.Fatal(links_sql.Error)
		} else if !tp.Valid && links_sql.Error == nil {
			t.Fatalf("expected error for period %s", tp.Period)
		}

		rows, err = TestClient.Query(links_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		defer rows.Close()
	}
}

func TestAsSignedInUser(t *testing.T) {
	links_sql := NewTopLinks().AsSignedInUser(test_user_id)
	if links_sql.Error != nil {
		t.Fatal(links_sql.Error)
	}

	rows, err := TestClient.Query(links_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	cols, err := rows.ColumnTypes()
	if err != nil {
		t.Fatal(err)
	}

	if len(cols) == 0 {
		t.Fatal("no columns")
	} else if len(cols) != 12 {
		t.Fatal("incorrect col count")
	}

	var test_cols = []struct {
		Want string
	}{
		{"id"},
		{"url"},
		{"sb"},
		{"sd"},
		{"cats"},
		{"summary"},
		{"summary_count"},
		{"tag_count"},
		{"like_count"},
		{"img_url"},
		{"is_liked"},
		{"is_copied"},
	}

	for i, col := range cols {
		if col.Name() != test_cols[i].Want {
			t.Fatalf("column %d: got %s, want %s", i, col.Name(), test_cols[i].Want)
		}
	}
}

func TestPage(t *testing.T) {
	var links_sql = NewTopLinks()

	want1 := strings.Replace(links_sql.Text, UNPAGINATED_LIMIT_CLAUSE, fmt.Sprintf(" LIMIT %d;", LINKS_PAGE_LIMIT+1), 1)
	want2 := strings.Replace(links_sql.Text, UNPAGINATED_LIMIT_CLAUSE, fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT+1, LINKS_PAGE_LIMIT), 1)
	want3 := strings.Replace(links_sql.Text, UNPAGINATED_LIMIT_CLAUSE, fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT+1, 2*LINKS_PAGE_LIMIT), 1)

	var test_cases = []struct {
		Page int
		Want string
	}{
		{0, links_sql.Text},
		{1, want1},
		{2, want2},
		{3, want3},
	}

	for _, tc := range test_cases {
		got := links_sql.Page(tc.Page).Text
		if got != tc.Want {
			t.Fatalf("input page %d, got %s, want %s", tc.Page, got, tc.Want)
		}

		links_sql = NewTopLinks()
	}
}

// Contributors
func TestNewContributors(t *testing.T) {
	contributors_sql := NewContributors()
	if contributors_sql.Error != nil {
		t.Fatal(contributors_sql.Error)
	}

	rows, err := TestClient.Query(contributors_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	cols, err := rows.ColumnTypes()
	if err != nil {
		t.Fatal(err)
	}

	if len(cols) == 0 {
		t.Fatal("no columns")
	} else if len(cols) != 2 {
		t.Fatalf("wrong columns (got %d, want 2)", len(cols))
	}

	var test_cols = []struct {
		Want string
	}{
		{"count"},
		{"submitted_by"},
	}

	for i, col := range cols {
		if col.Name() != test_cols[i].Want {
			t.Fatalf("column %d: got %s, want %s", i, col.Name(), test_cols[i].Want)
		}
	}
}

func TestContributorsFromCats(t *testing.T) {
	contributors_sql := NewContributors().FromCats([]string{"umvc3"})
	contributors_sql.Text = strings.Replace(
		contributors_sql.Text,
		`SELECT
count(l.id) as count, l.submitted_by
FROM Links l`,
		`SELECT
count(l.id) as count, l.global_cats
FROM Links l`,
	1)

	rows, err := TestClient.Query(contributors_sql.Text)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cat, count string
		if err := rows.Scan(&count, &cat); err != nil {
			t.Fatal(err)
		} else if !strings.Contains(strings.ToLower(cat), "umvc3") {
			t.Fatalf("got %s, should contain %s", cat, "umvc3")
		}
	}
}

func TestContributorsDuringPeriod(t *testing.T) {
	var test_periods = [7]struct {
		Period string
		Valid  bool
	}{
		{"day", true},
		{"week", true},
		{"month", true},
		{"year", true},
		{"all", false},
		{"shouldfail", false},
	}

	// Period only
	for _, period := range test_periods {
		contributors_sql := NewContributors().DuringPeriod(period.Period)
		if period.Valid && contributors_sql.Error != nil {
			t.Fatal(contributors_sql.Error)
		} else if !period.Valid && contributors_sql.Error == nil {
			t.Fatalf("expected error for period %s", period.Period)
		}

		rows, err := TestClient.Query(contributors_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		defer rows.Close()
	}

	// Period and Cats
	for _, period := range test_periods {
		contributors_sql := NewContributors().DuringPeriod(period.Period).FromCats([]string{"umvc3"})
		if period.Valid && contributors_sql.Error != nil {
			t.Fatal(contributors_sql.Error)
		} else if !period.Valid && contributors_sql.Error == nil {
			t.Fatalf("expected error for period %s", period.Period)
		}

		rows, err := TestClient.Query(contributors_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		defer rows.Close()
	}
}
