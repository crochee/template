package storage

import (
	"strings"

	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func NewIgnoreSelectLogger(debug bool) ignoreSelectLogger {
	return ignoreSelectLogger{debug: debug}
}

type ignoreSelectLogger struct {
	debug bool
}

func (ignoreSelectLogger) Name() string {
	return "IgnoreSelectLogger"
}

func (i ignoreSelectLogger) Initialize(db *gorm.DB) error {
	queryCallback := db.Callback().Query()
	if err := queryCallback.Before("gorm:query").Register("ignore_query", i.ignoreQueryOrRow); err != nil {
		return err
	}

	rowCallback := db.Callback().Row()
	if err := rowCallback.Before("gorm:row").Register("ignore_row", i.ignoreQueryOrRow); err != nil {
		return err
	}

	rawCallback := db.Callback().Raw()
	if err := rawCallback.Before("gorm:raw").Register("ignore_raw", i.ignoreExecSelect); err != nil {
		return err
	}
	return nil
}

func (i ignoreSelectLogger) ignoreQueryOrRow(db *gorm.DB) {
	if db.Error != nil {
		return
	}
	if i.debug {
		return
	}
	db.Logger = db.Logger.LogMode(glogger.Warn)
}

func (i ignoreSelectLogger) ignoreExecSelect(db *gorm.DB) {
	if db.Error != nil {
		return
	}
	sql := db.Statement.SQL.String()
	sql = strings.ToUpper(sql)

	if !i.debug && strings.Contains(sql, "SELECT") {
		db.Logger = db.Logger.LogMode(glogger.Warn)
		return
	}
}
