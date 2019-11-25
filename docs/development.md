# How to run tests

* Run the test script:

```
./scripts/test.sh
```

# How to release

* Make sure the version number in [`internal/gorson/version/version.go`](/internal/gorson/version/version.go) matches the version you're releasing.
* Run the release script to cross-compile binaries for distribution:
```
./scripts/release.sh
```
* create a new release in github: https://github.com/pbs/gorson/releases/new
* add the binaries created locally in your `./bin` directory