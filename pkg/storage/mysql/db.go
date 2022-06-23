package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/crochee/devt/pkg/logger"
	"github.com/crochee/devt/pkg/logger/gormx"
)

type option struct {
	debug bool

	maxOpenConn int
	maxIdleConn int

	user     string
	password string
	IP       string
	port     string
	database string
	charset  string

	timeout         time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	connMaxLifetime time.Duration
}

type Option func(*option)

// New init DB
func New(ctx context.Context, opts ...Option) (*DB, error) {
	o := &option{
		debug:       true,
		maxOpenConn: 100,
		maxIdleConn: 80,
		IP:          "127.0.0.1",
		port:        "3306",
		charset:     "utf8mb4",
	}
	for _, f := range opts {
		f(o)
	}
	client, err := gorm.Open(mysql.Open(Dsn(o.user, o.password, o.IP, o.port,
		o.database, o.charset, o.timeout, o.readTimeout, o.writeTimeout)),
		&gorm.Config{
			SkipDefaultTransaction: false,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 不考虑表名单复数变化
			},
			FullSaveAssociations: false,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
			PrepareStmt: true,
		},
	)
	if err != nil {
		return nil, err
	}
	session := &gorm.Session{Context: ctx}
	if o.debug { // 是否显示sql语句
		session.Logger = client.Logger.LogMode(glogger.Info)
	}
	client = client.Session(session)

	var sqlDB *sql.DB
	if sqlDB, err = client.DB(); err != nil {
		return nil, err
	}
	// 连接池配置
	sqlDB.SetMaxOpenConns(o.maxOpenConn)        // 默认值0，无限制
	sqlDB.SetMaxIdleConns(o.maxIdleConn)        // 默认值2
	sqlDB.SetConnMaxLifetime(o.connMaxLifetime) // 默认值0，永不过期

	c := &DB{DB: client, debug: o.debug}
	runtime.SetFinalizer(c, closeClient)
	return c, nil
}

func closeClient(c *DB) {
	s, err := c.DB.DB()
	if err != nil {
		log.Println(err)
		return
	}
	if err = s.Close(); err != nil {
		log.Println(err)
	}
}

func Dsn(user, password, ip, port, database, charset string, timeout, readTimeout, writeTimeout time.Duration) string {
	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		user, password, ip, port, database, charset, true, "UTC")
	if timeout != 0 {
		uri += fmt.Sprintf("&timeout=%s", timeout)
	}
	if readTimeout != 0 {
		uri += fmt.Sprintf("&readTimeout=%s", readTimeout)
	}
	if writeTimeout != 0 {
		uri += fmt.Sprintf("&writeTimeout=%s", writeTimeout)
	}
	return uri
}

type opt struct {
	slowThreshold time.Duration
	colorful      bool
	levelFunc     func(glogger.LogLevel, bool) glogger.LogLevel
}

type Opt func(*opt)

type DB struct {
	*gorm.DB
	debug bool
}

func (d *DB) Debug() bool {
	return d.debug

}

// With options to set orm logger
func (d *DB) With(ctx context.Context, opts ...Opt) *DB {
	l := logger.From(ctx)
	o := &opt{
		slowThreshold: 10 * time.Second,
		colorful:      false,
		levelFunc:     getLevel,
	}
	for _, f := range opts {
		f(o)
	}
	return &DB{DB: d.Session(&gorm.Session{
		Context: ctx,
		Logger: gormx.NewLog(l, d.debug, glogger.Config{
			SlowThreshold: o.slowThreshold,
			Colorful:      o.colorful,
			LogLevel:      o.levelFunc(glogger.Warn, d.debug),
		}),
	}),
		debug: d.debug,
	}
}

func getLevel(l glogger.LogLevel, debug bool) glogger.LogLevel {
	if debug {
		return glogger.Info
	}
	return l
}
