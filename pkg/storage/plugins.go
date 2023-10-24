package storage

import (
	"strings"

	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

type IgnoreSelectLogger struct {
}

func (IgnoreSelectLogger) Name() string {
	return "IgnoreSelectLogger"
}

func (IgnoreSelectLogger) Initialize(db *gorm.DB) error {
	queryCallback := db.Callback().Query()
	if err := queryCallback.Before("gorm:query").Register("ignore_query", ignoreQueryOrRow); err != nil {
		return err
	}

	rowCallback := db.Callback().Row()
	if err := rowCallback.Before("gorm:row").Register("ignore_row", ignoreQueryOrRow); err != nil {
		return err
	}

	rawCallback := db.Callback().Raw()
	if err := rawCallback.Before("gorm:raw").Register("ignore_raw", ignoreExecSelect); err != nil {
		return err
	}
	return nil
}

func ignoreQueryOrRow(db *gorm.DB) {
	if db.Error != nil {
		return
	}
	db.Logger = db.Logger.LogMode(glogger.Warn)
}

func ignoreExecSelect(db *gorm.DB) {
	if db.Error != nil {
		return
	}
	sql := db.Statement.SQL.String()
	sql = strings.ToUpper(sql)

	if strings.Contains(sql, "SELECT") {
		db.Logger = db.Logger.LogMode(glogger.Warn)
		return
	}
}
