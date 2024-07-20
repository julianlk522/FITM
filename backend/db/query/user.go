package handler

import (
	"fmt"
	"strings"
)

// User Treasure Map Links
const BASE_FIELDS = `SELECT 
	Links.id as link_id, 
	url, 
	submitted_by as login_name, 
	submit_date, 
	categories, 
	COALESCE(global_summary,"") as summary, 
	COALESCE(summary_count,0) as summary_count, 
	COALESCE(like_count,0) as like_count, 
	COALESCE(img_url,"") as img_url
`

// Authenticated: add IsLiked, IsCopied, IsTagged
const AUTH_FIELDS = `, 
COALESCE(is_liked,0) as is_liked, 
COALESCE(is_tagged,0) as is_tagged,
COALESCE(is_copied,0) as is_copied`

const AUTH_FROM = ` LEFT JOIN
	(
	SELECT id, count(*) as is_liked, user_id, link_id as like_link_id2
	FROM 'Link Likes'
	WHERE user_id = 'REQ_USER_ID'
	GROUP BY id
	)
ON like_link_id2 = link_id 
LEFT JOIN 
(
	SELECT id as tag_id, link_id as tlink_id, count(*) as is_tagged 
	FROM Tags
	WHERE Tags.submitted_by = 'REQ_LOGIN_NAME'
	GROUP BY tag_id
)
ON tlink_id = link_id
LEFT JOIN
	(
	SELECT id as copy_id, count(*) as is_copied, user_id as cuser_id, link_id as clink_id
	FROM 'Link Copies'
	WHERE cuser_id = 'REQ_USER_ID'
	GROUP BY copy_id
	)
ON clink_id = link_id`



// Submitted links (global categories replaced with user-assigned)
type GetTmapSubmitted struct {
	Query
}

func NewGetTmapSubmitted(req_user_id string, req_login_name string) *GetTmapSubmitted {
	var sql string
	
	if req_user_id != "" {
		req_user_auth_from := strings.Replace(AUTH_FROM, "REQ_LOGIN_NAME", req_login_name, 1)
		sql = BASE_FIELDS + AUTH_FIELDS + SUBMITTED_FROM + req_user_auth_from + SUBMITTED_WHERE

		sql = strings.ReplaceAll(sql, "REQ_USER_ID", req_user_id)
	} else {
		sql = BASE_FIELDS + SUBMITTED_FROM + SUBMITTED_WHERE
	}
	
	return &GetTmapSubmitted{Query: Query{Text: sql}}
}

const SUBMITTED_FROM = ` FROM Links
JOIN
	(
	SELECT categories, link_id as tag_link_id
	FROM Tags
	WHERE submitted_by = 'LOGIN_NAME'
	)
ON link_id = tag_link_id
LEFT JOIN
	(
	SELECT count(*) as like_count, link_id as like_link_id
	FROM 'Link Likes'
	GROUP BY link_id
	)
ON like_link_id = link_id
LEFT JOIN
	(
	SELECT count(*) as summary_count, link_id as summary_link_id
	FROM Summaries
	GROUP BY link_id
	)
ON summary_link_id = link_id`

const SUBMITTED_WHERE = ` WHERE submitted_by = 'LOGIN_NAME'`

func (q *GetTmapSubmitted) FromCategories(categories []string) *GetTmapSubmitted {
	var cat_clause string
	for _, cat := range categories {
		cat_clause += fmt.Sprintf(` AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text += cat_clause

	return q
}

func (q *GetTmapSubmitted) ForUser(login_name string) *GetTmapSubmitted {
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)
	q.Text += ";"

	return q
}



// Tagged links submitted by other users (global categories replaced with user-assigned)
type GetTmapTagged struct {
	Query
}

func NewGetTmapTagged(req_user_id string, req_login_name string) *GetTmapTagged {
	var sql string

	if req_user_id != "" {
		req_user_auth_from := strings.Replace(AUTH_FROM, "REQ_LOGIN_NAME", req_login_name, 1)
		sql = BASE_FIELDS + AUTH_FIELDS + TAGGED_FROM + req_user_auth_from + TAGGED_WHERE

		sql = strings.ReplaceAll(sql, "REQ_USER_ID", req_user_id)
	} else {
		sql = BASE_FIELDS + TAGGED_FROM + TAGGED_WHERE
	}	

	return &GetTmapTagged{Query: Query{Text: sql}}
}	

const TAGGED_FROM = SUBMITTED_FROM
const TAGGED_WHERE = ` WHERE submitted_by != 'LOGIN_NAME'`

func (q *GetTmapTagged) FromCategories(categories []string) *GetTmapTagged {
	var and_clause string
	for _, cat := range categories {
		and_clause += fmt.Sprintf(` AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text += and_clause

	return q
}

func (q *GetTmapTagged) ForUser(login_name string) *GetTmapTagged {
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)
	q.Text += ";"

	return q
}



// Copied links submitted by other users
type GetTmapCopied struct {
	Query
}

func NewGetTmapCopied(req_user_id string, req_login_name string) *GetTmapCopied {
	var sql string

	if req_user_id != "" {
		req_user_auth_from := strings.Replace(AUTH_FROM, "REQ_LOGIN_NAME", req_login_name, 1)
		sql = COPIED_FIELDS + AUTH_FIELDS + COPIED_FROM + req_user_auth_from + COPIED_WHERE

		sql = strings.ReplaceAll(sql, "REQ_USER_ID", req_user_id)
	} else {
		sql = COPIED_FIELDS + COPIED_FROM + COPIED_WHERE
	}	

	return &GetTmapCopied{Query: Query{Text: sql}}
}

var COPIED_FIELDS = strings.Replace(BASE_FIELDS, "categories", `COALESCE(global_cats,"") as categories`, 1)

const COPIED_FROM = ` FROM Links
JOIN
	(
	SELECT link_id as copy_link_id, user_id as copier_id
	FROM 'Link Copies'
	JOIN Users
	ON Users.id = copier_id
	WHERE Users.login_name = 'LOGIN_NAME'
	)
ON copy_link_id = link_id
LEFT JOIN
	(
	SELECT count(*) as like_count, link_id as like_link_id
	FROM 'Link Likes'
	GROUP BY link_id
	)
ON like_link_id = link_id
LEFT JOIN
	(
	SELECT count(*) as summary_count, link_id as summary_link_id
	FROM Summaries
	GROUP BY link_id
	)
ON summary_link_id = link_id`

const COPIED_WHERE = ` WHERE link_id NOT IN
	(
	SELECT link_id
	FROM TAGS
	WHERE submitted_by = 'LOGIN_NAME'
	)`

func (q *GetTmapCopied) FromCategories(categories []string) *GetTmapCopied {
	var cat_clause string

	for _, cat := range categories {
		cat_clause += fmt.Sprintf(` AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text += cat_clause

	return q
}

func (q *GetTmapCopied) ForUser(login_name string) *GetTmapCopied {
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)
	q.Text += ";"

	return q
}