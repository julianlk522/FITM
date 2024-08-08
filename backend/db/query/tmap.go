package db

import (
	"fmt"
	"strings"
)

// PROFILE
func NewTmapProfile(login_name string) string {
	return fmt.Sprintf(`SELECT 
	login_name, 
	COALESCE(about,"") as about, 
	COALESCE(pfp,"") as pfp, 
	created
FROM Users 
WHERE login_name = '%s';`, login_name)
}

// LINKS
const BASE_FIELDS = `SELECT 
	Links.id as link_id, 
	url, 
	submitted_by as login_name, 
	submit_date, 
	cats, 
	0 as cats_from_user,
	COALESCE(global_summary,"") as summary, 
	COALESCE(summary_count,0) as summary_count, 
	COALESCE(like_count,0) as like_count, 
	tag_count, 
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
type TmapSubmitted struct {
	Query
}

func NewTmapSubmitted(login_name string) *TmapSubmitted {	
	q := &TmapSubmitted{Query: Query{Text: BASE_FIELDS + SUBMITTED_FROM + SUBMITTED_WHERE + BASE_ORDER}}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

const SUBMITTED_FROM = ` FROM Links
JOIN
	(
	SELECT categories as cats, link_id as tag_link_id
	FROM Tags
	WHERE submitted_by = 'LOGIN_NAME'
	)
ON tag_link_id = link_id
LEFT JOIN
	(
	SELECT count(*) as tag_count, link_id as tag_link_id2
	FROM Tags
	GROUP BY tag_link_id2
	)
ON tag_link_id2 = link_id
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

func (q *TmapSubmitted) FromCategories(categories []string) *TmapSubmitted {
	var cat_clause string
	for _, cat := range categories {
		cat_clause += fmt.Sprintf(` 
		AND ',' || cats || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text = strings.Replace(q.Text, BASE_ORDER, cat_clause + BASE_ORDER, 1)

	return q
}

var SUBMITTED_FROM_LINES = strings.Split(SUBMITTED_FROM, "\n")
var SUBMITTED_FROM_LAST_LINE = SUBMITTED_FROM_LINES[len(SUBMITTED_FROM_LINES)-1]

func (q *TmapSubmitted) AsSignedInUser(req_user_id string, req_login_name string) *TmapSubmitted {
	
	// 2 replacers required: cannot be achieved with 1 since REQ_USER_ID/REQ_LOGIN_NAME replacements must be applied to auth fields/from after they are inserted
	fields_replacer := strings.NewReplacer(BASE_FIELDS, BASE_FIELDS + AUTH_FIELDS, SUBMITTED_FROM_LAST_LINE, SUBMITTED_FROM_LAST_LINE + AUTH_FROM)
	auth_replacer := strings.NewReplacer("REQ_USER_ID", req_user_id, "REQ_LOGIN_NAME", req_login_name)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}



// Copied links submitted by other users (global cats replaced with user-assigned if user has tagged)
type TmapCopied struct {
	Query
}

func NewTmapCopied(login_name string) *TmapCopied {
	q := &TmapCopied{Query: Query{Text: COPIED_FIELDS + COPIED_FROM + COPIED_WHERE + BASE_ORDER}}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

var COPIED_FIELDS = strings.Replace(
	strings.Replace(
		BASE_FIELDS, 
		"cats", 
		"COALESCE(user_cats,global_cats) as cats", 
		1), 
	"0 as cats_from_user", 
	`COALESCE(cats_from_user,0) as cats_from_user`, 
	1)

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
	SELECT categories as user_cats, categories IS NOT NULL as cats_from_user, link_id as tag_link_id
	FROM Tags
	WHERE submitted_by = 'LOGIN_NAME'
	)
ON tag_link_id = link_id
LEFT JOIN
	(
	SELECT count(*) as tag_count, link_id as tag_link_id2
	FROM Tags
	GROUP BY tag_link_id2
	)
ON tag_link_id2 = link_id
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

func (q *TmapCopied) FromCategories(categories []string) *TmapCopied {
	var cat_clause string

	for _, cat := range categories {
		cat_clause += fmt.Sprintf(` 
		AND ',' || cats || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text = strings.Replace(q.Text, BASE_ORDER, cat_clause + BASE_ORDER, 1)
	return q
}

func (q *TmapCopied) ForUser(login_name string) *TmapCopied {
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

var COPIED_FROM_LINES = strings.Split(COPIED_FROM, "\n")
var COPIED_FROM_LAST_LINE = COPIED_FROM_LINES[len(COPIED_FROM_LINES)-1]

func (q *TmapCopied) AsSignedInUser(req_user_id string, req_login_name string) *TmapCopied {
	fields_replacer := strings.NewReplacer(COPIED_FIELDS, COPIED_FIELDS + AUTH_FIELDS, COPIED_FROM_LAST_LINE, COPIED_FROM_LAST_LINE + AUTH_FROM)
	auth_replacer := strings.NewReplacer("REQ_USER_ID", req_user_id, "REQ_LOGIN_NAME", req_login_name)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}



// Tagged links submitted by other users (global cats replaced with user-assigned)
type TmapTagged struct {
	Query
}

func NewTmapTagged(login_name string) *TmapTagged {
	q := &TmapTagged{Query: Query{Text: BASE_FIELDS + TAGGED_FROM + TAGGED_WHERE + BASE_ORDER}}
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

func (q *TmapTagged) FromCategories(categories []string) *TmapTagged {
	var cat_clause string
	for _, cat := range categories {
		cat_clause += fmt.Sprintf(` 
		AND ',' || cats || ',' LIKE '%%,%s,%%'`, cat)
	}

	q.Text = strings.Replace(q.Text, BASE_ORDER, cat_clause + BASE_ORDER, 1)

	return q
}

func (q *TmapTagged) ForUser(login_name string) *TmapTagged {
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

var TAGGED_FROM_LAST_LINE = SUBMITTED_FROM_LAST_LINE

func (q *TmapTagged) AsSignedInUser(req_user_id string, req_login_name string) *TmapTagged {
	fields_replacer := strings.NewReplacer(BASE_FIELDS, BASE_FIELDS + AUTH_FIELDS, TAGGED_FROM_LAST_LINE, TAGGED_FROM_LAST_LINE + AUTH_FROM)
	auth_replacer := strings.NewReplacer("REQ_USER_ID", req_user_id, "REQ_LOGIN_NAME", req_login_name)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}