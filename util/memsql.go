package util

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
	"time"
)

var db *sqlx.DB

func ConnectToMemSQL(config Config) error {
	var err error

	connParams := strings.Join([]string{
		// convert timestame and date to time.Time
		"parseTime=true",
		// don't use the binary protocol
		"interpolateParams=true",
		// set a sane connection timeout rather than the default infinity
		"timeout=10s",
	}, "&")

	connString := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/information_schema?%s",
		config.MasterUser,
		config.MasterPass,
		config.MasterHost,
		config.MasterPort,
		connParams,
	)

	log.Printf("Connecting to MemSQL %s", connString)

	db, err = sqlx.Open("mysql", connString)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(time.Hour * 6)
	db.SetMaxIdleConns(10)

	return db.Ping()
}

func DBGetVariable(varName string) (string, error) {
	var res struct {
		Name  string `db:"Variable_name"`
		Value string `db:"Value"`
	}
	err := db.Get(&res, "SHOW VARIABLES LIKE ?", varName)
	return res.Value, err
}
