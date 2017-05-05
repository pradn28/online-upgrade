package steps

import (
	"errors"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/memsql/online-upgrade/util"
)

// Set Custom error messages
var (
	ErrMasterID      = errors.New("update_config: Failed to retreive Master [masterID]")
	ErrAggregatorIDs = errors.New("update_config: Failed to retreive Aggregator [memsqlIDs]")
)

// UpdateConfig updates MemSQL database configuration
// Use UpdateConfig to set specific variable values required for upgrading
// auto_attach, aggregator_failure_detection, leaf_failure_detection
func UpdateConfig(state string) error {

	v := []string{"auto_attach", "aggregator_failure_detection", "leaf_failure_detection"}
	// Create new spinner to show activity of snapshotting
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	// For each variable update config on the MASTER with the correct state
	masterID, err := util.GetNodeIDs("MASTER")
	if err != nil || len(masterID) != 1 {
		return ErrMasterID
	}

	for _, variableName := range v {
		// Configure spinner
		s.Prefix = " "
		s.Suffix = fmt.Sprintf(" Updating %s to %s on MASTER Aggregator", variableName, state)
		s.FinalMSG = fmt.Sprintf(" ✓ [%s] %s on MASTER\n", state, variableName)
		s.Start()

		// Update configs
		err := util.OpsMemsqlUpdateConfig(masterID[0], variableName, state, "--set-global")
		if err != nil {
			return err
		}
		s.Stop()

	}
	fmt.Println("MASTER Aggregator config updated")

	// Get child aggregator IDs
	variableName := "aggregator_failure_detection"
	memsqlIDs, err := util.GetNodeIDs("AGGREGATOR")
	if err != nil {
		return ErrAggregatorIDs
	}

	// For each child AGGREGATOR update aggregator_failure_detection
	for i := range memsqlIDs {
		aggID := memsqlIDs[i]

		// Configure spinner
		s.Prefix = " "
		s.Suffix = fmt.Sprintf(" Updating %s to %s on CHILD AGGREGATOR [%s]", variableName, state, aggID[0:7])
		s.FinalMSG = fmt.Sprintf(" ✓ [%s] %s on AGGREGATOR [%s]\n", state, variableName, aggID[0:7])
		s.Start()
		// For each CA update the config
		// TODO: store existing settings and reset to previous
		err := util.OpsMemsqlUpdateConfig(aggID, variableName, state, "--set-global")
		if err != nil {
			return err
		}
		s.Stop()

	}
	fmt.Println("CHILD AGGREGATOR configs updated")

	return nil
}
