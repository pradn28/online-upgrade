# MemSQL Online Upgrade 

MemSQL Online Upgrade is a Go project for upgrading MemSQL from 5.x to a newer 5.x version.

## Overview

`memsql-online-upgrade` will execute various steps to upgrade your cluster without downtime. While the first step will verify your cluster is healthy, it is recommended that you double check your cluster is in a healthy state prior to starting the upgrade. It is also recommended you backup your cluster prior to the upgrade.

## Requirements 

MemSQL must be running in High Availability mode and running MemSQL Ops 5.7 or higher.

## Usage

Run `memsql-online-upgrade` from the master aggregator on the cluster that you intend to upgrade. 

```
$ ./memsql-online-upgrade
```

You can optionally specify a host, username, and port when executing `memsql-online-upgrade`. If you don't specify these paramters, the following default values will be used:

* Host: `127.0.0.1`
* Username: `root`
* Port: `3306`

For a full list of available options, use the `-h` flag when executing `memsql-online-upgrade`.

If you want to upgrade to a specific build version, you can use the flag `-version-hash` followed by the desired version hash. By default, `memsql-online-upgrade` will use the latest released 5.x version.

#### Logging

When `memsql-online-upgrade` is executed, a log file is created in the current working directory. It will either create or append to the default log file (**online-upgrade.log**) unless a file location is specified with `-log-path`.

#### Health Check

Health check will preform the following steps to ensure your cluster is ready and willing to be upgraded:

* **Redundancy level check**: The redundancy level must be 2 for High Availability.
* **Agents online**: All agents must be in an `ONLINE` state.
* **Node check**: All MemSQL nodes must be `ONLINE` and `CONNECTED`.
* **Ops check**: MemSQL Ops must be running, and be version 5.7 or newer.

If any of these steps fail, `memsql-online-upgrade` will exit.

#### Snapshots

Snapshots will be taken for all user databases. If any of these Snapshots fail, `memsql-online-upgrade` will exit. You can find additional information in the log file.

#### Update Config

As part of the upgrade process, some system variables will be modified. Some variables will be set to `OFF` prior to the upgrade, and then back to `ON` after the upgrade is complete. If you require these variables to be set to `OFF` after the upgrade, you will need change them manually. The affected variables are listed below:

* `auto_attach`
* `aggregator_failure_detection`
* `leaf_failure_detection`

#### Detach Leaves

Each leaf in the cluster's first Availability Group (AG) will be detached. Before this operation starts, a few more checks are performed:

* **Orphan Check**: Ensures there are no orphan partitions in the cluster.
* **Master/Slave Check**: Ensures that all masters and slaves are present.

If all checks pass, we will detach the leaves in the first AG. All error and success messages are passed to `stdout` and the log file. 

#### Upgrade Nodes

The upgrade step stops each leaf, upgrades it, and then starts the leaf again. `memsql-online-upgrade` leverages the MemSQL Ops command `memsql-ops memsql-upgrade` and executes it on each leaf, one at a time.

You can observe additional information by tailing the log file. Additionally, you can also execute `watch -n 3 memsql-ops memsql-list` from a separate session on the master. 

If at any time the upgrade fails, the current leaf being upgraded will rollback to the previous version and halt the upgrade. At this point you would need to continue the upgrade manually. Check the output of `show leaves`, `show cluster status`, and `memsql-ops memsql-list` to verify the state of the cluster. 

After all the leaves in each Availability Group are complete, redundancy is restored to all user databases before continuing. 

#### Rebalance

Upon completion of the upgrade, all users databases are rebalanced.

#### Upgrade Aggregators

After all the leaves are successfully upgraded, each of the child aggregators are upgraded one at a time. Once all of the child aggregators have been upgraded, the master aggregator is upgraded.

#### Update Configs

The last steps of the process is to reset the variables we updated at the beginning of the upgrade. 

## Development

Developing is easy, and there are just a few things to consider. First, it's important to understand the `Go` language. `memsql-online-upgrade` is tested using Docker, so we recommend that you have Docker set up and available. Setting up and configuring Docker is outside the scope of this README. You must also provide an enterprise license key in the Dockerfile for testing. 

