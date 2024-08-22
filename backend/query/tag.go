package query

import (
	"fmt"
	"strings"
)

const (
	TAGS_PAGE_LIMIT              = 20
	TOP_OVERLAP_SCORES_LIMIT     = 20
	TOP_GLOBAL_CATS_LIMIT        = 15
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
	FROM Links
LEFT JOIN
	(
	SELECT count(*) as summary_count, link_id as slink_id
	FROM Summaries
	GROUP BY slink_id
	)
ON slink_id = links_id`

func NewTagPageLink(link_id string, user_id string) *TagPageLink {
	return (&TagPageLink{Query: Query{Text: TAG_PAGE_LINK_BASE}})._FromID(link_id)._AsSignedInUser(user_id)
}

func (l *TagPageLink) _FromID(link_id string) *TagPageLink {
	l.Text = strings.Replace(
		l.Text,
		"FROM Links",
		fmt.Sprintf(
			`FROM Links 
			WHERE id = '%s'
			)`,
			link_id,
		),
		1)

	return l
}

func (l *TagPageLink) _AsSignedInUser(user_id string) *TagPageLink {
	l.Text += fmt.Sprintf(` 
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

// Tag Rankings (cat overlap scores)
type TagRankings struct {
	Query
}

const TOP_OVERLAP_SCORES_BASE_FIELDS = `SELECT
(julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) AS lifespan_overlap, 
cats`

const TOP_OVERLAP_SCORES_PUBLIC_FIELDS = `, 
Tags.submitted_by, 
last_updated `

var TOP_OVERLAP_SCORES_BASE = TOP_OVERLAP_SCORES_BASE_FIELDS + ` 
FROM Tags 
INNER JOIN Links 
ON Links.id = Tags.link_id
WHERE link_id = 'LINK_ID'
ORDER BY lifespan_overlap DESC` + fmt.Sprintf(`
LIMIT %d`, TOP_OVERLAP_SCORES_LIMIT)

func NewTagRankings(link_id string) *TagRankings {
	return (&TagRankings{Query: Query{Text: TOP_OVERLAP_SCORES_BASE}})._ForLink(link_id)
}

func (o *TagRankings) _ForLink(link_id string) *TagRankings {
	o.Text = strings.Replace(
		o.Text,
		"LINK_ID",
		link_id,
		1)

	return o
}

func (o *TagRankings) Public() *TagRankings {
	o.Text = strings.Replace(
		o.Text,
		TOP_OVERLAP_SCORES_BASE_FIELDS,
		TOP_OVERLAP_SCORES_BASE_FIELDS + TOP_OVERLAP_SCORES_PUBLIC_FIELDS,
		1,
	)

	return o
}

// Global Cat Counts
type GlobalCatCounts struct {
	Query
}

const GLOBAL_CAT_COUNTS_BASE = `WITH RECURSIVE split(id, global_cats, str) AS 
	(
	SELECT id, '', global_cats||',' 
	FROM Links
	UNION ALL SELECT
	id,
	substr(str, 0, instr(str, ',')),
	substr(str, instr(str, ',') + 1)
	FROM split
	WHERE str != ''
	)
SELECT global_cats, count(global_cats) as count
FROM split
WHERE global_cats != ''
GROUP BY global_cats
ORDER BY count DESC, global_cats ASC;`

func NewTopGlobalCatCounts() *GlobalCatCounts {
	return (&GlobalCatCounts{Query: Query{Text: GLOBAL_CAT_COUNTS_BASE}})._Limit(TOP_GLOBAL_CATS_LIMIT)
}

func (t *GlobalCatCounts) DuringPeriod(period string) *GlobalCatCounts {
	clause, err := GetPeriodClause(period)
	if err != nil {
		t.Error = err
		return t
	}

	t.Text = strings.Replace(
		t.Text,
		"FROM Links",
		fmt.Sprintf(
			`FROM Links
			WHERE %s`,
			clause,
		),
		1)

	return t
}

func (t *GlobalCatCounts) _Limit(limit int) *GlobalCatCounts {

	t.Text = strings.Replace(
		t.Text,
		";",
		fmt.Sprintf(
			`
LIMIT %d;`,
			limit,
		),
		1)

	return t
}
