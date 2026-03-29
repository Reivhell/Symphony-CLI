package version

// Version is the running Symphony CLI semver string (e.g. "0.1.0", "dev").
// main.go sets these from cmd before cmd.Execute().
var (
	Version   = "0.1.0"
	Commit    = "none"
	BuildDate = "unknown"
)
