package util_test

import (
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemsql(t *testing.T) {
	// Need to use ClusterInABox to test DBSetVariable
	defer testutil.ClusterInABox(t)()

	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")()

	// Requires a master aggregator
	t.Run("DBSetVariable", func(t *testing.T) {
		err := util.DBSetVariable("SET @@GLOBAL.redundancy_level = 2")
		assert.Nil(t, err)
	})

	t.Run("DBGetVariable", func(t *testing.T) {
		varval, err := util.DBGetVariable("redundancy_level")
		assert.Nil(t, err)
		assert.Equal(t, "2", varval)
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
