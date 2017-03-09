package util

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
	"testing"
	"time"
)

var dbConn *sqlx.DB

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

	dbConn, err = sqlx.Open("mysql", connString)
	if err != nil {
		return err
	}

	dbConn.SetConnMaxLifetime(time.Hour * 6)
	dbConn.SetMaxIdleConns(10)

	return dbConn.Ping()
}

// TestGetDB exposes the private dbConn variable for testing.
// Should not be used in any of the actual code - all queries should be
// materialized as tested functions in this package instead.
func TestGetDB(t *testing.T) *sqlx.DB {
	if dbConn == nil {
		t.Fatal("ConnectToMemSQL() must run before TestGetDB() can be called")
	}

	return dbConn
}

func DBGetVariable(varName string) (string, error) {
	var res struct {
		Name  string `db:"Variable_name"`
		Value string `db:"Value"`
	}
	err := dbConn.Get(&res, "SHOW VARIABLES LIKE ?", varName)
	return res.Value, err
}

// DBSetVariable sets database variables
func DBSetVariable(db string) error {
	_, err := dbConn.Exec(db)
	return err
}

func DBGetUserDatabases() ([]string, error) {
	rows, err := dbConn.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	dbs := make([]string, 0)
	for rows.Next() {
		var db string
		err := rows.Scan(&db)
		if err != nil {
			return nil, err
		}
		if db != "information_schema" &&
			db != "memsql" &&
			db != "sharding" {
			dbs = append(dbs, db)
		}
	}
	return dbs, nil
}

func DBSnapshotDatabase(db string) error {
	_, err := dbConn.Exec(fmt.Sprintf("SNAPSHOT DATABASE `%s`", db))
	return err
}
