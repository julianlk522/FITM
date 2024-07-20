package db

import (
	"fmt"
	"strings"
)

// SUMMARIES FOR LINK
type GetSummaryPageLink struct {
	Query
}

const GET_SUMMARY_PAGE_LINK_BASE_FIELDS = `SELECT links_id as link_id, url, submitted_by, submit_date, coalesce(categories,"") as categories, summary, COUNT('Link Likes'.id) as like_count, img_url`

const GET_SUMMARY_PAGE_LINK_BASE = GET_SUMMARY_PAGE_LINK_BASE_FIELDS + ` 
FROM 
	(
	SELECT id as links_id, url, submitted_by, submit_date, global_cats as categories, global_summary as summary, coalesce(img_url,"") as img_url 
		FROM Links`

func NewGetSummaryPageLink(ID string) *GetSummaryPageLink {
	new := &GetSummaryPageLink{Query: Query{Text: GET_SUMMARY_PAGE_LINK_BASE}}
	return new._FromID(ID)
}

func (l *GetSummaryPageLink) _FromID(ID string) *GetSummaryPageLink {
	l.Text += fmt.Sprintf(` WHERE id = '%s'
	) 
LEFT JOIN 'Link Likes' 
ON 'Link Likes'.link_id = links_id;`, ID)

	return l
}

func (l *GetSummaryPageLink) ForSignedInUser(user_id string ) *GetSummaryPageLink {
	l.Text = strings.Replace(l.Text, GET_SUMMARY_PAGE_LINK_BASE_FIELDS, GET_SUMMARY_PAGE_LINK_BASE_FIELDS + ", COALESCE(is_liked,0) as is_liked, COALESCE(is_copied,0) as is_copied", 1)

	l.Text = strings.Replace(l.Text, ";",
	fmt.Sprintf(` 
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
ON copy_link_id = link_id;`, user_id), 1)

	return l
}

type GetSummaries struct {
	Query
}

const GET_SUMMARIES_BASE_FIELDS = `SELECT sumid, text, sb as submitted_by, last_updated, COALESCE(count(sl.id),0) as like_count`

const GET_SUMMARIES_BASE = GET_SUMMARIES_BASE_FIELDS + ` 
FROM 
	(
	SELECT sumid, text, Users.login_name as sb, last_updated
	FROM 
		(
		SELECT id as sumid, text, submitted_by, last_updated
		FROM Summaries 
		) 
	JOIN Users 
	ON Users.id = submitted_by
	) 
LEFT JOIN 'Summary Likes' as sl 
ON sl.summary_id = sumid 
GROUP BY sumid;`

func NewGetSummaries(link_id string) *GetSummaries {
	new := &GetSummaries{Query: Query{Text: GET_SUMMARIES_BASE}}
	return new._FromID(link_id)
}

func (s *GetSummaries) _FromID(link_id string) *GetSummaries {
	s.Text = strings.Replace(s.Text, "FROM Summaries", fmt.Sprintf(`FROM Summaries 
	WHERE link_id = '%s'`, link_id), 1)

	return s
}

func (s *GetSummaries) ForSignedInUser(user_id string ) *GetSummaries {
	s.Text = strings.Replace(s.Text, GET_SUMMARIES_BASE_FIELDS, GET_SUMMARIES_BASE_FIELDS + ", COALESCE(is_liked,0) as is_liked", 1)

	s.Text = strings.Replace(s.Text, "LEFT JOIN 'Summary Likes' as sl", fmt.Sprintf(`
	LEFT JOIN
		(
		SELECT id, count(*) as is_liked, user_id, summary_id as slsumid
		FROM 'Summary Likes'
		WHERE user_id = '%s'
		GROUP BY id
		)
	ON slsumid = sumid
	LEFT JOIN 'Summary Likes' as sl`, user_id), 1)

	return s
}