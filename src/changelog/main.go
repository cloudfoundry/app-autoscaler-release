package main

import (
	"changelog/display"
	"changelog/github"
	"changelog/util"
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

	skipSubmodule := os.Getenv("SKIP_SUBMODULE")
	if skipSubmodule != "true" {
		fmt.Printf("Also querying submodule for changes...\n")
		//git ls-tree HEAD src/app-autoscaler | awk '{print $3}'
		toSha, err := util.GetShaOfSubmoduleAtCommit("HEAD")
		if err != nil {
			panic(err)
		}

		if toSha != "" {
			fromSha, err := util.GetShaOfSubmoduleAtCommit(commit)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Checking from %s to %s\n", fromSha, toSha)
			if fromSha != toSha {
				// get PRs from app-autoscaler too
				otherPRs, err := client.FetchPullRequestsAfterCommit(owner, "app-autoscaler", branch, fromSha, toSha)
				if err != nil {
					panic(err)
				}

				prs = append(prs, otherPRs...)
			}
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
