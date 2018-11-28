[![Stability](https://img.shields.io/badge/Stability-Under%20Active%20Development-Red.svg)](https://github.com/pbs/gorson)

# Warning: experimental

This is an experimental library, and is currently unsupported.

# Usage

`gorson` loads parameters from [AWS ssm parameter store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html), and adds them as shell environment variables.

## Download parameters from parameter store as a json file

```
$ gorson get /a/parameter/store/path/ > ./example.json
```

```
$ cat ./example.json

{
    "alpha": "the_alpha_value",
    "beta": "the_beta_value",
    "delta": "the_delta_value"
}
```

## Load parameters as environment variables from a json file

```
source <(gorson load ./example.json)
```

```
$ env | grep 'alpha\|beta\|delta'
alpha=the_alpha_value
delta=the_delta_value
beta=the_beta_value
```

## Upload parameters to parameter store from a json file

```
$ gorson put /a/parameter/store/path/ --file=./new-values.json
```

# Installation

Currently gorson ships binaries for OS X and Linux 64bit systems. You can download the latest release from [GitHub](https://github.com/pbs/gorson/releases)

## OS X

```
$ wget https://github.com/pbs/gorson/releases/download/0.0.1/gorson-0.0.1-darwin-amd64
```

## Linux

Download the binary
```
$ wget https://github.com/pbs/gorson/releases/download/0.0.1/gorson-0.0.1-linux-amd64
```

Move the binary to an installation path, make it executable, and add to path
```
mkdir -p /opt/gorson/bin
mv gorson-0.0.1-linux-amd64 /opt/gorson/bin/gorson
chmod +x /opt/gorson/bin/gorson
export PATH="$PATH:/opt/gorson/bin"
```

# Tests

`./scripts/test.sh`