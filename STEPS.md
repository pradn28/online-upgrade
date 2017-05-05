# Online Upgrade steps from 5.x -> 5.x

Note: User is prompted between major steps to continue.

<!-- pre-upgrade -->
pre-upgrade checks
 -> Check redundancy level
 -> ops version check
 -> all agents and nodes are online and connected
 -> check that the cluster is balanced

<!-- snapshot databases -->
snapshot all user databases
 -> connect to master
 -> for each database:
    -> SNAPSHOT $database

<!-- update config -->
store existing settings
set auto_attach, aggregator_failure_detection and leaf_failure_detection on master to OFF
 -> memsql-ops memsql-update-config...

store existing settings
disable aggregator_failure_detection on each agg
 -> memsql-ops memsql-update-config

<!-- Detach Leaves -->
detach AG 1 leaf nodes once the AG 2 slaves are caught up
 -> connect to master
 -> for leaf in leaves:
    DETACH LEAF $leaf:$port

    verify detach semantics re: locking master writes
    check sync replication?

<!-- memsql upgrade -->
memsql-ops memsql-stop <memsql-id>

memsql-ops memsql-upgrade --skip-snapshot [ --memsql-id id ]...

memsql-ops memsql-start <memsql-id>

<!-- attach leaves -->
attach AG 1 leaf nodes (no rebalance)
 -> connect to master
    -> for leaf in leaves:
        ATTACH LEAF $leaf:$port NO REBALANCE

wait for the new leaves to catch up and come online
 -> each leaf in AG 1 needs to be online

restore redundancy on user databases
 -> from master

<!-- Detach Leaves -->
detach AG 2 leaf nodes (once partitions are caught up)
 -> from master

<!-- memsql upgrade -->
memsql-ops memsql-stop <memsql-id>

memsql-ops memsql-upgrade --skip-snapshot <memsql ids for leaves in availability group B>

memsql-ops memsql-start <memsql-id>

attach AG 2 leaf nodes (no rebalance)
 -> connect to master
    -> for leaf in leaves:
        ATTACH LEAF $leaf:$port NO REBALANCE

wait for all leaves to be happy
 -> each leaf in AG 1+2 needs to be online

restore redundancy on user databases
 -> from master

<!-- rebalance cluster -->
rebalance partitions on user databases (once they are caught up)
 -> from master

<!-- upgrade aggregators -->
We will prompt the user before each aggregator upgrade. We will also give them some explanation of what is going to happen.

for each child agg (or set of child aggs):
    memsql-ops memsql-upgrade --skip-snapshot --memsql-id <child agg ids>

memsql-ops memsql-stop <master id>

memsql-ops memsql-upgrade --skip-snapshot --memsql-id <master id>

<!-- cleanup -->
wait for all nodes to be happy
 -> each leaf in AG 1+2 needs to be online
 -> aggs should be online + happy

enable auto_attach, aggregator_failure_detection and leaf_failure_detection on master
 -> memsql-ops memsql-update-config...

enable aggregator_failure_detection on each agg
 -> memsql-ops memsql-update-config...

