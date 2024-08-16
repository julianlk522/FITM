package query

import (
	"fmt"
	"strings"
)

const (
	LINKS_PAGE_LIMIT = 20
	CATEGORY_CONTRIBUTORS_LIMIT int = 5
)

// Links
var UNPAGINATED_LIMIT_CLAUSE = fmt.Sprintf(` 
LIMIT %d;`, LINKS_PAGE_LIMIT)

func _PaginateLimitClause(page int) string {
	if page == 1 {
		return fmt.Sprintf(" LIMIT %d;", LINKS_PAGE_LIMIT + 1)
	}
	return fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT + 1, (page - 1) * LINKS_PAGE_LIMIT)
}

type TopLinks struct {
	Query
}

var TOP_LINKS_BASE = `SELECT 
links_id as link_id, 
url, 
sb, 
sd, 
cats, 
summary, 
COALESCE(summary_count,0) as summary_count,
tag_count,
like_count, 
img_url
FROM 
	(
	SELECT 
		Links.id as links_id, 
		url, 
		submitted_by as sb, 
		Links.submit_date as sd, 
		COALESCE(global_cats,"") as cats, 
		COALESCE(global_summary,"") as summary, 
		COALESCE(like_count,0) as like_count, 
		COALESCE(img_url,"") as img_url 
	FROM LINKS 
	LEFT JOIN 
		(
		SELECT link_id as likes_link_id, count(*) as like_count 
		FROM 'Link Likes'
		GROUP BY likes_link_id
		) 
	ON Links.id = likes_link_id
	)
LEFT JOIN
	(
	SELECT count(*) as summary_count, link_id as slink_id
	FROM Summaries
	GROUP BY slink_id
	)
ON slink_id = links_id
LEFT JOIN 
	(
	SELECT count(*) as tag_count, link_id as tlink_id
	FROM Tags
	GROUP BY tlink_id
	)
ON tlink_id = links_id
GROUP BY links_id 
ORDER BY like_count DESC, summary_count DESC, link_id DESC` + UNPAGINATED_LIMIT_CLAUSE

func NewTopLinks() *TopLinks {
	return (&TopLinks{Query: Query{Text: TOP_LINKS_BASE}})
}

func (l *TopLinks) FromLinkIDs(link_ids []string) *TopLinks {
	link_ids_str := strings.Join(link_ids, ",")

	l._Where(fmt.Sprintf(`links_id IN (%s)`, link_ids_str))
	return l
}

func (l *TopLinks) DuringPeriod(period string) (*TopLinks) {
	clause , err := GetPeriodClause(period)
	if err != nil {
		l.Error = err
		return l
	}
	l._Where(clause)
	return l
}

func (l *TopLinks) Page(page int) *TopLinks {
	if page == 0 {
		return l
	}
	
	l.Text  = strings.Replace(l.Text, UNPAGINATED_LIMIT_CLAUSE, _PaginateLimitClause(page), 1)
	
	return l
}

func (l *TopLinks) _Where(clause string) *TopLinks {

	// Swap previous WHERE for AND, if any
	l.Text = strings.Replace(l.Text, "WHERE", "AND", 1)

	l.Text = strings.Replace(l.Text, "ON Links.id = likes_link_id", fmt.Sprintf("ON Links.id = likes_link_id WHERE %s", clause), 1)

	return l
}



// Link IDs
type LinkIDs struct {
	Query
}

const LINK_IDS_BASE = "SELECT id FROM Links"

func NewLinkIDs(categories_str string) *LinkIDs {
	categories := strings.Split(categories_str, ",")

	return (&LinkIDs{Query: Query{Text: LINK_IDS_BASE}})._FromCategories(categories)
}

func (l *LinkIDs) _FromCategories(categories []string) *LinkIDs {
	l.Text += fmt.Sprintf(` WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'`, categories[0])
	for i := 1; i < len(categories); i++ {
		l.Text += fmt.Sprintf(` AND ',' || global_cats || ',' LIKE '%%,%s,%%'`, categories[i])
	}

	l.Text += ` GROUP BY id`

	return l
}



// Subcats
type Subcats struct {
	Query
}

const SUBCATS_BASE = "SELECT global_cats FROM Links"

func NewSubcats(categories []string) *Subcats {
	return (&Subcats{Query: Query{Text: SUBCATS_BASE}})._FromCategories(categories)
}

func (c *Subcats) _FromCategories(categories []string) *Subcats {
	c.Text += fmt.Sprintf(" WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[0])
	for i := 1; i < len(categories); i++ {
		c.Text += fmt.Sprintf(" AND ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[i])
	}
	
	c.Text +=" GROUP BY global_cats;"

	return c
}

func (c *Subcats) DuringPeriod(period string) (*Subcats) {
	clause , err := GetPeriodClause(period)
	if err != nil {
		c.Error = err
		return c
	}
	c._Where(clause)
	return c
}

func (c *Subcats) _Where(clause string) *Subcats {

	// Swap previous WHERE for AND, if any
	c.Text = strings.Replace(c.Text, "WHERE", "AND", 1)

	// Prepend new clause
	c.Text = strings.Replace(c.Text, SUBCATS_BASE, fmt.Sprintf("%s WHERE %s", SUBCATS_BASE, clause), 1)

	return c
}


// Cat counts
type CatCount struct {
	Query
}

const CAT_COUNT_BASE = "SELECT count(*) as link_count FROM Links"

func NewCatCount(categories []string) *CatCount {
	return (&CatCount{Query: Query{Text: CAT_COUNT_BASE}})._FromCategories(categories)
}


func (c *CatCount) _FromCategories(categories []string) *CatCount {
	c.Text += fmt.Sprintf(" WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[0])
	for i := 1; i < len(categories); i++ {
		c.Text += fmt.Sprintf(" AND ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[i])
	}
	c.Text +=";"

	return c
}



// Cats contributors (single or multiple cats)
type CatsContributors struct {
	Query
}

const CATS_CONTRIBUTORS_BASE = `SELECT 
	count(*), submitted_by 
FROM Links`

func NewCatsContributors(categories []string) *CatsContributors {
	return (&CatsContributors{Query: Query{Text: CATS_CONTRIBUTORS_BASE}})._FromCategories(categories)
}

func (c *CatsContributors) _FromCategories(categories []string) *CatsContributors {
	c.Text += fmt.Sprintf(" WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[0])
	for i := 1; i < len(categories); i++ {
		c.Text += fmt.Sprintf(" AND ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[i])
	}
	
	c.Text += fmt.Sprintf(` 
		GROUP BY submitted_by 
		ORDER BY count(*) DESC, submitted_by ASC
		LIMIT %d;`, CATEGORY_CONTRIBUTORS_LIMIT)
	return c
}

func (c *CatsContributors) DuringPeriod(period string) (*CatsContributors) {
	clause , err := GetPeriodClause(period)
	if err != nil {
		c.Error = err
		return c
	}

	// Swap previous WHERE for AND, if any
	c.Text = strings.Replace(c.Text, "WHERE", "AND", 1)

	// Prepend new clause
	c.Text = strings.Replace(
		c.Text, 
		CATS_CONTRIBUTORS_BASE, 
		fmt.Sprintf(
			"%s WHERE %s", 
			CATS_CONTRIBUTORS_BASE, 
			clause), 
	1)

	return c
}
