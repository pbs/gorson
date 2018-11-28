# How to set up a development environment

* Clone this repository into a Go [workspace](https://golang.org/doc/code.html#Organization):

```
git clone git@github.com:pbs/gorson.git $GOPATH/src/github.com/pbs/gorson
```

* Using [`dep`](https://golang.github.io/dep/), ensure local dependencies are installed

```
cd $GOPATH/src/github.com/pbs/gorson
dep ensure
```

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