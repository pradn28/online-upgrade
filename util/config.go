package util

import "flag"

// Config struct for commandline flags
type Config struct {
	MasterHost       string
	MasterPort       int
	MasterUser       string
	MasterPass       string
	LogPath          string
	VersionHash      string
	SkipVersionCheck bool
	OutputWidth      int
}

var configData = Config{
	MasterHost:       "127.0.0.1",
	MasterPort:       3306,
	MasterUser:       "root",
	MasterPass:       "",
	LogPath:          "online-upgrade.log",
	VersionHash:      "",
	SkipVersionCheck: false,
	OutputWidth:      80,
}

func init() {
	flag.StringVar(
		&configData.MasterHost, "host", configData.MasterHost,
		"The Master Aggregator host")

	flag.IntVar(
		&configData.MasterPort, "port", configData.MasterPort,
		"The Master Aggregator port")

	flag.StringVar(
		&configData.MasterUser, "user", configData.MasterUser,
		"The Master Aggregator user")

	flag.StringVar(
		&configData.MasterPass, "password", configData.MasterPass,
		"The Master Aggregator password")

	flag.StringVar(
		&configData.LogPath, "log-path", configData.LogPath,
		"Where to write out the full audit log")

	flag.StringVar(
		&configData.VersionHash, "version-hash", configData.VersionHash,
		"Version hash for MemSQL version you want to upgrade to")

	flag.BoolVar(
		&configData.SkipVersionCheck, "skip-version-check", configData.SkipVersionCheck,
		"Skip version check during upgrade")

	flag.IntVar(
		&configData.OutputWidth, "output-width", configData.OutputWidth,
		"The output column width for messages. Default:80")

}

// ParseFlags parses command line args
func ParseFlags() Config {
	flag.Parse()
	return configData
}
