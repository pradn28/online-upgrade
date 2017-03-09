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

As part of the upgrade process, we must set some variables. We will set them off prior to the upgrade and back on again after the upgrade is complete.

- auto_attach, aggregator_failure_detection, leaf_failure_detection
<!-- TODO -->
<!-- Please add info on new steps here and dont forget your tests -->


## Development
<!-- TODO -->
<!-- everything below here -->
Lots of neat stuff goes here

## Deployment
<!-- TODO -->
<!-- everything below here -->
Quick and dirty.
```
go build -o memsql-online-upgrade main.go
```
Ship it!
