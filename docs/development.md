# How to run tests

* Run the test script:

```bash
./scripts/test.sh
```

## How to release

From the `main` branch, run the release script:

```bash
./scripts/release.sh
```

Then, create a pull request with your version bump and get it merged to `main`.

The release script will:

* update the version number everywhere relevant in the codebase
* check out a git branch
* commit the version change
* tag the commit with the new version number
* push the branch
* create a Github release
