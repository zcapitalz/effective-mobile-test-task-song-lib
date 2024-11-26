package repos

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func inConditionWithSubquery(property string, query sq.SelectBuilder) sq.Sqlizer {
	sql, args, _ := query.ToSql()
	subQuery := fmt.Sprintf("%s IN (%s)", property, sql)
	return sq.Expr(subQuery, args...)
}
