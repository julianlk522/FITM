package query

import (
	"fmt"
	"regexp"
	"strings"
)

// PROFILE
func NewTmapProfile(login_name string) string {
	return fmt.Sprintf(`SELECT 
	login_name, 
	COALESCE(about,'') as about, 
	COALESCE(pfp,'') as pfp, 
	created
FROM Users 
WHERE login_name = '%s';`, login_name)
}

// LINKS
const BASE_CTES = `SummaryCount AS (
    SELECT link_id, COUNT(*) AS summary_count
    FROM Summaries
    GROUP BY link_id
),
LikeCount AS (
    SELECT link_id, COUNT(*) AS like_count
    FROM 'Link Likes'
    GROUP BY link_id
),
TagCount AS (
    SELECT link_id, COUNT(*) AS tag_count
    FROM Tags
    GROUP BY link_id
)`

const USER_CATS_CTE = `UserCats AS (
    SELECT link_id, cats as user_cats
    FROM user_cats_fts
    WHERE submitted_by = 'LOGIN_NAME'
)`

const BASE_FIELDS = `
SELECT 
	l.id AS link_id,
    l.url,
    l.submitted_by AS login_name,
    l.submit_date,
    uc.user_cats,
    0 AS cats_from_user,
    COALESCE(l.global_summary, '') AS summary,
    COALESCE(sc.summary_count, 0) AS summary_count,
    COALESCE(lc.like_count, 0) AS like_count,
    COALESCE(tc.tag_count, 0) AS tag_count,
    COALESCE(l.img_url, '') AS img_url`

const BASE_FROM = `
FROM Links l
INNER JOIN UserCats uc ON l.id = uc.link_id
LEFT JOIN TagCount tc ON l.id = tc.link_id
LEFT JOIN LikeCount lc ON l.id = lc.link_id
LEFT JOIN SummaryCount sc ON l.id = sc.link_id`

const ORDER = `
ORDER BY lc.like_count DESC, sc.summary_count DESC, l.id DESC;`

// Submitted links (global cats replaced with user-assigned)
type TmapSubmitted struct {
	Query
}

