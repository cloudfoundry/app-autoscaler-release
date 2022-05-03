package main

import (
	"bytes"
	"changelog/display"
	"changelog/github"
	"flag"
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

type cliOpts struct {
	token                  string `names:"-t, --token" usage:"Github token"`
	previousReleaseTagName string `names:"-n, --prev-rel-tag" usage:"Tag name of the previous release"`
	outputFile             string `names:"-o, --out-file" usage:"Output file"`
	nextReleaseTagNameFile string `names:"-v, --verion-file" usage:"Output file for the tag name of the next release"`
}

func parseCliOpts(args []string) (cliOpts, error) {
	var opts cliOpts
	var t *flag.FlagSet = flag.CommandLine
	err := t.ParseStruct(&opts, args...)

	return opts, nil
}

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
