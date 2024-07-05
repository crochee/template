package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	NotFound           = gorm.ErrRecordNotFound
	ErrNotRowsAffected = errors.New("0 rows affected")
	ErrDuplicate       = "1062: Duplicate"
)

type option struct {
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
	logger          glogger.Interface
	plugins         []gorm.Plugin
}

type Option func(*option)

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

func WithLogger(logger glogger.Interface) Option {
	return func(o *option) {
		o.logger = logger
	}
}

func WithPlugins(plugins ...gorm.Plugin) Option {
	return func(o *option) {
		o.plugins = plugins
	}
}

// New init DB
func New(ctx context.Context, opts ...Option) (*DB, error) {
	o := &option{
		maxOpenConn: 100,
		maxIdleConn: 80,
		ip:          "127.0.0.1",
		port:        "3306",
		charset:     "utf8mb4",
	}
	for _, f := range opts {
		f(o)
	}
	client, err := gorm.Open(
		&mysql.Dialector{Config: &mysql.Config{
			DSN: Dsn(
				o.user,
				o.password,
				o.ip,
				o.port,
				o.database,
				o.charset,
				o.timeout,
				o.readTimeout,
				o.writeTimeout,
			),
			DisableWithReturning: true,
		}},
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
			Logger:      o.logger,
		},
	)
	if err != nil {
		return nil, err
	}

	// Replace preload callback
	client.Callback().Query().Replace("gorm:preload", Preload)

	// 插件注入
	for _, plugin := range o.plugins {
		if err = client.Use(plugin); err != nil {
			return nil, err
		}
	}
	// 注入context
	client = client.WithContext(ctx)

	var sqlDB *sql.DB
	if sqlDB, err = client.DB(); err != nil {
		return nil, err
	}
	// 连接池配置
	sqlDB.SetMaxOpenConns(o.maxOpenConn)        // 默认值0，无限制
	sqlDB.SetMaxIdleConns(o.maxIdleConn)        // 默认值2
	sqlDB.SetConnMaxLifetime(o.connMaxLifetime) // 默认值0，永不过期

	return &DB{DB: client}, nil
}

func Dsn(
	user, password, ip, port, database, charset string,
	timeout, readTimeout, writeTimeout time.Duration,
) string {
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

type DB struct {
	*gorm.DB
}

func (d *DB) Close() error {
	s, err := d.DB.DB()
	if err != nil {
		return err
	}
	return s.Close()
}
