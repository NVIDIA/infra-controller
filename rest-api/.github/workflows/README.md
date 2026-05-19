<!--
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# GitHub Actions Workflows for Docker Build and Push

This directory contains GitHub Actions workflows for building and pushing Docker images to NVIDIA Container Registry (NVCR).

## Workflows Overview

### 1. `main-build.yml` - Main Build Workflow
The primary workflow that orchestrates the entire build and push process.

**Triggers:**
- Push to `main` branch (builds and pushes)
- Push of version tags (e.g., `v1.2.3`) (builds and pushes)
- Pull requests to `main` (builds only, does not push)
- Push to `feat/my-task` branch (builds only, does not push)
- Manual trigger via workflow_dispatch, will adhere to above rules

**Build-Only Mode:**
The workflow supports building without pushing to test Dockerfiles before secrets are configured. Pushes to any branch other than `main` will execute in build-only mode.

### 2. `prepare-build-info.yml` - Build Information Preparation
Reusable workflow that generates version information and build metadata.

**Outputs:**
- `version`: Value from VERSION file
- `short_sha`: 7-character git commit SHA
- `full_sha`: Full git commit SHA
- `target_registry`: Target NVCR registry path
- `release_tag`: Tags in the format of `v*.*.*`

### 3. `build-push-docker.yml` - Docker Build and Push
Reusable workflow that builds and pushes all Docker images.

**Images Built:**
1. `nico-rest-api`
2. `nico-rest-db`
3. `nico-rest-site-manager`
4. `nico-rest-workflow`
5. `nico-rest-site-agent`
6. `nico-rest-cert-manager`
7. `nico-flow`
8. `nico-psm`
9. `nico-nsm`

**Note:** This workflow uses production-optimized multi-stage Dockerfiles located in `docker/production/`.
## Required Secrets

You must configure the following secrets in your GitHub repository:

### NVCR_USERNAME
- **Description**: NVIDIA Container Registry username
- **Value**: `$oauthtoken` (for NGC API tokens) or your NVCR username
- **Required**: Yes

### NVCR_TOKEN
- **Description**: NVIDIA Container Registry API token
- **Value**: Your NGC API token
- **Required**: Yes

### How to Configure Secrets

1. **Get your NGC API Token:**
   - Go to https://ngc.nvidia.com/
   - Sign in with your NVIDIA account
   - Navigate to "Setup" -> "Generate API Key"
   - Copy your API key

2. **Add secrets to GitHub:**
   - Go to your GitHub repository
   - Click "Settings" -> "Secrets and variables" -> "Actions"
   - Click "New repository secret"
   - Add both secrets:
     - Name: `NVCR_USERNAME`, Value: `$oauthtoken`
     - Name: `NVCR_TOKEN`, Value: `<your-ngc-api-key>`

## Configuration

### Customize Registry Path

Edit `.github/workflows/prepare-build-info.yml` and update the `target_registry` value:

```yaml
# Line ~57
target_registry="nvcr.io/dsx/your-team/nico"
```

Replace `your-team` with appropriate NGC team name. Available options are:
- *carbide-dev*: For all development images
- *carbide*: Production promoted images

### Customize Runner

By default, workflows use `ubuntu-latest`. To use custom runners, edit `main-build.yml`:

```yaml
prepare:
  uses: ./.github/workflows/prepare-build-info.yml
  with:
    runner: your-custom-runner  # Change this

build-and-push:
  uses: ./.github/workflows/build-push-docker.yml
  with:
    runner: your-custom-runner  # Change this
```

### Enable/Disable Push

Images are pushed only when:
- Event is a commit or tag push
- Branch is `main` or a version git tag is pushed
- Developers add a commit message in a branch with `push-container` in commit message
- Secrets are properly configured (login step is skipped if secrets are missing)

**Build-Only Mode (No Push):**
The workflow automatically runs in build-only mode on:
- Feature branches (like `feature/docker-build-push-nvcr`) when commit message doesn't include `push-container`

