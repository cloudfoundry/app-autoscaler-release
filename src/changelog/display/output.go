package display

import (
	"changelog/github"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func GenerateOutput(prs []github.PullRequest, previousVersion string) (string, string, error) {
	var sb strings.Builder
	sb.WriteString("# Changelog for app-autoscaler-release\n\n")

	var breakingChanges []github.PullRequest
	var enhancements []github.PullRequest
	var bugFixes []github.PullRequest
	var dependencyUpdates []github.PullRequest
	var chores []github.PullRequest
	var other []github.PullRequest

	for _, pr := range prs {
		if ArrayContains(pr.Labels, "exclude-from-changelog") {
			// exclude
		} else if ArrayContains(pr.Labels, "breaking-change") {
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

	v, err := semver.NewVersion(previousVersion)
	if err != nil {
		return "", "", err
	}

	var recommendedVersion semver.Version
	if len(breakingChanges) > 0 {
		recommendedVersion = v.IncMajor()
	} else if len(enhancements) > 0 {
		recommendedVersion = v.IncMinor()
	} else {
		recommendedVersion = v.IncPatch()
	}

	return sb.String(), recommendedVersion.String(), nil
}

func Header(sb *strings.Builder, header string) {
	fmt.Fprintf(sb, "\n## %s\n\n", header)
}

func DisplayPRs(sb *strings.Builder, prs []github.PullRequest) {
	for _, p := range prs {
		fmt.Fprintf(sb, "* [%s](%s) - **%s**\n", p.Title, strings.ReplaceAll(p.Url, "\"", ""), p.Author)
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
