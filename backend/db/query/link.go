package db

import (
	"fmt"
	"strings"
)

// TODO: move constants to shared location, remove this duplicate
const LINKS_PAGE_LIMIT = 20
var LIMIT_CLAUSE = fmt.Sprintf(" LIMIT %d;", LINKS_PAGE_LIMIT)


// LINKS
type GetTopLinks struct {
	Query
}

const get_top_links_base = `SELECT 
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

func NewGetTopLinks() *GetTopLinks {
	new := &GetTopLinks{Query: Query{Text: get_top_links_base}}
	return new._Limit()
}

func (l *GetTopLinks) FromLinkIDs(link_ids []string) *GetTopLinks {
	link_ids_str := strings.Join(link_ids, ",")

	l._Where(fmt.Sprintf(`links_id IN (%s)`, link_ids_str))
	return l
}

func (l *GetTopLinks) DuringPeriod(period string) (*GetTopLinks) {
	clause , err := GetPeriodClause(period)
	if err != nil {
		l.Error = err
		return l
	}
	l._Where(clause)
	return l
}

func (l *GetTopLinks) Page(page int) *GetTopLinks {
	if page == 0 {
		return l
	}
	
	l.Text  = strings.Replace(l.Text, LIMIT_CLAUSE, fmt.Sprintf(" LIMIT %d OFFSET %d;", LINKS_PAGE_LIMIT +1, (page - 1) * LINKS_PAGE_LIMIT), 1)
	
	return l
}

func (l *GetTopLinks) _Limit() *GetTopLinks {
	l.Text = strings.Replace(l.Text, ";", LIMIT_CLAUSE, 1)
	return l
}

func (l *GetTopLinks) _Where(clause string) *GetTopLinks {

	// Swap previous WHERE for AND, if any
	l.Text = strings.Replace(l.Text, "WHERE", "AND", 1)

	// Prepend new clause
	l.Text = strings.Replace(l.Text, "ON Links.id = likes_link_id", fmt.Sprintf("ON Links.id = likes_link_id WHERE %s", clause), 1)

	return l
}



// LINK IDs
// e.g., for checking if a link exists
type GetLinkIDs struct {
	Query
}

const get_link_ids_base = "SELECT id FROM Links"

func NewGetLinkIDs(categories_str string) *GetLinkIDs {
	categories := strings.Split(categories_str, ",")

	new := &GetLinkIDs{Query: Query{Text: get_link_ids_base}}
	new._FromCategories(categories)
	return new
}

func (l *GetLinkIDs) _FromCategories(categories []string) *GetLinkIDs {
	l.Text += fmt.Sprintf(` WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'`, categories[0])
	for i := 1; i < len(categories); i++ {
		l.Text += fmt.Sprintf(` AND ',' || global_cats || ',' LIKE '%%,%s,%%'`, categories[i])
	}

	l.Text += ` GROUP BY id`

	return l
}



// LINK SUBCATEGORIES
type GetSubcategories struct {
	Query
}

const get_subcategories_base = "SELECT global_cats FROM Links"

func NewGetSubcategories(categories []string) *GetSubcategories {
	new := &GetSubcategories{Query: Query{Text: get_subcategories_base}}
	new._FromCategories(categories)
	return new
}

func (c *GetSubcategories) _FromCategories(categories []string) *GetSubcategories {
	c.Text += fmt.Sprintf(" WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[0])
	for i := 1; i < len(categories); i++ {
		c.Text += fmt.Sprintf(" AND ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[i])
	}
	
	c.Text +=" GROUP BY global_cats;"

	return c
}

func (c *GetSubcategories) DuringPeriod(period string) (*GetSubcategories) {
	clause , err := GetPeriodClause(period)
	if err != nil {
		c.Error = err
		return c
	}
	c._Where(clause)
	return c
}

func (c *GetSubcategories) _Where(clause string) *GetSubcategories {

	// Swap previous WHERE for AND, if any
	c.Text = strings.Replace(c.Text, "WHERE", "AND", 1)

	// Prepend new clause
	c.Text = strings.Replace(c.Text, get_subcategories_base, fmt.Sprintf("%s WHERE %s", get_subcategories_base, clause), 1)

	return c
}


// LINK COUNTS
type GetLinkCount struct {
	Query
}

const get_category_counts_base = "SELECT count(*) as link_count FROM Links"

func NewGetLinkCount(categories []string) *GetLinkCount {
	new := &GetLinkCount{Query: Query{Text: get_category_counts_base}}
	new._FromCategories(categories)
	return new
}


func (c *GetLinkCount) _FromCategories(categories []string) *GetLinkCount {
	c.Text += fmt.Sprintf(" WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[0])
	for i := 1; i < len(categories); i++ {
		c.Text += fmt.Sprintf(" AND ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[i])
	}

	c.Text +=";"

	return c
}



// CATEGORY CONTRIBUTORS
type GetCategoryContributors struct {
	Query
}

const get_category_contributors_base = `SELECT count(*), submitted_by FROM Links`

func NewGetCategoryContributors(categories []string) *GetCategoryContributors {
	new := &GetCategoryContributors{Query: Query{Text: get_category_contributors_base}}
	new._FromCategories(categories)
	return new
}

func (c *GetCategoryContributors) _FromCategories(categories []string) *GetCategoryContributors {
	c.Text += fmt.Sprintf(" WHERE ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[0])
	for i := 1; i < len(categories); i++ {
		c.Text += fmt.Sprintf(" AND ',' || global_cats || ',' LIKE '%%,%s,%%'", categories[i])
	}
	
	c.Text +=" GROUP BY submitted_by ORDER BY count(*) DESC, submitted_by ASC;"

	return c
}

func (c *GetCategoryContributors) DuringPeriod(period string) (*GetCategoryContributors) {
	clause , err := GetPeriodClause(period)
	if err != nil {
		c.Error = err
		return c
	}
	c._Where(clause)
	return c
}

func (c *GetCategoryContributors) Limit(limit int) *GetCategoryContributors {
	c.Text = strings.Replace(c.Text, ";", fmt.Sprintf(" LIMIT %d;", limit), 1)
	return c
}


func (c *GetCategoryContributors) _Where(clause string) *GetCategoryContributors {

	// Swap previous WHERE for AND, if any
	c.Text = strings.Replace(c.Text, "WHERE", "AND", 1)

	// Prepend new clause
	c.Text = strings.Replace(c.Text, get_category_contributors_base, fmt.Sprintf("%s WHERE %s", get_category_contributors_base, clause), 1)

	return c
}
