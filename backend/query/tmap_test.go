package query

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/julianlk522/fitm/model"
)

var (
	test_login_name = "jlk"
	test_user_id    = "3"
	test_cats       = []string{"go", "coding"}

	test_req_user_id    = "13"
	test_req_login_name = "bradley"
)

// Profile
func TestNewTmapProfile(t *testing.T) {
	profile_sql := NewTmapProfile(test_login_name)

	var profile model.Profile
	if err := TestClient.QueryRow(profile_sql).Scan(
		&profile.LoginName,
		&profile.About,
		&profile.PFP,
		&profile.Created,
	); err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
}

// Submitted
func TestNewTmapSubmitted(t *testing.T) {

	// first retrieve all IDs of links submitted by user
	var submitted_ids []string

	submitted_ids_sql := fmt.Sprintf(
		`SELECT id FROM Links WHERE submitted_by = '%s'`,
		test_req_login_name,
	)

	rows, err := TestClient.Query(submitted_ids_sql)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		submitted_ids = append(submitted_ids, id)
	}

	// execute query and confirm all submitted links are present
	submitted_sql := NewTmapSubmitted(test_req_login_name)
	if submitted_sql.Error != nil {
		t.Fatal(submitted_sql.Error)
	}

	rows, err = TestClient.Query(submitted_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var l model.TmapLink
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.LikeCount,
			&l.TagCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		} else if l.SubmittedBy != test_req_login_name {
			t.Fatalf("SubmittedBy != test login_name (%s)", test_req_login_name)
		} else if l.TagCount == 0 {
			t.Fatalf("TagCount == 0: %+v", l)
		}

		// remove from submitted_ids if returned by query
		for i := 0; i < len(submitted_ids); i++ {
			if l.ID == submitted_ids[i] {
				submitted_ids = append(submitted_ids[0:i], submitted_ids[i+1:]...)
				break
			}
		}
	}

	// if any IDs are left in submitted_ids then they were incorrectly
	// omitted by query
	if len(submitted_ids) > 0 {
		t.Fatalf("not all submitted links returned, see missing IDs: %+v", submitted_ids)
	}
}

func TestNewTmapSubmittedFromCats(t *testing.T) {
	submitted_sql := NewTmapSubmitted(test_login_name).FromCats(test_cats)
	if submitted_sql.Error != nil {
		t.Fatal(submitted_sql.Error)
	}

	rows, err := TestClient.Query(submitted_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var l model.TmapLink
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.LikeCount,
			&l.TagCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		} else if !strings.Contains(l.Cats, test_cats[0]) || !strings.Contains(l.Cats, test_cats[1]) {
			t.Fatalf("got %s, should contain %s", l.Cats, test_cats)
		} else if l.TagCount == 0 {
			t.Fatal("TagCount == 0")
		}
	}

	// test "." properly escaped
	submitted_sql = NewTmapSubmitted(test_login_name).FromCats([]string{"YouTube", "c. viper"})
	if submitted_sql.Error != nil {
		t.Fatal(submitted_sql.Error)
	} else if strings.Contains(submitted_sql.Text, ".") && !strings.Contains(submitted_sql.Text, `'.'`) {
		t.Fatal("failed to ecape period in cat 'c. viper'")
	}
}

func TestNewTmapSubmittedAsSignedInUser(t *testing.T) {
	submitted_sql := NewTmapSubmitted(test_login_name).AsSignedInUser(test_req_user_id, test_req_login_name)
	if submitted_sql.Error != nil {
		t.Fatal(submitted_sql.Error)
	}

	rows, err := TestClient.Query(submitted_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	// just test first row since column counts will be the same
	if rows.Next() {
		var l model.TmapLinkSignedIn
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.TagCount,
			&l.LikeCount,
			&l.ImgURL,
			&l.IsLiked,
			&l.IsCopied,
		); err != nil {
			t.Fatal(err)
		}
	}
}

