[![Stability](https://img.shields.io/badge/Stability-Under%20Active%20Development-Red.svg)](https://github.com/pbs/gorson) [![Main Workflow](https://github.com/pbs/gorson/workflows/Main%20Workflow/badge.svg)](https://github.com/pbs/gorson/actions?query=workflow%3A%22Main+Workflow%22)

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

There's also a `--format` flag to pass in which format you want the parameters to export as.

```
$ gorson get --format yaml /a/parameter/store/path/ > ./example.yml
```

```
$ cat ./example.yml

alpha: "the_alpha_value"
beta: "the_beta_value"
delta: "the_delta_value"
```

```
$ gorson get --format env /a/parameter/store/path/ > ./.env
```

```
$ cat ./.env

alpha="the_alpha_value"
beta="the_beta_value"
delta="the_delta_value"
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
$ wget https://github.com/pbs/gorson/releases/download/7/gorson-7-darwin-amd64
```

## Linux

Download the binary
```
$ w https://github.com/pbs/gorson/releases/download/7/gorson-7-darwin-amd64
```

Move the binary to an installation path, make it executable, and add to path
```
mkdir -p /opt/gorson/bin
mv gorson-7-linux-amd64 /opt/gorson/bin/gorson
chmod +x /opt/gorson/bin/gorson
export PATH="$PATH:/opt/gorson/bin"
```

# Notes

These environment variables will affect the AWS session behavior:

https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html


* `AWS_PROFILE`: use a named profile from your `~/.aws/config` file (see https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html)
* `AWS_REGION`: use a specific AWS region (see https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html)

```
AWS_PROFILE=example-profile AWS_REGION=us-east-1 gorson get /a/parameter/store/path/
```

# Development

See [docs/development.md](docs/development.md)
