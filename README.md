[![Stability](https://img.shields.io/badge/Stability-Under%20Active%20Development-Red.svg)](https://github.com/pbs/gorson) [![Main Workflow](https://github.com/pbs/gorson/workflows/Main%20Workflow/badge.svg)](https://github.com/pbs/gorson/actions?query=workflow%3A%22Main+Workflow%22)

# Warning: experimental

This is an experimental library, and is currently unsupported.

# Usage

`gorson` loads parameters from [AWS ssm parameter store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html), and adds them as shell environment variables.

## Download parameters from parameter store as a json file

```bash
gorson get /a/parameter/store/path/ > ./example.json
```

```bash
$ cat ./example.json

{
    "alpha": "the_alpha_value",
    "beta": "the_beta_value",
    "delta": "the_delta_value"
}
```

There's also a `--format` flag to pass in which format you want the parameters to export as.

```bash
gorson get --format yaml /a/parameter/store/path/ > ./example.yml
```

```bash
$ cat ./example.yml

alpha: "the_alpha_value"
beta: "the_beta_value"
delta: "the_delta_value"
```

```bash
gorson get --format env /a/parameter/store/path/ > ./.env
```

```bash
$ cat ./.env

alpha="the_alpha_value"
beta="the_beta_value"
delta="the_delta_value"
```

## Load parameters as environment variables from a json file

```bash
source <(gorson load ./example.json)
```

```bash
$ env | grep 'alpha\|beta\|delta'
alpha=the_alpha_value
delta=the_delta_value
beta=the_beta_value
```

## Upload parameters to parameter store from a json file

```bash
gorson put /a/parameter/store/path/ --file=./new-values.json
```

## Delete parameter difference on put

```bash
$ gorson put /a/parameter/store/path/ --file=./different-values.json --delete

The following are present in the file, but not in parameter store:
/a/parameter/store/path/gamma
Are you sure you'd like to delete these parameters?
Type yes to proceed:

```

## Auto-approve prompts

If you would like to answer 'yes' to any prompts that require it, append `--auto-approve`.

## Deactivate color

If you would prefer the output of commands to be colorless, append `--no-color`.

# Installation

Currently gorson ships binaries for OS X and Linux 64bit systems. You can download the latest release from [GitHub](https://github.com/pbs/gorson/releases)

## OS X

```bash
wget https://github.com/pbs/gorson/releases/download/6/gorson-6-darwin-amd64
```

## Linux

Download the binary

```bash
wget https://github.com/pbs/gorson/releases/download/6/gorson-6-darwin-amd64
```

Move the binary to an installation path, make it executable, and add to path

```bash
mkdir -p /opt/gorson/bin
mv gorson-6-linux-amd64 /opt/gorson/bin/gorson
chmod +x /opt/gorson/bin/gorson
export PATH="$PATH:/opt/gorson/bin"
```

## asdf

Install using asdf

Add asdf plugin

```bash
asdf plugin add gorson https://github.com/pbs/asdf-pbs.git
```

List available versions

```bash
asdf list-all gorson
```

Install a particular version

```bash
asdf install gorson 6
```

Make a particular version your default

```bash
asdf global gorson 6
```

# Notes

These environment variables will affect the AWS session behavior:

<https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html>

* `AWS_PROFILE`: use a named profile from your `~/.aws/config` file (see <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html>)
* `AWS_REGION`: use a specific AWS region (see <https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html>)

```bash
AWS_PROFILE=example-profile AWS_REGION=us-east-1 gorson get /a/parameter/store/path/
```

# Development

See [docs/development.md](docs/development.md)
