package query

import (
	"database/sql"
	"testing"

	"fmt"
	"strings"

	"oitm/model"
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
		{"link_id"},
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

func TestFromIDs(t *testing.T) {
	links_sql := NewTopLinks().FromLinkIDs([]string{"1", "2", "3"})

	if links_sql.Error != nil {
		t.Fatal(links_sql.Error)
	}

	rows, err := TestClient.Query(links_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var l model.Link
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.Summary,
			&l.SummaryCount,
			&l.TagCount,
			&l.LikeCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		}

		if l.ID != "1" && l.ID != "2" && l.ID != "3" {
			t.Fatalf("got %s, want 1, 2, or 3", l.ID)
		}
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

	for _, period := range test_periods {
		links_sql := NewTopLinks().DuringPeriod(period.Period)
		if period.Valid && links_sql.Error != nil {
			t.Fatal(links_sql.Error)
		} else if !period.Valid && links_sql.Error == nil {
			t.Fatalf("expected error for period %s", period.Period)
		}

		rows, err := TestClient.Query(links_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		defer rows.Close()
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

func Test_LinksWhere(t *testing.T) {
	l1 := NewTopLinks()
	l2 := NewTopLinks()

	clause1 := "links_id IN ('1', '2', '3')"
	clause2 := "julianday('now') - julianday(submit_date) < 31"

	l1._Where(clause1)
	l2._Where(clause2)

	rows1, err := TestClient.Query(l1.Text)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	defer rows1.Close()

	rows2, err := TestClient.Query(l2.Text)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	defer rows2.Close()
}

// Link IDs
func TestNewLinkIDs(t *testing.T) {
	ids_sql := NewLinkIDs("umvc3")

	if ids_sql.Error != nil {
		t.Fatal(ids_sql.Error)
	}

	rows, err := TestClient.Query(ids_sql.Text)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	defer rows.Close()

	cols, err := rows.ColumnTypes()
	if err != nil {
		t.Fatal(err)
	}

	if len(cols) == 0 {
		t.Fatal("no columns")
	} else if len(cols) != 1 {
		t.Fatalf("wrong column count (got %d, want 1)", len(cols))
	}

	if cols[0].Name() != "id" {
		t.Fatalf("got %s for column name, want %s", cols[0].Name(), "id")
	}
}

func Test_LinkIDsFromCats(t *testing.T) {
	ids_sql := NewLinkIDs("umvc3")
	ids_sql.Text = strings.Replace(ids_sql.Text, "id", "global_cats", 1)

	rows, err := TestClient.Query(ids_sql.Text)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(c, "umvc3") {
			t.Fatalf("got %s, should contain %s", c, "umvc3")
		}
	}
}

// Subcats
func TestNewSubcats(t *testing.T) {
	subcats_sql := NewSubcats([]string{"umvc3"})
	if subcats_sql.Error != nil {
		t.Fatal(subcats_sql.Error)
	}

	rows, err := TestClient.Query(subcats_sql.Text)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(c, "umvc3") {
			t.Fatalf("got %s, should contain %s", c, "umvc3")
		}
	}
}

func TestSubcatsDuringPeriod(t *testing.T) {
	var test_periods = [7]struct {
		Period string
		Valid  bool
	}{
		{"day", true},
		{"week", true},
		{"month", true},
		{"year", true},
		{"all", false},
		{"yabbadabbadoo", false},
	}

	for _, period := range test_periods {
		links_sql := NewSubcats([]string{"umvc3"}).DuringPeriod(period.Period)
		if period.Valid && links_sql.Error != nil {
			t.Fatal(links_sql.Error)
		} else if !period.Valid && links_sql.Error == nil {
			t.Fatalf("expected error for period %s", period.Period)
		}

		rows, err := TestClient.Query(links_sql.Text)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		defer rows.Close()
	}
}

func Test_SubcatsWhere(t *testing.T) {
	subcats_sql := NewSubcats([]string{"umvc3"})

	clause := "julianday('now') - julianday(submit_date) < 31"

	subcats_sql._Where(clause)

	rows, err := TestClient.Query(subcats_sql.Text)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	defer rows.Close()
}

// Cats counts
func TestNewCatCount(t *testing.T) {
	var count int

	cc_sql := NewCatsCount([]string{"umvc3"})
	if cc_sql.Error != nil {
		t.Fatal(cc_sql.Error)
	}

	err := TestClient.QueryRow(cc_sql.Text).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	// query.NewCatCount() is extremely basic. not worth checking if
	// correct counts are returned since that would require pretty much
	// rewriting the query verbatim
}

func Test_CatCountsFromCats(t *testing.T) {
	var test_cats = []string{"umvc3", "flowers"}
	var single_cat_count, multiple_cats_count int
	var cats string

	// single cat
	cc_sql := NewCatsCount(test_cats[:1])
	if cc_sql.Error != nil {
		t.Fatal(cc_sql.Error)
	}
	cc_sql.Text = strings.Replace(
		cc_sql.Text,
		"count(*) as link_count",
		"count(*) as link_count, global_cats",
		1)

	err := TestClient.QueryRow(cc_sql.Text).Scan(&single_cat_count, &cats)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	} else if !strings.Contains(cats, "umvc3") {
		t.Fatalf("got %s, should contain %s", cats, "umvc3")
	}

	// multiple cats
	cc_sql = NewCatsCount(test_cats)
	cc_sql.Text = strings.Replace(
		cc_sql.Text,
		"count(*) as link_count",
		"count(*) as link_count, global_cats",
		1)

	err = TestClient.QueryRow(cc_sql.Text).Scan(&multiple_cats_count, &cats)
	if err != nil {
		t.Fatal(err)
	} else if !(strings.Contains(cats, "umvc3") && strings.Contains(cats, "flowers")) {
		t.Fatalf(
			"got %s, should contain %s and %s",
			cats,
			"umvc3",
			"flowers",
		)
	}

	if multiple_cats_count >= single_cat_count {
		t.Fatalf(
			"got same counts (%s only: %d, %s: %d), multiple cat counts should be fewer", test_cats[0],
			single_cat_count,
			strings.Join(test_cats, ","),
			multiple_cats_count,
		)
	}
}

// Cats contributors
func TestNewCatsContributors(t *testing.T) {
	contributors_sql := NewCatsContributors([]string{"umvc3"})
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
		{"count(*)"},
		{"submitted_by"},
	}

	for i, col := range cols {
		if col.Name() != test_cols[i].Want {
			t.Fatalf("column %d: got %s, want %s", i, col.Name(), test_cols[i].Want)
		}
	}
}

func Test_ContributorsFromCats(t *testing.T) {
	contributors_sql := NewCatsContributors([]string{"umvc3"})
	contributors_sql.Text = strings.Replace(
		contributors_sql.Text,
		"count(*), submitted_by",
		"global_cats",
		1)

	rows, err := TestClient.Query(contributors_sql.Text)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(cat, "umvc3") {
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

	for _, period := range test_periods {
		contributors_sql := NewCatsContributors([]string{"umvc3"}).DuringPeriod(period.Period)
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
