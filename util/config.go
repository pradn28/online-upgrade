package util

import "flag"

type Config struct {
	MasterHost string
	MasterPort int
	MasterUser string
	MasterPass string
	LogPath    string
}

var configData = Config{
	MasterHost: "127.0.0.1",
	MasterPort: 3306,
	MasterUser: "root",
	MasterPass: "",
	LogPath:    "online-upgrade.log",
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
}

func ParseFlags() Config {
	flag.Parse()
	return configData
}
