package util_test

import (
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemsql(t *testing.T) {
	defer testutil.ClusterSinglebox(t)()

	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")()

	t.Run("DBGetVariable", func(t *testing.T) {
		varval, err := util.DBGetVariable("version_compile_os")
		assert.Nil(t, err)
		assert.Equal(t, "Linux", varval)
	})

	t.Run("DBGetUserDatabases", func(t *testing.T) {
		dbs, err := util.DBGetUserDatabases()
		assert.Nil(t, err)
		assert.Equal(t, []string{"testing"}, dbs)
	})

	t.Run("DBSnapshotDatabase", func(t *testing.T) {
		err := util.DBSnapshotDatabase("testing")
		assert.Nil(t, err)
	})
}
