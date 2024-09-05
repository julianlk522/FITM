package query

import (
	"fmt"
	"strings"
)

const (
	LINKS_PAGE_LIMIT                = 20
	CONTRIBUTORS_PAGE_LIMIT         = 10
)

// Links
var UNPAGINATED_LIMIT_CLAUSE = fmt.Sprintf(
	` 
	LIMIT %d;`, 
	LINKS_PAGE_LIMIT,
)

func _PaginateLimitClause(page int) string {
	if page == 1 {
		return fmt.Sprintf(" LIMIT %d;", LINKS_PAGE_LIMIT+1)
	}
	return fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT+1, (page-1)*LINKS_PAGE_LIMIT)
}

type TopLinks struct {
	Query
}

const LINKS_BASE_CTES = 
`WITH LikeCount AS (
    SELECT link_id, COUNT(*) AS like_count 
    FROM 'Link Likes'
    GROUP BY link_id
),
SummaryCount AS (
    SELECT link_id, COUNT(*) AS summary_count
    FROM Summaries
    GROUP BY link_id
),
TagCount AS (
    SELECT link_id, COUNT(*) AS tag_count
    FROM Tags
    GROUP BY link_id
)`

const LINKS_AUTH_CTES = `,
IsLiked AS (
	SELECT link_id, COUNT(*) AS is_liked
	FROM 'Link Likes'
	WHERE user_id = 'REQ_USER_ID'
	GROUP BY link_id
),
IsCopied AS (
	SELECT link_id, COUNT(*) AS is_copied
	FROM 'Link Copies'
	WHERE user_id = 'REQ_USER_ID'
	GROUP BY link_id
)`

const LINKS_BASE_FIELDS = ` 
SELECT 
	l.id, 
    l.url, 
    l.submitted_by AS sb, 
    l.submit_date AS sd, 
    COALESCE(l.global_cats, '') AS cats, 
    COALESCE(l.global_summary, '') AS summary, 
    COALESCE(s.summary_count, 0) AS summary_count,
    COALESCE(t.tag_count, 0) AS tag_count,
    COALESCE(ll.like_count, 0) AS like_count, 
    COALESCE(l.img_url, '') AS img_url`

const LINKS_AUTH_FIELDS = `,
	COALESCE(il.is_liked,0) AS is_liked,
	COALESCE(ic.is_copied,0) AS is_copied`

const LINKS_BASE_FROM = ` 
FROM 
    LINKS l
LEFT JOIN LikeCount ll ON l.id = ll.link_id
LEFT JOIN SummaryCount s ON l.id = s.link_id
LEFT JOIN TagCount t ON l.id = t.link_id`

const LINKS_AUTH_FROM = `
LEFT JOIN IsLiked il ON l.id = il.link_id
LEFT JOIN IsCopied ic ON l.id = ic.link_id`

const LINKS_CATS_FROM = `INNER JOIN CatsFilter f ON l.id = f.link_id`

const LINKS_ORDER_BY = ` 
ORDER BY 
    like_count DESC, 
    summary_count DESC, 
    l.id DESC`

func NewTopLinks() *TopLinks {
	return (&TopLinks{
		Query: Query{
			Text: 
				LINKS_BASE_CTES +
				LINKS_BASE_FIELDS +
				LINKS_BASE_FROM +
				LINKS_ORDER_BY +
				UNPAGINATED_LIMIT_CLAUSE,
		},
	})
}

func (l *TopLinks) FromCats(cats []string) *TopLinks {
	if len(cats) == 0 || cats[0] == "" {
		l.Error = fmt.Errorf("no cats provided")
		return l
	}

	// build clause from cats
	clause := fmt.Sprintf(`WHERE global_cats MATCH '%s'`, cats[0])
	for i := 1; i < len(cats); i++ {
		clause += fmt.Sprintf(` 
		AND global_cats MATCH '%s'`, cats[i])
	}

	// build CTE from clause
	cats_cte := fmt.Sprintf(
			`,
		CatsFilter AS (
			SELECT link_id
			FROM global_cats_fts
			%s
		)`, clause,
	)

	// prepend CTE
	l.Text = strings.Replace(
		l.Text,
		LINKS_BASE_FIELDS,
		cats_cte + "\n" + LINKS_BASE_FIELDS,
	1)

	// append join
	l.Text = strings.Replace(
		l.Text,
		LINKS_ORDER_BY,
		"\n" + LINKS_CATS_FROM + LINKS_ORDER_BY,
		1,
	)
	
	return l
}

