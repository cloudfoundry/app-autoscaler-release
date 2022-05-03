package main

import (
	"changelog/display"
	"changelog/github"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	owner  = "cloudfoundry"
	repo   = "app-autoscaler-release"
	branch = "main"
)

type cliOpts struct {
	token                  string // would like usage: `names:"-t, --token" usage:"Github token"`
	previousReleaseTagName string // would like usage: `names:"-n, --prev-rel-tag" usage:"Tag name of the previous release"`
	latestCommitSHAId      string // would like usage: `names:"-h, --latest-commit-sha-id" usage:"SHA id of the latest commit to include in the release"`
	changelogFile          string // would like usage: `names:"-o, --changelog-file" usage:"Output file to write the changelog into"`
	nextReleaseTagNameFile string // would like usage: `names:"-v, --version-file" usage:"Output file for the tag name of the next release"`
}

func parseCliOpts() cliOpts {
	var opts cliOpts

	// ToDo: Remove access to env-variable which was for purposes of an easy transition.
	flag.StringVar(&opts.token, "token", os.Getenv("GITHUB_TOKEN"), "Github token")
	flag.StringVar(&opts.previousReleaseTagName, "prev-rel-tag", os.Getenv("PREVIOUS_VERSION"),
		"Tag name of the previous release")
	flag.StringVar(&opts.latestCommitSHAId, "last-commit-sha-id", os.Getenv("GIT_COMMIT_SHA_ID"),
		"SHA id of the last commit to include into the release")
	flag.StringVar(&opts.changelogFile, "changelog-file", os.Getenv("OUTPUT_FILE"),
		"Output file to write the changelog into")
	flag.StringVar(&opts.nextReleaseTagNameFile, "version-file", os.Getenv("RECOMMENDED_VERSION_FILE"),
		"Output file for the tag name of the next release")

	flag.Parse()

	return opts
}

func main() {
	opts := parseCliOpts()

	client := github.New(opts.token)
	mapReleaseSHAIdToReleaseTagName, err := client.FetchSHAIDsOfReleases(owner, repo)
	if err != nil {
		panic(err)
	}

	// ToDo: Refactor in own function?
	var previousReleaseSHAId string
	for releaseSHAId, releaseTagName := range mapReleaseSHAIdToReleaseTagName {
		releaseTagNameIsPreviousRelease := releaseTagName == "v"+opts.previousReleaseTagName ||
			releaseTagName == opts.previousReleaseTagName
		if releaseTagNameIsPreviousRelease {
			previousReleaseSHAId = releaseSHAId
		}
	}

	var allPullRequests []github.PullRequest

	shaForPreviousReleaseFound := previousReleaseSHAId != ""
	if shaForPreviousReleaseFound {
		prs, err := client.FetchPullRequests(owner, repo, branch, previousReleaseSHAId,
			opts.latestCommitSHAId)
		if err != nil {
			panic(err)
		}
		allPullRequests = append(allPullRequests, prs...)
	}

	changelog, nextVersion, err := display.GenerateOutput(allPullRequests, opts.previousReleaseTagName)
	if err != nil {
		panic(err)
	}

	if opts.changelogFile != "" {
		err := ioutil.WriteFile(opts.changelogFile, []byte(changelog), 0600)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println(changelog)
	}

	if opts.nextReleaseTagNameFile != "" {
		err := ioutil.WriteFile(opts.nextReleaseTagNameFile, []byte(nextVersion), 0600)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println(nextVersion)
	}

	fmt.Printf("Total PRs %d\n", len(allPullRequests))
}
