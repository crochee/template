package storage

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
)

func Preload(db *gorm.DB) {
	if db.Error == nil && len(db.Statement.Preloads) > 0 {
		if db.Statement.Schema == nil {
			db.AddError(fmt.Errorf("%w when using preload", gorm.ErrModelValueRequired))
			return
		}

		preloadMap := map[string]map[string][]interface{}{}
		for name := range db.Statement.Preloads {
			preloadFields := strings.Split(name, ".")
			if preloadFields[0] == clause.Associations {
				for _, rel := range db.Statement.Schema.Relationships.Relations {
					if rel.Schema == db.Statement.Schema {
						if _, ok := preloadMap[rel.Name]; !ok {
							preloadMap[rel.Name] = map[string][]interface{}{}
						}

						if value := strings.TrimPrefix(strings.TrimPrefix(name, preloadFields[0]), "."); value != "" {
							preloadMap[rel.Name][value] = db.Statement.Preloads[name]
						}
					}
				}
			} else {
				if _, ok := preloadMap[preloadFields[0]]; !ok {
					preloadMap[preloadFields[0]] = map[string][]interface{}{}
				}

				if value := strings.TrimPrefix(strings.TrimPrefix(name, preloadFields[0]), "."); value != "" {
					preloadMap[preloadFields[0]][value] = db.Statement.Preloads[name]
				}
			}
		}

		preloadNames := make([]string, 0, len(preloadMap))
		for key := range preloadMap {
			preloadNames = append(preloadNames, key)
		}
		sort.Strings(preloadNames)

		preloadDB := db.Session(&gorm.Session{Context: db.Statement.Context, NewDB: true, SkipHooks: db.Statement.SkipHooks, Initialized: true})
		db.Statement.Settings.Range(func(k, v interface{}) bool {
			preloadDB.Statement.Settings.Store(k, v)
			return true
		})

		if err := preloadDB.Statement.Parse(db.Statement.Dest); err != nil {
			return
		}
		preloadDB.Statement.ReflectValue = db.Statement.ReflectValue

		for _, name := range preloadNames {
			if rel := preloadDB.Statement.Schema.Relationships.Relations[name]; rel != nil {
				db.AddError(preload(preloadDB.Table("").Session(&gorm.Session{Context: db.Statement.Context, SkipHooks: db.Statement.SkipHooks}), rel, append(db.Statement.Preloads[name], db.Statement.Preloads[clause.Associations]...), preloadMap[name]))
			} else {
				db.AddError(fmt.Errorf("%s: %w for schema %s", name, gorm.ErrUnsupportedRelation, db.Statement.Schema.Name))
			}
		}
	}
}