###### Step 1:

Get the code. If you are reading this, you must already have it or you are viewing the repo.

```
git clone git@github.com:memsql/online-upgrade.git
```

##### Step 2:

Ensure Docker is set up and ready by executing the following command:

```
docker ps
```

If this command is successful, then Docker is ready. Docker is required for testing purposes.

When you run `make test` or other options, a Docker container is created, and all tests in the project are executed in the container.

We recommend that you first run `make test-image-shell`. This command does a few things. First, it reads the Dockerfile in this project. The Dockerfile will set up the required Docker container for testing. See the Dockerfile for more details. Once all the required files are downloaded and the container is ready, two containers will be started using the image that was just created, and then they will be linked together. The first container will run in the background and the second container will drop you in a Bash shell. At this point, you can manually run some tests or just exit. The `test-image-shell` is useful for testing a single test. 

After these steps are complete, you should be ready to write some code and then test it.

#### Step 3:

It's important to understand the codebase. There are three main directories to focus on:

* `./steps`: Contains all the individual upgrade steps and their supporting tests. 
* `./testutil`: Contains functions related directly with testing.
* `./util`: Contains functions that support the steps.

For example, `./steps/pre_upgrade.go` will run several functions to verify the cluster is ready to be upgraded. Within the `pre_upgrade.go` file, the `func PreUpgrade()` function calls the util package for `util.OpsAgentList()` and returns a slice of AgentInfo (`[]AgentInfo`).

Each file in the `steps` and `util` directory have corresponding tests that either test a single function or an entire step, which encompasses more than one function.

Next, there are important files that are used to run and build `memsql-online-upgrade`:

* `Dockerfile`: Required to build the Docker container for testing.
* `Makefile`: Required to set up and execute testing.
* `main.go`: The main file the calls all the steps in their proper order. This file is also used to for build purposes.
* `glide.yaml` and `.lock`: These files are used for Go package management.  

##### Step 4:

Build and test your new function or step. This process is out of scope for this README.

#### Testing

There are several ways to get stated with testing. First, start by running `make test-image-shell` from within the `memsql-online-upgrade` repo directory (`go-workspace/src/github.com/memsql/online-upgrade/`).

```
[online-upgrade]$ make test-image-shell
```

Note that running `make` will download and install everything required for testing and will start executing all available tests. All test files should have the format `*_test.go`.

On subsequent tests you can either run `make test` or `make test-image-shell`. These files are explained below.

By default, `make` and `make test` will find all available tests by utilizing the output `glide nv`. The `glide nv` command, an alias for `glide novendor`, will be passed into go test `go test $(glide nv)` and will run over all directories of the project except the vendor directory.

If you only want to run a single test, you can run `make test` and specify the test argument, which expects a package folder or a specific test. See the examples below.

**Testing a package (e.g. `steps`)**:
```
make test test=./steps/...
```

**Executing a single test:**
```
make test test=./steps/detach_leaf_test.go
```

If you want to execute a specific test manually, simply run `go test <test>` from within the test shell. Start the test shell and run your desired test. The `test-image-shell` will spin up both master and child containers and drop you at a shell prompt. From the shell prompt, you can simply run your test.

Start by executing the command below:

```
make test-image-shell
```

Now you should see a prompt like this:

```
root@online-upgrade-master:/go/src/github.com/memsql/online-upgrade#
```

Now, execute a test. Note `-p 1` is required for running multiple tests at the same time due to `go test` invoking parallelization by default.

```
go test ./steps/pre_upgrade_test.go
```
```
go test -p 1 ./steps/... ./util/...
```

The test will run and either produce and error or a pass message like below.

```
ok  	command-line-arguments	59.890s
```

You can also pass in a `-v` to see activity during the test.

For more information about testing, you can also read the Makefile. As previously mentioned, a High Availability cluster is started during testing, which requires several steps to create. 


## Deployment

To deploy quickly, execute the following command:

```
go build -o memsql-online-upgrade main.go

```

This command produces an executable binary. Make sure you only execute this command in Linux.
