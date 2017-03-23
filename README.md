# Online Upgrade 

Online Upgrade is a Go project for upgrading MemSQL from 5.x -> 5.x

## Usage

memsql-online-upgrade should be run from the master aggregator on the cluster that you intend to upgrade. 

```
$ ./memsql-online-upgrade
```
`memsql-online-upgrade` will use the default host ("127.0.0.1"), username ("root") and port (3306), if not specified on the command line. For a full list of available options, you can add the `-h` flag to `memsql-online-upgrade`.

Need a specific version? You can use the flag `-version-hash` followed by the version hash you want to upgrade to. The default is to upgrade to the latest version. 

## Overview

memsql-online-upgrade will execute various steps to upgrade you cluster without downtime. While the first step will verify you cluster is healthy, it is recommended that you double check your cluster is in a healthy status prior to starting the upgrade. 

#### Logging
Upon execution of the script, memsql-online-upgrade will create a log file in the current working directory. It will either create or append to the default log file ("online-upgrade.log") unless a file location is specified with `-log-path`.

#### Health Check
Health check will preform the following steps to ensure your cluster is ready and willing to be upgraded.
- Redundancy level check. Redundancy level must be 2 (HA)
- Agents Online. Check that all agents are in an ONLINE state.
- Node Check. Verify all MemSQL node are ONLINE and CONNECTED

If any of these steps fail, memsql-online-upgrade will exit.

#### Snapshots
Snapshots will be taken for all user databases. If any of these Snapshots fail, memsql-online-upgrade will exit. You can find additional information in the log file.

#### Update Config
As part of the upgrade process, we must set some variables. We will set them "off" prior to the upgrade and back "on" again after the upgrade is complete.

- auto_attach, aggregator_failure_detection, leaf_failure_detection

#### Detach Leaves
Now we can start the process to detach the leaves in the first AG.

But, before we do, we need to check a few more things.
* Orphan Check - This will ensure there are not any orphan partitions in the cluster.
* Master/Slave Check - Ensure that all masters and slaves are present.

If all checks pass, we will detach the leaves in the first AG.

All error and success messages are passed to stdout and the log file. 

#### Upgrade Nodes

The upgrade step will, stop each leaf, upgrade it and start the leaf back up again. 
Online upgrade will leverage `memsql-ops memsql-upgrade` for each leaf. One at a time.
You can observe additional information by tailing the log file. You can also `watch -n 3 memsql-ops memsql-list` from a separate session on the master. 

If at any time the upgrade fails, the current leaf being upgraded will rollback to the previous version and halt the upgrade. At this point you would need to continue the upgrade manually. Check the output of `show leaves`, `show cluster status`, and `memsql-ops memsql-list` to verify the state of the cluster. 

After all the leaves in each availability group are complete, we will restore redundancy to all user databases before continuing. 

#### Rebalance
Upon completion of the upgrade, we will rebalance all users databases.

#### Upgrade Aggregators
After all the leaves are successfully upgraded, we will upgrade each of the child aggregators one at a time followed by the master aggregator. 

#### Update Configs
The last steps of the process is to reset the variables we updated at the beginning of the upgrade. 

## Development

Developing is easy. There are just a few things we need to know. First, you should know `Go`. We also leverage Docker to test, so you should have Docker setup and available. That setup is outside the scope of this README.  

###### Step 1:
Get the code. If you are reading this, you must already have it or you are viewing the repo.

```
git clone git@github.com:memsql/online-upgrade.git
```

##### Step 2:
Ensure Docker is setup and ready.
```
docker ps
```
You'll know at this point if docker is ready. So why do you need Docker. Testing.

When you run `make test` or other options, which we'll discussed later in Testing, we spin up a docker container and run all the tests in the project.

Now, it is advised that you first run `make test-image-shell`. This will do a few different things. It will, first and foremost, read the Dockerfile in this project. The Dockerfile will setup the required Docker container for testing. See the Dockerfile for more details. Once all the require files are downloaded and the container is ready, we will start two containers with the image we just created and link them together. The first container will run in the background and the second one will drop you in a Bash shell. At this point you can manually runs some tests or just exit. The `test-image-shell` is usefully for testing a single test. 

After that is complete, you should be ready to write some code and of course test it.

#### Step 3:
Understanding the codebase.
There are currently three main directories to focus on:

* ./steps - contains all the individual upgrade steps and their supporting tests 
* ./testuitl - contains function related directly with testing
* ./util - contains functions that support each step

As an example `./steps/pre_upgrade.go` will run several functions to verify the cluster is ready to be upgraded. Within the pre_upgrade step, `func PreUpgrade()` we call to the util package for `util.OpsAgentList()` which will return a slice of AgentInfo ([]AgentInfo).

Each of the steps and utils have a test to either test the function or the entire step which includes several other functions. 

Next thing to understand is some of the necessary file to run and build memsql-online-upgrade.

* Dockerfile - required to build the docker container for testing.
* Makefile - required to setup and execute testing.
* main.go - is the main file the calls all the steps in order. Which is also the file we use to build. 
* glide.yaml and .lock - are used for go package management. 

This rest of the file you should already be familiar with. 

##### Step 4:

Build and test your new function or step. This will not be explained in this README. See testing section below on testing. 

  

#### Testing
There are several ways to get stated with testing. Lets start with running `make` from within the memsql-online-upgrade repo directory (go-workspace/src/github.com/memsql/online-upgrade/).

```
[online-upgrade]$ make
```
Note that running `make` will download and install everything required for testing and will start testing all available tests. All test files should have the format `*_test.go`.

On subsequent tests you can either run `make test`, `make test-one` or `make test-image-shell`. These are all explained below.

By default make and make test will find all available tests by utilizing the output `glide nv`. The `glide nv` command, an alias for `glide novendor`, will be passed into go test `go test $(glide nv)` and will run over all directories of the project except the vendor directory.

When you only want to run a single test, we can run `make test-one`, which excepts a package folder or a specific test. See the examples below.

Testing a package (e.g. steps)
```
make test-one test=./steps/...
```

And a single test
```
make test-one test=./steps/detach_leaf_test.go
```

Want to test manually? Sure. You can do that too. You can simply run `go test <test>` from within the test shell. Start the test shell and run your test. The `test-image-shell` will spin up both master and child containers and drop you at a shell prompt. From here you can simply run your test. See the example below.

```
make test-image-shell
```
Now you should see a prompt like this.
```
root@online-upgrade-master:/go/src/github.com/memsql/online-upgrade#
```
Let go ahead and run a test. Note `-p 1` is required for running multiple tests at the same time due to `go test` invoking parallelization by default.
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
You can also pass in a `-v` to see whats going on during the test.

For more information about testing, you can also take a peak at the Makefile. As was previously mentioned, we are spinning up an HA cluster during testing and there are several steps required to make this happen. 


## Deployment
Quick and dirty.
```
go build -o memsql-online-upgrade main.go

```
This will produce a executable binary. Make sure you do this in Linux.

Ship it!

