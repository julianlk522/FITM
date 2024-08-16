package model

import (
	"strings"
)

func HasTooLongCats(cats string) bool {
	split_cats := strings.Split(cats, ",")

	for _, cat := range(split_cats) {
		if len(cat) > CAT_CHAR_LIMIT {
			return true
		}
	}

	return false
}

func IsTooManyCats(cats string) bool {
	num_cats := strings.Count(cats, ",") + 1
	// +1 since "a" (no commas) would be one cat
	// and "a,b" (one comma) would be two
	return num_cats > NUM_CATS_LIMIT
}