func (l *TopLinks) DuringPeriod(period string) *TopLinks {
	clause, err := GetPeriodClause(period)
	if err != nil {
		l.Error = err
		return l
	}
	
	l.Text = strings.Replace(
		l.Text,
		LINKS_ORDER_BY,
		"\n" + "WHERE " + clause + LINKS_ORDER_BY,
		1,
	)

	return l
}

func (l *TopLinks) AsSignedInUser(req_user_id string) *TopLinks {

	// append auth CTEs
	l.Text = strings.Replace(
		l.Text,
		LINKS_BASE_CTES,
		LINKS_BASE_CTES + LINKS_AUTH_CTES,
		1,
	)

	// append auth fields
	l.Text = strings.Replace(
		l.Text,
		LINKS_BASE_FIELDS,
		LINKS_BASE_FIELDS + LINKS_AUTH_FIELDS,
		1,
	)

	// apend auth FROM
	l.Text = strings.Replace(
		l.Text,
		LINKS_BASE_FROM,
		LINKS_BASE_FROM + LINKS_AUTH_FROM,
		1,
	)

	// swap all "REQ_USER_ID" with req_user_id
	l.Text = strings.ReplaceAll(
		l.Text,
		"REQ_USER_ID",
		req_user_id,
	)

	return l
}

func (l *TopLinks) Page(page int) *TopLinks {
	if page == 0 {
		return l
	}

	l.Text = strings.Replace(
		l.Text, 
		UNPAGINATED_LIMIT_CLAUSE, 
		_PaginateLimitClause(page), 
	1)

	return l
}

// Cats contributors (single or multiple cats)
type Contributors struct {
	Query
}

const CONTRIBUTORS_FIELDS = `SELECT count(*), submitted_by 
FROM Links`

var CONTRIBUTORS_GBOBL = fmt.Sprintf(
	` 
	GROUP BY submitted_by
	ORDER BY count(*) DESC, submitted_by ASC
	LIMIT %d;`, CONTRIBUTORS_PAGE_LIMIT,
)

func NewContributors() *Contributors {
	return (&Contributors{
		Query: Query{
			Text: CONTRIBUTORS_FIELDS + CONTRIBUTORS_GBOBL,
		},
	})
}

func (c *Contributors) FromCats(cats []string) *Contributors {
	clause := fmt.Sprintf(" WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'", cats[0])
	for i := 1; i < len(cats); i++ {
		clause += fmt.Sprintf(" AND ',' || global_cats || ',' LIKE '%%,%s,%%'", cats[i])
	}

	// Swap previous WHERE for AND, if any
	c.Text = strings.Replace(c.Text, "WHERE", "AND", 1)

	c.Text = strings.Replace(
		c.Text, 
		CONTRIBUTORS_FIELDS,
		CONTRIBUTORS_FIELDS + clause, 
		1,
	)

	return c
}

func (c *Contributors) DuringPeriod(period string) *Contributors {
	clause, err := GetPeriodClause(period)
	if err != nil {
		c.Error = err
		return c
	}

	// Swap previous WHERE for AND, if any
	c.Text = strings.Replace(c.Text, "WHERE", "AND", 1)

	// Prepend new clause
	c.Text = strings.Replace(
		c.Text,
		CONTRIBUTORS_FIELDS,
		fmt.Sprintf(
			"%s WHERE %s",
			CONTRIBUTORS_FIELDS,
			clause),
		1)

	return c
}
