# How to run tests

* Run the test script:

```
./scripts/test.sh
```

# How to release

From the `master` branch, run the release script:

```
./scripts/release.sh
```

Then, create a pull request with your version bump and get it merged to `master`.

The release script will:

This will:

* update the version number everywhere relevant in the codebase
* check out a git branch
* commit the version change
* tag the commit with the new version number
* push the branch
* create a Github release
