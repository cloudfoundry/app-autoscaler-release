package cf

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

const (
	ProductName = "app-autoscaler"
)

// getBuildInfo extracts git repository URL and commit ID from build info
func getBuildInfo() (repoURL, commitID string) {
	repoURL = "unknown"
	commitID = "unknown"

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.remote":
			repoURL = setting.Value
		case "vcs.revision":
			commitID = setting.Value
		}
	}

	return
}

// GetUserAgent returns a custom HTTP User-Agent string in the format:
// app-autoscaler/{version} ({gitRepo}; {commitId}) Go/{goVersion} {os}/{arch}
func GetUserAgent() string {
	repoURL, commitID := getBuildInfo()

	version := commitID
	if version == "unknown" || version == "" {
		version = "dev"
	}

	systemInfo := fmt.Sprintf("%s; %s", repoURL, commitID)
	platformInfo := fmt.Sprintf("Go/%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	return fmt.Sprintf("%s/%s (%s) %s", ProductName, version, systemInfo, platformInfo)
}
