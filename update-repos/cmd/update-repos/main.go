package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	baseDir := "."
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}

	updater := NewRepoUpdater("0.0.1", baseDir)

	slog.Info("Starting repository updater", "version", updater.version, "baseDir", baseDir)

	if err := updater.Run(); err != nil {
		slog.Error("Update process failed", "error", err)
		os.Exit(1)
	}
}

type RepoUpdater struct {
	version     string
	logger      *slog.Logger
	failedRepos []string
	baseDir     string
}

func NewRepoUpdater(version, baseDir string) *RepoUpdater {
	return &RepoUpdater{
		version:     version,
		logger:      slog.Default(),
		failedRepos: nil,
		baseDir:     baseDir,
	}
}

func (r *RepoUpdater) runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	return cmd.Run()
}

func (r *RepoUpdater) runGitCommandWithOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()

	return strings.TrimSpace(string(output)), err
}

func (r *RepoUpdater) gitStatusHasChanges(dir string) (bool, error) {
	output, err := r.runGitCommandWithOutput(dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}

	return len(output) > 0, nil
}

func (r *RepoUpdater) updateRepo(repoDir string) error {
	repoName := filepath.Base(repoDir)
	r.logger.Info("Processing repository", "name", repoName)

	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		r.logger.Info("SKIP: not a git repository", "name", repoName)
		return nil
	}

	if err := r.runGitCommand(repoDir, "rev-parse", "--verify", "HEAD"); err != nil {
		r.logger.Info("SKIP: has no commits", "name", repoName)
		return nil
	}

	currentBranch, err := r.runGitCommandWithOutput(repoDir, "branch", "--show-current")
	if err != nil || currentBranch == "" {
		r.logger.Warn("in detached HEAD state", "name", repoName)
	}

	hasChanges, err := r.gitStatusHasChanges(repoDir)
	if err != nil {
		r.logger.Error("Failed to check status", "name", repoName)
		r.failedRepos = append(r.failedRepos, fmt.Sprintf("%s: Failed to check status", repoName))
		return fmt.Errorf("failed to check status")
	}

	if hasChanges {
		r.logger.Warn("has uncommitted changes - stashing", "name", repoName)
		stashMsg := fmt.Sprintf("Auto-stash before update %s", time.Now().Format("2006-01-02 15:04:05"))
		if err := r.runGitCommand(repoDir, "stash", "push", "-m", stashMsg); err != nil {
			r.logger.Error("Failed to stash changes", "name", repoName)
			r.failedRepos = append(r.failedRepos, fmt.Sprintf("%s: Failed to stash changes", repoName))
			return fmt.Errorf("failed to stash changes")
		}
	}

	if err := r.runGitCommand(repoDir, "fetch", "origin"); err != nil {
		r.logger.Error("Failed to fetch from origin", "name", repoName)
		r.failedRepos = append(r.failedRepos, fmt.Sprintf("%s: Failed to fetch", repoName))
		return fmt.Errorf("failed to fetch")
	}

	var mainBranch string
	if err := r.runGitCommand(repoDir, "rev-parse", "--verify", "origin/main"); err == nil {
		mainBranch = "main"
	} else if err := r.runGitCommand(repoDir, "rev-parse", "--verify", "origin/master"); err == nil {
		mainBranch = "master"
	} else {
		r.logger.Error("Neither main nor master branch exists on origin", "name", repoName)
		r.failedRepos = append(r.failedRepos, fmt.Sprintf("%s: No main/master branch", repoName))
		return fmt.Errorf("no main/master branch")
	}

	if err := r.runGitCommand(repoDir, "checkout", mainBranch); err != nil {
		r.logger.Error("Failed to checkout branch", "name", repoName, "branch", mainBranch)
		r.failedRepos = append(r.failedRepos, fmt.Sprintf("%s: Failed to checkout %s", repoName, mainBranch))
		return fmt.Errorf("failed to checkout %s", mainBranch)
	}

	if err := r.runGitCommand(repoDir, "pull", "origin", mainBranch); err != nil {
		r.logger.Error("Failed to pull latest changes", "name", repoName)
		r.failedRepos = append(r.failedRepos, fmt.Sprintf("%s: Failed to pull", repoName))
		return fmt.Errorf("failed to pull")
	}

	r.logger.Info("Successfully updated repository", "name", repoName, "branch", mainBranch)
	return nil
}

func (r *RepoUpdater) Run() error {
	r.logger.Info("Starting repository update process", "baseDir", r.baseDir)

	if err := os.Chdir(r.baseDir); err != nil {
		r.logger.Error("Cannot access base directory", "baseDir", r.baseDir)
		return fmt.Errorf("cannot access base directory %s", r.baseDir)
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		r.logger.Error("Cannot read directory contents")
		return fmt.Errorf("cannot read directory contents")
	}

	var processed, successful int

	for _, entry := range entries {
		if entry.IsDir() {
			processed++
			if err := r.updateRepo(entry.Name()); err == nil {
				successful++
			}
			if err := os.Chdir(r.baseDir); err != nil {
				r.logger.Error("Cannot return to base directory", "baseDir", r.baseDir)
				return fmt.Errorf("cannot return to base directory")
			}
		}
	}

	r.logger.Info("=== SUMMARY ===")
	r.logger.Info("Repository update summary", "processed", processed, "successful", successful, "failed", processed-successful)

	if len(r.failedRepos) > 0 {
		r.logger.Error("Failed repositories", "failures", r.failedRepos)
		return fmt.Errorf("some repositories failed to update")
	}

	r.logger.Info("All repositories updated successfully!")
	return nil
}
