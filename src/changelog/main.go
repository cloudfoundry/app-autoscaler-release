package main

import (
	"changelog/display"
	"changelog/github"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	owner  = "cloudfoundry"
	repo   = "app-autoscaler-release"
	branch = "main"
)

func main() {
	// FIXME these should be flags
	client := github.New(os.Getenv("GITHUB_TOKEN"))
	previousVersion := os.Getenv("PREVIOUS_VERSION")
	latestVersion := os.Getenv("GIT_COMMIT_SHA_ID")
	outputFile := os.Getenv("OUTPUT_FILE")
	recommendedVersionFile := os.Getenv("RECOMMENDED_VERSION_FILE")

	mapReleaseSHAIdToReleaseTagName, err := client.FetchSHAIDsOfReleases(owner, repo)
	if err != nil {
		panic(err)
	}

	// ToDo: Refactor in own function?
	var previousReleaseSHAId string
	for releaseSHAId, releaseTagName := range mapReleaseSHAIdToReleaseTagName {
		// e.g. search for 1.0.0 or v1.0.0
		if releaseTagName == "v"+previousVersion || releaseTagName == previousVersion {
			previousReleaseSHAId = releaseSHAId
		}
	}

	var allPullRequests []github.PullRequest

	shaForPreviousReleaseFound := previousReleaseSHAId != ""
	if shaForPreviousReleaseFound {
		latestCommitSHA := latestVersion

		prs, err := client.FetchPullRequestsAfterCommit(owner, repo, branch, previousReleaseSHAId, latestCommitSHA)
		if err != nil {
			panic(err)
		}
		allPullRequests = append(allPullRequests, prs...)
	}

	changelog, nextVersion, err := display.GenerateOutput(allPullRequests, previousVersion)
	if err != nil {
		panic(err)
	}

	if outputFile != "" {
		err := ioutil.WriteFile(outputFile, []byte(changelog), 0600)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println(changelog)
	}

	if recommendedVersionFile != "" {
		err := ioutil.WriteFile(recommendedVersionFile, []byte(nextVersion), 0600)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println(nextVersion)
	}

	fmt.Printf("Total PRs %d\n", len(allPullRequests))
}
