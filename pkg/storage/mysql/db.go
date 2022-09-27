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

type option struct {
	debug bool

	maxOpenConn int
	maxIdleConn int

	user     string
	password string
	ip       string
	port     string
	database string
	charset  string

	timeout         time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	connMaxLifetime time.Duration
}

type Option func(*option)

func WithDBDebug(debug bool) Option {
	return func(o *option) {
		o.debug = debug
	}
}

func WithMaxOpenConn(maxOpenConn int) Option {
	return func(o *option) {
		o.maxOpenConn = maxOpenConn
	}
}

func WithMaxIdleConn(maxIdleConn int) Option {
	return func(o *option) {
		o.maxIdleConn = maxIdleConn
	}
}

func WithUser(user string) Option {
	return func(o *option) {
		o.user = user
	}
}

func WithPassword(password string) Option {
	return func(o *option) {
		o.password = password
	}
}

func WithIP(ip string) Option {
	return func(o *option) {
		o.ip = ip
	}
}

func WithPort(port string) Option {
	return func(o *option) {
		o.port = port
	}
}

func WithDatabase(db string) Option {
	return func(o *option) {
		o.database = db
	}
}

func WithCharset(charset string) Option {
	return func(o *option) {
		o.charset = charset
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *option) {
		o.timeout = timeout
	}
}

func WithReadTimeout(readTimeout time.Duration) Option {
	return func(o *option) {
		o.readTimeout = readTimeout
	}
}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(o *option) {
		o.writeTimeout = writeTimeout
	}
}

func WithMaxLifetime(connMaxLifetime time.Duration) Option {
	return func(o *option) {
		o.connMaxLifetime = connMaxLifetime
	}
}

// New init DB
func New(ctx context.Context, opts ...Option) (*DB, error) {
	o := &option{
		debug:       true,
		maxOpenConn: 100,
		maxIdleConn: 80,
		ip:          "127.0.0.1",
		port:        "3306",
		charset:     "utf8mb4",
	}
	for _, f := range opts {
		f(o)
	}
	client, err := gorm.Open(mysql.Open(Dsn(o.user, o.password, o.ip, o.port,
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

	c := &DB{DB: client, Debug: o.debug}
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
