# MariaDB Galera Cluster operator

Concept for implementation of boilerplate largely based on:
- https://github.com/coreos/etcd-operator
- https://github.com/kubernetes/sample-controller

## Requirements:

Automated cluster bootstraping
Resilient to single __node__ failure
Undisrupted during __pod__ (re)integration
Handles backups automatically
Handles restores automaticaly
Should recover fresh nodes quickly from local snapshot
Exposing metrics for collection / alerting

## Means to achieve:


### Bootstraping

Use StatefulSet to have a predictable network identity. Assume 3 replicas at minimum which can be treated as the core nodes to which additional nodes can initiate during bootstrap.
Leader election:
Snapshots need to ensure only one pod is writing to the shared storage, leader election will indicate which POD is a master pod and as such allowed to save snapshot. Other PODs will be acting as hot standby.

### Snapshoting

From elected master a quick method to save database dump to snapshot folder is required. For that use of xtrabackup seems inevitable.

  __non-client facing snapshot node ?__
  __incremental snapshots__
  __what if snapshot pod changes? incremental snap corruption ?__

### Backups

After snapshot is complete, it should be transferred to external storage according to expected retention rules 
(ie. daily, hourly etc., keeping always last week, every second for a month, every seventh for a year)

### Recovery

Be able to define that in case of lack of or corrupted snapshot, dump can be downloaded from some location 
to seed the database. Seed process needs to be a part of init and blocking for startup of other pods in StatefulSet 
https://kubernetes.io/docs/tutorials/stateful-application/basic-stateful-set/#ordered-pod-creation

  __point in time recovery ?__
  __corrupted snapshot ?__

### Resources

Default values for CPU and memory allocation, small alocation, 

  __alerting when potentialy too small ?__

### Scheduling

Databases should never be scheduled on the same physical node. To achieve that a Pod-AntiAffinity needs 
to be configured so that two pods for db can never be scheduled side by side.

### Upgrades

Test for seamless upgrades to newer version of MariaDB engine

### Growing storage space

Needs to accommodate for uninterrupted storage space growth. Applying with modified size and deleting pods 
sequentially should be enough, make use of snapshot and IST to recover data on new pod.

  __could it grow automaticaly ?__
  __define both starting and max PV size ?__
  __would it require PVC creation outside of POD volumeClaimTemplates to avoid reset of size on `oc apply`?__

### Monitoring

CPU
Memory
IO
MySQL Metrics
Galera metrics

### Other notes

Readiness probe to check for status (ie. exclude new node and donor untill IST is done)
What should happen when MariaDBCluster object is deleted, can w have some kind of deletion protection ?
Should be marked as critical cluster pod as failure during bootstrap may be an issue / handle mid-bootstrap interruptions

### Future considerations

Implement ProxySQL for a more intelligent routing to backends (avoiding direct use of service via kube-proxy)
Use PodPreset to inject credentials automatically
