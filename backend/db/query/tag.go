package db

import (
	"fmt"
	"strings"
)

// Tags for Link
type GetTagPageLink struct {
	Query
}

const GET_TAG_PAGE_LINK_BASE = `SELECT links_id as link_id, url, submitted_by, submit_date, coalesce(categories,"") as categories, summary, COUNT('Link Likes'.id) as like_count, img_url, COALESCE(is_liked,0) as is_liked, COALESCE(is_copied,0) as is_copied
FROM 
	(
	SELECT id as links_id, url, submitted_by, submit_date, global_cats as categories, global_summary as summary, coalesce(img_url,"") as img_url 
		FROM Links`

func NewGetTagPageLink(ID string, user_id string) *GetTagPageLink {
	new := &GetTagPageLink{Query: Query{Text: GET_TAG_PAGE_LINK_BASE}}
	return new._FromID(ID)._ForSignedInUser(user_id)
}

func (l *GetTagPageLink) _FromID(ID string) *GetTagPageLink {
	l.Text += fmt.Sprintf(` WHERE id = '%s'
	) 
`, ID)

	return l
}

func (l *GetTagPageLink) _ForSignedInUser(user_id string ) *GetTagPageLink {
	l.Text += fmt.Sprintf(`LEFT JOIN 'Link Likes'
	ON 'Link Likes'.link_id = links_id
	LEFT JOIN
		(
		SELECT id as like_id, count(*) as is_liked, user_id as luser_id, link_id as like_link_id2
		FROM 'Link Likes'
		WHERE luser_id = '%[1]s'
		GROUP BY like_id
		)
	ON like_link_id2 = link_id
	LEFT JOIN
		(
		SELECT id as copy_id, count(*) as is_copied, user_id as cuser_id, link_id as copy_link_id
		FROM 'Link Copies'
		WHERE cuser_id = '%[1]s'
		GROUP BY copy_id
		)
	ON copy_link_id = link_id;`, user_id)

	return l
}


// Earliest Tags for Link
type GetEarliestTags struct {
	Query
}

const GET_EARLIEST_TAGS_BASE = `SELECT (julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) * 100 AS lifespan_overlap, categories, Tags.submitted_by, last_updated 
	FROM Tags 
	INNER JOIN Links 
	ON Links.id = Tags.link_id `

func NewGetEarliestTags(link_id string) *GetEarliestTags {
	new := &GetEarliestTags{Query: Query{Text: GET_EARLIEST_TAGS_BASE}}
	return new._FromLink(link_id)
}

func (t *GetEarliestTags) _FromLink(link_id string) *GetEarliestTags {

	t.Text += fmt.Sprintf(` WHERE link_id = '%s'
	ORDER BY lifespan_overlap DESC`, link_id)

	return t
}

func (t *GetEarliestTags) Limit(limit int) *GetEarliestTags {

	t.Text += fmt.Sprintf(` LIMIT %d;`, limit)
	return t
}


// All Global Categories
type GetAllGlobalCategories struct {
	Query
}

const GET_ALL_GLOBAL_CATEGORIES_BASE = `SELECT global_cats
		FROM Links
		WHERE global_cats != ""`

func NewGetAllGlobalCategories() *GetAllGlobalCategories {
	new := &GetAllGlobalCategories{Query: Query{Text: GET_ALL_GLOBAL_CATEGORIES_BASE}}
	return new
}

func (t *GetAllGlobalCategories) FromPeriod(period string) *GetAllGlobalCategories {
	clause, err := GetPeriodClause(period)
	if err != nil {
		t.Error = err
		return t
	}

	t.Text += fmt.Sprintf(` AND %s;`, clause)

	return t
}


// Top Tags (Overlap Scores)
type GetTopOverlapScores struct {
	Query
}

const GET_TOP_OVERLAP_SCORES_BASE = `
SELECT (julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) AS lifespan_overlap, categories 
	FROM Tags 
	INNER JOIN Links 
	ON Links.id = Tags.link_id
	ORDER BY lifespan_overlap DESC`


func NewGetTopOverlapScores(link_id string) *GetTopOverlapScores {
	new := &GetTopOverlapScores{Query: Query{Text: GET_TOP_OVERLAP_SCORES_BASE}}
	return new._FromLink(link_id)
}

func (o *GetTopOverlapScores) _FromLink(link_id string) *GetTopOverlapScores {
	o. Text = strings.Replace(o.Text, "ORDER BY lifespan_overlap DESC", fmt.Sprintf("WHERE link_id = '%s' ORDER BY lifespan_overlap DESC", link_id), 1)

	return o
}

func (o *GetTopOverlapScores) Limit(limit int) *GetTopOverlapScores {
	o.Text += fmt.Sprintf(` LIMIT %d;`, limit)

	return o
}