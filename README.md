# Online Upgrade steps from 5.x -> 5.x

pre-upgrade checks
 -> is HA setup
 -> ops is happy + healthy
 -> all leaves are online

snapshot all user databases
 -> connect to master
 -> for each database:
    -> SNAPSHOT $database

disable auto_attach, aggregator_failure_detection and leaf_failure_detection on master
 -> memsql-ops memsql-update-config...

disable aggregator_failure_detection on each agg
 -> memsql-ops memsql-update-config

detach AG 1 leaf nodes once the AG 2 slaves are caught up
 -> connect to master
 -> for leaf in leaves:
    DETACH LEAF $leaf:$port

    TODO: verify detach semantics re: locking master writes
    TODO: check sync replication?

memsql-ops memsql-stop <master id>

memsql-ops memsql-upgrade --skip-snapshot [ --memsql-id id ]...

memsql-ops memsql-start <master id>

attach AG 1 leaf nodes (no rebalance)
 -> connect to master
    -> for leaf in leaves:
        ATTACH LEAF $leaf:$port NO REBALANCE

wait for the new leaves to catch up and come online
 -> each leaf in AG 1 needs to be online

restore redundancy on user databases
 -> from master

detach AG 2 leaf nodes (once partitions are caught up)
 -> from master

memsql-ops memsql-stop <master id>

memsql-ops memsql-upgrade --skip-snapshot <memsql ids for leaves in availability group B>

memsql-ops memsql-start <master id>

attach AG 2 leaf nodes (no rebalance)
 -> connect to master
    -> for leaf in leaves:
        ATTACH LEAF $leaf:$port NO REBALANCE

wait for all leaves to be happy
 -> each leaf in AG 1+2 needs to be online

restore redundancy on user databases
 -> from master

rebalance partitions on user databases (once they are caught up)
 -> from master

for each child agg (or set of child aggs):
    memsql-ops memsql-upgrade --skip-snapshot --memsql-id <child agg ids>

memsql-ops memsql-stop <master id>

memsql-ops memsql-upgrade --skip-snapshot --memsql-id <master id>

wait for all nodes to be happy
 -> each leaf in AG 1+2 needs to be online
 -> aggs should be online + happy

enable auto_attach, aggregator_failure_detection and leaf_failure_detection on master
 -> memsql-ops memsql-update-config...

enable aggregator_failure_detection on each agg
 -> memsql-ops memsql-update-config...
