package query

import (
	"fmt"
	"strings"
)

const SUMMARIES_PAGE_LIMIT = 20

// Summaries Page link
type SummaryPageLink struct {
	Query
}

const SUMMARY_PAGE_LINK_BASE_FIELDS = `SELECT 
links_id as link_id, 
url, 
sb, 
sd, 
cats, 
summary, 
COALESCE(like_count,0) as like_count,
tag_count,   
img_url`

const SUMMARY_PAGE_LINK_BASE = SUMMARY_PAGE_LINK_BASE_FIELDS + ` 
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
	) 
LEFT JOIN
	(
	SELECT count(*) as like_count, link_id as llink_id
	FROM "Link Likes"
	GROUP BY llink_id
	)
ON llink_id = links_id
LEFT JOIN 
	(
	SELECT count(*) as tag_count, link_id as tlink_id
	FROM Tags
	GROUP BY tlink_id
	)
ON tlink_id = links_id;`

func NewSummaryPageLink(ID string) *SummaryPageLink {
	return (&SummaryPageLink{Query: Query{Text: SUMMARY_PAGE_LINK_BASE}})._FromID(ID)
}

func (l *SummaryPageLink) _FromID(ID string) *SummaryPageLink {
	l.Text = strings.Replace(
		l.Text,
		"FROM Links",
		fmt.Sprintf(
			`FROM Links 
			WHERE id = '%s'`,
			ID),
		1)

	return l
}

func (l *SummaryPageLink) AsSignedInUser(user_id string) *SummaryPageLink {
	l.Text = strings.Replace(l.Text, SUMMARY_PAGE_LINK_BASE_FIELDS, SUMMARY_PAGE_LINK_BASE_FIELDS+`, 
		COALESCE(is_liked,0) as is_liked, 
		COALESCE(is_copied,0) as is_copied`, 1)

	l.Text = strings.Replace(l.Text, ";",
		fmt.Sprintf(` 
			LEFT JOIN
				(
				SELECT id as like_id, count(*) as is_liked, user_id as luser_id, link_id as like_link_id2
				FROM "Link Likes"
				WHERE luser_id = '%[1]s'
				GROUP BY like_id
				)
			ON like_link_id2 = link_id
			LEFT JOIN
				(
				SELECT id as copy_id, count(*) as is_copied, user_id as cuser_id, link_id as copy_link_id
				FROM "Link Copies"
				WHERE cuser_id = '%[1]s'
				GROUP BY copy_id
				)
			ON copy_link_id = link_id;`, user_id),
		1)

	return l
}

// Summaries for link
type Summaries struct {
	Query
}

const SUMMARIES_BASE_FIELDS = `SELECT 
	sumid, 
	text, 
	ln, 
	last_updated, 
	COALESCE(count(sl.id),0) as like_count`

const SUMMARIES_FROM = ` 
FROM 
	(
	SELECT sumid, text, Users.login_name as ln, last_updated
	FROM 
		(
		SELECT id as sumid, text, submitted_by as sb, last_updated
		FROM Summaries 
		) 
	JOIN Users 
	ON Users.id = sb
	) 
LEFT JOIN "Summary Likes" as sl 
ON sl.summary_id = sumid 
GROUP BY sumid`

var SUMMARIES_LIMIT = fmt.Sprintf(`
LIMIT %d;`, SUMMARIES_PAGE_LIMIT)

func NewSummariesForLink(link_id string) *Summaries {
	return (&Summaries{
		Query: Query{
			Text: 
				SUMMARIES_BASE_FIELDS +
				SUMMARIES_FROM +
				SUMMARIES_LIMIT,
		},
	})._FromID(link_id)
}

func (s *Summaries) _FromID(link_id string) *Summaries {
	s.Text = strings.Replace(
		s.Text,
		"FROM Summaries",
		fmt.Sprintf(
			`FROM Summaries 
			WHERE link_id = '%s'`,
			link_id),
		1)

	return s
}

func (s *Summaries) AsSignedInUser(user_id string) *Summaries {
	s.Text = strings.Replace(s.Text, SUMMARIES_BASE_FIELDS, SUMMARIES_BASE_FIELDS+`, 
	COALESCE(is_liked,0) as is_liked`, 1)

	s.Text = strings.Replace(s.Text, `LEFT JOIN "Summary Likes" as sl`, fmt.Sprintf(`
	LEFT JOIN
		(
		SELECT id, count(*) as is_liked, user_id, summary_id as slsumid
		FROM "Summary Likes"
		WHERE user_id = '%s'
		GROUP BY id
		)
	ON slsumid = sumid
	LEFT JOIN "Summary Likes" as sl`, user_id), 1)

	return s
}
