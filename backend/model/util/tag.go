package model

import (
	e "oitm/error"
	"strings"
)

func IsTooManyCats(cats string) bool {
	num_cats := strings.Count(cats, ",") + 1
	// +1 since "a" (no commas) would be one cat
	// and "a,b" (one comma) would be two
	return num_cats > e.NEW_TAG_CAT_LIMIT
}