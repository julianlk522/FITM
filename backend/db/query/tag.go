package db

import (
	"fmt"
	"strings"
)

// Tags Page link
type TagPageLink struct {
	Query
}

const TAG_PAGE_LINK_BASE = `SELECT 
	links_id as link_id, 
	url, 
	sb, 
	sd, 
	cats, 
	summary,
	summary_count, 
	COUNT('Link Likes'.id) as like_count, 
	img_url, 
	COALESCE(is_liked,0) as is_liked, 
	COALESCE(is_copied,0) as is_copied
FROM 
	(
	SELECT 
		id as links_id, 
		url, 
		submitted_by as sb, 
		submit_date as sd, 
		COALESCE(global_cats,"") as cats, 
		global_summary as summary, 
		COALESCE(img_url,"") as img_url 
	FROM Links`

func NewTagPageLink(ID string, user_id string) *TagPageLink {
	return (&TagPageLink{Query: Query{Text: TAG_PAGE_LINK_BASE}})._FromID(ID)._ForSignedInUser(user_id)
}

func (l *TagPageLink) _FromID(ID string) *TagPageLink {
	l.Text += fmt.Sprintf(` WHERE id = '%s') `, ID)

	return l
}

func (l *TagPageLink) _ForSignedInUser(user_id string ) *TagPageLink {
	l.Text += fmt.Sprintf(` 
	LEFT JOIN
		(
		SELECT count(*) as summary_count, link_id as slink_id
		FROM Summaries
		GROUP BY slink_id
		)
	ON slink_id = links_id
	LEFT JOIN 'Link Likes'
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
type TagRankings struct {
	Query
}

const TAG_RANKINGS_BASE = `SELECT 
	(julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) * 100 AS lifespan_overlap, 
	categories, 
	Tags.submitted_by, 
	last_updated 
FROM Tags 
INNER JOIN Links 
ON Links.id = Tags.link_id `

func NewTagRankingsForLink(link_id string) *TagRankings {
	return (&TagRankings{Query: Query{Text: TAG_RANKINGS_BASE}})._FromLink(link_id)
}

func (t *TagRankings) _FromLink(link_id string) *TagRankings {

	t.Text += fmt.Sprintf(` WHERE link_id = '%s'
	ORDER BY lifespan_overlap DESC`, link_id)

	return t
}

func (t *TagRankings) Limit(limit int) *TagRankings {

	t.Text += fmt.Sprintf(` LIMIT %d;`, limit)
	return t
}


// All Global Cats
type GlobalCats struct {
	Query
}

const GLOBAL_CATS_BASE = `SELECT global_cats
	FROM Links
	WHERE global_cats != ""`

func NewAllGlobalCats() *GlobalCats {
	return &GlobalCats{Query: Query{Text: GLOBAL_CATS_BASE}}
}

func (t *GlobalCats) DuringPeriod(period string) *GlobalCats {
	clause, err := GetPeriodClause(period)
	if err != nil {
		t.Error = err
		return t
	}

	t.Text += fmt.Sprintf(` AND %s;`, clause)

	return t
}


// Top Tags (Overlap Scores)
type TopOverlapScores struct {
	Query
}

const TOP_OVERLAP_SCORES_BASE = `SELECT 
	(julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) AS lifespan_overlap, 
	categories 
FROM Tags 
INNER JOIN Links 
ON Links.id = Tags.link_id
ORDER BY lifespan_overlap DESC`


func NewTopOverlapScores(link_id string) *TopOverlapScores {
	return (&TopOverlapScores{Query: Query{Text: TOP_OVERLAP_SCORES_BASE}})._FromLink(link_id)
}

func (o *TopOverlapScores) _FromLink(link_id string) *TopOverlapScores {
	o. Text = strings.Replace(o.Text, "ORDER BY lifespan_overlap DESC", fmt.Sprintf("WHERE link_id = '%s' ORDER BY lifespan_overlap DESC", link_id), 1)

	return o
}

func (o *TopOverlapScores) Limit(limit int) *TopOverlapScores {
	o.Text += fmt.Sprintf(` LIMIT %d;`, limit)

	return o
}