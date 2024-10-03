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
    FROM "Link Likes"
    GROUP BY link_id
),
TagCount AS (
    SELECT link_id, COUNT(*) AS tag_count
    FROM Tags
    GROUP BY link_id
)`

const POSSIBLE_USER_CATS_CTE = `PossibleUserCats AS (
    SELECT 
		link_id, 
		cats AS user_cats,
		(cats IS NOT NULL) AS cats_from_user
    FROM user_cats_fts
    WHERE submitted_by = 'LOGIN_NAME'
)`

const POSSIBLE_USER_SUMMARY_CTE = `PossibleUserSummary AS (
    SELECT
        link_id, text as user_summary
    FROM Summaries
    INNER JOIN Users u ON u.id = submitted_by
	WHERE u.login_name = 'LOGIN_NAME'
)`

const GLOBAL_CATS_CTE = `GlobalCatsFTS AS (
    SELECT
        link_id,
        global_cats
    FROM global_cats_fts
)`

const USER_COPIES_CTE = `UserCopies AS (
    SELECT lc.link_id
    FROM "Link Copies" lc
    INNER JOIN Users u ON u.id = lc.user_id
    WHERE u.login_name = 'LOGIN_NAME'
)`

const AUTH_CTES = `IsLiked AS (
	SELECT link_id, COUNT(*) AS is_liked
	FROM "Link Likes"
	WHERE user_id = 'REQ_USER_ID'
	GROUP BY id
),
IsCopied AS (
	SELECT link_id, COUNT(*) AS is_copied
	FROM "Link Copies"
	WHERE user_id = 'REQ_USER_ID'
	GROUP BY id
)`

const BASE_FIELDS = `
SELECT 
	l.id AS link_id,
    l.url,
    l.submitted_by AS login_name,
    l.submit_date,
    COALESCE(puc.user_cats, l.global_cats) AS cats,
    COALESCE(puc.cats_from_user,0) AS cats_from_user,
    COALESCE(pus.user_summary, l.global_summary, '') AS summary,
    COALESCE(sc.summary_count, 0) AS summary_count,
    COALESCE(lc.like_count, 0) AS like_count,
    COALESCE(tc.tag_count, 0) AS tag_count,
    COALESCE(l.img_url, '') AS img_url`

const AUTH_FIELDS = `, 
	COALESCE(is_liked,0) as is_liked, 
	COALESCE(is_copied,0) as is_copied`

const FROM = `FROM Links l`

const BASE_JOINS = `
LEFT JOIN PossibleUserCats puc ON l.id = puc.link_id
LEFT JOIN PossibleUserSummary pus ON l.id = pus.link_id
LEFT JOIN TagCount tc ON l.id = tc.link_id
LEFT JOIN LikeCount lc ON l.id = lc.link_id
LEFT JOIN SummaryCount sc ON l.id = sc.link_id`

const GLOBAL_CATS_JOIN = "LEFT JOIN GlobalCatsFTS gc ON l.id = gc.link_id"

const COPIED_JOIN = "INNER JOIN UserCopies uc ON l.id = uc.link_id"

const AUTH_JOINS = `
LEFT JOIN IsLiked il ON l.id = il.link_id
LEFT JOIN IsCopied ic ON l.id = ic.link_id`

const NO_NSFW_CATS_WHERE = `
WHERE l.id NOT IN (
	SELECT link_id FROM global_cats_fts WHERE global_cats MATCH 'NSFW'
)`

const ORDER = `
ORDER BY lc.like_count DESC, sc.summary_count DESC, l.id DESC;`

// Submitted links (global cats replaced with user-assigned)
type TmapSubmitted struct {
	Query
}

func NewTmapSubmitted(login_name string) *TmapSubmitted {
	q := &TmapSubmitted{
		Query: Query{
			Text: "WITH " + BASE_CTES + ",\n" +
				POSSIBLE_USER_CATS_CTE + ",\n" +
				POSSIBLE_USER_SUMMARY_CTE +
				BASE_FIELDS + "\n" +
				FROM +
				BASE_JOINS +
				NO_NSFW_CATS_WHERE +
				SUBMITTED_WHERE +
				ORDER,
		},
	}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

const SUBMITTED_WHERE = `
AND l.submitted_by = 'LOGIN_NAME'`

func (q *TmapSubmitted) FromCats(cats []string) *TmapSubmitted {
	q.Text = FromUserOrGlobalCats(q.Text, cats)
	return q
}

func (q *TmapSubmitted) AsSignedInUser(req_user_id string, req_login_name string) *TmapSubmitted {

	// 2 replacers required: cannot be achieved with 1 since REQ_USER_ID/REQ_LOGIN_NAME replacements must be applied to auth fields/from _after_ they are inserted
	fields_replacer := strings.NewReplacer(
		BASE_CTES, BASE_CTES+",\n"+AUTH_CTES,
		BASE_FIELDS, BASE_FIELDS+AUTH_FIELDS,
		BASE_JOINS, BASE_JOINS+AUTH_JOINS,
	)
	auth_replacer := strings.NewReplacer(
		"REQ_USER_ID", req_user_id,
		"REQ_LOGIN_NAME", req_login_name,
	)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}

func (q *TmapSubmitted) NSFW() *TmapSubmitted {

	// remove NSFW clause
	q.Text = strings.Replace(
		q.Text,
		NO_NSFW_CATS_WHERE,
		"",
		1,
	)

	// swap AND to WHERE in WHERE clause
	q.Text = strings.Replace(
		q.Text,
		"AND l.submitted_by",
		"WHERE l.submitted_by",
		1,
	)
	return q
}

// Copied links submitted by other users (global cats replaced with user-assigned if user has tagged)
type TmapCopied struct {
	Query
}

func NewTmapCopied(login_name string) *TmapCopied {
	q := &TmapCopied{
		Query: Query{
			Text: "WITH " + USER_COPIES_CTE + ",\n" +
				BASE_CTES + ",\n" +
				POSSIBLE_USER_CATS_CTE + ",\n" +
				POSSIBLE_USER_SUMMARY_CTE +
				BASE_FIELDS + "\n" +
				FROM + "\n" +
				COPIED_JOIN +
				BASE_JOINS +
				NO_NSFW_CATS_WHERE +
				COPIED_WHERE +
				ORDER,
		},
	}
	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)

	return q
}

const COPIED_WHERE = ` 
AND submitted_by != 'LOGIN_NAME'`

func (q *TmapCopied) FromCats(cats []string) *TmapCopied {
	q.Text = FromUserOrGlobalCats(q.Text, cats)
	return q
}

func (q *TmapCopied) AsSignedInUser(req_user_id string, req_login_name string) *TmapCopied {
	fields_replacer := strings.NewReplacer(
		BASE_CTES, BASE_CTES+",\n"+AUTH_CTES,
		BASE_FIELDS, BASE_FIELDS+AUTH_FIELDS,
		COPIED_JOIN, COPIED_JOIN+AUTH_JOINS,
	)
	auth_replacer := strings.NewReplacer(
		"REQ_USER_ID", req_user_id,
		"REQ_LOGIN_NAME", req_login_name,
	)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}

func (q *TmapCopied) NSFW() *TmapCopied {

	// remove NSFW clause
	q.Text = strings.Replace(
		q.Text,
		NO_NSFW_CATS_WHERE,
		"",
		1,
	)

	// swap AND to WHERE in WHERE clause
	q.Text = strings.Replace(
		q.Text,
		"AND submitted_by !=",
		"WHERE submitted_by !=",
		1,
	)
	return q
}

// Tagged links submitted by other users (global cats replaced with user-assigned)
type TmapTagged struct {
	Query
}

func NewTmapTagged(login_name string) *TmapTagged {
	q := &TmapTagged{
		Query: Query{
			Text: "WITH " + BASE_CTES + ",\n" +
				USER_CATS_CTE + ",\n" +
				POSSIBLE_USER_SUMMARY_CTE + ",\n" +
				USER_COPIES_CTE +
				TAGGED_FIELDS + "\n" +
				FROM +
				TAGGED_JOINS +
				NO_NSFW_CATS_WHERE +
				TAGGED_WHERE +
				ORDER,
		},
	}

	q.Text = strings.ReplaceAll(q.Text, "LOGIN_NAME", login_name)
	return q
}

const USER_CATS_CTE = `UserCats AS (
    SELECT link_id, cats as user_cats
    FROM user_cats_fts
    WHERE submitted_by = 'LOGIN_NAME'
)`

var TAGGED_FIELDS = strings.Replace(
	strings.Replace(
		BASE_FIELDS,
		"COALESCE(puc.user_cats, l.global_cats) AS cats",
		"uct.user_cats",
		1,
	),
	`COALESCE(puc.cats_from_user,0) AS cats_from_user`,
	"1 AS cats_from_user",
	1,
)

var TAGGED_JOINS = strings.Replace(
	BASE_JOINS,
	"LEFT JOIN PossibleUserCats puc ON l.id = puc.link_id",
	"INNER JOIN UserCats uct ON l.id = uct.link_id",
	1,
) + "\n" + strings.Replace(
	COPIED_JOIN,
	"INNER",
	"LEFT",
	1,
)

const TAGGED_WHERE = `
AND submitted_by != 'LOGIN_NAME'
AND l.id NOT IN
	(SELECT link_id FROM UserCopies)`

func (q *TmapTagged) FromCats(cats []string) *TmapTagged {
	escaped := GetCatsWithEscapedChars(cats)
	var cat_clause string
	for _, cat := range escaped {
		cat_clause += fmt.Sprintf(
			"\nAND uct.user_cats MATCH '%s'", cat)
	}

	q.Text = strings.Replace(
		q.Text,
		ORDER,
		cat_clause+ORDER,
		1,
	)
	return q
}

func (q *TmapTagged) AsSignedInUser(req_user_id string, req_login_name string) *TmapTagged {
	fields_replacer := strings.NewReplacer(
		BASE_CTES, BASE_CTES+",\n"+AUTH_CTES,
		TAGGED_FIELDS, TAGGED_FIELDS+AUTH_FIELDS,
		TAGGED_JOINS, TAGGED_JOINS+AUTH_JOINS,
	)
	auth_replacer := strings.NewReplacer(
		"REQ_USER_ID", req_user_id,
		"REQ_LOGIN_NAME", req_login_name,
	)

	q.Text = fields_replacer.Replace(q.Text)
	q.Text = auth_replacer.Replace(q.Text)

	return q
}

func (q *TmapTagged) NSFW() *TmapTagged {

	// remove NSFW clause
	q.Text = strings.Replace(
		q.Text,
		NO_NSFW_CATS_WHERE,
		"",
		1,
	)

	// swap AND to WHERE in WHERE clause
	q.Text = strings.Replace(
		q.Text,
		"AND submitted_by !=",
		"WHERE submitted_by !=",
		1,
	)
	return q
}

func FromUserOrGlobalCats(q string, cats []string) string {
	escaped := GetCatsWithEscapedChars(cats)
	var cat_match string
	cat_match += fmt.Sprintf("'%s", escaped[0])
	for i := 1; i < len(escaped); i++ {
		cat_match += fmt.Sprintf(" AND %s", escaped[i])
	}
	cat_match += "'"

	puc_WHERE := regexp.MustCompile(`WHERE submitted_by = '.+'`).FindString(q)
	q = strings.Replace(
		q,
		puc_WHERE,
		puc_WHERE+"\nAND cats MATCH "+cat_match,
		1,
	)

	gc_CTE := strings.Replace(
		GLOBAL_CATS_CTE,
		"FROM global_cats_fts",
		"FROM global_cats_fts"+"\nWHERE global_cats MATCH "+cat_match,
		1,
	)
	q = strings.Replace(
		q,
		BASE_FIELDS,
		",\n"+gc_CTE+BASE_FIELDS,
		1,
	)

	q = strings.Replace(
		q,
		BASE_JOINS,
		BASE_JOINS+"\n"+GLOBAL_CATS_JOIN,
		1,
	)

	and_clause := `
	AND (
	gc.global_cats IS NOT NULL
	OR
	puc.user_cats IS NOT NULL
)`
	q = strings.Replace(
		q,
		ORDER,
		and_clause+ORDER,
		1,
	)

	return q
}
