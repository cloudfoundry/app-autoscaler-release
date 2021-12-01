package util

import "strings"

var Runner CommandRunner

func GetShaOfSubmoduleAtCommit(commit string) (string, error) {
	cmd := Command{Name: "git", Args: []string{"ls-tree", commit, "../app-autoscaler"}}
	if Runner == nil {
		Runner = DefaultCommandRunner{}
	}
	output, err := Runner.RunWithoutRetry(&cmd)
	if err != nil {
		return "", err
	}

	parts := strings.Split(output, " ")
	if len(parts) > 1 {
		return strings.Split(parts[2], "\t")[0], nil
	}
	return "", nil
}
