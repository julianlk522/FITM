package query

import (
	"fmt"
	"strings"
)

const (
	LINKS_PAGE_LIMIT        = 20
	CONTRIBUTORS_PAGE_LIMIT = 10
)

// Links
var LINKS_UNPAGINATED_LIMIT_CLAUSE = fmt.Sprintf(
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

const LINKS_BASE_CTES = `WITH LikeCount AS (
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

const LINKS_FROM = `
FROM
	Links l`

const LINKS_BASE_JOINS = `
LEFT JOIN LikeCount ll ON l.id = ll.link_id
LEFT JOIN SummaryCount s ON l.id = s.link_id
LEFT JOIN TagCount t ON l.id = t.link_id`

const LINKS_AUTH_JOINS = `
LEFT JOIN IsLiked il ON l.id = il.link_id
LEFT JOIN IsCopied ic ON l.id = ic.link_id`

const LINKS_NO_NSFW_CATS_WHERE = `
WHERE l.id NOT IN (
	SELECT link_id FROM global_cats_fts WHERE global_cats MATCH 'NSFW'
)`

const LINKS_ORDER_BY = ` 
ORDER BY 
    like_count DESC, 
    summary_count DESC, 
    l.id DESC`

func NewTopLinks() *TopLinks {
	return (&TopLinks{
		Query: Query{
			Text: LINKS_BASE_CTES +
				LINKS_BASE_FIELDS +
				LINKS_FROM +
				LINKS_BASE_JOINS +
				LINKS_NO_NSFW_CATS_WHERE +
				LINKS_ORDER_BY +
				LINKS_UNPAGINATED_LIMIT_CLAUSE,
		},
	})
}

func (l *TopLinks) FromCats(cats []string) *TopLinks {
	if len(cats) == 0 || cats[0] == "" {
		l.Error = fmt.Errorf("no cats provided")
		return l
	}

	cats = GetCatsWithEscapedPeriods(cats)

	// build clause from cats
	clause := fmt.Sprintf(`WHERE global_cats MATCH '%s`, cats[0])
	for i := 1; i < len(cats); i++ {
		clause += fmt.Sprintf(`
		AND %s`, cats[i])
	}
	clause += `'`

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
		LINKS_BASE_CTES,
		LINKS_BASE_CTES+cats_cte,
		1)

	// append join
	l.Text = strings.Replace(
		l.Text,
		LINKS_FROM,
		LINKS_FROM+"\n"+LINKS_CATS_JOIN,
		1,
	)

	return l
}

const LINKS_CATS_JOIN = `INNER JOIN CatsFilter f ON l.id = f.link_id`

func (l *TopLinks) DuringPeriod(period string) *TopLinks {
	clause, err := GetPeriodClause(period)
	if err != nil {
		l.Error = err
		return l
	}

	l.Text = strings.Replace(
		l.Text,
		LINKS_ORDER_BY,
		"\n"+"AND "+clause+LINKS_ORDER_BY,
		1,
	)

	return l
}

func (l *TopLinks) SortBy(order_by string) *TopLinks {

	// acceptable order_by values:
	// newest
	// rating (default)

	updated_order_by_clause := `
	ORDER BY `
	switch order_by {
	case "newest":
		updated_order_by_clause += "submit_date DESC, like_count DESC, summary_count DESC"
	case "rating":
		updated_order_by_clause += "like_count DESC, summary_count DESC, submit_date DESC"
	default:
		l.Error = fmt.Errorf("invalid order_by value")
		return l
	}

	l.Text = strings.Replace(
		l.Text,
		LINKS_ORDER_BY,
		updated_order_by_clause,
		1,
	)

	return l
}

func (l *TopLinks) AsSignedInUser(req_user_id string) *TopLinks {

	// append auth CTEs
	l.Text = strings.Replace(
		l.Text,
		LINKS_BASE_CTES,
		LINKS_BASE_CTES+LINKS_AUTH_CTES,
		1,
	)

	// append auth fields
	l.Text = strings.Replace(
		l.Text,
		LINKS_BASE_FIELDS,
		LINKS_BASE_FIELDS+LINKS_AUTH_FIELDS,
		1,
	)

	// apend auth FROM
	l.Text = strings.Replace(
		l.Text,
		LINKS_BASE_JOINS,
		LINKS_BASE_JOINS+LINKS_AUTH_JOINS,
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

func (l *TopLinks) NSFW() *TopLinks {

	// remove NSFW clause
	l.Text = strings.Replace(
		l.Text,
		LINKS_NO_NSFW_CATS_WHERE,
		"",
		1,
	)

	// replace AND with WHERE in .DuringPeriod clause
	l.Text = strings.Replace(
		l.Text,
		"AND submit_date",
		"WHERE submit_date",
		1,
	)

	return l
}

func (l *TopLinks) Page(page int) *TopLinks {
	if page == 0 {
		return l
	}

	l.Text = strings.Replace(
		l.Text,
		LINKS_UNPAGINATED_LIMIT_CLAUSE,
		_PaginateLimitClause(page),
		1)

	return l
}

// Cats contributors (single or multiple cats)
type Contributors struct {
	Query
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

func NewContributors() *Contributors {
	return (&Contributors{
		Query: Query{
			Text: CONTRIBUTORS_FIELDS + CONTRIBUTORS_GBOBL,
		},
	})
}

const CONTRIBUTORS_CATS_FROM = `INNER JOIN CatsFilter f ON l.id = f.link_id`

func (c *Contributors) FromCats(cats []string) *Contributors {
	cats = GetCatsWithEscapedPeriods(cats)

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
