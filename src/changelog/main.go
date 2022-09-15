package main

import (
	"changelog/display"
	"changelog/github"
	"flag"
	"fmt"
	"log"
	"os"
)

type cliOpts struct {
	token                  string // would like usage: `names:"-t, --token" usage:"Github token"`
	owner                  string // would like usage: `names:"-o, --owner" usage:"Repository owner"`
	repo                   string // would like usage: `names:"-r, --repo" usage:"Repository name"`
	branch                 string // would like usage: `names:"-b, --branch" usage:"branch name"`
	previousReleaseTagName string // would like usage: `names:"-n, --prev-rel-tag" usage:"Tag name of the previous release"`
	latestCommitSHAId      string // would like usage: `names:"-h, --latest-commit-sha-id" usage:"SHA id of the latest commit to include in the release"`
	changelogFile          string // would like usage: `names:"-o, --changelog-file" usage:"Output file to write the changelog into"`
	nextReleaseTagNameFile string // would like usage: `names:"-v, --version-file" usage:"Output file for the tag name of the next release"`
}

func parseCliOpts() cliOpts {
	var opts cliOpts

	// ToDo: Remove access to env-variable which was for purposes of an easy transition.
	flag.StringVar(&opts.token, "token", os.Getenv("GITHUB_TOKEN"), "Github token")
	flag.StringVar(&opts.owner, "owner", "cloudfoundry", "Repository owner")
	flag.StringVar(&opts.repo, "repo", "app-autoscaler-release", "Repository owner")
	flag.StringVar(&opts.branch, "branch", "main", "Branch the release notes are for")
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

	client := github.New(opts.token, opts.owner, opts.repo, opts.branch)
	tagsToSha, err := client.GetTagsToSha()
	if err != nil {
		log.Fatalf("ERROR: Failed connect with github client:%s", err.Error())
	}

	// ToDo: Refactor in own function?
	previousReleaseSHAId := tagsToSha[opts.previousReleaseTagName]
	oldNameing := tagsToSha["v"+opts.previousReleaseTagName]
	if oldNameing != "" {
		previousReleaseSHAId = oldNameing
	}
	if previousReleaseSHAId == "" {
		log.Fatalf("ERROR: Could not find the previous release sha for tag: '%s'", opts.previousReleaseTagName)
	}

	var allPullRequests []github.PullRequest

	prs, err := client.FetchPullRequests(previousReleaseSHAId, opts.latestCommitSHAId)
	if err != nil {
		log.Fatalf("ERROR: failed to get pull requests: '%s'", opts.previousReleaseTagName)
	}
	allPullRequests = append(allPullRequests, prs...)

	changelog, nextVersion, err := display.GenerateOutput(allPullRequests, opts.previousReleaseTagName)
	if err != nil {
		log.Fatalf("ERROR: failed to generate md file: '%s'", opts.previousReleaseTagName)
	}

	if opts.changelogFile != "" {
		err := os.WriteFile(opts.changelogFile, []byte(changelog), 0600)
		if err != nil {
			log.Fatalf("ERROR: failed write md file '%s':%s", opts.previousReleaseTagName, err)
		}
	} else {
		fmt.Println(changelog)
	}

	if opts.nextReleaseTagNameFile != "" {
		err := os.WriteFile(opts.nextReleaseTagNameFile, []byte(nextVersion), 0600)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println(nextVersion)
	}

	fmt.Printf("Total PRs %d\n", len(allPullRequests))
}
