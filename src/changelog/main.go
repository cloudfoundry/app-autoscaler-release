package main

import (
	"changelog/display"
	"changelog/github"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	owner  string = "cloudfoundry"
	repo   string = "app-autoscaler-release"
	branch string = "main"
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

	var commit string
	for k, v := range commitsFromReleases {
		if v == "v"+previousVersion {
			commit = k
		} else if v == previousVersion {
			commit = k
		}
	}

	latestCommitSHA, err := client.FetchLatestReleaseCommitFromBranch(owner, repo, branch, commitsFromReleases)
	if err != nil {
		panic(err)
	}

	prs, err := client.FetchPullRequestsAfterCommit(owner, repo, branch, commit, latestCommitSHA)
	if err != nil {
		panic(err)
	}

	// get PRs from app-autoscaler too
	//otherPRs, err := client.FetchPullRequestsAfterCommit(owner, "app-autoscaler", branch, commit, latestCommitSHA)
	//if err != nil {
	//	panic(err)
	//}

	//for _, pr := range otherPRs {
	//	fmt.Printf("Got %s\n", pr.Url)
	//}

	//fmt.Printf("Got %d other prs\n", len(otherPRs))

	if latestCommitSHA == "" {
		prs = filterPrs(prs)
	}

	changelog, nextVersion, err := display.GenerateOutput(prs, previousVersion)
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
}

func filterPrs(prs []github.PullRequest) []github.PullRequest {
	var filtered []github.PullRequest
	for _, pr := range prs {
		if pr.Number > 245 {
			filtered = append(filtered, pr)
		}
	}
	return filtered
}
