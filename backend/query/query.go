package query

import (
	"fmt"
	"strings"

	e "github.com/julianlk522/fitm/error"
)

type Query struct {
	Text  string
	Error error
}

func GetPeriodClause(period string) (clause string, err error) {
	var days int
	switch period {
	case "day":
		days = 1
	case "week":
		days = 7
	case "month":
		days = 30
	case "year":
		days = 365
	default:
		return "", e.ErrInvalidPeriod
	}

	return fmt.Sprintf("submit_date >= date('now', '-%d days')", days), nil
}

func GetCatsWithEscapedChars(cats []string) []string {
	no_pers := GetCatsWithEscapedPeriods(cats)
	no_forward_slashes := GetCatsWithEscapedForwardSlashes(no_pers)
	no_hyphens := GetCatsWithEscapedHyphens(no_forward_slashes)

	return no_hyphens
}

func GetCatsWithEscapedPeriods(cats []string) []string {
	var escaped []string
	for i := 0; i < len(cats); i++ {
		if strings.Contains(cats[i], ".") {
			escaped = append(escaped, strings.ReplaceAll(cats[i], `.`, `"."`))
		} else {
			escaped = append(escaped, cats[i])
		}
	}

	return escaped
}

func GetCatsWithEscapedForwardSlashes(cats []string) []string {
	var escaped []string
	for i := 0; i < len(cats); i++ {
		if strings.Contains(cats[i], "/") {
			escaped = append(escaped, strings.ReplaceAll(cats[i], `/`, `"/"`))
		} else {
			escaped = append(escaped, cats[i])
		}
	}

	return escaped
}

func GetCatsWithEscapedHyphens(cats []string) []string {
	var escaped []string
	for i := 0; i < len(cats); i++ {
		if strings.Contains(cats[i], "-") {
			escaped = append(escaped, strings.ReplaceAll(cats[i], `-`, `"-"`))
		} else {
			escaped = append(escaped, cats[i])
		}
	}

	return escaped
}
