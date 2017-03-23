package steps

import (
	"fmt"
	"log"
	"time"

	"github.com/briandowns/spinner"
	"github.com/memsql/online-upgrade/util"
)

// AttachLeaves will attach all leaves in a specifed availability group.
func AttachLeaves(group int) error {
	// Create new spinner to show activity of attach
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	// Get a list of leaves
	leaves, err := util.DBShowLeaves()
	if err != nil {
		return err
	}
	leafCount := len(leaves)

	// Attach leaves in a specified AG (host, port)
	for i := range leaves {
		l := leaves[i]
		// If leaf is in the specified AG, attach it
		if l.AG == group {
			// Configure spinner
			s.Prefix = " "
			s.Suffix = fmt.Sprintf(" Attaching %s:%d", l.Host, l.Port)
			s.FinalMSG = fmt.Sprintf(" ✓ Attach complete for %s:%d\n", l.Host, l.Port)
			s.Start()
			// Attach now
			attachErr := util.DBAttachLeaf(l.Host, l.Port)
			if err != nil {
				return attachErr
			}
			log.Printf("Attached %s:%d", l.PairHost, l.PairPort)
			s.Stop()
		}
	}

	// Check all leaves are back online
	// Configure spinner
	s.Prefix = " "
	s.Suffix = fmt.Sprintf(" Waiting for all %d leaves to be online", leafCount)
	s.FinalMSG = fmt.Sprintf(" ✓ All %d leaves are now online\n", leafCount)
	s.Start()
	// Wait for leaves to be online and connected
	attachWaitErr := util.OpsWaitMemsqlsOnlineConnected(leafCount)
	if err != nil {
		return attachWaitErr
	}
	log.Printf("All %d leaves are online", leafCount)
	s.Stop()

	return nil
}
