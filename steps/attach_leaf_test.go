package steps_test

import (
	"testing"

	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAttachLeaves will test the attaching of leaves
// in a specified availablity group.
func TestAttachLeaves(t *testing.T) {
	// Build an HA test cluster
	defer testutil.ClusterHA(t)()
	// Create a testing database
	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")()

	// Start with detaching leaves in AG 1
	t.Run("DetachLeaves", func(t *testing.T) {
		err := steps.DetachLeaves(1)
		assert.Nil(t, err)
	})

	// Check show leaves to validate leaves were detached
	t.Run("CheckShowLeaves", func(t *testing.T) {
		leaves, _ := util.DBShowLeaves()
		assert.Len(t, leaves, 2)
		assert.Equal(t, "detached", leaves[0].State)
		assert.Equal(t, "online", leaves[1].State)
	})

	// Now lets attach and check again
	t.Run("AttachLeaves", func(t *testing.T) {
		err := steps.AttachLeaves(1)
		assert.Nil(t, err)
	})

	// Check show leaves to validate leaves are attched
	t.Run("CheckShowLeaves", func(t *testing.T) {
		leaves, _ := util.DBShowLeaves()
		assert.Len(t, leaves, 2)
		assert.Equal(t, "online", leaves[0].State)
		assert.Equal(t, "online", leaves[1].State)
	})

}
