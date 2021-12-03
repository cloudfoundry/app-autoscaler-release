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
	outputFile := os.Getenv("OUTPUT_FILE")
	recommendedVersionFile := os.Getenv("RECOMMENDED_VERSION_FILE")

	commitsFromReleases, err := client.FetchCommitsFromReleases(owner, repo)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("all releases %+v", commitsFromReleases )

	var commit string
	for k, v := range commitsFromReleases {
		// e.g. search for 1.0.0 or v1.0.0
		if v == "v"+previousVersion || v == previousVersion {
			commit = k
		}
	}

	var allPullRequests []github.PullRequest

	if commit != "" {
		latestCommitSHA, err := client.FetchLatestReleaseCommitFromBranch(owner, repo, branch, commitsFromReleases)
		if err != nil {
			panic(err)
		}

		prs, err := client.FetchPullRequestsAfterCommit(owner, repo, branch, commit, latestCommitSHA)
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
