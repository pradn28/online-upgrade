package util

import (
	"strings"
)

// LineWrapper returns a string with line breaks at the specified line length
func LineWrapper(s string, maxLength int) string {

	// Convert incoming string to a slice
	inSlice := strings.Fields(s)

	allSlices := [][]string{}
	outSlice := []string{}
	lineLength := 0

	for i := range inSlice {
		// Get the length of the word and +1 for the space after the word
		wordLength := len(inSlice[i]) + 1
		// Add word to the slice if it does not exceed the max length
		if (lineLength + wordLength) <= maxLength {
			lineLength += wordLength
			outSlice = append(outSlice, inSlice[i])
		} else {
			// If the word would cause the line to exceed the maxLength,
			// add outSlice to allSlices and start new outSlice with this word
			lineLength = wordLength
			allSlices = append(allSlices, outSlice)
			outSlice = []string{inSlice[i]}
		}
	}

	// Add remaining words to allSlices
	allSlices = append(allSlices, outSlice)

	// Create string from each slice and add newlines
	var outString string
	for x := range allSlices {
		eachString := strings.Join(allSlices[x], " ")
		outString += eachString
		outString += "\n"
	}

	return outString
}
