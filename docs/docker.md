# Setting up Docker

This repository has been built with development of `gorson` done entirely in Docker in mind. Below are the steps required to get
this package up and running through Docker.

## Build Docker image

```bash
$ docker build -t gorson .
```

## Run tests on Docker image

```bash
$ docker run -it -v `pwd`:/app gorson ./scripts/test.sh
```

## Create releases

```bash
$ docker run -it -v `pwd`:/app gorson ./scripts/docker-release.sh
```