func NewTmapSubmitted(login_name string) *TmapSubmitted {
	q := &TmapSubmitted{
		Query: Query{
			Text: 
			"WITH " + BASE_CTES + ",\n" +
			USER_CATS_CTE +
			BASE_FIELDS + 
			BASE_FROM + "\n" +
			SUBMITTED_WHERE + 
			ORDER,
		},
	}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

const SUBMITTED_WHERE = `WHERE l.submitted_by = 'LOGIN_NAME'`

func (q *TmapSubmitted) FromCats(cats []string) *TmapSubmitted {

	// escape any "." in cats
	for i := 0; i < len(cats); i++ {
		if strings.Contains(cats[i], ".") {
			cats[i] = strings.Replace(cats[i], `.`, `"."`, 1)
		}
	}

	var cat_clause string
	for _, cat := range cats {
		cat_clause += fmt.Sprintf("\nAND uc.user_cats MATCH '%s'", cat)
	}

	// find line with "WHERE l.submitted_by = 'xxx'"
	// append after
	sw_line := regexp.MustCompile(`WHERE l.submitted_by = '.+'`).FindString(q.Text)

	q.Text = strings.Replace(
		q.Text,
		sw_line,
		sw_line + cat_clause,
		1,
	)

	return q
}

func (q *TmapSubmitted) AsSignedInUser(req_user_id string, req_login_name string) *TmapSubmitted {

	// 2 replacers required: cannot be achieved with 1 since REQ_USER_ID/REQ_LOGIN_NAME replacements must be applied to auth fields/from _after_ they are inserted
	fields_replacer := strings.NewReplacer(
		BASE_CTES, BASE_CTES + ",\n" + AUTH_CTES,
		BASE_FIELDS, BASE_FIELDS + AUTH_FIELDS,
		BASE_FROM, BASE_FROM + AUTH_JOIN,
	)
	auth_replacer := strings.NewReplacer(
		"REQ_USER_ID", req_user_id, 
		"REQ_LOGIN_NAME", req_login_name,
	)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}

const AUTH_CTES = `IsLiked AS (
	SELECT link_id, COUNT(*) AS is_liked
	FROM 'Link Likes'
	WHERE user_id = 'REQ_USER_ID'
	GROUP BY id
),
IsCopied AS (
	SELECT link_id, COUNT(*) AS is_copied
	FROM 'Link Copies'
	WHERE user_id = 'REQ_USER_ID'
	GROUP BY id
)`

const AUTH_FIELDS = `, 
	COALESCE(is_liked,0) as is_liked, 
	COALESCE(is_copied,0) as is_copied`

const AUTH_JOIN = `
LEFT JOIN IsLiked il ON l.id = il.link_id
LEFT JOIN IsCopied ic ON l.id = ic.link_id`

// Copied links submitted by other users (global cats replaced with user-assigned if user has tagged)
type TmapCopied struct {
	Query
}

func NewTmapCopied(login_name string) *TmapCopied {
	q := &TmapCopied{
		Query: Query{
			Text: 
				"WITH " + USER_COPIES_CTE + ",\n" +
				POSSIBLE_USER_CATS_CTE + ",\n" +
				BASE_CTES +
				COPIED_FIELDS + 
				COPIED_FROM +
				COPIED_WHERE + 
				ORDER,
		},
	}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

const USER_COPIES_CTE = `UserCopies AS (
    SELECT lc.link_id
    FROM 'Link Copies' lc
    INNER JOIN Users u ON u.id = lc.user_id
    WHERE u.login_name = 'LOGIN_NAME'
)`

const POSSIBLE_USER_CATS_CTE = `PossibleUserCats AS (
    SELECT 
		link_id, 
		cats AS user_cats,
		(cats IS NOT NULL) AS cats_from_user
    FROM user_cats_fts
    WHERE submitted_by = 'LOGIN_NAME'
)`

var COPIED_FIELDS = strings.Replace(
	strings.Replace(
		BASE_FIELDS,
		"uc.user_cats", "COALESCE(puc.user_cats, l.global_cats) AS cats",
		1,
	), 
	"0 AS cats_from_user", `COALESCE(puc.cats_from_user,0) AS cats_from_user`,
	1,
)

var COPIED_FROM = strings.Replace(
	BASE_FROM,
	"INNER JOIN UserCats uc ON l.id = uc.link_id",
	COPIED_JOIN,
	1,
)

const COPIED_JOIN = `INNER JOIN UserCopies uc ON l.id = uc.link_id
LEFT JOIN PossibleUserCats puc ON l.id = puc.link_id`

const COPIED_WHERE = ` 
WHERE submitted_by != 'LOGIN_NAME'`

func (q *TmapCopied) FromCats(cats []string) *TmapCopied {

	// escape any "." in cats
	for i := 0; i < len(cats); i++ {
		if strings.Contains(cats[i], ".") {
			cats[i] = strings.Replace(cats[i], `.`, `"."`, 1)
		}
	}

	var cats_clause string
	for i := range cats {
		cats_clause += fmt.Sprintf(
			"\nAND cats MATCH '%s'", cats[i],
		)
	}

	q.Text = strings.Replace(
		q.Text, 
		ORDER, 
		cats_clause + ORDER,  
		1,
	)

	return q
}

func (q *TmapCopied) AsSignedInUser(req_user_id string, req_login_name string) *TmapCopied {
	fields_replacer := strings.NewReplacer(
		BASE_CTES, BASE_CTES + ",\n" + AUTH_CTES,
		COPIED_FIELDS, COPIED_FIELDS + AUTH_FIELDS, 
		COPIED_FROM, COPIED_FROM + AUTH_JOIN,
	)
	auth_replacer := strings.NewReplacer(
		"REQ_USER_ID", req_user_id, 
		"REQ_LOGIN_NAME", req_login_name,
	)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}

// Tagged links submitted by other users (global cats replaced with user-assigned)
type TmapTagged struct {
	Query
}

func NewTmapTagged(login_name string) *TmapTagged {
	q := &TmapTagged{
		Query: Query{
			Text: 
				"WITH " + USER_COPIES_CTE + ",\n" +
				USER_CATS_CTE + ",\n" +
				BASE_CTES +
				BASE_FIELDS + 
				BASE_FROM +
				TAGGED_WHERE + 
				ORDER,
		},
	}

	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)
	return q
}

const TAGGED_WHERE = ` WHERE submitted_by != 'LOGIN_NAME'
	AND l.id NOT IN
		(SELECT link_id FROM UserCopies)`

func (q *TmapTagged) FromCats(cats []string) *TmapTagged {

	// escape any "." in cats
	for i := 0; i < len(cats); i++ {
		if strings.Contains(cats[i], ".") {
			cats[i] = strings.Replace(cats[i], `.`, `"."`, 1)
		}
	}

	var cat_clause string
	for _, cat := range cats {
		cat_clause += fmt.Sprintf(
			"\nAND uc.user_cats MATCH '%s'", cat)
	}

	q.Text = strings.Replace(
		q.Text, 
		ORDER, 
		cat_clause + ORDER, 
		1,
	)

	return q
}

func (q *TmapTagged) AsSignedInUser(req_user_id string, req_login_name string) *TmapTagged {
	fields_replacer := strings.NewReplacer(
		BASE_CTES, BASE_CTES + ",\n" + AUTH_CTES,
		BASE_FIELDS, BASE_FIELDS + AUTH_FIELDS, 
		BASE_FROM, BASE_FROM + AUTH_JOIN,
	)
	auth_replacer := strings.NewReplacer(
		"REQ_USER_ID", req_user_id, 
		"REQ_LOGIN_NAME", req_login_name,
	)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}
