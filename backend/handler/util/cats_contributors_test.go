package handler

import (
	"fmt"
	"oitm/query"
	"strings"
	"testing"
)

func TestScanCatsContributors(t *testing.T) {

	// single cat
	contributors_sql := query.NewCatsContributors(test_single_cat)
	if contributors_sql.Error != nil {
		t.Fatal(contributors_sql.Error)
	}

	test_cats_str := test_single_cat[0]
	contributors := ScanCatsContributors(contributors_sql, test_cats_str)

	if len(*contributors) == 0 {
		t.Fatal("no contributors")
	}

	// verify that each contributor submitted the correct number of links
	var ls int
	for _, contributor := range *contributors {
		err := TestClient.QueryRow(
			fmt.Sprintf(
				`SELECT count(*)
				FROM Links
				WHERE submitted_by = '%s'
				AND ',' || global_cats || ',' LIKE '%%,%s,%%';`,
				contributor.LoginName,
				test_cats_str),
		).Scan(&ls)
		if err != nil {
			t.Fatal(err)
		} else if ls != contributor.LinksSubmitted {
			t.Fatalf(
				"expected %d links submitted, got %d (contributor: %s)", contributor.LinksSubmitted,
				ls,
				contributor.LoginName,
			)
		}
	}

	// multiple cats
	contributors_sql = query.NewCatsContributors(test_multiple_cats)
	if contributors_sql.Error != nil {
		t.Fatal(contributors_sql.Error)
	}

	test_cats_str = strings.Join(test_multiple_cats, ",")
	contributors = ScanCatsContributors(contributors_sql, test_cats_str)

	if len(*contributors) == 0 {
		t.Fatal("no contributors")
	}

	// verify that each contributor submitted the correct number of links
	for _, contributor := range *contributors {
		err := TestClient.QueryRow(
			fmt.Sprintf(
				`SELECT count(*)
				FROM Links
				WHERE submitted_by = '%s'
				AND ',' || global_cats || ',' LIKE '%%,%s,%%'
				AND ',' || global_cats || ',' LIKE '%%,%s,%%';`,
				contributor.LoginName,
				test_multiple_cats[0],
				test_multiple_cats[1]),
		).Scan(&ls)
		if err != nil {
			t.Fatal(err)
		} else if ls != contributor.LinksSubmitted {
			t.Fatalf(
				"expected %d links submitted, got %d (contributor: %s)", contributor.LinksSubmitted,
				ls,
				contributor.LoginName,
			)
		}
	}
}
