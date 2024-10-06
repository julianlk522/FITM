package query

import (
	"fmt"
	"strings"

	e "github.com/julianlk522/fitm/error"
)

type Query struct {
	Text  string
	Args  []interface{}
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
	chars_replacer := strings.NewReplacer(
		".", `"."`,
		"/", `"/"`,
		"-", `"-"`,
	)
	for i := 0; i < len(cats); i++ {
		cats[i] = chars_replacer.Replace(cats[i])
	}
	return cats
}