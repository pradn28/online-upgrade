package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// GetUserConfirmation waits for user to confirm
func GetUserConfirmation(message, instruction, confirmString string) {
	// Grab any config information specified on the command line
	// If no config information is passed in, we will use the defaults in 'config.go'
	config := ParseFlags()
	// Setup reader to watch stdin for user input
	reader := bufio.NewReader(os.Stdin)

	// Setup divider based on output width from config
	dividerLength := config.OutputWidth - 2
	divider := strings.Repeat("-", dividerLength)
	divider = fmt.Sprintf("+%s+", divider)

	// Get string with new lines at spcified width.
	message = LineWrapper(message, config.OutputWidth)

	// Print Message and Instructions to continue.
	fmt.Println(divider)
	fmt.Print(message)
	fmt.Println(divider)

	fmt.Print(instruction)
	for {
		response, _ := reader.ReadString('\n')
		response = strings.TrimRight(response, "\n")

		if response == confirmString {
			break
		} else if response != "" {
			fmt.Printf("You entered \"%s\"\n", response)
			fmt.Printf("You must type %s to continue: ", confirmString)
		} else {
			fmt.Printf("You must type %s to continue: ", confirmString)
		}
	}
}
