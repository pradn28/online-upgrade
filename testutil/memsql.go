package testutil

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/memsql/online-upgrade/util"
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
