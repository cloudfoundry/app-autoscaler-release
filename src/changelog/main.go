package main

import (
	"changelog/github"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var (
	owner  string = "cloudfoundry"
	repo   string = "app-autoscaler-release"
	branch string = "main"
)

func main() {
	var sb strings.Builder
	sb.WriteString("# Changelog for app-autoscaler-release\n\n")

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

	latestReleaseCommit, err := client.FetchLatestReleaseCommitFromBranch(owner, repo, branch, commitsFromReleases)
	if err != nil {
		panic(err)
	}

	prs, err := client.FetchPullRequestsAfterCommit(owner, repo, branch, commit, latestReleaseCommit, []string{})
	if err != nil {
		panic(err)
	}

	var breakingChanges []github.PullRequest
	var enhancements []github.PullRequest
	var bugFixes []github.PullRequest
	var dependencyUpdates []github.PullRequest
	var chores []github.PullRequest
	var other []github.PullRequest

	for _, pr := range prs {
		if latestReleaseCommit != "" || pr.Number > 245 {
			if ArrayContains(pr.Labels, "breaking-change") {
				breakingChanges = append(breakingChanges, pr)
			} else if ArrayContains(pr.Labels, "enhancement") {
				enhancements = append(enhancements, pr)
			} else if ArrayContains(pr.Labels, "dependencies") {
				dependencyUpdates = append(dependencyUpdates, pr)
			} else if ArrayContains(pr.Labels, "bug") {
				bugFixes = append(bugFixes, pr)
			} else if ArrayContains(pr.Labels, "chore") {
				chores = append(chores, pr)
			} else {
				other = append(other, pr)
			}
		}
	}

	if len(breakingChanges) > 0 {
		Header(&sb, "Breaking Changes")
		DisplayPRs(&sb, breakingChanges)
	}

	if len(enhancements) > 0 {
		Header(&sb, "Enhancements")
		DisplayPRs(&sb, enhancements)
	}

	if len(bugFixes) > 0 {
		Header(&sb, "Bug Fixes")
		DisplayPRs(&sb, bugFixes)
	}

	if len(chores) > 0 {
		Header(&sb, "Chores")
		DisplayPRs(&sb, chores)
	}

	if len(dependencyUpdates) > 0 {
		Header(&sb, "Dependency Updates")
		DisplayPRs(&sb, dependencyUpdates)
	}

	if len(other) > 0 {
		Header(&sb, "Other")
		DisplayPRs(&sb, other)
	}

	if outputFile != "" {
		err := ioutil.WriteFile(outputFile, []byte(sb.String()), 0600)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println(sb.String())
	}

	v, err := semver.NewVersion(previousVersion)
	if err != nil {
		panic(err)
	}

	var recommendedVersion semver.Version
	if len(breakingChanges) > 0 {
		recommendedVersion = v.IncMajor()
	} else if len(enhancements) > 0 {
		recommendedVersion = v.IncMinor()
	} else {
		recommendedVersion = v.IncPatch()
	}

	if recommendedVersionFile != "" {
		err := ioutil.WriteFile(recommendedVersionFile, []byte(recommendedVersion.String()), 0600)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println(recommendedVersion.String())
	}
}

func Header(sb *strings.Builder, header string) {
	sb.WriteString(fmt.Sprintf("\n## %s\n\n", header))
}

func DisplayPRs(sb *strings.Builder, prs []github.PullRequest) {
	for _, p := range prs {
		sb.WriteString(fmt.Sprintf("* [%s](%s) - `%s`\n", p.Title, strings.ReplaceAll(p.Url, "\"", ""), p.Author))
	}
}

func ArrayContains(array []string, in string) bool {
	for _, i := range array {
		if i == in {
			return true
		}
	}
	return false
}
