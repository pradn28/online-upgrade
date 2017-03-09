package steps

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/memsql/online-upgrade/util"
)

// SnapshotDatabases will snapshot all user databases prior to upgrade
func SnapshotDatabases() error {
	dbs, err := util.DBGetUserDatabases()
	if err != nil {
		return err
	}
	// Create new spinner to show activity of snapshotting
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	for i := range dbs {
		db := dbs[i]

		// Configure spinner
		s.Prefix = " "
		s.Suffix = fmt.Sprintf(" Snapshotting  %s", db)
		s.FinalMSG = fmt.Sprintf(" ✓ Snapshot complete for %s\n", db)
		s.Start()

		err := util.DBSnapshotDatabase(db)

		if err != nil {
			return fmt.Errorf("Failed to snapshot database %s: %s", db, err)
		}

		s.Stop()
	}
	return nil
}
