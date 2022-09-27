package mysql

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"go_template/internal/store"
	"go_template/pkg/storage/mysql"
)

var (
	dbClient *mysql.DB
)

// Init init database
func Init(ctx context.Context) (err error) {
	dbClient, err = mysql.New(ctx,
		mysql.WithDBDebug(viper.GetString("mode") != gin.ReleaseMode),
		mysql.WithUser(viper.GetString("mysql.user")),
		mysql.WithPassword(viper.GetString("mysql.password")),
		mysql.WithIP(viper.GetString("mysql.ip")),
		mysql.WithPort(viper.GetString("mysql.port")),
		mysql.WithDatabase(viper.GetString("mysql.name")),
		mysql.WithCharset(viper.GetString("mysql.charset")),
		mysql.WithMaxOpenConn(viper.GetInt("mysql.max_open_conns")),
		mysql.WithMaxIdleConn(viper.GetInt("mysql.max_idle_conns")),
		mysql.WithMaxLifetime(time.Duration(viper.GetInt("mysql.conn_max_lifetime"))*time.Second),
	)
	return
}

// DB 若想使用SELECT打印日志，请使用DB(ctx,database.WithLog())
func DB(ctx context.Context, opts ...mysql.Opt) *mysql.DB {
	return dbClient.With(ctx, opts...)
}

// GetMysqlFactory create mysql factory with context.Context
func GetMysqlFactory(ctx context.Context) store.Factory {
	return &dataStore{db: DB(ctx)}
}

type dataStore struct {
	db *mysql.DB
}

func (d *dataStore) Begin() store.Factory {
	d.db.Begin()
	return &dataStore{db: &mysql.DB{
		DB:    d.db.Begin(),
		Debug: d.db.Debug,
	}}
}

func (d *dataStore) Commit() {
	d.db.Commit()
}

func (d *dataStore) Rollback() {
	d.db.Rollback()
}

func (d *dataStore) Auth() store.AuthorControlStore {
	return newAuthorControl(d.db)
}

func (d *dataStore) Flow() store.ChangeFlowStore {
	return newResourceChangeFlow(d.db)
}

func (d *dataStore) Pkg() store.ResourcePkgStore {
	return newResourcePkg(d.db)
}
