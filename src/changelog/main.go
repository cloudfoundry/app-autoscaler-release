package main

import (
	"changelog/display"
	"changelog/github"
	"flag"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager/v3"
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
	var logger = lager.NewLogger("changelog")

	opts := parseCliOpts()

	client := github.New(opts.token, opts.owner, opts.repo, opts.branch)
	tagsToSha, err := client.GetTagsToSha()
	if err != nil {
		logger.Fatal("failed connect with github client", err, lager.Data{"owner": opts.owner, "repo": opts.repo, "branch": opts.branch})
	}

	// ToDo: Refactor in own function?
	previousReleaseSHAId := tagsToSha[opts.previousReleaseTagName]
	oldNameing := tagsToSha["v"+opts.previousReleaseTagName]
	if oldNameing != "" {
		previousReleaseSHAId = oldNameing
	}
	if previousReleaseSHAId == "" {
		logger.Fatal("could not find the previous release", nil, lager.Data{"previousReleaseTag": opts.previousReleaseTagName})
	}

	var allPullRequests []github.PullRequest

	prs, err := client.FetchPullRequests(previousReleaseSHAId, opts.latestCommitSHAId)
	if err != nil {
		logger.Fatal("failed to get pull requests", err, lager.Data{"previousReleaseSHAId": previousReleaseSHAId, "latestCommitSHAId": opts.latestCommitSHAId})
	}
	allPullRequests = append(allPullRequests, prs...)

	changelog, nextVersion, err := display.GenerateOutput(allPullRequests, opts.previousReleaseTagName)
	if err != nil {
		logger.Fatal("failed to generate md file", nil, lager.Data{"numberPRs": len(allPullRequests), "previousReleaseTag": opts.previousReleaseTagName})
	}

	if opts.changelogFile != "" {
		err := os.WriteFile(opts.changelogFile, []byte(changelog), 0600)
		if err != nil {
			logger.Fatal("failed write md file", err, lager.Data{"file": opts.changelogFile})
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
