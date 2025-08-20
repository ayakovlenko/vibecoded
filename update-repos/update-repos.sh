#!/usr/bin/env bash

set -e

LOG_FILE="update_repos.log"
FAILED_REPOS=()

log() {
	echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

update_repo() {
	local repo_dir="$1"
	local repo_name
	repo_name=$(basename "$repo_dir")

	log "Processing repository: $repo_name"

	cd "$repo_dir" || {
		log "ERROR: Cannot access directory $repo_dir"
		FAILED_REPOS+=("$repo_name: Cannot access directory")
		return 1
	}

	if [ ! -d ".git" ]; then
		log "SKIP: $repo_name is not a git repository"
		return 0
	fi

	if ! git rev-parse --verify HEAD >/dev/null 2>&1; then
		log "SKIP: $repo_name has no commits"
		return 0
	fi

	local current_branch
	current_branch=$(git branch --show-current 2>/dev/null || echo "")
	if [ -z "$current_branch" ]; then
		log "WARNING: $repo_name is in detached HEAD state"
	fi

	if git status --porcelain | grep -q .; then
		log "WARNING: $repo_name has uncommitted changes - stashing"
		if ! git stash push -m "Auto-stash before update $(date)" 2>/dev/null; then
			log "ERROR: Failed to stash changes in $repo_name"
			FAILED_REPOS+=("$repo_name: Failed to stash changes")
			return 1
		fi
	fi

	if ! git fetch origin 2>/dev/null; then
		log "ERROR: Failed to fetch from origin in $repo_name"
		FAILED_REPOS+=("$repo_name: Failed to fetch")
		return 1
	fi

	if ! git rev-parse --verify origin/main >/dev/null 2>&1; then
		if ! git rev-parse --verify origin/master >/dev/null 2>&1; then
			log "ERROR: Neither main nor master branch exists on origin in $repo_name"
			FAILED_REPOS+=("$repo_name: No main/master branch")
			return 1
		else
			local main_branch="master"
		fi
	else
		local main_branch="main"
	fi

	if ! git checkout "$main_branch" 2>/dev/null; then
		log "ERROR: Failed to checkout $main_branch in $repo_name"
		FAILED_REPOS+=("$repo_name: Failed to checkout $main_branch")
		return 1
	fi

	if ! git pull origin "$main_branch" 2>/dev/null; then
		log "ERROR: Failed to pull latest changes in $repo_name"
		FAILED_REPOS+=("$repo_name: Failed to pull")
		return 1
	fi

	log "SUCCESS: Updated $repo_name to latest $main_branch"
	return 0
}

main() {
	local base_dir="${1:-$(pwd)}"

	log "Starting repository update process in: $base_dir"
	log "Log file: $LOG_FILE"

	cd "$base_dir" || {
		log "ERROR: Cannot access base directory $base_dir"
		exit 1
	}

	local processed=0
	local successful=0

	for dir in */; do
		if [ -d "$dir" ]; then
			processed=$((processed + 1))
			if update_repo "$dir"; then
				successful=$((successful + 1))
			fi
			cd "$base_dir"
		fi
	done

	log "=== SUMMARY ==="
	log "Processed: $processed repositories"
	log "Successful: $successful repositories"
	log "Failed: $((processed - successful)) repositories"

	if [ ${#FAILED_REPOS[@]} -gt 0 ]; then
		log "Failed repositories:"
		for failure in "${FAILED_REPOS[@]}"; do
			log "  - $failure"
		done
		exit 1
	fi

	log "All repositories updated successfully!"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
	main "$@"
fi
