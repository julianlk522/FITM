package db

import (
	"fmt"
	"strings"
)

// PROFILE
func NewGetTmapProfile(login_name string) string {
	return fmt.Sprintf(`SELECT login_name, COALESCE(about,"") as about, COALESCE(pfp,"") as pfp, COALESCE(created,"") as created 
	FROM Users 
	WHERE login_name = '%s';`, login_name)
}

// LINKS
const BASE_FIELDS = `SELECT 
	Links.id as link_id, 
	url, 
	submitted_by as login_name, 
	submit_date, 
	categories, 
	0 as cats_from_user,
	COALESCE(global_summary,"") as summary, 
	COALESCE(summary_count,0) as summary_count, 
	COALESCE(like_count,0) as like_count, 
	COALESCE(img_url,"") as img_url`

const BASE_ORDER = ` 
ORDER BY like_count DESC, summary_count DESC, link_id DESC;`

// Authenticated: add IsLiked, IsCopied, IsTagged
const AUTH_FIELDS = `, 
	COALESCE(is_liked,0) as is_liked, 
	COALESCE(is_tagged,0) as is_tagged,
	COALESCE(is_copied,0) as is_copied`

const AUTH_FROM = ` 
LEFT JOIN
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



// Submitted links (global cats replaced with user-assigned)
type GetTmapSubmitted struct {
	Query
}

func NewGetTmapSubmitted(login_name string) *GetTmapSubmitted {	
	q := &GetTmapSubmitted{Query: Query{Text: BASE_FIELDS + SUBMITTED_FROM + SUBMITTED_WHERE + BASE_ORDER}}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
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

const SUBMITTED_WHERE = ` 
WHERE submitted_by = 'LOGIN_NAME'`

func (q *GetTmapSubmitted) FromCategories(categories []string) *GetTmapSubmitted {
	var cat_clause string
	for _, cat := range categories {
		cat_clause += fmt.Sprintf(` 
		AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text = strings.Replace(q.Text, BASE_ORDER, cat_clause + BASE_ORDER, 1)

	return q
}

var SUBMITTED_FROM_LINES = strings.Split(SUBMITTED_FROM, "\n")
var SUBMITTED_FROM_LAST_LINE = SUBMITTED_FROM_LINES[len(SUBMITTED_FROM_LINES)-1]

func (q *GetTmapSubmitted) AsSignedInUser(req_user_id string, req_login_name string) *GetTmapSubmitted {
	
	// 2 replacer required: cannot be achieved with 1 since REQ_USER_ID/REQ_LOGIN_NAME replacements must be applied to auth fields/from after they are inserted
	fields_replacer := strings.NewReplacer(BASE_FIELDS, BASE_FIELDS + AUTH_FIELDS, SUBMITTED_FROM_LAST_LINE, SUBMITTED_FROM_LAST_LINE + AUTH_FROM)
	auth_replacer := strings.NewReplacer("REQ_USER_ID", req_user_id, "REQ_LOGIN_NAME", req_login_name)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}



// Copied links submitted by other users (global categories replaced with user-assigned if user has tagged)
type GetTmapCopied struct {
	Query
}

func NewGetTmapCopied(login_name string) *GetTmapCopied {
	q := &GetTmapCopied{Query: Query{Text: COPIED_FIELDS + COPIED_FROM + COPIED_WHERE + BASE_ORDER}}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

var COPIED_FIELDS = strings.Replace(strings.Replace(BASE_FIELDS, "categories", `COALESCE(categories,global_cats) as categories`, 1), "0 as cats_from_user", `COALESCE(cats_from_user,0) as cats_from_user`, 1)

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
	SELECT categories, categories IS NOT NULL as cats_from_user, link_id as tag_link_id
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

const COPIED_WHERE = ` 
WHERE submitted_by != 'LOGIN_NAME'`

func (q *GetTmapCopied) FromCategories(categories []string) *GetTmapCopied {
	var cat_clause string

	for _, cat := range categories {
		cat_clause += fmt.Sprintf(` 
		AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text = strings.Replace(q.Text, BASE_ORDER, cat_clause + BASE_ORDER, 1)
	return q
}

func (q *GetTmapCopied) ForUser(login_name string) *GetTmapCopied {
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

var COPIED_FROM_LINES = strings.Split(COPIED_FROM, "\n")
var COPIED_FROM_LAST_LINE = COPIED_FROM_LINES[len(COPIED_FROM_LINES)-1]

func (q *GetTmapCopied) AsSignedInUser(req_user_id string, req_login_name string) *GetTmapCopied {
	fields_replacer := strings.NewReplacer(COPIED_FIELDS, COPIED_FIELDS + AUTH_FIELDS, COPIED_FROM_LAST_LINE, COPIED_FROM_LAST_LINE + AUTH_FROM)
	auth_replacer := strings.NewReplacer("REQ_USER_ID", req_user_id, "REQ_LOGIN_NAME", req_login_name)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}



// Tagged links submitted by other users (global categories replaced with user-assigned)
type GetTmapTagged struct {
	Query
}

func NewGetTmapTagged(login_name string) *GetTmapTagged {
	q := &GetTmapTagged{Query: Query{Text: BASE_FIELDS + TAGGED_FROM + TAGGED_WHERE + BASE_ORDER}}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}	

const TAGGED_FROM = SUBMITTED_FROM
const TAGGED_WHERE = ` WHERE submitted_by != 'LOGIN_NAME'
	AND link_id NOT IN
		(
		SELECT link_id
		FROM 'Link Copies'
		JOIN Users
		ON Users.id = 'Link Copies'.user_id
		WHERE Users.login_name = 'LOGIN_NAME'
		)`

func (q *GetTmapTagged) FromCategories(categories []string) *GetTmapTagged {
	var cat_clause string
	for _, cat := range categories {
		cat_clause += fmt.Sprintf(` 
		AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text = strings.Replace(q.Text, BASE_ORDER, cat_clause + BASE_ORDER, 1)

	return q
}

func (q *GetTmapTagged) ForUser(login_name string) *GetTmapTagged {
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

var TAGGED_FROM_LAST_LINE = SUBMITTED_FROM_LAST_LINE

func (q *GetTmapTagged) AsSignedInUser(req_user_id string, req_login_name string) *GetTmapTagged {
	fields_replacer := strings.NewReplacer(BASE_FIELDS, BASE_FIELDS + AUTH_FIELDS, TAGGED_FROM_LAST_LINE, TAGGED_FROM_LAST_LINE + AUTH_FROM)
	auth_replacer := strings.NewReplacer("REQ_USER_ID", req_user_id, "REQ_LOGIN_NAME", req_login_name)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}