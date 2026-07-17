# Release process

Releases are built by the GitHub Action in `.github/workflows/release.yml`. It builds the SDL frontend with SDL2 statically linked for:

- Linux amd64: `izapple2-linux-amd64.tar.gz`
- Windows amd64: `izapple2-windows-amd64.zip`
- macOS universal (Intel and Apple Silicon): `izapple2-macos-universal.tar.gz`

Each archive contains the `izapple2` (or `izapple2.exe`) binary, the README and the LICENSE.

The Linux binary is built on Ubuntu 22.04 and runs on any distribution with glibc 2.35 or newer.

## Steps

1. Make sure master is green and up to date.
2. Tag the release and push the tag:

    ```shell
    git tag vX.Y.Z
    git push origin vX.Y.Z
    ```

3. Wait for the "Release" workflow to finish. It creates a **draft** release with the three archives attached and auto-generated release notes.
4. Review the draft release on GitHub (notes and binaries) and click "Publish release".

## Testing the build without releasing

Run the "Release" workflow manually from the GitHub Actions tab (workflow_dispatch). It builds the three artifacts, downloadable from the workflow run page, but does not create a release.
