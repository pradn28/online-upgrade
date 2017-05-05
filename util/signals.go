package util

import (
	"fmt"
	"os"
	"os/signal"
)

// CatchSignals will setup Notify to relay specific OS signals
// to a specified channel and handle the signals that are on the channel
func CatchSignals(c chan os.Signal, s os.Signal) {
	// Setup notify to add signal to the channel
	signal.Notify(c, s)

	// TODO: Provide option to re-setup notify.
	// Currently if the user hits control-c we just simply stop watching for signals.
	// If the user hits control-c anytime later in the script, we will exit
	// without prompting.

	// Watch channel for signals and prompt user
	go func() {
		<-c
		fmt.Println("\nPress control-c again to exit")
		// Stop notifying and allow next signal to pass through
		signal.Stop(c)
	}()
}
