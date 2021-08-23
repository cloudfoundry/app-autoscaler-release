package main

import (
	"changelog/display"
	"changelog/github"
	"changelog/util"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

	// fmt.Printf("commitsFromReleases=%+v\n", commitsFromReleases)

	latestCommitSHA, err := client.FetchLatestReleaseCommitFromBranch(owner, repo, branch, commitsFromReleases)
	if err != nil {
		panic(err)
	}

	prs, err := client.FetchPullRequestsAfterCommit(owner, repo, branch, commit, latestCommitSHA)
	if err != nil {
		panic(err)
	}

	skipSubmodule := os.Getenv("SKIP_SUBMODULE")
	if skipSubmodule != "true" {
		//git ls-tree HEAD src/app-autoscaler | awk '{print $3}'
		toSha, err := getShaOfSubmoduleAtCommit("HEAD")
		if err != nil {
			panic(err)
		}

		fromSha, err := getShaOfSubmoduleAtCommit(commit)
		if err != nil {
			panic(err)
		}

		if fromSha != toSha {
			// get PRs from app-autoscaler too
			otherPRs, err := client.FetchPullRequestsAfterCommit(owner, "app-autoscaler", branch, fromSha, toSha)
			if err != nil {
				panic(err)
			}

			prs = append(prs, otherPRs...)
		}
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

	fmt.Printf("Total PRs %d\n", len(prs))
}

func getShaOfSubmoduleAtCommit(commit string) (string, error) {
	cmd := util.Command{Name: "git", Args: []string{"ls-tree", commit, "../app-autoscaler"}}
	runner := util.DefaultCommandRunner{}

	output, err := runner.RunWithoutRetry(&cmd)
	if err != nil {
		return "", err
	}

	parts := strings.Split(output, " ")
	return parts[2], nil
}