// Copied
func TestNewTmapCopied(t *testing.T) {
	copied_sql := NewTmapCopied(test_login_name)
	if copied_sql.Error != nil {
		t.Fatal(copied_sql.Error)
	}

	rows, err := TestClient.Query(copied_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var l model.TmapLink
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.LikeCount,
			&l.TagCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		} else if l.TagCount == 0 {
			t.Fatal("TagCount == 0")
		}

		// check that tmap owner has copied
		var link_id string
		err := TestClient.QueryRow(
			fmt.Sprintf(`SELECT id
				FROM 'Link Copies'
				WHERE link_id = %s
				AND user_id = %s`,
				l.ID,
				test_user_id),
		).Scan(&link_id)

		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNewTmapCopiedFromCats(t *testing.T) {
	copied_sql := NewTmapCopied(test_login_name).FromCats(test_cats)
	if copied_sql.Error != nil {
		t.Fatal(copied_sql.Error)
	}

	rows, err := TestClient.Query(copied_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var l model.TmapLink
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.LikeCount,
			&l.TagCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		} else if !strings.Contains(l.Cats, test_cats[0]) || !strings.Contains(l.Cats, test_cats[1]) {
			t.Fatalf("got %s, should contain %s", l.Cats, test_cats)
		} else if l.TagCount == 0 {
			t.Fatal("TagCount == 0")
		}

		// check that tmap owner has copied
		var link_id string
		err := TestClient.QueryRow(
			fmt.Sprintf(`SELECT id
				FROM 'Link Copies'
				WHERE link_id = %s
				AND user_id = %s`,
				l.ID,
				test_user_id),
		).Scan(&link_id)

		if err != nil {
			t.Fatal(err)
		}
	}

	// test "." properly escaped
	copied_sql = NewTmapCopied(test_login_name).FromCats([]string{"YouTube", "c. viper"})
	if copied_sql.Error != nil {
		t.Fatal(copied_sql.Error)
	} else if strings.Contains(copied_sql.Text, ".") && !strings.Contains(copied_sql.Text, `'.'`) {
		t.Fatal("failed to ecape period in cat 'c. viper'")
	}
}

func TestNewTmapCopiedAsSignedInUser(t *testing.T) {
	copied_sql := NewTmapCopied(test_login_name).AsSignedInUser(test_req_user_id, test_req_login_name)
	if copied_sql.Error != nil {
		t.Fatal(copied_sql.Error)
	}

	rows, err := TestClient.Query(copied_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	// test first row only since column counts will be the same
	if rows.Next() {
		var l model.TmapLinkSignedIn
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.TagCount,
			&l.LikeCount,
			&l.ImgURL,
			&l.IsLiked,
			&l.IsCopied,
		); err != nil {
			t.Fatal(err)
		}
	}
}

// Tagged
func TestNewTmapTagged(t *testing.T) {
	tagged_sql := NewTmapTagged(test_login_name)
	if tagged_sql.Error != nil {
		t.Fatal(tagged_sql.Error)
	}

	rows, err := TestClient.Query(tagged_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var l model.TmapLink
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.LikeCount,
			&l.TagCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		} else if l.TagCount == 0 {
			t.Fatal("TagCount == 0")
		}

		// check that tmap owner has tagged
		var link_id string
		err := TestClient.QueryRow(
			fmt.Sprintf(`SELECT id
				FROM Tags
				WHERE link_id = %s
				AND submitted_by = '%s'`,
				l.ID,
				test_login_name),
		).Scan(&link_id)

		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNewTmapTaggedFromCats(t *testing.T) {
	tagged_sql := NewTmapTagged(test_login_name).FromCats(test_cats)
	if tagged_sql.Error != nil {
		t.Fatal(tagged_sql.Error)
	}

	rows, err := TestClient.Query(tagged_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var l model.TmapLink
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.LikeCount,
			&l.TagCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		} else if !strings.Contains(l.Cats, test_cats[0]) || !strings.Contains(l.Cats, test_cats[1]) {
			t.Fatalf("got %s, should contain %s", l.Cats, test_cats)
		} else if l.TagCount == 0 {
			t.Fatal("TagCount == 0")
		}

		// check that tmap owner has tagged
		var link_id string
		err := TestClient.QueryRow(
			fmt.Sprintf(`SELECT id
				FROM Tags
				WHERE link_id = %s
				AND submitted_by = %s`,
				l.ID,
				test_login_name),
		).Scan(&link_id)

		if err != nil {
			t.Fatal(err)
		}
	}

	// test "." properly escaped
	tagged_sql = NewTmapTagged(test_login_name).FromCats([]string{"YouTube", "c. viper"})
	if tagged_sql.Error != nil {
		t.Fatal(tagged_sql.Error)
	} else if strings.Contains(tagged_sql.Text, ".") && !strings.Contains(tagged_sql.Text, `'.'`) {
		t.Fatal("failed to ecape period in cat 'c. viper'")
	}
}

