package util_test

import (
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemsqlHA(t *testing.T) {
	defer testutil.ClusterHA(t)()

	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")()
	defer testutil.CreateDatabase(t, "sharding_1a2s3d4f5g6h")()

	// Requires a master aggregator
	t.Run("DBShowLeaves", func(t *testing.T) {
		leaves, err := util.DBShowLeaves()
		assert.Nil(t, err)
		assert.Len(t, leaves, 2) // HA Cluster
		assert.Equal(t, "online", leaves[0].State)
	})

	t.Run("MasterSlaveCheck", func(t *testing.T) {
		partitions, err := util.DBShowClusterStatus()
		assert.Nil(t, err)
		masterSlaveErr := util.MasterSlaveCheck(partitions)
		assert.Nil(t, masterSlaveErr)
	})

	t.Run("OrphanCheck", func(t *testing.T) {
		partitions, err := util.DBShowClusterStatus()
		assert.Nil(t, err)
		orphanErr := util.OrphanCheck(partitions)
		assert.Nil(t, orphanErr)
	})

	// DBRestoreRedundancy
	t.Run("DBRestoreRedundancy", func(t *testing.T) {
		err := util.DBRestoreRedundancy("testing")
		assert.Nil(t, err)
	})

	// DBRebalancePartitions
	t.Run("DBRebalancePartitions", func(t *testing.T) {
		err := util.DBRebalancePartitions("testing")
		assert.Nil(t, err)
	})

	t.Run("DBGetUserDatabases", func(t *testing.T) {
		dbs, err := util.DBGetUserDatabases()
		assert.Nil(t, err)
		assert.Equal(t, []string{"testing"}, dbs)
	})
}

func TestMemsql(t *testing.T) {
	defer testutil.ClusterInABox(t)()

	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")()

	t.Run("DBShowClusterStatus", func(t *testing.T) {
		rows, err := util.DBShowClusterStatus()
		assert.Nil(t, err)
		assert.Equal(t, "testing", rows[0].Database)
	})

	t.Run("DBSetVariable", func(t *testing.T) {
		err := util.DBSetVariable("SET @@GLOBAL.aggregator_failure_detection = OFF")
		assert.Nil(t, err)
	})

	t.Run("DBGetVariable", func(t *testing.T) {
		varval, err := util.DBGetVariable("redundancy_level")
		assert.Nil(t, err)
		assert.Equal(t, "1", varval)
	})

	t.Run("DBSnapshotDatabase", func(t *testing.T) {
		err := util.DBSnapshotDatabase("testing")
		assert.Nil(t, err)
	})
}
