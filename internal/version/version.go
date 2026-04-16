package version

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Short() string {
	return Version
}

func Full() string {
	return Version + " (" + Commit + ") built " + Date
}
