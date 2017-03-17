package steps_test

import (
	"testing"

	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetachLeaves will test the detaching of leaves
// in a specified availablity group. This test requires
// a High Availability Cluster
func TestDetachLeaves(t *testing.T) {
	// Build an HA test cluster
	defer testutil.ClusterHA(t)()
	// Create a testing database
	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")()

	// Detach leaves in AG 1
	t.Run("DetachLeaves", func(t *testing.T) {
		err := steps.DetachLeaves(1)
		assert.Nil(t, err)
	})

	// Check show leaves to validate leaves were detached
	t.Run("CheckShowLeaves", func(t *testing.T) {
		leaves, _ := util.DBShowLeaves()
		assert.Len(t, leaves, 4)
		assert.Equal(t, "detached", leaves[0].State)
		assert.Equal(t, "detached", leaves[1].State)
		assert.Equal(t, "online", leaves[2].State)
		assert.Equal(t, "online", leaves[3].State)
	})

}
