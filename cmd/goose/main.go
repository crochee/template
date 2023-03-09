// This is custom goose binary with sqlite3 support only.

package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose"

	"template/config"
	"template/internal/store/mysql"
)

var (
	flags      = flag.NewFlagSet("goose", flag.ExitOnError)
	dir        = flags.String("dir", ".", "directory with migration files")
	verbose    = flags.Bool("v", false, "enable verbose mode")
	help       = flags.Bool("h", false, "print help")
	configFile = flag.String("f", "./config/template.yaml", "the config file")
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	goose.SetVerbose(*verbose)

	args := flags.Args()
	if *help {
		flags.Usage()
		return nil
	}
	if len(args) == 0 {
		flags.Usage()
		return nil
	}
	command := args[0]
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}
	// 初始化配置
	if err := config.LoadConfig(*configFile); err != nil {
		log.Fatal(err)
	}

	db, err := mysql.NewMysqlClient(context.Background())
	if err != nil {
		return err
	}
	defer db.DB.Close()
	var arguments []string
	if len(args) > 1 {
		arguments = append(arguments, args[1:]...)
	}
	var d *sql.DB
	if d, err = db.DB.DB.DB(); err != nil {
		return err
	}
	return goose.Run(command, d, *dir, arguments...)
}

func usage() {
	fmt.Println(`Usage: goose [OPTIONS] COMMAND

Examples:
    goose status
    goose create init sql
    goose create add_some_column sql
    goose create fetch_user_data go
    goose up

Options:`)
	flags.PrintDefaults()
	fmt.Println(`
Commands:
    up                   Migrate the DB to the most recent version available
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations`)
}
