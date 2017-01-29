package util

import (
	"fmt"
	"log"
	"os"
)

func SetupLogging(config Config) (func(), error) {
	f, err := os.OpenFile(config.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("Opening log file")

	return func() {
		err := f.Close()
		if err != nil {
			fmt.Printf("Failed to close log file")
		}
	}, nil
}
