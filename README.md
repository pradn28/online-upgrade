# Online Upgrade 

Online Upgrade is a Go project for upgrading MemSQL from 5.x -> 5.x

## Usage

memsql-online-upgrade should be run from the master aggregator on the cluster that you intend to upgrade. 

```
$ ./memsql-online-upgrade
```
`memsql-online-upgrade` will use the default host ("127.0.0.1"), username ("root") and port (3306), if not specified on the command line. For a full list of available options, you can the `-h` flag to `memsql-online-upgrade`.

## Overview

memsql-online-upgrade will execute various steps to upgrade you cluster without downtime.

#### Logging
Upon execution of the script, memsql-online-upgrade will create a log file in the current working directory. It will either create or append to the default log file ("online-upgrade.log") unless a file location is specified with `-log-path`.

#### Health Check
Health check will preform the following steps to ensure your cluster is ready and willing to be upgraded.
- Redundancy level check. Redundancy level must be 2 (HA)
- Agents Online. Check that all agents are in an ONLINE state.
- Node Check. Verify all MemSQL node are ONLINE and CONNECTED

If any of these steps fail, memsql-online-upgrade will exit.

#### Snapshot
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

<!-- Please add info on new steps here and dont forget your tests -->


## Development
<!-- TODO update development procedures -->
<!-- everything below here -->
Lots of neat stuff goes here

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
go test -p 1 ./steps/... ./util/...
```
The test will run and either produce and error or a pass message like below.
```
ok  	command-line-arguments	59.890s
```
You can also pass in a `-v` to see whats going on during the test.

For more information about testing, you can also take a peak at the Makefile. As was previously mentioned, we are spinning up an HA cluster during testing and there are several steps required to make this happen. 


## Deployment
<!-- TODO update deployment process -->
Quick and dirty.
```
go build -o memsql-online-upgrade main.go
```
Ship it!
