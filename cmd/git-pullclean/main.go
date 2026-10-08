package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var repoName string

func init() {
	flag.StringVar(&repoName, "repo", "origin", "repository name(default='origin')")
	flag.Parse()
}

func runGitCommand(args []string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("$ git %s\n", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func getGitCommandOutput(args []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes)), nil
}

func getGitRootDirectory() (string, error) {
	gitArgs := []string{"rev-parse", "--show-toplevel"}
	return getGitCommandOutput(gitArgs, 10*time.Second)
}

func getCurrentBranchName() (string, error) {
	gitArgs := []string{"branch", "--show-current"}
	return getGitCommandOutput(gitArgs, 10*time.Second)
}

func isRepositoryClean() error {
	gitArgs := []string{"diff", "--quiet"}
	return runGitCommand(gitArgs, 10*time.Second)
}

func pullIntoCurrentBranch(branch string) error {
	// set so long time for bigger repository
	gitArgs := []string{"pull", "--rebase", repoName, branch}
	return runGitCommand(gitArgs, 600*time.Second)
}

func fetchPrune() error {
	// set longer time for bigger repository
	gitArgs := []string{"fetch", "--prune"}
	return runGitCommand(gitArgs, 600*time.Second)
}

func cleanupLfsFiles() error {
	root, err := getGitRootDirectory()
	if err != nil {
		return err
	}

	// check if the repository use Git LFS
	lfsDir := filepath.Join(root, ".git", "lfs")
	if _, err := os.Stat(lfsDir); os.IsNotExist(err) {
		// this repository does not use Git LFS
		return nil
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error happened in LFS dir. Skip LFS cleanup")
		return nil
	}

	gitArgs := []string{"lfs", "prune"}
	return runGitCommand(gitArgs, 10*time.Second)
}

func pullAndCleanup() error {
	branch, err := getCurrentBranchName()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// git pull --rebase
	if err := pullIntoCurrentBranch(branch); err != nil {
		return fmt.Errorf("failed to pull commits into '%s' branch: %w", branch, err)
	}

	// git fetch --prune
	if err := fetchPrune(); err != nil {
		return fmt.Errorf("failed to remove unnecessary tracking references: %w", err)
	}

	// clean LFS files if the repository uses Git LFs
	if err := cleanupLfsFiles(); err != nil {
		return fmt.Errorf("failed to clean up LFS files: %w", err)
	}

	return nil
}

func isProtectedBranch(branch string) bool {
	protected := []string{"main", "master", "develop"}
	return slices.Contains(protected, branch)
}

func collectMergedBranches() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "branch", "--merged")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	mergedBranches := []string{}
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		// ignore current branch or workspace branches
		if strings.HasPrefix(branch, "*") || strings.HasPrefix(branch, "+") {
			continue
		}
		if isProtectedBranch(branch) {
			continue
		}

		mergedBranches = append(mergedBranches, branch)
	}

	return mergedBranches, nil
}

func deleteMergedBranches() error {
	currentBranch, err := getCurrentBranchName()
	if err != nil {
		return fmt.Errorf("failed to get current branch name: %w", err)
	}

	// Delete merged branches only if the current branch is main/master/develop
	if !isProtectedBranch(currentBranch) {
		return nil
	}

	mergedBranches, err := collectMergedBranches()
	if err != nil {
		return fmt.Errorf("failed to collect merged branches: %w", err)
	}

	deletedBranches := 0
	timeout := 5 * time.Second
	for _, branch := range mergedBranches {
		fmt.Printf("Delete branch='%s'\n", branch)

		gitArgs := []string{"branch", "-d", branch}
		if err := runGitCommand(gitArgs, timeout); err != nil {
			fmt.Fprintf(os.Stderr, "failed to delete branch='%s'\n", branch)
			continue
		}

		deletedBranches++
	}

	fmt.Printf("Delete %d branches\n", deletedBranches)
	return nil
}

func _main() int {
	if err := isRepositoryClean(); err != nil {
		fmt.Fprintf(os.Stderr, "This repository is not clean. Please commit changes first\n")
		return 1
	}

	if err := pullAndCleanup(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to pull and cleanup repository: %v\n", err)
		return 1
	}

	if err := deleteMergedBranches(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to delete merged branches: %v\n", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(_main())
}
