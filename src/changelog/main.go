package main

import (
	"bytes"
	"changelog/display"
	"changelog/github"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
		latestCommitSHA := localGitRepoFetchLatestCommitSHAId(branch)

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

func localGitRepoFetchLatestCommitSHAId(branchName string) string {
	rev := fmt.Sprintf("origin/%s", branchName)
	gitRevParseCmd := exec.Command("git", "rev-parse", rev)

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	gitRevParseCmd.Stdout = &stdOut
	gitRevParseCmd.Stderr = &stdErr

	if err := gitRevParseCmd.Run(); err != nil {
		log.Fatalf("failed to get SHA-ID of latest git-commit: %s\n\t%s", err, stdErr.String())
	}

	return stdOut.String()
}
