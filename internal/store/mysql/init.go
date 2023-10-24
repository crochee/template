package mysql

import (
	"context"
	"time"

	"github.com/spf13/viper"

	"template/internal/store"
	"template/pkg/logger/gormx"
	"template/pkg/storage"
)

// NewMysqlClient create mysql factory with context.Context
func NewMysqlClient(ctx context.Context) (*dataStore, error) {
	c, err := storage.New(ctx,
		storage.WithUser(viper.GetString("mysql.user")),
		storage.WithPassword(viper.GetString("mysql.password")),
		storage.WithIP(viper.GetString("mysql.ip")),
		storage.WithPort(viper.GetString("mysql.port")),
		storage.WithDatabase(viper.GetString("mysql.name")),
		storage.WithCharset(viper.GetString("mysql.charset")),
		storage.WithMaxOpenConn(viper.GetInt("mysql.max_open_conns")),
		storage.WithMaxIdleConn(viper.GetInt("mysql.max_idle_conns")),
		storage.WithMaxLifetime(time.Duration(viper.GetInt("mysql.conn_max_lifetime"))*time.Second),
		storage.WithLogger(storage.NewLog(gormx.NewGormWriterFrom)),
		storage.WithPlugins(
			storage.IgnoreSelectLogger{},
		))
	if err != nil {
		return nil, err
	}
	return &dataStore{DB: c}, nil
}

type dataStore struct {
	DB *storage.DB
}

func (d *dataStore) Begin() store.Store {
	return &dataStore{DB: &storage.DB{
		DB: d.DB.Begin(),
	}}
}

func (d *dataStore) Commit() {
	d.DB.Commit()
}

func (d *dataStore) Rollback() {
	d.DB.Rollback()
}

func (d *dataStore) Area() store.AreaStore {
	return newArea(d.DB)
}
