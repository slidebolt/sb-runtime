# Git Workflow for sb-runtime

This repository contains the Slidebolt runtime environment, providing core execution utilities and logging integration for services and plugins.

## Dependencies
- **Internal:**
  - `sb-contract`: Core interfaces and shared structures.
- **External:** 
  - Standard Go library.

## Build Process
- **Type:** Pure Go Library (Shared Module).
- **Consumption:** Imported as a module dependency in other Go projects via `go.mod`.
- **Artifacts:** No standalone binary or executable is produced.
- **Validation:** 
  - Validated through unit tests: `go test -v ./...`
  - Validated by its consumers during their respective build/test cycles.

## Pre-requisites & Publishing
As a core runtime library, `sb-runtime` should be updated whenever the core `sb-contract` is changed.

**Before publishing:**
1. Determine current tag: `git tag | sort -V | tail -n 1`
2. Ensure all local tests pass: `go test -v ./...`

**Publishing Order:**
1. Ensure `sb-contract` is tagged and pushed (e.g., `v1.0.4`).
2. Update `sb-runtime/go.mod` to reference the latest `sb-contract` tag.
3. Determine next semantic version for `sb-runtime` (e.g., `v1.0.4`).
4. Commit and push the changes to `main`.
5. Tag the repository: `git tag v1.0.4`.
6. Push the tag: `git push origin main v1.0.4`.
7. Update dependent repositories using `go get github.com/slidebolt/sb-runtime@v1.0.4`.

## Update Workflow & Verification
1. **Modify:** Update runtime logic in `runtime.go` or `log.go`.
2. **Verify Local:**
   - Run `go mod tidy`.
   - Run `go test ./...`.
3. **Commit:** Ensure the commit message clearly describes the runtime change.
4. **Tag & Push:** (Follow the Publishing Order above).
