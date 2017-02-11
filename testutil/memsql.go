package testutil

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/memsql/online-upgrade/util"
)

func CreateDatabase(t *testing.T, db string) func() {
	conn := util.TestGetDB(t)
	_, err := conn.Exec(fmt.Sprintf("CREATE DATABASE `%s`", db))
	require.Nil(t, err)
	return func() {
		_, err := conn.Exec(fmt.Sprintf("DROP DATABASE `%s`", db))
		require.Nil(t, err)
	}
}
