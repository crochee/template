package storage

import (
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Status string

func (Status) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{StatusQueryClause{Field: f}}
}

type StatusQueryClause struct {
	Field *schema.Field
}

func (StatusQueryClause) Name() string {
	return ""
}

func (StatusQueryClause) Build(clause.Builder) {
}

func (StatusQueryClause) MergeClause(*clause.Clause) {
}

func (StatusQueryClause) ModifyStatement(stmt *gorm.Statement) {
	if _, ok := stmt.Clauses["soft_delete_enabled"]; ok || stmt.Statement.Unscoped {
		return
	}
	if c, ok := stmt.Clauses["WHERE"]; ok {
		if where, ok := c.Expression.(clause.Where); ok && len(where.Exprs) > 1 {
			for _, expr := range where.Exprs {
				if orCond, ok := expr.(clause.OrConditions); ok && len(orCond.Exprs) == 1 {
					where.Exprs = []clause.Expression{clause.And(where.Exprs...)}
					c.Expression = where
					stmt.Clauses["WHERE"] = c
					break
				}
			}
		}
	}
	stmt.Clauses["soft_delete_enabled"] = clause.Clause{}
}

func (Status) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{StatusDeleteDeleteClause{Field: f}}
}

type StatusDeleteDeleteClause struct {
	Field *schema.Field
}

func (StatusDeleteDeleteClause) Name() string {
	return ""
}

func (StatusDeleteDeleteClause) Build(clause.Builder) {
}

func (StatusDeleteDeleteClause) MergeClause(*clause.Clause) {
}

func (s StatusDeleteDeleteClause) ModifyStatement(stmt *gorm.Statement) {
	if stmt.SQL.Len() != 0 || stmt.Statement.Unscoped {
		return
	}

	var clauseSet clause.Set
	curTime := stmt.NowFunc()
	if !strings.Contains(s.Field.Comment, "skip_delete") {
		clauseSet = append(clauseSet, clause.Assignment{Column: clause.Column{Name: s.Field.DBName}, Value: "deleted"})
	}

	if _, ok := stmt.Schema.FieldsByName["Deleted"]; ok {
		assignment := clause.Assignment{
			Column: clause.Column{Name: "deleted"},
			Value:  gorm.Expr("`id`"),
		}
		_, idOk := stmt.Schema.FieldsByDBName["id"]
		if !idOk {
			assignment.Value = uint64(curTime.UnixNano())
			return
		}
		clauseSet = append(clauseSet, assignment)
	}
	if _, ok := stmt.Schema.FieldsByName["DeletedAt"]; ok {
		clauseSet = append(clauseSet, clause.Assignment{Column: clause.Column{Name: "deleted_at"}, Value: curTime})
	}

	stmt.AddClause(clauseSet)
	if stmt.Schema != nil {
		_, queryValues := schema.GetIdentityFieldValuesMap(stmt.Context, stmt.ReflectValue, stmt.Schema.PrimaryFields)
		column, values := schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)

		if len(values) > 0 {
			stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
		}

		if stmt.ReflectValue.CanAddr() && stmt.Dest != stmt.Model && stmt.Model != nil {
			_, queryValues = schema.GetIdentityFieldValuesMap(stmt.Context, reflect.ValueOf(stmt.Model), stmt.Schema.PrimaryFields)
			column, values = schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)

			if len(values) > 0 {
				stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
			}
		}
	}

	if _, ok := stmt.Clauses["WHERE"]; !stmt.AllowGlobalUpdate && !ok {
		_ = stmt.AddError(gorm.ErrMissingWhereClause)
	} else {
		StatusQueryClause(s).ModifyStatement(stmt)
	}

	stmt.AddClauseIfNotExists(clause.Update{})
	stmt.Build("UPDATE", "SET", "WHERE")
}