This allows you to:
- Test Dockerfiles before secrets are configured
- Verify that builds succeed in CI
- See build logs and results in GitHub Actions

## Docker Image Tagging

Pushes to `main` branch will receive the following tags:

1. **VERSION + Git Short SHA**: `<image>:1.0.0=<short-sha>` (e.g., `nico-rest-api:1.0.0-abc1234`)
   - Primary tag for reproducibility and traceability
   - Unique for each commit in `main` branch

2. **Latest Tag**: `<image>:latest`
   - Points to the most recent successful build on `main` branch

Version git tags will receive the following docker tag:

1. **Version Tag**: `<image>:<tag>` (e.g., `nico-rest-api:v1.2.3`)
   - Based on tags pushed with v* format

Commits to branches containing `push-container` will receive Git short SHA tag:

1. **Branch Name + Git Short SHA Tag**: `<image>:fix-api-<short-sha>` (e.g., `nico-rest-api:fix-api-ghi678`)
   - Allows developers to push their in-progress work to dev environments

## Usage Examples

### Manual Trigger on Push to Feature Branch
```bash
git add .
git commit -m "Updated code. push-container"
git push origin feat/my-feature
```

### Trigger on Tag
```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

### Manual Trigger
1. Go to "Actions" tab in GitHub
2. Select "Build and Push All Docker Images"
3. Click "Run workflow"
4. Select branch and click "Run workflow"

## Pull Images

After successful build and push:

```bash
# Log in to NVCR
docker login nvcr.io
# Username: $oauthtoken
# Password: <your-ngc-api-key>

# Pull by SHA (recommended for production)
docker pull nvcr.io/dsx/your-team/nico-rest-api:1.2.3-abc1234

# Pull by version
docker pull nvcr.io/dsx/your-team/nico-rest-api:v1.2.3

# Pull latest
docker pull nvcr.io/dsx/your-team/nico-rest-api:latest
```

## Build Cache

The workflows use GitHub Actions cache to speed up builds:
- Cache is scoped per image
- Uses layer caching via Docker Buildx
- Reduces build times on subsequent runs

## Troubleshooting

### Authentication Failures
- Verify `NVCR_USERNAME` is set to `$oauthtoken`
- Verify `NVCR_TOKEN` contains your valid NGC API key
- Ensure your NGC API key has push permissions

### Build Failures
- Check that all Dockerfiles exist and are valid
- Verify Docker build contexts are correct
- Review build logs in GitHub Actions

### Push Not Happening
- Ensure you're not on a pull request (PRs don't push by default)
- Verify `push_enabled: true` in the workflow
- Check that secrets are configured correctly

### Registry Path Issues
- Update `target_registry` in `prepare-build-info.yml`
- Format should be: `nvcr.io/<org>/<project>`
- No trailing slashes

## Multi-Architecture Support

To add ARM64 support, modify the `platforms` field in `build-push-docker.yml`:

```yaml
platforms: linux/amd64,linux/arm64
```

Note: ARM64 builds may require ARM64 runners or QEMU emulation.

## Advanced: Conditional Builds

To build only specific images based on file changes, you can add path filters:

```yaml
- name: Filter paths
  uses: dorny/paths-filter@v3
  id: filter
  with:
    filters: |
      api:
        - 'docker/production/Dockerfile.nico-rest-api'
      db:
        - 'docker/production/Dockerfile.nico-rest-db'

- name: Build API
  if: steps.filter.outputs.api == 'true'
  # ... build steps
```

## Security Considerations

1. **Secrets Management**: Never commit secrets to the repository
2. **API Token Rotation**: Regularly rotate your NGC API tokens
3. **Access Control**: Limit who can trigger workflows manually
4. **Image Scanning**: Consider adding vulnerability scanning steps
5. **Tag Immutability**: Use SHA tags for production deployments

## Support

For issues or questions:
- Check GitHub Actions logs for detailed error messages
- Review NVCR documentation: https://docs.nvidia.com/ngc/
- Verify Docker build locally before pushing
