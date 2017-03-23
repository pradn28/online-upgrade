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

	// Build test cluster
	defer testutil.ClusterHA(t)()

	// Create testing database
	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")

	// Run update config to set variables to OFF
	t.Run("UpdateConfig", func(t *testing.T) {
		err := steps.UpdateConfig("OFF")
		assert.Nil(t, err)
	})

	// Detach leaves in AG 1
	t.Run("DetachLeaves", func(t *testing.T) {
		err := steps.DetachLeaves(1)
		assert.Nil(t, err)
	})

	//Run the upgrade on the leaves in AG 1
	t.Run("UpgradeMemSQL", func(t *testing.T) {
		err := steps.UpgradeLeaves(1)
		assert.Nil(t, err)
	})

	// Attach leaves in AG 1
	t.Run("AttachLeaves", func(t *testing.T) {
		err := steps.AttachLeaves(1)
		assert.Nil(t, err)
	})

	// Verify all leaves are online
	t.Run("CheckShowLeaves", func(t *testing.T) {
		leaves, _ := util.DBShowLeaves()
		assert.Len(t, leaves, 2)
		assert.Equal(t, "online", leaves[0].State)
		assert.Equal(t, "online", leaves[1].State)
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

}
