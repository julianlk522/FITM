package db

import (
	"fmt"
	"strings"
)

// TODO: move constants to shared location, remove this duplicate
const LINKS_PAGE_LIMIT = 20
var LIMIT_CLAUSE = fmt.Sprintf(" LIMIT %d;", LINKS_PAGE_LIMIT)


// LINKS
type TopLinks struct {
	Query
}

const TOP_LINKS_BASE = `SELECT 
links_id as link_id, 
url, 
link_author as submitted_by, 
sd, 
categories, 
summary, 
coalesce(count(Summaries.id),0) as summary_count, 
like_count, 
img_url
FROM 
	(
	SELECT Links.id as links_id, url, submitted_by as link_author, Links.submit_date as sd, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url 
	FROM LINKS 
	LEFT JOIN 
		(
		SELECT link_id as likes_link_id, count(*) as like_count 
		FROM 'Link Likes'
		GROUP BY likes_link_id
		) 
	ON Links.id = likes_link_id
	)
LEFT JOIN Summaries 
ON Summaries.link_id = links_id 
GROUP BY links_id 
ORDER BY like_count DESC, summary_count DESC, link_id DESC;`

func NewTopLinks() *TopLinks {
	return (&TopLinks{Query: Query{Text: TOP_LINKS_BASE}})._Limit()
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
	
	l.Text  = strings.Replace(l.Text, LIMIT_CLAUSE, fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT +1, (page - 1) * LINKS_PAGE_LIMIT), 1)
	
	return l
}

func (l *TopLinks) _Limit() *TopLinks {
	l.Text = strings.Replace(l.Text, ";", LIMIT_CLAUSE, 1)
	return l
}

func (l *TopLinks) _Where(clause string) *TopLinks {

	// Swap previous WHERE for AND, if any
	l.Text = strings.Replace(l.Text, "WHERE", "AND", 1)

	l.Text = strings.Replace(l.Text, "ON Links.id = likes_link_id", fmt.Sprintf("ON Links.id = likes_link_id WHERE %s", clause), 1)

	return l
}



// LINK IDs
// e.g., for checking if a link exists
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



// LINK SUBCATEGORIES
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


// CAT COUNTS
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



// CATS CONTRIBUTORS
// (single or multiple cats)
type CatsContributors struct {
	Query
}

const CATS_CONTRIBUTORS_BASE = `SELECT count(*), submitted_by FROM Links`

func NewCatsContributors(categories []string) *CatsContributors {
	return (&CatsContributors{Query: Query{Text: CATS_CONTRIBUTORS_BASE}})._FromCategories(categories)
}

func (c *CatsContributors) _FromCategories(categories []string) *CatsContributors {
	c.Text += fmt.Sprintf(" WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[0])
	for i := 1; i < len(categories); i++ {
		c.Text += fmt.Sprintf(" AND ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[i])
	}
	
	c.Text +=" GROUP BY submitted_by ORDER BY count(*) DESC, submitted_by ASC;"

	return c
}

func (c *CatsContributors) DuringPeriod(period string) (*CatsContributors) {
	clause , err := GetPeriodClause(period)
	if err != nil {
		c.Error = err
		return c
	}
	c._Where(clause)
	return c
}

func (c *CatsContributors) Limit(limit int) *CatsContributors {
	c.Text = strings.Replace(c.Text, ";", fmt.Sprintf(" LIMIT %d;", limit), 1)
	return c
}


func (c *CatsContributors) _Where(clause string) *CatsContributors {

	// Swap previous WHERE for AND, if any
	c.Text = strings.Replace(c.Text, "WHERE", "AND", 1)

	// Prepend new clause
	c.Text = strings.Replace(c.Text, CATS_CONTRIBUTORS_BASE, fmt.Sprintf("%s WHERE %s", CATS_CONTRIBUTORS_BASE, clause), 1)

	return c
}
