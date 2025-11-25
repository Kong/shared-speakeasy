# Release Process

This repository contains multiple Go modules as subdirectories. Each module is released independently using a dedicated GitHub Actions workflow.

## Prerequisites

- You must have write access to the repository
- The module must have a `go.mod` file in its subdirectory
- Version must follow semantic versioning format: `vMAJOR.MINOR.PATCH` (e.g., `v1.2.0`)

## Release Steps

1. **Navigate to Actions**
   - Go to the [Actions tab](../../actions) in the GitHub repository
   - Select the "Tag & Release Go Submodule" workflow

2. **Trigger the Workflow**
   - Click "Run workflow"
   - Fill in the required inputs:
     - **Module**: The subdirectory path of the module (e.g., `hooks/mesh_defaults`)
     - **Version**: The version to release (e.g., `v1.2.0`)

3. **Workflow Validation**
   The workflow automatically validates:
   - Module path exists and contains a `go.mod` file
   - Version follows semantic versioning format (`vMAJOR.MINOR.PATCH`)
   - Tag doesn't already exist

4. **Release Creation**
   If validation passes, the workflow will:
   - Create a Git tag in the format: `{module}/{version}` (e.g., `hooks/mesh_defaults/v0.0.4`)
   - Create a GitHub release with the tag
   - Attach automated release notes
