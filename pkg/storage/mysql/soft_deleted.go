package mysql

import (
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Deleted uint64

func (d Deleted) UpdateClauses(field *schema.Field) []clause.Interface {
	return []clause.Interface{SoftDeletedQueryClause{Field: field}}
}

type SoftDeletedUpdateClause struct {
	Field *schema.Field
}

func (s SoftDeletedUpdateClause) Name() string {
	return ""
}

func (s SoftDeletedUpdateClause) Build(clause.Builder) {
}

func (s SoftDeletedUpdateClause) MergeClause(*clause.Clause) {
}

func (s SoftDeletedUpdateClause) ModifyStatement(stmt *gorm.Statement) {
	if _, ok := stmt.Clauses["soft_delete_enabled"]; ok {
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
	stmt.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: s.Field.DBName}, Value: 0},
	}})
	stmt.Clauses["soft_delete_enabled"] = clause.Clause{}
}

func (d Deleted) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{SoftDeletedQueryClause{Field: f}}
}

type SoftDeletedQueryClause struct {
	Field *schema.Field
}

func (s SoftDeletedQueryClause) Name() string {
	return ""
}

func (s SoftDeletedQueryClause) Build(clause.Builder) {
}

func (s SoftDeletedQueryClause) MergeClause(*clause.Clause) {
}

func (s SoftDeletedQueryClause) ModifyStatement(stmt *gorm.Statement) {
	if _, ok := stmt.Clauses["soft_delete_enabled"]; ok {
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
	stmt.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: s.Field.DBName}, Value: 0},
	}})
	stmt.Clauses["soft_delete_enabled"] = clause.Clause{}
}

func (d Deleted) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{SoftDeleteDeletedClause{Field: f}}
}

type SoftDeleteDeletedClause struct {
	Field *schema.Field
}

func (s SoftDeleteDeletedClause) Name() string {
	return ""
}

func (s SoftDeleteDeletedClause) Build(clause.Builder) {

}

func (s SoftDeleteDeletedClause) MergeClause(*clause.Clause) {
}

func (s SoftDeleteDeletedClause) ModifyStatement(stmt *gorm.Statement) {
	if stmt.SQL.String() != "" {
		return
	}
	assignment := clause.Assignment{
		Column: clause.Column{Name: s.Field.DBName},
		Value:  gorm.Expr("`id`"),
	}
	curTime := stmt.DB.NowFunc()
	_, idOk := stmt.Schema.FieldsByDBName["id"]
	if !idOk {
		assignment.Value = uint64(curTime.UnixNano())
		return
	}
	clauseSet := clause.Set{assignment}
	if _, ok := stmt.Schema.FieldsByName["DeletedAt"]; ok {
		clauseSet = append(clauseSet, clause.Assignment{Column: clause.Column{Name: "deleted_at"}, Value: curTime})
	}
	if filed, ok := stmt.Schema.FieldsByName["Status"]; ok {
		if !strings.Contains(filed.Comment, "skip_delete") {
			clauseSet = append(clauseSet, clause.Assignment{Column: clause.Column{Name: "status"}, Value: "deleted"})
		}
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
	if _, ok := stmt.Clauses["WHERE"]; !stmt.DB.AllowGlobalUpdate && !ok {
		_ = stmt.DB.AddError(gorm.ErrMissingWhereClause)
	} else {
		SoftDeletedQueryClause(s).ModifyStatement(stmt)
	}

	stmt.AddClauseIfNotExists(clause.Update{})
	stmt.Build("UPDATE", "SET", "WHERE")
}
