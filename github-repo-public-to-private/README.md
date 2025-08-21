A self-contained Python script that uses the GitHub API to change repository
visibility from public to private. Supports batch operations on multiple
repositories and works with both GitHub.com and GitHub Enterprise instances.

Usage:

```bash
# Set your GitHub token
export GITHUB_TOKEN=your_personal_access_token

# Make multiple repos private
./change_repo_visibility.py username repo1 repo2 repo3

# For GitHub Enterprise
export GITHUB_BASE_URL=https://github.enterprise.com/api/v3
./change_repo_visibility.py username repo1 repo2
```

Limitations:

- Errors `403 - Repository was archived so is read-only` are not handled

