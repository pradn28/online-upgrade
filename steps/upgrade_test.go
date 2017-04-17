package steps_test

import (
	"os"
	"testing"

	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	versionHash = os.Getenv("MEMSQL_VERSION_5_7_2")
)

func TestUpgrade(t *testing.T) {

	// Get command line arguments and defer reset of args
	osArgs := os.Args
	defer func() { os.Args = osArgs }()

	// Set version-hash argument
	os.Args = append(os.Args, "-version-hash", versionHash)

	// Build a test HA cluster and return a
	// function, which we will defer, to terminate it.
	defer testutil.ClusterHA(t)()

	// Create a testing database
	database := "testing"
	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, database)()

	// Create row store tables
	rowTableNames := []string{"testing_one", "testing_two"}
	testutil.CreateLoremTable(t, "row", database, rowTableNames)
	// Create column store tables
	columnarTableNames := []string{"testing_three", "testing_four"}
	testutil.CreateLoremTable(t, "column", database, columnarTableNames)

	// Run update config to set variables to OFF
	t.Run("UpdateConfig", func(t *testing.T) {
		err := steps.UpdateConfig("OFF")
		assert.Nil(t, err)
	})

	// Make a done channel to signal other channels when we are done
	done := make(chan struct{})

	// make channel for lorem ipsum strings
	loremData := make(chan []string)

	testutil.GenerateLoremStrings(loremData, done)
	testutil.InsertLoremData(loremData, t, database, rowTableNames)
	testutil.InsertLoremData(loremData, t, database, columnarTableNames)

	// Detach leaves in AG 1
	t.Run("DetachLeaves", func(t *testing.T) {
		err := steps.DetachLeaves(1)
		assert.Nil(t, err)
	})

	// Run the upgrade on the leaves in AG 1
	t.Run("UpgradeMemSQL", func(t *testing.T) {
		err := steps.UpgradeLeaves(1)
		assert.Nil(t, err)
	})

	// Attach leaves in AG 1
	t.Run("AttachLeaves", func(t *testing.T) {
		err := steps.AttachLeaves(1)
		assert.Nil(t, err)
	})

	// Run restore redundancy on all user databases
	t.Run("RestoreRedundancy", func(t *testing.T) {
		err := steps.RestoreRedundancy()
		assert.Nil(t, err)
	})

	// Verify all leaves are online
	t.Run("CheckShowLeaves", func(t *testing.T) {
		leaves, _ := util.DBShowLeaves()
		assert.Len(t, leaves, 2)
		assert.Equal(t, "online", leaves[0].State)
		assert.Equal(t, "online", leaves[1].State)
	})

	// Detach leaves in AG 2
	t.Run("DetachLeaves", func(t *testing.T) {
		err := steps.DetachLeaves(2)
		assert.Nil(t, err)
	})

	// Run the upgrade on the leaves in AG 2
	t.Run("UpgradeMemSQL", func(t *testing.T) {
		err := steps.UpgradeLeaves(2)
		assert.Nil(t, err)
	})

	// Close done channel
	close(done)

	// Attach leaves in AG 2
	t.Run("AttachLeaves", func(t *testing.T) {
		err := steps.AttachLeaves(2)
		assert.Nil(t, err)
	})

	// Run restore redundancy on all user databases
	t.Run("RestoreRedundancy", func(t *testing.T) {
		err := steps.RestoreRedundancy()
		assert.Nil(t, err)
	})

	// Verify all leaves are online
	t.Run("CheckShowLeaves", func(t *testing.T) {
		leaves, _ := util.DBShowLeaves()
		assert.Len(t, leaves, 2)
		assert.Equal(t, "online", leaves[0].State)
		assert.Equal(t, "online", leaves[1].State)
	})

	// Run rebalance partitions on all user databases
	t.Run("RebalancePartitions", func(t *testing.T) {
		err := steps.RebalancePartitions()
		assert.Nil(t, err)
	})

	// Run upgrade on Aggs
	t.Run("UpgradeAggs", func(t *testing.T) {
		err := steps.UpgradeAggregators()
		assert.Nil(t, err)
	})

	// Run upgrade on Master
	t.Run("UpgradeMaster", func(t *testing.T) {
		err := steps.UpgradeMaster()
		assert.Nil(t, err)
	})

	// Update Configs to ON
	t.Run("UpdateConfig", func(t *testing.T) {
		err := steps.UpdateConfig("ON")
		assert.Nil(t, err)
	})

}
