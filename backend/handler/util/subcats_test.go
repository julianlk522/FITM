package handler

import (
	"oitm/model"
	"oitm/query"
	"testing"
)

func TestScanSubcats(t *testing.T) {
	subcats_sql := query.NewSubcats(test_single_cat)
	// NewSubcats().Error tested in query/link_test.go
	// TODO: move to query/subcats_test.go and update this comment

	// single cat
	subcats := ScanSubcats(subcats_sql, test_single_cat)
	if len(subcats) == 0 {
		t.Fatal("no subcats (single test cat)")
	}

	// multiple cats
	subcats = ScanSubcats(subcats_sql, test_multiple_cats)
	if len(subcats) == 0 {
		t.Fatal("no subcats (multiple test cats)")
	}
}

func TestGetCountsOfSubcatsFromCats(t *testing.T) {
	var cats, subcats = test_multiple_cats, test_single_cat

	counts, err := GetCountsOfSubcatsFromCats(subcats, cats)
	if err != nil {
		t.Fatal(err)
	}

	if len(*counts) == 0 {
		t.Fatalf("no counts returned for subcat %s of cat %s", subcats, cats)
	}

	for _, c := range *counts {
		if c.Count == 0 {
			t.Fatalf("subcat %s of cat %s returned count 0", subcats, cats)
		}
	}

	// not worth testing that cat counts are correct since it would require
	// basically rewriting the NewCatCount() query verbatim
}

func TestSortAndLimitCatCounts(t *testing.T) {

	// 16 items: should be limited to CATEGORY_PAGE_LIMIT (15)
	test_cat_counts := []model.CatCount{
		{Category: "dog", Count: 1},
		{Category: "elephant", Count: 3},
		{Category: "bird", Count: 0},
		{Category: "cat", Count: 1},
		{Category: "fish", Count: 0},
		{Category: "cow", Count: 0},
		{Category: "horse", Count: 10},
		{Category: "goat", Count: 0},
		{Category: "sheep", Count: 3},
		{Category: "chicken", Count: 0},
		{Category: "penguin", Count: 10},
		{Category: "goose", Count: 0},
		{Category: "turkey", Count: 0},
		{Category: "pig", Count: 5},
		{Category: "duck", Count: 0},
		{Category: "octopus", Count: 1000},
	}

	SortAndLimitCatCounts(&test_cat_counts)

	if len(test_cat_counts) > CATEGORY_PAGE_LIMIT {
		t.Fatalf(
			"expected %d subcats, got %d", 
			CATEGORY_PAGE_LIMIT, 
			len(test_cat_counts),
		)
	}

	// Confirm correct order

	if test_cat_counts[0].Category != "octopus" {
		t.Fatalf(
			"expected first subcat to be 'octopus' with count 1000, got %s with count %d", 
			test_cat_counts[0].Category, 
			test_cat_counts[0].Count)

	// horse before penguin due to alphabetical order when same count
	} else if test_cat_counts[1].Category != "horse" {
		t.Fatalf(
			"expected second subcat to be 'horse' with count 10, got %s with count %d", 
			test_cat_counts[1].Category, 
			test_cat_counts[1].Count)
	} else if test_cat_counts[2].Category != "penguin" {
		t.Fatalf(
			"expected third subcat to be 'penguin' with count 10, got %s with count %d", 
			test_cat_counts[2].Category, 
			test_cat_counts[2].Count)
	}
}
