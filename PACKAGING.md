#

## Package Matrix

### Linux
- `ondevice_$version_linux-$arch.tgz` (amd64, i386 and armhf)
- `ondevice_$version_$arch.deb`, `ondevice-daemon_$version_all.deb` (amd64, i386 and armhf):w
- TODO `ondevice_$version_$arch.rpm`

### MacOS
- [homebrew tap](https://github.com/ondevice/homebrew-tap)
- TODO binary install (`ondevice_$version_mac-amd64.tgz`)


## Release Checklist

- check github issues and milestones (unless it's a bugfix release)
- test code locally
- package *nightly* builds:
  - push your changes, wait for the build to succeed on travis-ci.org (and the resulting artifacts to show up on https://repo.ondevice.io/builds/ )  
    alternatively, run `make package` locally (make sure to have a clean working directory)
  - install the resulting packages on different test machines (especially if packaging details have changed)
    - use these packages for some days, just to be sure
- create `stable` packages:
  - update `CHANGELOG.md` (as well as `build/deb/debian/changelog`)
  - update `$VERSION` in `Makefile`
  - update `build/deb/debian/changelog` (using the contents of CHANGELOG.md)
  - run `make package` and see if things work
  - `git commit`
  - `git tag v$VERSION`
  - `git push origin master v$VERSION`
  - wait for travis to build things
- release process
  - create a github release, copying the info from `CHANGELOG.md` (the build artifacts should already be there)
  - copy the `//repo.ondevice.io/build/$buildNr/` dir to `/client/$version/`
  - update the `/client/stable` symlink
  - update the debian repos
  - upgrade homebrew tap (by updating `url` as well as `sha256`)


## TODO

- rpm
- arch (pacman)
- msys2 (pacman), chocolately?
