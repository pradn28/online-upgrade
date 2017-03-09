package steps_test

import (
	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestPreUpgrade(t *testing.T) {
	// Build test cluster
	defer testutil.ClusterInABox(t)()

	// Create a testing database
	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")()

	// Set redundancy level prior to testing
	// PreUpgrade checks the redundancy level. Upgrade requires HA.
	require.Nil(t, util.DBSetVariable("SET @@GLOBAL.redundancy_level = 2"))

	// Run pre-upgrade check
	t.Run("PreUpgrade", func(t *testing.T) {
		err := steps.PreUpgrade()
		assert.Nil(t, err)
	})
}
