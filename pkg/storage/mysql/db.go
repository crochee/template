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

	"go_template/pkg/logger"
	"go_template/pkg/logger/gormx"
)

type Option struct {
	Debug bool

	MaxOpenConn int
	MaxIdleConn int

	User     string
	Password string
	IP       string
	Port     string
	Database string
	Charset  string

	Timeout         time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ConnMaxLifetime time.Duration
}

// New init DB
func New(ctx context.Context, opts ...func(*Option)) (*DB, error) {
	o := &Option{
		Debug:       true,
		MaxOpenConn: 100,
		MaxIdleConn: 80,
		IP:          "127.0.0.1",
		Port:        "3306",
		Charset:     "utf8mb4",
	}
	for _, f := range opts {
		f(o)
	}
	client, err := gorm.Open(mysql.Open(Dsn(o.User, o.Password, o.IP, o.Port,
		o.Database, o.Charset, o.Timeout, o.ReadTimeout, o.WriteTimeout)),
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
	if o.Debug { // 是否显示sql语句
		session.Logger = client.Logger.LogMode(glogger.Info)
	}
	client = client.Session(session)

	var sqlDB *sql.DB
	if sqlDB, err = client.DB(); err != nil {
		return nil, err
	}
	// 连接池配置
	sqlDB.SetMaxOpenConns(o.MaxOpenConn)        // 默认值0，无限制
	sqlDB.SetMaxIdleConns(o.MaxIdleConn)        // 默认值2
	sqlDB.SetConnMaxLifetime(o.ConnMaxLifetime) // 默认值0，永不过期

	c := &DB{DB: client, Debug: o.Debug}
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
	debug         bool
}

type Opt func(*opt)

type DB struct {
	*gorm.DB
	Debug bool
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
	c := &DB{DB: d.Session(&gorm.Session{
		Context: ctx,
		Logger: gormx.NewLog(l, o.debug || d.Debug, glogger.Config{
			SlowThreshold: o.slowThreshold,
			Colorful:      o.colorful,
			LogLevel:      o.levelFunc(glogger.Info, d.Debug),
		}),
	}),
		Debug: d.Debug,
	}
	return c
}

func getLevel(l glogger.LogLevel, debug bool) glogger.LogLevel {
	if debug {
		return glogger.Info
	}
	return l
}

func WithNoInfoHandle(o *opt) {
	o.levelFunc = noInfoHandle
}

func WithDebug(o *opt) {
	o.debug = true
}

func noInfoHandle(level glogger.LogLevel, _ bool) glogger.LogLevel {
	if level > glogger.Warn {
		return glogger.Warn
	}
	return level
}
