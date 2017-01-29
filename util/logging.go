package util

import (
	"log"
)

func SetupLogging() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
