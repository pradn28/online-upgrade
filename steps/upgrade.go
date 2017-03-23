package steps

import (
	"fmt"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/memsql/online-upgrade/util"
)

var s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)

// UpgradeLeaves will upgrade leaves in a specified AG
func UpgradeLeaves(availabilityGroup int) error {
	// Get a list of MemSQL nodes
	memsqls, err := util.OpsMemsqlList("--availability-group", strconv.Itoa(availabilityGroup))
	if err != nil {
		return err
	}
	// For each leaf in the specified AG,
	// upgrade it to the specified version or latest
	for i := range memsqls {
		m := memsqls[i]
		err := upgradeMemsql(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpgradeAggregators will upgrade all Child Aggregators
func UpgradeAggregators() error {
	// Get a list of MemSQL nodes
	memsqls, err := util.OpsMemsqlList("--memsql-role", "AGGREGATOR")
	if err != nil {
		return err
	}
	// For each Aggregator, upgrade it to the specified version or latest
	for i := range memsqls {
		m := memsqls[i]
		err := upgradeMemsql(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpgradeMaster upgrades the Master Aggregator
func UpgradeMaster() error {
	// Get a list of MemSQL nodes
	memsqls, err := util.OpsMemsqlList("--memsql-role", "MASTER")
	if err != nil {
		return err
	}
	// Upgrade the MA to the specified version or latest
	for i := range memsqls {
		m := memsqls[i]
		err := upgradeMemsql(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func upgradeMemsql(m util.MemsqlInfo) error {

	currentVersion, err := util.OpsMemsqlGetVersion(m.MemsqlID)
	if err != nil {
		return err
	}

	// Stop MemSQL
	s.Prefix = " "
	s.Suffix = fmt.Sprintf(" Stopping %s (%s:%d)", m.MemsqlID[0:7], m.Host, m.Port)
	s.FinalMSG = fmt.Sprintf(" ✓ [STOPPED] %s (%s:%d)\n", m.MemsqlID[0:7], m.Host, m.Port)
	s.Start()
	stopErr := util.OpsNodeManagement("memsql-stop", m.MemsqlID, "--no-prompt")
	if stopErr != nil {
		return stopErr
	}
	s.Stop()

	// Start Upgrade
	s.Suffix = fmt.Sprintf(" Upgrading %s from version %s",
		m.MemsqlID[0:7],
		currentVersion,
	)
	s.Start()
	upgradeErr := util.OpsMemsqlUpgrade(
		m.MemsqlID,
		"--no-prompt",
		"--skip-snapshot",
		"--no-backup-data-directories",
	)
	if upgradeErr != nil {
		return upgradeErr
	}

	// Check Version
	newVersion, err := util.OpsMemsqlGetVersion(m.MemsqlID)
	if err != nil {
		return err
	}

	s.FinalMSG = fmt.Sprintf(" ✓ [UPGRADED] %s now running version: %s\n",
		m.MemsqlID[0:7],
		newVersion,
	)
	s.Stop()

	// Start MemSQL
	// Technically not required as the upgrade starts the leaf. But just to be sure.
	s.Suffix = fmt.Sprintf(" Starting %s (%s:%d)", m.MemsqlID[0:7], m.Host, m.Port)
	s.FinalMSG = fmt.Sprintf(" ✓ [STARTED] %s (%s:%d)\n", m.MemsqlID[0:7], m.Host, m.Port)

	s.Start()
	startErr := util.OpsNodeManagement("memsql-start", m.MemsqlID, "--no-prompt")
	if err != nil {
		return startErr
	}
	s.Stop()

	return err
}
