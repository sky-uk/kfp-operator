# KFP SDK

This module provides an extension of the KFP SDK and a corresponding docker image.

## Setup

Note that we use the [dynamic versioning plugin](https://pypi.org/project/poetry-dynamic-versioning/) for Poetry to version this module.
The version differs from those of the resulting containers (which are based on `git describe`) because Poetry would otherwise reject it. This will be installed automatically when running the `build` make target.

## Run tests
```bash
make test
```

## Build the container image

```bash
make docker-build
```