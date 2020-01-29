# connectivity-manager

This repository contains the code for the Connectivity Manager component. The component resides in the Mngt Cluster and is in charge of managing the `status` of each App Cluster. It receives `ClusterAlive` checks from the different App Clusters and sets the clusters' status depending on these checks.

## Getting Started

When a new platform is installed, a `grace-period` is set (120 s by default) and the component takes 2 parameters: `threshold` and `offlinePolicy`.
* **`threshold`**: period after which an App Cluster that has lost communication with the Mngt Cluster (that is, it has stopped receiving `ClusterAlive` signals) will change its status from `ONLINE`/`ONLINE_CORDON` to `OFFLINE`/`OFFLINE_CORDON`.
* **`offlinePolicy`**: parameter that defines the policy that will be triggered when a cluster has lost communication with the Mngt Cluster for a `grace-period` amount of time. It can be set to `none` or `drain`:
  * *`none`*: no policy will be triggered.
  * *`drain`*: the App Cluster will be drained (a `drain` signal will be sent to the conductor) after the `grace-period` expires and all the applications running on it will be redeployed somewhere else (when possible).
  
### Cluster status lifecycle
* When an App Cluster is created and no `ClusterAlive signals` are being sent yet, the cluster status will be `UNKNOWN`.
* When one of those checks arrives to the `connectivity-manager`, its status will change to `ONLINE` for as long as the `ClusterAlive` signals are being received.
* If the component doesn't get any `ClusterAlive` signals for longer than `threshold`, the cluster status will be set to `OFFLINE` (if the previous status was `ONLINE`) or `OFFLINE_CORDON` (if the previous status was `ONLINE_CORDON`).
+ If the component doesn't get any `ClusterAlive` signals for longer than `grace-period`, the cluster status will be set to `OFFLINE_CORDON` and the `offlinePolicy` will be triggered.

### Prerequisites

* [conductor](https://github.com/nalej/conductor)
* [connectivity-checker](https://github.com/nalej/connectivity-checker)

### Build and compile

In order to build and compile this repository use the provided Makefile:

```
make all
```

This operation generates the binaries for this repo, download dependencies,
run existing tests and generate ready-to-deploy Kubernetes files.

### Run tests

Tests are executed using Ginkgo. To run all the available tests:

```
make test
```

No tests are available for this repository at the moment.

### Update dependencies

Dependencies are managed using Godep. For an automatic dependencies download use:

```
make dep
```

In order to have all dependencies up-to-date run:

```
dep ensure -update -v
```

## Contributing

Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the available versions, see the [tags on this repository](https://github.com/nalej/connectivity-manager/tags). 

## Authors

See also the list of [contributors](https://github.com/nalej/connectivity-manager/contributors) who participated in this project.

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.
