



# GitHub Actions

This directory contains GitHub Actions workflows for building, testing, and linting the gochecks project.

## Available Workflows

- `.github/workflows/go.yml`: Builds multi-platform binaries and packages on push or pull request to main
- `.github/workflows/lint.yml`: Runs linters on push or pull request to main

## Adding Tests

To add tests to the workflow, update the `test` job in `.github/workflows/go.yml` with the appropriate test commands.

