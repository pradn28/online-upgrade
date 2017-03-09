package steps_test

import (
	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestUpdateConfig(t *testing.T) {
	// Build test cluster
	defer testutil.ClusterInABox(t)()

	// Create a testing database
	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, "testing")()

	// Run update config to set variables to OFF
	t.Run("UpdateConfig", func(t *testing.T) {
		err := steps.UpdateConfig("OFF")
		assert.Nil(t, err)
	})

	// Verify that update config actually set the variables
	t.Run("DBGetVariable", func(t *testing.T) {
		varval, err := util.DBGetVariable("auto_attach")
		assert.Nil(t, err)
		assert.Equal(t, "OFF", varval)
	})

	// Run again to double check by setting to ON
	t.Run("UpdateConfig", func(t *testing.T) {
		err := steps.UpdateConfig("ON")
		assert.Nil(t, err)
	})

	// Verify again that the variable was set back to ON
	t.Run("DBGetVariable", func(t *testing.T) {
		varval, err := util.DBGetVariable("auto_attach")
		assert.Nil(t, err)
		assert.Equal(t, "ON", varval)
	})
}
