package handler

import (
	"errors"
	"fmt"
)

type Query struct {
	Text string
	Error error
}

func GetPeriodClause(period string) (clause string, err error) {
	var days int
	switch period {
		case "day":
			days = 1
		case "week":
			days = 7
		case "month":
			days = 30
		case "year":
			days = 365
		default:
			return "", errors.New("invalid period")
	}

	return fmt.Sprintf("julianday('now') - julianday(submit_date) < %d", days), nil
}

func WithPeriodClause(sql string, period string) (string) {
	clause, err := GetPeriodClause(period)
	if err != nil {
		return sql
	}
	return sql + fmt.Sprintf(" WHERE %s", clause)
}