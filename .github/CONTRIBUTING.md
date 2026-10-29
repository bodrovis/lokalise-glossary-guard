# Contributing

1. [Fork the repository.][fork]
2. [Create a topic branch.][branch]
3. Implement your feature or bug fix.
4. Don't forget to add tests and make sure they pass (`go test -shuffle=on -count=1 -race ./...`).
5. `gofmt -s -w . && gofumpt -w . && go vet ./... && staticcheck ./... && golangci-lint run`.
6. `go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest -fix ./....`.
7. If necessary, add documentation for your feature or bug fix.
8. Commit and push your changes.
9. [Submit a pull request.][pr]

[fork]: http://help.github.com/fork-a-repo/
[branch]: https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-branches
[pr]: https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-requests

## Release process

Lokalise Glossary Guard uses GoReleaser and GitHub Actions to build, sign, and publish releases. GitHub releases are immutable, so each release must use a new version tag.

1. Make sure all changes are committed and tests pass.
2. Update `.github/release-notes.md` with the changes included in the new release.
3. Commit and push the updated release notes:

   ```bash
   git add .
   git commit -m "Prepare v1.4.1 release"
   git push origin main
   ```

4. Create and push a new version tag:

   ```bash
   git tag -a v1.4.1 -m "Release v1.4.1"
   git push origin v1.4.1
   ```

5. GitHub Actions will automatically build, sign, and publish the release using the prepared release notes.

**Do not create GitHub Releases manually.** Published releases cannot be updated or replaced. If a release needs to be rebuilt, use a new version tag.

View published releases [here](https://github.com/bodrovis/lokalise-glossary-guard/releases).
