package query

import (
	"fmt"
	"strings"
)

const (
	TAG_RANKINGS_PAGE_LIMIT   = 20
	GLOBAL_CATS_PAGE_LIMIT    = 20
	SPELLFIX_MATCHES_LIMIT    =  3
)

// Tags Page link
type TagPageLink struct {
	Query
}

const TAG_PAGE_LINK_BASE_FIELDS = `SELECT 
	links_id as link_id, 
	url, 
	sb, 
	sd, 
	cats, 
	summary,
	COALESCE(summary_count,0) as summary_count, 
	COUNT('Link Likes'.id) as like_count, 
	img_url`

const TAG_PAGE_LINK_AUTH_FIELDS = `,
	COALESCE(is_liked,0) as is_liked, 
	COALESCE(is_copied,0) as is_copied`

const TAG_PAGE_LINK_BASE_FROM = `
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
LEFT JOIN 'Link Likes'
	ON 'Link Likes'.link_id = links_id
LEFT JOIN
	(
	SELECT count(*) as summary_count, link_id as slink_id
	FROM Summaries
	GROUP BY slink_id
	)
ON slink_id = links_id`

func NewTagPageLink(link_id string) *TagPageLink {
	return (&TagPageLink{Query: Query{Text: TAG_PAGE_LINK_BASE_FIELDS + TAG_PAGE_LINK_BASE_FROM}})._FromID(link_id)
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

func (l *TagPageLink) AsSignedInUser(user_id string) *TagPageLink {
	l.Text = strings.Replace(
		l.Text,
		TAG_PAGE_LINK_BASE_FIELDS,
		TAG_PAGE_LINK_BASE_FIELDS+TAG_PAGE_LINK_AUTH_FIELDS,
		1,
	)

	l.Text += fmt.Sprintf(` 
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
(julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) * 100 AS lifespan_overlap, 
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
LIMIT %d`, TAG_RANKINGS_PAGE_LIMIT)

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
		TOP_OVERLAP_SCORES_BASE_FIELDS+TOP_OVERLAP_SCORES_PUBLIC_FIELDS,
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
ORDER BY count DESC, LOWER(global_cats) ASC;`

func NewTopGlobalCatCounts() *GlobalCatCounts {
	return (&GlobalCatCounts{
		Query: Query{
			Text: GLOBAL_CAT_COUNTS_BASE,
		},
	})._Limit(GLOBAL_CATS_PAGE_LIMIT)
}

func (t *GlobalCatCounts) SubcatsOfCats(cats_params string) *GlobalCatCounts {
	cats := strings.Split(cats_params, ",")
	
	// build match clause
	match_cats := make([]string, len(cats))
	copy(match_cats, cats)
	// escape any "." in match_cats
	for i := 0; i < len(match_cats); i++ {
		if strings.Contains(match_cats[i], ".") {
			match_cats[i] = strings.Replace(match_cats[i], `.`, `"."`, 1)
		}
	}
	match_clause := fmt.Sprintf(`WHERE global_cats MATCH '%s`, match_cats[0])
	for i := 1; i < len(match_cats); i++ {
		match_clause += fmt.Sprintf(" AND %s", match_cats[i])
	}
	match_clause += `'`

	// build NOT IN clause
	for i := range cats {
		cats[i] = "'" + cats[i] + "'"
	}
	not_in_clause := strings.Join(cats, ", ")

	t.Text = strings.Replace(
		t.Text,
		"WHERE global_cats != ''",
		fmt.Sprintf(
			`WHERE global_cats != ''
	AND global_cats NOT IN (%s)
	AND id IN (
		SELECT link_id 
		FROM global_cats_fts
		%s
			)`,
			not_in_clause,
			match_clause,
		),
		1)

	return t
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

// Global Cats Spellfix Matches For Snippet
type SpellfixMatches struct {
	Query
}

func NewSpellfixMatchesForSnippet(snippet string) *SpellfixMatches {

	// oddly, "WHERE word MATCH '%s OR %s*' doesn't work here
	return (&SpellfixMatches{
		Query: Query{
			Text: fmt.Sprintf(
				`SELECT word, rank
				FROM global_cats_spellfix
				WHERE 
					(word MATCH '%[1]s*'
					OR word MATCH '%[1]s')
				AND (score / rank) <= 100
				ORDER BY (score / rank)
				LIMIT %[2]d;`,
				snippet,
				SPELLFIX_MATCHES_LIMIT,
			),
		},
	})
}