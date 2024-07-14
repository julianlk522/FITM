package handler

import (
	"errors"
)

func AppendPeriodClause(sql *string, period string) (err error)  {
	switch period {
		case "day":
			*sql += ` WHERE julianday('now') - julianday(submit_date) <= 2`
		case "week":
			*sql += ` WHERE julianday('now') - julianday(submit_date) <= 8`
		case "month":
			*sql += ` WHERE julianday('now') - julianday(submit_date) <= 31`
		case "year":
			*sql += ` WHERE julianday('now') - julianday(submit_date) <= 366`
		default:
			return errors.New("invalid period")
	}

	return nil
}