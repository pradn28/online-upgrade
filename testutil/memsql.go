package testutil

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/memsql/online-upgrade/util"
	"github.com/stretchr/testify/require"
)

// CreateDatabase will create a database with a name specified
// and return a DROP DATABASE function; which should be deferred
func CreateDatabase(t *testing.T, db string) func() {
	conn := util.TestGetDB(t)
	_, err := conn.Exec(fmt.Sprintf("CREATE DATABASE `%s`", db))
	require.Nil(t, err)
	log.Println("Testing database created")
	return func() {
		_, err := conn.Exec(fmt.Sprintf("DROP DATABASE `%s`", db))
		require.Nil(t, err)
		log.Println("Testing database dropped")
	}
}

// CreateTable Create table for testing
func CreateTable(t *testing.T, db, table, columns string) {
	conn := util.TestGetDB(t)
	_, err := conn.Exec(fmt.Sprintf("CREATE TABLE `%s`.`%s` %s", db, table, columns))
	require.Nil(t, err)
	log.Printf("Created Table: %s", table)
}

// InsertRow inserts into table
func InsertRow(t *testing.T, db, table string, columns, values []string) error {

	columnNames := strings.Join(columns, ", ")
	columnValues := strings.Join(values, ", ")

	conn := util.TestGetDB(t)

	_, err := conn.Exec(fmt.Sprintf(
		"INSERT INTO `%s`.`%s` (%s) values (%s)",
		db, table, columnNames, columnValues,
	))
	require.Nil(t, err)
	return err
}
