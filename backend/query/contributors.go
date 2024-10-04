package query

import (
	"fmt"
	"strings"
)

const CONTRIBUTORS_PAGE_LIMIT = 10

type Contributors struct {
	Query
}

func NewContributors() *Contributors {
	return (&Contributors{
		Query: Query{
			Text: 
				CONTRIBUTORS_FIELDS + 
				CONTRIBUTORS_GBOBL,
		},
	})
}

const CONTRIBUTORS_FIELDS = `SELECT
count(l.id) as count, l.submitted_by
FROM Links l`

var CONTRIBUTORS_GBOBL = fmt.Sprintf(
	` 
	GROUP BY l.submitted_by
	ORDER BY count DESC, l.submitted_by ASC
	LIMIT %d;`, CONTRIBUTORS_PAGE_LIMIT,
)

func (c *Contributors) FromCats(cats []string) *Contributors {
	cats = GetCatsWithEscapedChars(cats)

	clause := fmt.Sprintf("WHERE global_cats MATCH '%s", cats[0])
	for i := 1; i < len(cats); i++ {
		clause += fmt.Sprintf(" AND %s", cats[i])
	}
	clause += "'"

	// build CTE
	cats_cte := fmt.Sprintf(
		`WITH CatsFilter AS (
	SELECT link_id
	FROM global_cats_fts
	%s
		)`, clause,
	)

	// prepend CTE
	c.Text = strings.Replace(
		c.Text,
		CONTRIBUTORS_FIELDS,
		cats_cte+"\n"+CONTRIBUTORS_FIELDS,
		1)

	// append join
	c.Text = strings.Replace(
		c.Text,
		CONTRIBUTORS_FIELDS,
		CONTRIBUTORS_FIELDS+"\n"+CONTRIBUTORS_CATS_FROM,
		1,
	)

	return c
}

const CONTRIBUTORS_CATS_FROM = `INNER JOIN CatsFilter f ON l.id = f.link_id`

func (c *Contributors) DuringPeriod(period string) *Contributors {
	clause, err := GetPeriodClause(period)
	if err != nil {
		c.Error = err
		return c
	}

	clause = strings.Replace(
		clause,
		"submit_date",
		"l.submit_date",
		1,
	)

	// Prepend new clause
	c.Text = strings.Replace(
		c.Text,
		CONTRIBUTORS_GBOBL,
		"\n"+"WHERE "+clause+CONTRIBUTORS_GBOBL,
		1)

	return c
}