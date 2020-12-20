# Setting up Docker

This repository has been built with development of `gorson` done entirely in Docker in mind. Below are the steps required to get
this package up and running through Docker.

## Build Docker image

```bash
docker build -t gorson .
```

or

```bash
docker-compose build
```

## Run tests on Docker image

```bash
docker run -it -v "$PWD":/app gorson ./scripts/test.sh
```

or

```bash
docker-compose run --rm builder ./scripts/test.sh
```

## Create releases

```bash
docker run -it -v "$PWD":/app gorson ./scripts/docker-release.sh
```

or

```bash
docker-compose run --rm builder ./scripts/docker-release.sh
```