func TestNewTmapTaggedAsSignedInUser(t *testing.T) {
	tagged_sql := NewTmapTagged(test_login_name).AsSignedInUser(test_req_user_id, test_req_login_name)
	if tagged_sql.Error != nil {
		t.Fatal(tagged_sql.Error)
	}

	rows, err := TestClient.Query(tagged_sql.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	// test first row only since column counts will be the same
	if rows.Next() {
		var l model.TmapLinkSignedIn
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.TagCount,
			&l.LikeCount,
			&l.ImgURL,
			&l.IsLiked,
			&l.IsCopied,
		); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFromUserOrGlobalCats(t *testing.T) {

	// submitted
	tmap_submitted := NewTmapSubmitted(test_login_name)
	_, err := TestClient.Query(tmap_submitted.Text)
	if err != nil {
		t.Fatal(err)
	}

	tmap_submitted.Text = FromUserOrGlobalCats(tmap_submitted.Text, test_cats)
	rows, err := TestClient.Query(tmap_submitted.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	// make sure links only have cats from test_cats
	for rows.Next() {
		var l model.TmapLink
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.LikeCount,
			&l.TagCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		} else if !strings.Contains(l.Cats, test_cats[0]) || !strings.Contains(l.Cats, test_cats[1]) {
			t.Fatalf("got %s, should contain %s", l.Cats, test_cats)
		}
	}

	// copied
	tmap_copied := NewTmapCopied(test_login_name)
	_, err = TestClient.Query(tmap_copied.Text)
	if err != nil {
		t.Fatal(err)
	}

	tmap_copied.Text = FromUserOrGlobalCats(tmap_copied.Text, test_cats)
	rows, err = TestClient.Query(tmap_copied.Text)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var l model.TmapLink
		if err := rows.Scan(
			&l.ID,
			&l.URL,
			&l.SubmittedBy,
			&l.SubmitDate,
			&l.Cats,
			&l.CatsFromUser,
			&l.Summary,
			&l.SummaryCount,
			&l.LikeCount,
			&l.TagCount,
			&l.ImgURL,
		); err != nil {
			t.Fatal(err)
		} else if !strings.Contains(l.Cats, test_cats[0]) || !strings.Contains(l.Cats, test_cats[1]) {
			t.Fatalf("got %s, should contain %s", l.Cats, test_cats)
		}
	}
}

func TestGetCatsWithEscapedPeriods(t *testing.T) {
	var test_cats = struct{
		Cats []string
		ExpectedResults []string
	}{
		Cats: []string{"YouTube", "c. viper"},
		ExpectedResults: []string{"YouTube", `c'.' viper`},
	}
	
	got := GetCatsWithEscapedPeriods(test_cats.Cats)
	for i, res := range got {
		if res != test_cats.ExpectedResults[i] {
			t.Fatalf("got %s, want %s", got, test_cats.ExpectedResults)
		}
	}
}
