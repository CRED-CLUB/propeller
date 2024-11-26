---
title: Contributing

layout: default
nav_order: 8
---
# Contribution Guide

Pre-requisites
1. Go 1.23 or later
2. Redis 6.0 or later
3. NATS 2.0 or later

Clone the repository

```shell
git clone https://github.com/CRED-CLUB/propeller.git
cd propeller
```

Run `make help` to list the available make targets

```shell
make help
```
```
Usage:
	build                          Build the binary
	build-sample-client            Build the test client binary
	clean                          Clean all generated files
	dev-dependencies-down          Bring down the dependencies (redis, NATS)
	dev-dependencies-up            Bring up the dependencies (redis, NATS)
	docker-down                    Bring down the propeller docker
	docker-up                      Bring up the propeller docker
	goimports                      Run goimports and format files
	goimports-check                Check goimports without modifying the files
	golint-check                   Run golint check
	help                           Display this help screen
	proto-clean                    Clean generated proto files
	proto-generate                 Generate proto bindings
	proto-lint                     Proto lint
```

Run `make build` to build the binary in `bin/` path

```shell
make build
```

There are utility targets to bring up dependent docker containers (Redis and NATS), like

```shell
make dev-dependencies-up
```

And to shut them down

```shell
make dev-dependencies-down
```

The above utilities are useful when developing and running `propeller` through an IDE.

`docker-up` and `docker-down` would also bring up/down `propeller` in a docker container.

Ensure the configuration is updated at `config/propeller.toml`

Run the binary

```shell
./bin/propeller
```

Refer [Testing Guide](https://cred-club.github.io/propeller/testing.html) for details on executing the APIs

Other options are self-explanatory.

**That's it! Go ahead and raise your first PR!**

---
