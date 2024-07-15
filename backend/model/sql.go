package model

import (
	"errors"
	"fmt"
	"strings"
)

const get_links_base_sql = `SELECT links_id as link_id, url, link_author as submitted_by, sd, categories, summary, coalesce(count(Summaries.id),0) as summary_count, like_count, img_url
FROM 
	(
	SELECT Links.id as links_id, url, submitted_by as link_author, Links.submit_date as sd, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url 
	FROM LINKS 
	LEFT JOIN 
		(
		SELECT link_id as likes_link_id, count(*) as like_count 
		FROM 'Link Likes'
		GROUP BY likes_link_id
		) 
	ON Links.id = likes_link_id
	)
LEFT JOIN Summaries 
ON Summaries.link_id = links_id 
GROUP BY links_id 
ORDER BY like_count DESC, summary_count DESC, link_id DESC;`

type GetLinksSQL struct {
	Text string
	Error error
}

func (l *GetLinksSQL) FromLinkIDs(link_ids []string) *GetLinksSQL {
	link_ids_str := strings.Join(link_ids, ",")

	l._AddWhere(fmt.Sprintf(`WHERE links_id IN (%s)`, link_ids_str))
	return l
}

func (l *GetLinksSQL) AddPeriod(period string) (*GetLinksSQL) {
	var too_many_days int
	switch period {
		case "day":
			too_many_days = 2
		case "week":
			too_many_days = 8
		case "month":
			too_many_days = 31
		case "year":
			too_many_days = 366
		default:
			l.Error = errors.New("invalid period")
			return l
	}

	l._AddWhere(fmt.Sprintf("WHERE julianday('now') - julianday(submit_date) <= %d", too_many_days))
	return l
}

func (l *GetLinksSQL) AddLimit(limit int) *GetLinksSQL {
	l.Text = strings.Replace(l.Text, ";", fmt.Sprintf(" LIMIT %d;", limit), 1)
	return l
}

func (l *GetLinksSQL) _AddWhere(clause string) *GetLinksSQL {

	// Swap previous WHERE for AND
	l.Text = strings.Replace(l.Text, "WHERE", "AND", 1)

	// Prepend new clause
	l.Text = strings.Replace(l.Text, "ON Links.id = likes_link_id", fmt.Sprintf("ON Links.id = likes_link_id %s", clause), 1)

	return l
}

func NewGetLinksSQL() *GetLinksSQL {
	return &GetLinksSQL{
		Text: get_links_base_sql,
	}
}