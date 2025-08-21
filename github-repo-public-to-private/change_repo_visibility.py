#!/usr/bin/env uv run
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "requests",
# ]
# ///

import requests
import sys
import os
from typing import Optional


def change_repo_visibility(
    owner: str,
    repo: str,
    token: str,
    base_url: str = "https://api.github.com",
    private: bool = True,
) -> bool:
    url = f"{base_url}/repos/{owner}/{repo}"
    headers = {
        "Authorization": f"token {token}",
        "Accept": "application/vnd.github.v3+json",
    }
    data = {"private": private}

    response = requests.patch(url, headers=headers, json=data)

    if response.status_code == 200:
        visibility = "private" if private else "public"
        print(f"Successfully changed {owner}/{repo} to {visibility}")
        return True
    else:
        print(
            f"Error: {response.status_code} - {response.json().get('message', 'Unknown error')}"
        )
        return False


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print(
            "Usage: python change_repo_visibility.py <owner> <repo1> [repo2] [repo3] ..."
        )
        print("Set GITHUB_TOKEN environment variable with your personal access token")
        print(
            "Optionally set GITHUB_BASE_URL for GitHub Enterprise (defaults to https://api.github.com)"
        )
        sys.exit(1)

    owner = sys.argv[1]
    repos = sys.argv[2:]
    token: Optional[str] = os.getenv("GITHUB_TOKEN")
    base_url: str = os.getenv("GITHUB_BASE_URL", "https://api.github.com")

    if not token:
        print("Error: GITHUB_TOKEN environment variable not set")
        sys.exit(1)

    success_count: int = 0
    total_count: int = len(repos)

    for repo in repos:
        print(f"Processing {owner}/{repo}...")
        if change_repo_visibility(owner, repo, token, base_url):
            success_count += 1
        print()

    print(
        f"Completed: {success_count}/{total_count} repositories successfully changed to private"
    )