func preload(tx *gorm.DB, rel *schema.Relationship, conds []interface{}, preloads map[string][]interface{}) error {
	var (
		reflectValue     = tx.Statement.ReflectValue
		relForeignKeys   []string
		relForeignFields []*schema.Field
		foreignFields    []*schema.Field
		foreignValues    [][]interface{}
		identityMap      = map[string][]reflect.Value{}
		inlineConds      []interface{}
	)

	if rel.JoinTable != nil {
		var (
			joinForeignFields    = make([]*schema.Field, 0, len(rel.References))
			joinRelForeignFields = make([]*schema.Field, 0, len(rel.References))
			joinForeignKeys      = make([]string, 0, len(rel.References))
		)

		for _, ref := range rel.References {
			if ref.OwnPrimaryKey {
				joinForeignKeys = append(joinForeignKeys, ref.ForeignKey.DBName)
				joinForeignFields = append(joinForeignFields, ref.ForeignKey)
				foreignFields = append(foreignFields, ref.PrimaryKey)
			} else if ref.PrimaryValue != "" {
				tx = tx.Where(clause.Eq{Column: ref.ForeignKey.DBName, Value: ref.PrimaryValue})
			} else {
				joinRelForeignFields = append(joinRelForeignFields, ref.ForeignKey)
				relForeignKeys = append(relForeignKeys, ref.PrimaryKey.DBName)
				relForeignFields = append(relForeignFields, ref.PrimaryKey)
			}
		}

		joinIdentityMap, joinForeignValues := schema.GetIdentityFieldValuesMap(tx.Statement.Context, reflectValue, foreignFields)
		if len(joinForeignValues) == 0 {
			return nil
		}

		joinResults := rel.JoinTable.MakeSlice().Elem()
		column, values := schema.ToQueryValues(clause.CurrentTable, joinForeignKeys, joinForeignValues)
		if err := tx.Where(IN{Column: column, Values: values}).Find(joinResults.Addr().Interface()).Error; err != nil {
			return err
		}

		// convert join identity map to relation identity map
		fieldValues := make([]interface{}, len(joinForeignFields))
		joinFieldValues := make([]interface{}, len(joinRelForeignFields))
		for i := 0; i < joinResults.Len(); i++ {
			joinIndexValue := joinResults.Index(i)
			for idx, field := range joinForeignFields {
				fieldValues[idx], _ = field.ValueOf(tx.Statement.Context, joinIndexValue)
			}

			for idx, field := range joinRelForeignFields {
				joinFieldValues[idx], _ = field.ValueOf(tx.Statement.Context, joinIndexValue)
			}

			if results, ok := joinIdentityMap[utils.ToStringKey(fieldValues...)]; ok {
				joinKey := utils.ToStringKey(joinFieldValues...)
				identityMap[joinKey] = append(identityMap[joinKey], results...)
			}
		}

		_, foreignValues = schema.GetIdentityFieldValuesMap(tx.Statement.Context, joinResults, joinRelForeignFields)
	} else {
		for _, ref := range rel.References {
			if ref.OwnPrimaryKey {
				relForeignKeys = append(relForeignKeys, ref.ForeignKey.DBName)
				relForeignFields = append(relForeignFields, ref.ForeignKey)
				foreignFields = append(foreignFields, ref.PrimaryKey)
			} else if ref.PrimaryValue != "" {
				tx = tx.Where(clause.Eq{Column: ref.ForeignKey.DBName, Value: ref.PrimaryValue})
			} else {
				relForeignKeys = append(relForeignKeys, ref.PrimaryKey.DBName)
				relForeignFields = append(relForeignFields, ref.PrimaryKey)
				foreignFields = append(foreignFields, ref.ForeignKey)
			}
		}

		identityMap, foreignValues = schema.GetIdentityFieldValuesMap(tx.Statement.Context, reflectValue, foreignFields)
		if len(foreignValues) == 0 {
			return nil
		}
	}

	// nested preload
	for p, pvs := range preloads {
		tx = tx.Preload(p, pvs...)
	}

	reflectResults := rel.FieldSchema.MakeSlice().Elem()
	column, values := schema.ToQueryValues(clause.CurrentTable, relForeignKeys, foreignValues)

	if len(values) != 0 {
		for _, cond := range conds {
			if fc, ok := cond.(func(*gorm.DB) *gorm.DB); ok {
				tx = fc(tx)
			} else {
				inlineConds = append(inlineConds, cond)
			}
		}

		if err := tx.Where(IN{Column: column, Values: values}).Find(reflectResults.Addr().Interface(), inlineConds...).Error; err != nil {
			return err
		}
	}

	fieldValues := make([]interface{}, len(relForeignFields))

	// clean up old values before preloading
	switch reflectValue.Kind() {
	case reflect.Struct:
		switch rel.Type {
		case schema.HasMany, schema.Many2Many:
			tx.AddError(rel.Field.Set(tx.Statement.Context, reflectValue, reflect.MakeSlice(rel.Field.IndirectFieldType, 0, 10).Interface()))
		default:
			tx.AddError(rel.Field.Set(tx.Statement.Context, reflectValue, reflect.New(rel.Field.FieldType).Interface()))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < reflectValue.Len(); i++ {
			switch rel.Type {
			case schema.HasMany, schema.Many2Many:
				tx.AddError(rel.Field.Set(tx.Statement.Context, reflectValue.Index(i), reflect.MakeSlice(rel.Field.IndirectFieldType, 0, 10).Interface()))
			default:
				tx.AddError(rel.Field.Set(tx.Statement.Context, reflectValue.Index(i), reflect.New(rel.Field.FieldType).Interface()))
			}
		}
	}

	for i := 0; i < reflectResults.Len(); i++ {
		elem := reflectResults.Index(i)
		for idx, field := range relForeignFields {
			fieldValues[idx], _ = field.ValueOf(tx.Statement.Context, elem)
		}

		items, ok := identityMap[utils.ToStringKey(fieldValues...)]
		if !ok {
			return fmt.Errorf("failed to assign association %#v, make sure foreign fields exists", elem.Interface())
		}

		for _, item := range items {
			reflectFieldValue := rel.Field.ReflectValueOf(tx.Statement.Context, item)
			if reflectFieldValue.Kind() == reflect.Ptr && reflectFieldValue.IsNil() {
				reflectFieldValue.Set(reflect.New(rel.Field.FieldType.Elem()))
			}

			reflectFieldValue = reflect.Indirect(reflectFieldValue)
			switch reflectFieldValue.Kind() {
			case reflect.Struct:
				tx.AddError(rel.Field.Set(tx.Statement.Context, item, elem.Interface()))
			case reflect.Slice, reflect.Array:
				if reflectFieldValue.Type().Elem().Kind() == reflect.Ptr {
					tx.AddError(rel.Field.Set(tx.Statement.Context, item, reflect.Append(reflectFieldValue, elem).Interface()))
				} else {
					tx.AddError(rel.Field.Set(tx.Statement.Context, item, reflect.Append(reflectFieldValue, elem.Elem()).Interface()))
				}
			}
		}
	}

	return tx.Error
}

type IN struct {
	Column interface{}
	Values []interface{}
}

func (in IN) Build(builder clause.Builder) {
	builder.WriteQuoted(in.Column)

	switch len(in.Values) {
	case 0:
		builder.WriteString(" IN (NULL)")
	case 1:
		if _, ok := in.Values[0].([]interface{}); !ok {
			builder.WriteString(" = ")
			builder.AddVar(builder, in.Values[0])
			break
		}

		fallthrough
	default:
		builder.WriteString(" IN (")
		for idx, value := range in.Values {
			if idx > 0 {
				builder.WriteByte(',')
			}
			switch v := value.(type) {
			case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
				builder.WriteString(fmt.Sprintf("%d", v))
			case string:
				builder.WriteString("'" + v + "'")
			default:
				builder.AddVar(builder, v)
			}
		}
		builder.WriteByte(')')
	}
}

func (in IN) NegationBuild(builder clause.Builder) {
	builder.WriteQuoted(in.Column)
	switch len(in.Values) {
	case 0:
		builder.WriteString(" IS NOT NULL")
	case 1:
		if _, ok := in.Values[0].([]interface{}); !ok {
			builder.WriteString(" <> ")
			builder.AddVar(builder, in.Values[0])
			break
		}

		fallthrough
	default:
		builder.WriteString(" NOT IN (")
		builder.AddVar(builder, in.Values...)
		builder.WriteByte(')')
	}
}
