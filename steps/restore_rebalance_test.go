package steps_test

import (
	"testing"

	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRestoreRedundancy test redundancy was able to be restored
func TestRestoreRedundancy(t *testing.T) {
	// Build test cluster
	defer testutil.ClusterHA(t)()

	// Create a testing database
	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "online")()
	defer testutil.CreateDatabase(t, "upgrade")()

	// Run restore redundancy on all user databases
	t.Run("RestoreRedundancy", func(t *testing.T) {
		err := steps.RestoreRedundancy()
		assert.Nil(t, err)
	})

	// Run rebalance partitions on all user databases
	t.Run("RebalancePartitions", func(t *testing.T) {
		err := steps.RebalancePartitions()
		assert.Nil(t, err)
	})

}
